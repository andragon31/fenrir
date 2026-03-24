package graph

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Prompt struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id,omitempty"`
	Text      string    `json:"text"`
	Module    string    `json:"module,omitempty"`
	NodeID    string    `json:"node_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionCheckpoint struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Trigger   string    `json:"trigger"`
	Summary   string    `json:"summary,omitempty"`
	Snapshot  string    `json:"snapshot"`
	CreatedAt time.Time `json:"created_at"`
}

func (g *Graph) SavePrompt(prompt *Prompt) (string, error) {
	if prompt.ID == "" {
		prompt.ID = "prm-" + uuid.New().String()
	}

	_, err := g.db.Exec(`
		INSERT INTO prompts (id, session_id, text, module, node_id)
		VALUES (?, ?, ?, ?, ?)
	`, prompt.ID, prompt.SessionID, prompt.Text, prompt.Module, prompt.NodeID)

	return prompt.ID, err
}

func (g *Graph) GetPrompts(sessionID string, limit int) ([]Prompt, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := g.db.Query(`
		SELECT id, session_id, text, module, node_id, created_at
		FROM prompts WHERE session_id = ?
		ORDER BY created_at DESC LIMIT ?
	`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []Prompt
	for rows.Next() {
		var p Prompt
		err := rows.Scan(&p.ID, &p.SessionID, &p.Text, &p.Module, &p.NodeID, &p.CreatedAt)
		if err != nil {
			continue
		}
		prompts = append(prompts, p)
	}
	return prompts, nil
}

func (g *Graph) SaveCheckpoint(checkpoint *SessionCheckpoint) (string, error) {
	if checkpoint.ID == "" {
		checkpoint.ID = "chk-" + uuid.New().String()
	}

	_, err := g.db.Exec(`
		INSERT INTO session_checkpoints (id, session_id, trigger, summary, snapshot)
		VALUES (?, ?, ?, ?, ?)
	`, checkpoint.ID, checkpoint.SessionID, checkpoint.Trigger, checkpoint.Summary, checkpoint.Snapshot)

	return checkpoint.ID, err
}

func (g *Graph) GetCheckpoint(sessionID string) (*SessionCheckpoint, error) {
	var c SessionCheckpoint
	err := g.db.QueryRow(`
		SELECT id, session_id, trigger, summary, snapshot, created_at
		FROM session_checkpoints WHERE session_id = ?
		ORDER BY created_at DESC LIMIT 1
	`, sessionID).Scan(&c.ID, &c.SessionID, &c.Trigger, &c.Summary, &c.Snapshot, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (g *Graph) GetCheckpoints(sessionID string, limit int) ([]SessionCheckpoint, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := g.db.Query(`
		SELECT id, session_id, trigger, summary, snapshot, created_at
		FROM session_checkpoints WHERE session_id = ?
		ORDER BY created_at DESC LIMIT ?
	`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkpoints []SessionCheckpoint
	for rows.Next() {
		var c SessionCheckpoint
		err := rows.Scan(&c.ID, &c.SessionID, &c.Trigger, &c.Summary, &c.Snapshot, &c.CreatedAt)
		if err != nil {
			continue
		}
		checkpoints = append(checkpoints, c)
	}
	return checkpoints, nil
}

type CompactNode struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Title      string   `json:"title"`
	Date       string   `json:"date"`
	Confidence float64  `json:"confidence"`
	Authority  string   `json:"authority"`
	Module     string   `json:"module"`
	Tags       []string `json:"tags"`
}

func (g *Graph) SearchCompact(query, nodeType, scope string, limit int) ([]CompactNode, error) {
	nodes, err := g.Search(query, nodeType, scope, limit, false)
	if err != nil {
		return nil, err
	}

	compact := make([]CompactNode, 0, len(nodes))
	for _, n := range nodes {
		compact = append(compact, CompactNode{
			ID:         n.ID,
			Type:       n.Type,
			Title:      n.Title,
			Date:       n.UpdatedAt.Format("2006-01-02"),
			Confidence: n.Confidence,
			Authority:  n.Weight,
			Module:     n.Scope,
			Tags:       n.Tags,
		})
	}

	return compact, nil
}

func (g *Graph) GetObservationFull(nodeID string) (map[string]interface{}, error) {
	var node Node
	var tagsJSON string
	err := g.db.QueryRow(`
		SELECT id, type, title, content, confidence, status, scope, weight, tags, session_id, created_at, updated_at
		FROM nodes WHERE id = ?
	`, nodeID).Scan(&node.ID, &node.Type, &node.Title, &node.Content, &node.Confidence, &node.Status, &node.Scope, &node.Weight, &tagsJSON, &node.SessionID, &node.CreatedAt, &node.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(tagsJSON), &node.Tags)

	var edges []Edge
	rows, err := g.db.Query(`
		SELECT id, from_id, to_id, relation, weight, metadata, created_at
		FROM edges WHERE from_id = ? OR to_id = ?
	`, nodeID, nodeID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var e Edge
			rows.Scan(&e.ID, &e.FromID, &e.ToID, &e.Relation, &e.Weight, &e.Metadata, &e.CreatedAt)
			edges = append(edges, e)
		}
	}

	result := map[string]interface{}{
		"node":  node,
		"edges": edges,
	}

	return result, nil
}
