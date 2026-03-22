package graph

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

type SyncStatus struct {
	Chunks    int    `json:"chunks"`
	LastSync  string `json:"last_sync"`
}

func (g *Graph) ExportChunks() error {
	dir := filepath.Join(g.dataDir, "chunks")
	os.MkdirAll(dir, 0755)

	rows, err := g.db.Query("SELECT id, type, title, content, updated_at FROM nodes")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, t, title, content, updated string
		rows.Scan(&id, &t, &title, &content, &updated)
		
		path := filepath.Join(dir, id+".json")
		data, _ := json.MarshalIndent(map[string]string{
			"id": id, "type": t, "title": title, "content": content, "updated": updated,
		}, "", "  ")
		os.WriteFile(path, data, 0644)
	}

	// Git add if in git-enabled environment
	if _, err := os.Stat(filepath.Join(g.dataDir, "..", ".git")); err == nil {
		exec.Command("git", "add", dir).Run()
	}

	return nil
}

func (g *Graph) ImportChunks() error {
	dir := filepath.Join(g.dataDir, "chunks")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	files, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	for _, f := range files {
		data, _ := os.ReadFile(f)
		var node Node
		json.Unmarshal(data, &node)
		g.SaveNode(&node)
	}
	return nil
}

func (g *Graph) ExportToJSON(path string) error {
	rows, err := g.db.Query("SELECT id, type, title, content, topic_key, confidence, status, scope, weight, tags, session_id, created_at, updated_at FROM nodes")
	if err != nil {
		return err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		var tagsJSON string
		rows.Scan(&n.ID, &n.Type, &n.Title, &n.Content, &n.TopicKey, &n.Confidence, &n.Status, &n.Scope, &n.Weight, &tagsJSON, &n.SessionID, &n.CreatedAt, &n.UpdatedAt)
		json.Unmarshal([]byte(tagsJSON), &n.Tags)
		nodes = append(nodes, n)
	}

	data, _ := json.MarshalIndent(nodes, "", "  ")
	return os.WriteFile(path, data, 0644)
}

func (g *Graph) ImportFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var nodes []Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		return err
	}

	for _, n := range nodes {
		g.SaveNode(&n)
	}
	return nil
}
