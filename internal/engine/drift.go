package engine

import (
	"fmt"

	"github.com/TU_ORG/fenrir/internal/graph"
)

type DriftEngine struct {
	graph *graph.Graph
}

type DriftResult struct {
	Module         string  `json:"module"`
	Score          float64 `json:"score"`
	Violations     int     `json:"violations"`
	Decisions     int     `json:"decisions"`
	SessionCount  int     `json:"session_count"`
	ChangesPerSession float64 `json:"changes_per_session"`
}

func NewDriftEngine(g *graph.Graph) *DriftEngine {
	return &DriftEngine{
		graph: g,
	}
}

func (e *DriftEngine) CalculateDrift(sessionID string) (*DriftResult, error) {
	var violations, decisions, sessions int

	e.graph.DB().QueryRow(`
		SELECT COUNT(*) FROM nodes 
		WHERE type = 'violation' AND session_id = ?
	`, sessionID).Scan(&violations)

	e.graph.DB().QueryRow(`
		SELECT COUNT(*) FROM nodes 
		WHERE type = 'decision' AND status = 'active'
	`).Scan(&decisions)

	e.graph.DB().QueryRow(`
		SELECT COUNT(*) FROM sessions
	`).Scan(&sessions)

	if sessions == 0 {
		sessions = 1
	}

	score := 0.0
	if decisions > 0 {
		score = float64(violations) / float64(decisions)
		if score > 1.0 {
			score = 1.0
		}
	}

	result := &DriftResult{
		Score:          score,
		Violations:     violations,
		Decisions:     decisions,
		SessionCount:  sessions,
		ChangesPerSession: float64(violations) / float64(sessions),
	}

	module := ""
	e.graph.DB().QueryRow(`
		SELECT scope FROM nodes 
		WHERE session_id = ? LIMIT 1
	`, sessionID).Scan(&module)

	if module != "" && module != "global" {
		result.Module = module
	}

	return result, nil
}

func (e *DriftEngine) GetProjectDrift() ([]DriftResult, error) {
	var modules []DriftResult

	rows, err := e.graph.DB().Query(`
		SELECT DISTINCT scope, COUNT(*) as session_count
		FROM nodes
		WHERE scope IS NOT NULL AND scope != '' AND scope != 'global'
		GROUP BY scope
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var module string
		var sessionCount int
		if err := rows.Scan(&module, &sessionCount); err != nil {
			continue
		}

		var violations int
		e.graph.DB().QueryRow(`
			SELECT COUNT(*) FROM nodes
			WHERE type = 'violation' AND scope LIKE ?
		`, "%"+module+"%").Scan(&violations)

		var decisions int
		e.graph.DB().QueryRow(`
			SELECT COUNT(*) FROM nodes
			WHERE type = 'decision' AND status = 'active' 
			AND (scope = 'global' OR scope LIKE ?)
		`, "%"+module+"%").Scan(&decisions)

		score := 0.0
		if decisions > 0 {
			score = float64(violations) / float64(decisions)
			if score > 1.0 {
				score = 1.0
			}
		}

		modules = append(modules, DriftResult{
			Module:         module,
			Score:          score,
			Violations:     violations,
			Decisions:     decisions,
			SessionCount:  sessionCount,
			ChangesPerSession: float64(violations) / float64(sessionCount),
		})
	}

	return modules, nil
}

func (e *DriftEngine) GetTrend(module string, weeks int) ([]float64, error) {
	var scores []float64

	rows, err := e.graph.DB().Query(`
		SELECT score FROM drift_scores
		WHERE module LIKE ? OR ? = ''
		ORDER BY updated_at DESC
		LIMIT ?
	`, "%"+module+"%", module, weeks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var score float64
		if err := rows.Scan(&score); err != nil {
			continue
		}
		scores = append(scores, score)
	}

	return scores, nil
}

func (e *DriftEngine) GetSeverity(score float64) string {
	switch {
	case score >= 0.6:
		return "critical"
	case score >= 0.3:
		return "warning"
	default:
		return "normal"
	}
}

func (e *DriftEngine) GetRecommendation(score float64) string {
	switch {
	case score >= 0.8:
		return "URGENT: Significant architectural drift detected. Review recent changes and align with architectural decisions."
	case score >= 0.6:
		return "WARNING: Moderate drift detected. Consider reviewing changes in affected modules."
	case score >= 0.3:
		return "NOTICE: Minor drift detected. Monitor the situation."
	default:
		return "GOOD: No significant drift detected. Architecture is stable."
	}
}

func (e *DriftEngine) FormatDrift(result *DriftResult) string {
	severity := e.GetSeverity(result.Score)
	recommendation := e.GetRecommendation(result.Score)

	return fmt.Sprintf(`
=== Drift Analysis ===

Module: %s
Score: %.2f (%s)
Violations: %d
Decisions: %d
Sessions: %d
Changes/Session: %.2f

%s
`, result.Module, result.Score, severity, result.Violations,
		result.Decisions, result.SessionCount, result.ChangesPerSession,
		recommendation)
}
