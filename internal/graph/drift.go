package graph

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DriftScore struct {
	Module     string    `json:"module"`
	Score      float64   `json:"score"`
	Violations int       `json:"violations"`
	Sessions   int       `json:"sessions"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ContextResult struct {
	ActiveSessions     int                    `json:"active_sessions"`
	RecentObservations int                    `json:"recent_observations"`
	Observations       []Node                 `json:"observations"`
	Predictions        []Prediction           `json:"predictions"`
	DriftAlerts        []DriftAlert           `json:"drift_alerts"`
	AutoInjected       map[string]interface{} `json:"auto_injected_context,omitempty"`
}

type Prediction struct {
	Module   string `json:"module"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type DriftAlert struct {
	Module string  `json:"module"`
	Score  float64 `json:"score"`
}

type Insight struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Count   int    `json:"count"`
	Details string `json:"details"`
}

type TraceEntry struct {
	NodeID    string `json:"node_id"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Depth     int    `json:"depth"`
}

func (g *Graph) GetContext(module string, limit int, includePredictions bool) (*ContextResult, error) {
	result := &ContextResult{}

	g.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE status = 'active'").Scan(&result.ActiveSessions)

	rows, err := g.Search("", "", module, limit, false)
	if err != nil {
		return nil, err
	}
	result.RecentObservations = len(rows)
	result.Observations = rows

	if includePredictions {
		predictions, err := g.GetPredictions(module)
		if err == nil {
			result.Predictions = predictions
		}

		alerts, err := g.GetDriftScores(module)
		if err == nil {
			for _, a := range alerts {
				if a.Score > 0.3 {
					result.DriftAlerts = append(result.DriftAlerts, DriftAlert{Module: a.Module, Score: a.Score})
				}
			}
		}
	}

	return result, nil
}

func (g *Graph) GetPredictions(module string) ([]Prediction, error) {
	query := `
		SELECT module, COUNT(*) as violations
		FROM (
			SELECT session_id, JSON_EXTRACT(files_modified, '$') as files, scope as module
			FROM nodes WHERE type = 'violation'
		)
		WHERE module LIKE ?
		GROUP BY module
		HAVING violations >= 3
	`

	rows, err := g.db.Query(query, "%"+module+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predictions []Prediction
	for rows.Next() {
		var p Prediction
		var count int
		if err := rows.Scan(&p.Module, &count); err != nil {
			continue
		}
		predictions = append(predictions, p)
	}

	// 2. Predict unstable modules via recurring bugfixes
	bugQuery := `
		SELECT scope, COUNT(*) as bugs
		FROM nodes
		WHERE type = 'bugfix' AND (scope LIKE ? OR scope = 'global')
		GROUP BY scope
		HAVING bugs >= 2
	`
	brows, err := g.db.Query(bugQuery, "%"+module+"%")
	if err == nil {
		defer brows.Close()
		for brows.Next() {
			var m string
			var count int
			if err := brows.Scan(&m, &count); err == nil {
				predictions = append(predictions, Prediction{
					Module:   m,
					Severity: "suggestion",
					Message:  fmt.Sprintf("Module might be unstable: %d bugfixes recently recorded in this area.", count),
				})
			}
		}
	}

	return predictions, nil
}

func (g *Graph) GetDriftScores(module string) ([]DriftScore, error) {
	sql := `SELECT module, score, violations, sessions, updated_at FROM drift_scores`
	var args []interface{}
	if module != "" {
		sql += " WHERE module = ?"
		args = append(args, module)
	}
	sql += " ORDER BY score DESC"

	rows, err := g.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []DriftScore
	for rows.Next() {
		var s DriftScore
		err := rows.Scan(&s.Module, &s.Score, &s.Violations, &s.Sessions, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	return scores, nil
}

func (g *Graph) GetInsights() ([]Insight, error) {
	// Simple insight: most modified modules
	sql := `
		SELECT scope as module, COUNT(*) as actions
		FROM nodes
		GROUP BY scope
		ORDER BY actions DESC
		LIMIT 10
	`
	rows, err := g.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []Insight
	for rows.Next() {
		var i Insight
		err := rows.Scan(&i.Title, &i.Count)
		if err != nil {
			continue
		}
		i.Type = "hotspot"
		i.Details = fmt.Sprintf("High activity detected in module: %s", i.Title)
		insights = append(insights, i)
	}
	return insights, nil
}

func (g *Graph) UpdateDriftScores(sessionID string, files []string) {
	if len(files) == 0 {
		return
	}

	var violations int
	g.db.QueryRow(`
		SELECT COUNT(*) FROM audit_log 
		WHERE session_id = ? AND result IN ('blocked', 'warning')
	`, sessionID).Scan(&violations)

	if violations == 0 {
		return
	}

	modules := make(map[string]bool)
	for _, f := range files {
		parts := strings.Split(filepath.ToSlash(f), "/")
		if len(parts) > 1 {
			module := strings.Join(parts[:2], "/")
			modules[module] = true
		} else if len(parts) == 1 {
			modules[parts[0]] = true
		}
	}

	scoreDelta := float64(violations) * 5.0 / float64(len(files))

	for m := range modules {
		g.db.Exec(`
			INSERT INTO drift_scores (id, module, score, violations, sessions)
			VALUES (?, ?, ?, ?, 1)
			ON CONFLICT(module) DO UPDATE SET
				score = score + ?,
				violations = violations + ?,
				sessions = sessions + 1,
				updated_at = CURRENT_TIMESTAMP
		`, "ds-"+uuid.New().String(), m, scoreDelta, violations, scoreDelta, violations)
	}

	g.db.Exec("UPDATE sessions SET drift_delta = ?, arch_violations = ? WHERE id = ?", scoreDelta, violations, sessionID)
}

func (g *Graph) Trace(target string, depth int) ([]TraceEntry, error) {
	query := `
		SELECT id, 'modified' as action, updated_at
		FROM nodes
		WHERE title LIKE ? OR content LIKE ?
		ORDER BY updated_at DESC
		LIMIT ?
	`
	rows, err := g.db.Query(query, "%"+target+"%", "%"+target+"%", 20)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []TraceEntry
	for rows.Next() {
		var e TraceEntry
		if err := rows.Scan(&e.NodeID, &e.Action, &e.Timestamp); err == nil {
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func (g *Graph) CalculateDrift(sessionID string) error {
	// 1. Get session nodes
	nodes, err := g.GetSessionNodes(sessionID)
	if err != nil || len(nodes) == 0 {
		return err
	}

	var sessionConfSum float64
	for _, n := range nodes {
		sessionConfSum += n.Confidence
	}
	sessionAvg := sessionConfSum / float64(len(nodes))

	// 2. Get global baseline
	globalAvg, _ := g.GetConfidenceAverage()

	// 3. If session confidence is significantly lower than baseline, signal drift
	if sessionAvg < globalAvg*0.8 {
		// Log a drift event (this would trigger UpdateDriftScores via session end logic)
		fmt.Printf("[Drift] Session %s confidence (%.2f) is below baseline (%.2f)\n", sessionID, sessionAvg, globalAvg)
	}

	return nil
}

func (g *Graph) GetConfidenceAverage() (float64, error) {
	var avg float64
	err := g.db.QueryRow("SELECT COALESCE(AVG(confidence), 0.7) FROM (SELECT confidence FROM nodes ORDER BY created_at DESC LIMIT 100)").Scan(&avg)
	return avg, err
}
