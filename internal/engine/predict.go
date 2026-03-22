package engine

import (
	"fmt"

	"github.com/andragon31/fenrir/internal/graph"
)

type PredictEngine struct {
	graph *graph.Graph
}

type Prediction struct {
	Type     string  `json:"type"`
	Module   string  `json:"module"`
	Severity string  `json:"severity"`
	Message  string  `json:"message"`
	Confidence float64 `json:"confidence"`
	Occurrences int   `json:"occurrences"`
}

func NewPredictEngine(g *graph.Graph) *PredictEngine {
	return &PredictEngine{
		graph: g,
	}
}

func (e *PredictEngine) GetPredictions(module string) ([]Prediction, error) {
	var predictions []Prediction

	predictions = append(predictions, e.predictViolations(module)...)
	predictions = append(predictions, e.predictRecurringBugs(module)...)
	predictions = append(predictions, e.predictDependencyIssues(module)...)

	return predictions, nil
}

func (e *PredictEngine) predictViolations(module string) []Prediction {
	var predictions []Prediction

	query := `
		SELECT scope, COUNT(*) as count
		FROM nodes
		WHERE type = 'violation'
		GROUP BY scope
		HAVING count >= 3
	`
	if module != "" {
		query = `
			SELECT scope, COUNT(*) as count
			FROM nodes
			WHERE type = 'violation' AND scope LIKE ?
			GROUP BY scope
			HAVING count >= 2
		`
	}

	rows, err := e.graph.DB().Query(query, "%"+module+"%")
	if err != nil {
		return predictions
	}
	defer rows.Close()

	for rows.Next() {
		var scope string
		var count int
		if err := rows.Scan(&scope, &count); err != nil {
			continue
		}

		predictions = append(predictions, Prediction{
			Type:       "violation_risk",
			Module:     scope,
			Severity:  "warning",
			Message:    fmt.Sprintf("This module has had %d violations in recent sessions. Consider reviewing architectural decisions.", count),
			Confidence: 0.7 + float64(count)*0.05,
			Occurrences: count,
		})
	}

	return predictions
}

func (e *PredictEngine) predictRecurringBugs(module string) []Prediction {
	var predictions []Prediction

	query := `
		SELECT title, scope, COUNT(*) as count
		FROM nodes
		WHERE type = 'bugfix'
		GROUP BY title, scope
		HAVING count >= 3
		ORDER BY count DESC
	`
	if module != "" {
		query = `
			SELECT title, scope, COUNT(*) as count
			FROM nodes
			WHERE type = 'bugfix' AND scope LIKE ?
			GROUP BY title, scope
			HAVING count >= 2
			ORDER BY count DESC
		`
	}

	rows, err := e.graph.DB().Query(query, "%"+module+"%")
	if err != nil {
		return predictions
	}
	defer rows.Close()

	for rows.Next() {
		var title, scope string
		var count int
		if err := rows.Scan(&title, &scope, &count); err != nil {
			continue
		}

		predictions = append(predictions, Prediction{
			Type:       "recurring_bug",
			Module:     scope,
			Severity:   "info",
			Message:    fmt.Sprintf("Bug '%s' has been fixed %d times. Consider a more permanent solution.", title, count),
			Confidence: 0.8,
			Occurrences: count,
		})
	}

	return predictions
}

func (e *PredictEngine) predictDependencyIssues(module string) []Prediction {
	var predictions []Prediction

	rows, err := e.graph.DB().Query(`
		SELECT DISTINCT title
		FROM nodes
		WHERE type = 'observation' AND title LIKE '%dependency%'
		AND title LIKE ?
		LIMIT 5
	`, "%"+module+"%")
	if err != nil {
		return predictions
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			continue
		}
		count++
	}

	if count >= 3 {
		predictions = append(predictions, Prediction{
			Type:       "dependency_instability",
			Module:     module,
			Severity:   "warning",
			Message:    "Multiple dependency-related observations in this module. Consider reviewing package stability.",
			Confidence: 0.6,
			Occurrences: count,
		})
	}

	return predictions
}

func (e *PredictEngine) GetSessionRisk(sessionID string) (string, float64) {
	var violations, warnings, completedSessions int

	e.graph.DB().QueryRow(`
		SELECT COUNT(*) FROM nodes
		WHERE type = 'violation' AND session_id = ?
	`, sessionID).Scan(&violations)

	e.graph.DB().QueryRow(`
		SELECT COUNT(*) FROM audit_log
		WHERE session_id = ? AND risk_level IN ('warning', 'high', 'critical')
	`, sessionID).Scan(&warnings)

	e.graph.DB().QueryRow(`
		SELECT COUNT(*) FROM sessions
		WHERE status = 'closed'
	`).Scan(&completedSessions)

	if completedSessions == 0 {
		completedSessions = 1
	}

	riskScore := (float64(violations)*0.5 + float64(warnings)*0.3) / float64(completedSessions)

	var riskLevel string
	switch {
	case riskScore >= 0.7:
		riskLevel = "high"
	case riskScore >= 0.4:
		riskLevel = "medium"
	default:
		riskLevel = "low"
	}

	return riskLevel, riskScore
}

func (e *PredictEngine) GetTeamHealth() (map[string]interface{}, error) {
	var totalSessions, closedSessions, totalViolations, totalDecisions int

	e.graph.DB().QueryRow("SELECT COUNT(*) FROM sessions").Scan(&totalSessions)
	e.graph.DB().QueryRow("SELECT COUNT(*) FROM sessions WHERE status = 'closed'").Scan(&closedSessions)
	e.graph.DB().QueryRow("SELECT COUNT(*) FROM nodes WHERE type = 'violation'").Scan(&totalViolations)
	e.graph.DB().QueryRow("SELECT COUNT(*) FROM nodes WHERE type = 'decision'").Scan(&totalDecisions)

	if totalSessions == 0 {
		totalSessions = 1
	}

	completionRate := float64(closedSessions) / float64(totalSessions)
	violationRate := float64(totalViolations) / float64(totalSessions)

	var healthScore float64
	if totalDecisions > 0 {
		healthScore = 1.0 - (float64(totalViolations) / float64(totalDecisions))
		if healthScore < 0 {
			healthScore = 0
		}
	}

	healthScore = (healthScore + completionRate) / 2

	return map[string]interface{}{
		"total_sessions":    totalSessions,
		"closed_sessions":    closedSessions,
		"total_violations":  totalViolations,
		"total_decisions":   totalDecisions,
		"completion_rate":    completionRate,
		"violation_rate":    violationRate,
		"health_score":      healthScore,
		"status":            e.getHealthStatus(healthScore),
	}, nil
}

func (e *PredictEngine) getHealthStatus(score float64) string {
	switch {
	case score >= 0.8:
		return "excellent"
	case score >= 0.6:
		return "good"
	case score >= 0.4:
		return "fair"
	default:
		return "poor"
	}
}
