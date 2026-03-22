package graph

import (
	"fmt"
)

type ImpactResult struct {
	Target       string   `json:"target"`
	AffectedNodes []string `json:"affected_nodes"`
	RiskScore    float64  `json:"risk_score"`
	Summary      string   `json:"summary"`
}

// PredictImpact performs a BFS on the graph to find nodes affected by a target change
func (g *Graph) PredictImpact(target string) (*ImpactResult, error) {
	result := &ImpactResult{
		Target: target,
	}

	// 1. Find initial nodes related to target
	rows, err := g.db.Query("SELECT id, title FROM nodes WHERE title LIKE ? OR content LIKE ?", "%"+target+"%", "%"+target+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	visited := make(map[string]bool)
	queue := []string{}

	for rows.Next() {
		var id, title string
		rows.Scan(&id, &title)
		queue = append(queue, id)
		visited[id] = true
		result.AffectedNodes = append(result.AffectedNodes, fmt.Sprintf("%s (%s)", title, id[:8]))
	}

	// 2. BFS to find connected nodes (depth 2)
	depth := 0
	for len(queue) > 0 && depth < 2 {
		levelSize := len(queue)
		for i := 0; i < levelSize; i++ {
			currID := queue[0]
			queue = queue[1:]

			// Find neighbors
			neighbors, _ := g.db.Query("SELECT to_id FROM edges WHERE from_id = ?", currID)
			if neighbors != nil {
				for neighbors.Next() {
					var toID string
					neighbors.Scan(&toID)
					if !visited[toID] {
						visited[toID] = true
						queue = append(queue, toID)
						
						// Get title for display
						var title string
						g.db.QueryRow("SELECT title FROM nodes WHERE id = ?", toID).Scan(&title)
						result.AffectedNodes = append(result.AffectedNodes, fmt.Sprintf("%s (%s)", title, toID[:8]))
					}
				}
				neighbors.Close()
			}
		}
		depth++
	}

	result.RiskScore = float64(len(result.AffectedNodes)) / 10.0
	if result.RiskScore > 1.0 {
		result.RiskScore = 1.0
	}

	result.Summary = fmt.Sprintf("Changing %s affects %d known architectural or design nodes. Risk score: %.2f", target, len(result.AffectedNodes), result.RiskScore)

	return result, nil
}
