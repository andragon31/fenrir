package graph

import (
	"time"

	"github.com/google/uuid"
)

type Spec struct {
	ID           string    `json:"id"`
	Capability   string    `json:"capability"`
	Title        string    `json:"title"`
	Requirement  string    `json:"requirement"`
	Scenarios    string    `json:"scenarios"`
	Status       string    `json:"status"`
	NodeID       string    `json:"node_id,omitempty"`
	ImportedFrom string    `json:"imported_from,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SpecDelta struct {
	ID          string    `json:"id"`
	PlanID      string    `json:"plan_id"`
	SpecID      string    `json:"spec_id"`
	DeltaType   string    `json:"delta_type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (g *Graph) SaveSpec(spec *Spec) (string, error) {
	if spec.ID == "" {
		spec.ID = "spec-" + uuid.New().String()
	}
	if spec.Status == "" {
		spec.Status = "active"
	}

	_, err := g.db.Exec(`
		INSERT INTO specs (id, capability, title, requirement, scenarios, status, node_id, imported_from, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(id) DO UPDATE SET
			capability=excluded.capability,
			title=excluded.title,
			requirement=excluded.requirement,
			scenarios=excluded.scenarios,
			status=excluded.status,
			imported_from=excluded.imported_from,
			updated_at=datetime('now')
	`, spec.ID, spec.Capability, spec.Title, spec.Requirement, spec.Scenarios, spec.Status, spec.NodeID, spec.ImportedFrom)

	return spec.ID, err
}

func (g *Graph) ListSpecs(capability, status string) ([]Spec, error) {
	query := "SELECT id, capability, title, requirement, scenarios, status, node_id, imported_from, created_at, updated_at FROM specs WHERE 1=1"
	args := []interface{}{}

	if capability != "" {
		query += " AND capability = ?"
		args = append(args, capability)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	} else {
		query += " AND status != 'deprecated'"
	}
	query += " ORDER BY capability, title"

	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var specs []Spec
	for rows.Next() {
		var s Spec
		err := rows.Scan(&s.ID, &s.Capability, &s.Title, &s.Requirement, &s.Scenarios, &s.Status, &s.NodeID, &s.ImportedFrom, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}
	return specs, nil
}

func (g *Graph) GetSpec(id string) (*Spec, error) {
	var s Spec
	err := g.db.QueryRow(`
		SELECT id, capability, title, requirement, scenarios, status, node_id, imported_from, created_at, updated_at
		FROM specs WHERE id = ?
	`, id).Scan(&s.ID, &s.Capability, &s.Title, &s.Requirement, &s.Scenarios, &s.Status, &s.NodeID, &s.ImportedFrom, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (g *Graph) CheckSpec(proposedChange, module string) (map[string]interface{}, error) {
	var affected []Spec
	var violated []Spec
	var suggestions []string

	rows, err := g.db.Query(`
		SELECT id, capability, title, requirement, scenarios, status
		FROM specs WHERE status = 'active'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changeLower := proposedChange
	for rows.Next() {
		var s Spec
		err := rows.Scan(&s.ID, &s.Capability, &s.Title, &s.Requirement, &s.Scenarios, &s.Status)
		if err != nil {
			continue
		}

		keywords := map[string]bool{}
		for _, kw := range []string{s.Capability, s.Title} {
			keywords[kw] = true
		}

		matched := false
		for kw := range keywords {
			if len(kw) > 3 && len(changeLower) > 3 {
				matched = true
				break
			}
		}

		if matched {
			affected = append(affected, s)
			if s.Status == "violated" {
				violated = append(violated, s)
			}
		}
	}

	if len(affected) == 0 {
		suggestions = append(suggestions, "Consider creating a new spec if this is a new requirement")
	}

	result := map[string]interface{}{
		"affected_specs":      affected,
		"violated_specs":      violated,
		"new_specs_suggested": suggestions,
	}

	return result, nil
}

func (g *Graph) SaveSpecDelta(delta *SpecDelta) (string, error) {
	if delta.ID == "" {
		delta.ID = "delta-" + uuid.New().String()
	}

	_, err := g.db.Exec(`
		INSERT INTO spec_deltas (id, plan_id, spec_id, delta_type, description)
		VALUES (?, ?, ?, ?, ?)
	`, delta.ID, delta.PlanID, delta.SpecID, delta.DeltaType, delta.Description)

	return delta.ID, err
}

func (g *Graph) GetSpecDeltas(planID string) ([]SpecDelta, error) {
	rows, err := g.db.Query(`
		SELECT id, plan_id, spec_id, delta_type, description, created_at
		FROM spec_deltas WHERE plan_id = ?
	`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deltas []SpecDelta
	for rows.Next() {
		var d SpecDelta
		err := rows.Scan(&d.ID, &d.PlanID, &d.SpecID, &d.DeltaType, &d.Description, &d.CreatedAt)
		if err != nil {
			continue
		}
		deltas = append(deltas, d)
	}
	return deltas, nil
}

func (g *Graph) UpdateSpecStatus(id, status string) error {
	_, err := g.db.Exec(`UPDATE specs SET status = ?, updated_at = datetime('now') WHERE id = ?`, status, id)
	return err
}
