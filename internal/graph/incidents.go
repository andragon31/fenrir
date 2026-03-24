package graph

import (
	"time"

	"github.com/google/uuid"
)

type Incident struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Severity    string     `json:"severity"`
	Status      string     `json:"status"`
	Module      string     `json:"module,omitempty"`
	PlanID      string     `json:"plan_id,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (g *Graph) SaveIncident(incident *Incident) (string, error) {
	if incident.ID == "" {
		incident.ID = "inc-" + uuid.New().String()
	}
	if incident.Severity == "" {
		incident.Severity = "medium"
	}
	if incident.Status == "" {
		incident.Status = "open"
	}

	_, err := g.db.Exec(`
		INSERT INTO incidents (id, type, title, description, severity, status, module, plan_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, incident.ID, incident.Type, incident.Title, incident.Description, incident.Severity, incident.Status, incident.Module, incident.PlanID)

	return incident.ID, err
}

func (g *Graph) GetIncidents(status, module string, limit int) ([]Incident, error) {
	if limit <= 0 {
		limit = 10
	}

	query := "SELECT id, type, title, description, severity, status, module, plan_id, resolved_at, created_at FROM incidents WHERE 1=1"
	args := []interface{}{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if module != "" {
		query += " AND module = ?"
		args = append(args, module)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []Incident
	for rows.Next() {
		var i Incident
		err := rows.Scan(&i.ID, &i.Type, &i.Title, &i.Description, &i.Severity, &i.Status, &i.Module, &i.PlanID, &i.ResolvedAt, &i.CreatedAt)
		if err != nil {
			continue
		}
		incidents = append(incidents, i)
	}
	return incidents, nil
}

func (g *Graph) ResolveIncident(id string) error {
	_, err := g.db.Exec(`
		UPDATE incidents SET status = 'resolved', resolved_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, id)
	return err
}

func (g *Graph) GetIncident(id string) (*Incident, error) {
	var i Incident
	err := g.db.QueryRow(`
		SELECT id, type, title, description, severity, status, module, plan_id, resolved_at, created_at
		FROM incidents WHERE id = ?
	`, id).Scan(&i.ID, &i.Type, &i.Title, &i.Description, &i.Severity, &i.Status, &i.Module, &i.PlanID, &i.ResolvedAt, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}
