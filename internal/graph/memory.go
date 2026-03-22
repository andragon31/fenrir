package graph

import (
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Node struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	TopicKey   string    `json:"topic_key,omitempty"` // Stable key for updates
	Confidence float64   `json:"confidence"`
	Status     string    `json:"status"`
	Scope      string    `json:"scope"`
	Weight     string    `json:"weight"`
	Tags       []string  `json:"tags"`
	SessionID  string    `json:"session_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Edge struct {
	ID        string    `json:"id"`
	FromID    string    `json:"from_id"`
	ToID      string    `json:"to_id"`
	Relation  string    `json:"relation"`
	Weight    float64   `json:"weight"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

func (g *Graph) SaveNode(node *Node) (string, error) {
	if node.ID == "" {
		node.ID = "obs-" + uuid.New().String()
	}
	if node.Confidence == 0 {
		node.Confidence = 1.0
	}
	if node.Status == "" {
		node.Status = "active"
	}
	if node.Scope == "" {
		node.Scope = "global"
	}
	if node.Weight == "" {
		node.Weight = "soft"
	}

	tagsJSON := "[]"
	if len(node.Tags) > 0 {
		tagsJSON = toJSON(node.Tags)
	}

	// Try UPSERT if topic_key is provided
	if node.TopicKey != "" {
		var existingID string
		err := g.db.QueryRow("SELECT id FROM nodes WHERE topic_key = ?", node.TopicKey).Scan(&existingID)
		if err == nil {
			node.ID = existingID // Update existing node if key matches
		}
	}

	tx, err := g.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO nodes (id, "type", title, content, topic_key, confidence, status, scope, weight, tags, session_id, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(id) DO UPDATE SET
			title=excluded.title,
			content=excluded.content,
			confidence=excluded.confidence,
			status=excluded.status,
			scope=excluded.scope,
			weight=excluded.weight,
			tags=excluded.tags,
			session_id=excluded.session_id,
			updated_at=datetime('now')
	`, node.ID, node.Type, node.Title, node.Content, node.TopicKey, node.Confidence, node.Status, node.Scope, node.Weight, tagsJSON, node.SessionID)

	if err != nil {
		return "", err
	}

	tx.Exec("DELETE FROM nodes_fts WHERE rowid = (SELECT rowid FROM nodes WHERE id = ?)", node.ID)
	_, err = tx.Exec(`
		INSERT INTO nodes_fts(rowid, title, content)
		SELECT rowid, title, content FROM nodes WHERE id = ?
	`, node.ID)
	
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	if node.Type == "decision" || node.Type == "policy" {
		g.cacheMutex.Lock()
		g.policyCache = nil
		g.cacheMutex.Unlock()
	}

	return node.ID, nil
}

func (g *Graph) SaveEdge(edge *Edge) (string, error) {
	if edge.ID == "" {
		edge.ID = "edge-" + uuid.New().String()
	}

	_, err := g.db.Exec(`
		INSERT INTO edges (id, from_id, to_id, relation, weight, metadata)
		VALUES (?, ?, ?, ?, ?, ?)
	`, edge.ID, edge.FromID, edge.ToID, edge.Relation, edge.Weight, edge.Metadata)

	return edge.ID, err
}

func (g *Graph) Search(query, nodeType, scope string, limit int, includeRelated bool) ([]Node, error) {
	var rows *sql.Rows
	var err error

	if query == "" {
		sqlQuery := "SELECT id, type, title, content, confidence, status, scope, weight, tags, session_id, created_at, updated_at FROM nodes WHERE 1=1"
		args := []interface{}{}

		if nodeType != "" {
			sqlQuery += " AND type = ?"
			args = append(args, nodeType)
		}
		if scope != "" {
			sqlQuery += " AND scope = ?"
			args = append(args, scope)
		}
		sqlQuery += " ORDER BY updated_at DESC LIMIT ?"
		args = append(args, limit)

		rows, err = g.db.Query(sqlQuery, args...)
	} else {
		sqlQuery := `
			SELECT n.id, n.type, n.title, n.content, n.confidence, n.status, n.scope, n.weight, n.tags, n.session_id, n.created_at, n.updated_at
			FROM nodes n
			JOIN nodes_fts ON n.rowid = nodes_fts.rowid
			WHERE nodes_fts MATCH ?
		`
		args := []interface{}{query}

		if nodeType != "" {
			sqlQuery += " AND n.type = ?"
			args = append(args, nodeType)
		}
		if scope != "" {
			sqlQuery += " AND n.scope = ?"
			args = append(args, scope)
		}
		sqlQuery += " ORDER BY rank LIMIT ?"
		args = append(args, limit)

		rows, err = g.db.Query(sqlQuery, args...)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var node Node
		var tagsJSON string
		err := rows.Scan(&node.ID, &node.Type, &node.Title, &node.Content, &node.Confidence, &node.Status, &node.Scope, &node.Weight, &tagsJSON, &node.SessionID, &node.CreatedAt, &node.UpdatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsJSON), &node.Tags)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (g *Graph) GetTimeline(nodeID string, depth int) ([]Node, []Edge, error) {
	var nodes []Node
	var edges []Edge
	visited := make(map[string]bool)

	if err := g.traverseGraph(nodeID, depth, visited, &nodes, &edges); err != nil {
		return nil, nil, err
	}

	return nodes, edges, nil
}

func (g *Graph) traverseGraph(nodeID string, depth int, visited map[string]bool, nodes *[]Node, edges *[]Edge) error {
	if depth < 0 || visited[nodeID] {
		return nil
	}
	visited[nodeID] = true

	var node Node
	var tagsJSON string
	err := g.db.QueryRow(`
		SELECT id, type, title, content, confidence, status, scope, weight, tags, session_id, created_at, updated_at
		FROM nodes WHERE id = ?
	`, nodeID).Scan(&node.ID, &node.Type, &node.Title, &node.Content, &node.Confidence, &node.Status, &node.Scope, &node.Weight, &tagsJSON, &node.SessionID, &node.CreatedAt, &node.UpdatedAt)
	if err != nil {
		return err
	}
	json.Unmarshal([]byte(tagsJSON), &node.Tags)
	*nodes = append(*nodes, node)

	rows, err := g.db.Query(`
		SELECT id, from_id, to_id, relation, weight, metadata, created_at
		FROM edges WHERE from_id = ? OR to_id = ?
	`, nodeID, nodeID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var edge Edge
		err := rows.Scan(&edge.ID, &edge.FromID, &edge.ToID, &edge.Relation, &edge.Weight, &edge.Metadata, &edge.CreatedAt)
		if err != nil {
			return err
		}
		*edges = append(*edges, edge)

		nextNodeID := edge.ToID
		if edge.ToID == nodeID {
			nextNodeID = edge.FromID
		}
		if !visited[nextNodeID] {
			g.traverseGraph(nextNodeID, depth-1, visited, nodes, edges)
		}
	}

	return nil
}

func (g *Graph) SanitizeSecrets(content string) string {
	// Simple patterns for keys
	patterns := []string{
		`sk-[A-Za-z0-9-]{32,}`,
		`AKIA[A-Za-z0-9]{16}`,
		`AIza[A-Za-z0-9_-]{35}`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		content = re.ReplaceAllString(content, "[SECRET]")
	}
	return content
}

func (g *Graph) SanitizePrivateTags(content string) string {
	return strings.ReplaceAll(content, "[PRIVATE]", "[REDACTED]")
}

func (g *Graph) UpdateConfidence(nodeID string, confidence float64, reason string) error {
	_, err := g.db.Exec(`
		UPDATE nodes 
		SET confidence = ?, 
		    content = content || "\n\n**Confidence Update:** " || ?
		WHERE id = ?
	`, confidence, reason, nodeID)
	return err
}

func (g *Graph) GetSessionNodes(sessionID string) ([]Node, error) {
	rows, err := g.db.Query(`
		SELECT id, type, title, content, confidence, status, scope, weight, tags, session_id, created_at, updated_at
		FROM nodes WHERE session_id = ?
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var node Node
		var tagsJSON string
		err := rows.Scan(&node.ID, &node.Type, &node.Title, &node.Content, &node.Confidence, &node.Status, &node.Scope, &node.Weight, &tagsJSON, &node.SessionID, &node.CreatedAt, &node.UpdatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsJSON), &node.Tags)
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
