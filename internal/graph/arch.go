package graph

import (
	"fmt"
	"strings"
)

type Decision struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Rationale string `json:"rationale"`
	Weight    string `json:"weight"`
	Scope     string `json:"scope"`
}

func (g *Graph) SaveDecision(decision, rationale, weight, scope string) (string, error) {
	node := &Node{
		Type:    "decision",
		Title:   decision,
		Content: rationale,
		Weight:  weight,
		Scope:   scope,
	}
	return g.SaveNode(node)
}

func (g *Graph) ListDecisions() ([]Decision, error) {
	rows, err := g.db.Query(`
		SELECT id, title, content, weight, scope
		FROM nodes
		WHERE type = 'decision'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []Decision
	for rows.Next() {
		var d Decision
		err := rows.Scan(&d.ID, &d.Title, &d.Rationale, &d.Weight, &d.Scope)
		if err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}
	return decisions, nil
}

func (g *Graph) VerifyAction(action, context string) (string, string, error) {
	g.cacheMutex.RLock()
	policies := g.policyCache
	g.cacheMutex.RUnlock()

	if len(policies) == 0 {
		rows, err := g.db.Query("SELECT title, content FROM nodes WHERE type IN ('decision', 'policy')")
		if err == nil {
			defer rows.Close()
			g.cacheMutex.Lock()
			for rows.Next() {
				var p Policy
				if err := rows.Scan(&p.Title, &p.Content); err == nil {
					g.policyCache = append(g.policyCache, p)
				}
			}
			policies = g.policyCache
			g.cacheMutex.Unlock()
		}
	}

	conflicts := []string{}
	actionLower := strings.ToLower(action)

	for _, p := range policies {
		if strings.Contains(actionLower, strings.ToLower(p.Title)) ||
			strings.Contains(actionLower, strings.ToLower(p.Content)) {
			conflicts = append(conflicts, fmt.Sprintf("%s: %s", p.Title, p.Content))
		}
	}

	if len(conflicts) > 0 {
		return "warning", fmt.Sprintf("Conflict with decisions: %s", strings.Join(conflicts, "; ")), nil
	}

	return "success", "Action allowed", nil
}

// GetArchitectureMermaid generates a Mermaid.js diagram of current architecture
func (g *Graph) GetArchitectureMermaid() (string, error) {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	rows, err := g.db.Query("SELECT id, title, type FROM nodes WHERE type IN ('decision', 'policy', 'arch')")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	nodeMap := make(map[string]bool)
	for rows.Next() {
		var id, title, nodeType string
		rows.Scan(&id, &title, &nodeType)
		shortID := id
		if len(id) > 8 {
			shortID = id[:8]
		}
		sb.WriteString(fmt.Sprintf("  %s[\"%s (%s)\"]\n", shortID, title, nodeType))
		nodeMap[id] = true
	}

	edgeRows, err := g.db.Query("SELECT from_id, to_id, relation FROM edges")
	if err != nil {
		return sb.String(), nil
	}
	defer edgeRows.Close()

	for edgeRows.Next() {
		var from, to, relation string
		edgeRows.Scan(&from, &to, &relation)
		if nodeMap[from] && nodeMap[to] {
			shortFrom := from
			if len(from) > 8 {
				shortFrom = from[:8]
			}
			shortTo := to
			if len(to) > 8 {
				shortTo = to[:8]
			}
			sb.WriteString(fmt.Sprintf("  %s -- %s --> %s\n", shortFrom, relation, shortTo))
		}
	}

	return sb.String(), nil
}
