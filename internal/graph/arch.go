package graph

type Decision struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Rationale string `json:"rationale"`
	Weight    string `json:"weight"`
	Scope     string `json:"scope"`
}

func (g *Graph) SaveDecision(decision, rationale, weight, scope string) (string, error) {
	node := &Node{
		Type:      "decision",
		Title:     decision,
		Content:   rationale,
		Weight:    weight,
		Scope:     scope,
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
	// TODO: Integrate with PolicyEngine or search decisions
	return "success", "Action allowed", nil
}
