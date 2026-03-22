package graph

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

const (
	MaxNodesWarning = 50000
	DBVersion       = 1
)

type Graph struct {
	db       *sql.DB
	dataDir  string
	nodeCount int
}

type Node struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
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

type Session struct {
	ID              string    `json:"id"`
	Goal            string    `json:"goal"`
	Status          string    `json:"status"`
	DriftDelta      float64   `json:"drift_delta"`
	ArchViolations  int       `json:"arch_violations"`
	PkgChecks       int       `json:"pkg_checks"`
	ReflectionDepth int       `json:"reflection_depth"`
	FilesModified   []string  `json:"files_modified"`
	Discoveries     string    `json:"discoveries"`
	Accomplished    string    `json:"accomplished"`
	OpenQuestions   string    `json:"open_questions"`
	Warnings        []string  `json:"warnings"`
	ToolCalls       int       `json:"tool_calls"`
	StartedAt       time.Time `json:"started_at"`
	ClosedAt        *time.Time `json:"closed_at"`
}

type SessionDNA struct {
	ID              string    `json:"id"`
	Goal            string    `json:"goal"`
	Status          string    `json:"status"`
	DriftDelta      float64   `json:"drift_delta"`
	ArchViolations  int       `json:"arch_violations"`
	PkgChecks       int       `json:"pkg_checks"`
	ReflectionDepth int       `json:"reflection_depth"`
	FilesModified   []string  `json:"files_modified"`
	Discoveries     string    `json:"discoveries"`
	Accomplished    string    `json:"accomplished"`
	OpenQuestions   string    `json:"open_questions"`
	Warnings        []string  `json:"warnings"`
	ToolCalls       int       `json:"tool_calls"`
	StartedAt       string    `json:"started_at"`
	ClosedAt        string    `json:"closed_at"`
}

type AuditLog struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	ToolCalled string    `json:"tool_called"`
	ActionType string    `json:"action_type"`
	Target     string    `json:"target"`
	RiskLevel  string    `json:"risk_level"`
	Result     string    `json:"result"`
	Metadata   string    `json:"metadata"`
	Timestamp  time.Time `json:"timestamp"`
}

type DriftScore struct {
	Module     string    `json:"module"`
	Score      float64   `json:"score"`
	Violations int       `json:"violations"`
	Sessions   int       `json:"sessions"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PkgCache struct {
	ID        string    `json:"id"`
	Ecosystem string    `json:"ecosystem"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Exists    bool      `json:"exists"`
	Trusted   bool      `json:"trusted"`
	CVECount  int       `json:"cve_count"`
	License   string    `json:"license"`
	Downloads int       `json:"downloads"`
	AgeDays   int       `json:"age_days"`
	Response  string    `json:"response"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PkgCheckResult struct {
	Exists          bool   `json:"exists"`
	Trusted         bool   `json:"trusted"`
	CVECount        int    `json:"cve_count"`
	License         string `json:"license"`
	Downloads       int    `json:"downloads"`
	AgeDays         int    `json:"age_days"`
	Warning         string `json:"warning"`
	SimilarPackages []string `json:"similar_legitimate"`
	CVEs            []CVE  `json:"cves"`
}

type CVE struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
}

type ContextResult struct {
	ActiveSessions    int         `json:"active_sessions"`
	RecentObservations int        `json:"recent_observations"`
	Observations      []Node      `json:"observations"`
	Predictions       []Prediction `json:"predictions"`
	DriftAlerts       []DriftAlert `json:"drift_alerts"`
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
	SessionID  string `json:"session_id"`
	Action     string `json:"action"`
	Timestamp  string `json:"timestamp"`
	Violations int    `json:"violations"`
}

type Decision struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Rationale string `json:"rationale"`
	Weight    string `json:"weight"`
	Scope     string `json:"scope"`
}

type SyncStatus struct {
	Chunks    int    `json:"chunks"`
	LastSync  string `json:"last_sync"`
}

type Stats struct {
	TotalNodes       int `json:"total_nodes"`
	TotalEdges       int `json:"total_edges"`
	TotalSessions    int `json:"total_sessions"`
	ActiveSessions   int `json:"active_sessions"`
	TotalDecisions   int `json:"total_decisions"`
	AuditEntries     int `json:"audit_entries"`
	CachedPackages   int `json:"cached_packages"`
}

type Issue struct {
	Type        string `json:"type"`
	Package     string `json:"package"`
	Description string `json:"description"`
}

func New(dataDir string) (*Graph, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "fenrir.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)

	g := &Graph{
		db:      db,
		dataDir: dataDir,
	}

	g.checkNodeCount()

	return g, nil
}

func (g *Graph) checkNodeCount() {
	var count int
	g.db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count)
	g.nodeCount = count

	if count > MaxNodesWarning {
		fmt.Printf("⚠️  Warning: %d nodes in graph (limit: %d)\n", count, MaxNodesWarning)
	}
}

func (g *Graph) Init() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			"type" TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			confidence REAL DEFAULT 1.0,
			status TEXT DEFAULT 'active',
			scope TEXT DEFAULT 'global',
			weight TEXT DEFAULT 'soft',
			tags TEXT,
			session_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE VIRTUAL TABLE IF NOT EXISTS nodes_fts USING fts5(
			title, content,
			content='nodes',
			content_rowid='rowid'
		)`,

		`CREATE TABLE IF NOT EXISTS edges (
			id TEXT PRIMARY KEY,
			from_id TEXT NOT NULL,
			to_id TEXT NOT NULL,
			relation TEXT NOT NULL,
			weight REAL DEFAULT 1.0,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (from_id) REFERENCES nodes(id),
			FOREIGN KEY (to_id) REFERENCES nodes(id)
		)`,

		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			goal TEXT,
			status TEXT DEFAULT 'active',
			drift_delta REAL DEFAULT 0.0,
			arch_violations INTEGER DEFAULT 0,
			pkg_checks INTEGER DEFAULT 0,
			reflection_depth INTEGER DEFAULT 0,
			files_modified TEXT,
			discoveries TEXT,
			accomplished TEXT,
			open_questions TEXT,
			warnings TEXT,
			tool_calls INTEGER DEFAULT 0,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			closed_at DATETIME
		)`,

		`CREATE TABLE IF NOT EXISTS audit_log (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			tool_called TEXT NOT NULL,
			action_type TEXT NOT NULL,
			target TEXT,
			risk_level TEXT DEFAULT 'low',
			result TEXT DEFAULT 'success',
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS pkg_cache (
			id TEXT PRIMARY KEY,
			ecosystem TEXT NOT NULL,
			name TEXT NOT NULL,
			version TEXT,
			"exists" INTEGER NOT NULL,
			trusted INTEGER DEFAULT 1,
			cve_count INTEGER DEFAULT 0,
			license TEXT,
			downloads INTEGER DEFAULT 0,
			age_days INTEGER DEFAULT 0,
			response TEXT,
			cached_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS drift_scores (
			id TEXT PRIMARY KEY,
			module TEXT NOT NULL,
			score REAL DEFAULT 0.0,
			violations INTEGER DEFAULT 0,
			sessions INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS meta (
			key TEXT PRIMARY KEY,
			value TEXT
		)`,
	}

	for _, schema := range schemas {
		if _, err := g.db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}

	if _, err := g.db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	if _, err := g.db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	indices := []string{
		"CREATE INDEX IF NOT EXISTS idx_edges_from ON edges(from_id)",
		"CREATE INDEX IF NOT EXISTS idx_edges_to ON edges(to_id)",
		"CREATE INDEX IF NOT EXISTS idx_edges_rel ON edges(relation)",
		"CREATE INDEX IF NOT EXISTS idx_audit_session ON audit_log(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_audit_risk ON audit_log(risk_level)",
		"CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type)",
		"CREATE INDEX IF NOT EXISTS idx_nodes_session ON nodes(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status)",
		"CREATE INDEX IF NOT EXISTS idx_pkg_cache_expires ON pkg_cache(expires_at)",
	}

	for _, idx := range indices {
		if _, err := g.db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

func (g *Graph) Close() error {
	return g.db.Close()
}

func (g *Graph) DB() *sql.DB {
	return g.db
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

	tx, err := g.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO nodes (id, type, title, content, confidence, status, scope, weight, tags, session_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, node.ID, node.Type, node.Title, node.Content, node.Confidence, node.Status, node.Scope, node.Weight, tagsJSON, node.SessionID)
	if err != nil {
		return "", err
	}

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

	g.nodeCount++
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

func (g *Graph) StartSession(goal, module string) (string, *ContextResult, error) {
	sessionID := "ses-" + uuid.New().String()

	_, err := g.db.Exec(`
		INSERT INTO sessions (id, goal, status)
		VALUES (?, ?, 'active')
	`, sessionID, goal)
	if err != nil {
		return "", nil, err
	}

	ctx, err := g.GetContext(module, 20, true)
	if err != nil {
		return "", nil, err
	}

	return sessionID, ctx, nil
}

func (g *Graph) EndSession(sessionID, goal, discoveries, accomplished, openQuestions string, filesModified []string) (*SessionDNA, error) {
	filesJSON, _ := json.Marshal(filesModified)

	_, err := g.db.Exec(`
		UPDATE sessions
		SET status = 'closed',
		    goal = COALESCE(?, goal),
		    discoveries = ?,
		    accomplished = ?,
		    open_questions = ?,
		    files_modified = ?,
		    closed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, goal, discoveries, accomplished, openQuestions, string(filesJSON), sessionID)
	if err != nil {
		return nil, err
	}

	return g.GetSessionDNA(sessionID)
}

func (g *Graph) GetSessionDNA(sessionID string) (*SessionDNA, error) {
	var dna SessionDNA
	var filesJSON, warningsJSON string

	err := g.db.QueryRow(`
		SELECT id, goal, status, drift_delta, arch_violations, pkg_checks, reflection_depth,
		       files_modified, discoveries, accomplished, open_questions, warnings, tool_calls,
		       started_at, closed_at
		FROM sessions WHERE id = ?
	`, sessionID).Scan(&dna.ID, &dna.Goal, &dna.Status, &dna.DriftDelta, &dna.ArchViolations,
		&dna.PkgChecks, &dna.ReflectionDepth, &filesJSON, &dna.Discoveries, &dna.Accomplished,
		&dna.OpenQuestions, &warningsJSON, &dna.ToolCalls, &dna.StartedAt, &dna.ClosedAt)
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(filesJSON), &dna.FilesModified)
	json.Unmarshal([]byte(warningsJSON), &dna.Warnings)

	if dna.ClosedAt != "" {
		dna.ClosedAt = dna.ClosedAt
	}

	return &dna, nil
}

func (g *Graph) ListSessions(limit int) ([]Session, error) {
	rows, err := g.db.Query(`
		SELECT id, goal, status, drift_delta, arch_violations, pkg_checks, started_at, closed_at
		FROM sessions ORDER BY started_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		err := rows.Scan(&s.ID, &s.Goal, &s.Status, &s.DriftDelta, &s.ArchViolations, &s.PkgChecks, &s.StartedAt, &s.ClosedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
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
		p.Severity = "warning"
		p.Message = fmt.Sprintf("This module has had %d violations in recent sessions", count)
		predictions = append(predictions, p)
	}

	return predictions, nil
}

func (g *Graph) SaveDecision(decision, rationale, weight, scope string) (string, error) {
	node := &Node{
		ID:         "dec-" + uuid.New().String(),
		Type:       "decision",
		Title:      decision,
		Content:    rationale,
		Weight:     weight,
		Scope:      scope,
		Confidence: 1.0,
		Status:     "active",
	}

	return g.SaveNode(node)
}

func (g *Graph) VerifyAction(proposedAction, context string) (string, string, error) {
	rows, err := g.db.Query(`
		SELECT id, title, content, weight, scope
		FROM nodes
		WHERE type = 'decision' AND status = 'active'
		AND (scope = 'global' OR scope LIKE ?)
	`, "%"+context+"%")
	if err != nil {
		return "allowed", "", nil
	}
	defer rows.Close()

	for rows.Next() {
		var id, title, content, weight, scope string
		if err := rows.Scan(&id, &title, &content, &weight, &scope); err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(proposedAction), strings.ToLower(title)) ||
		   strings.Contains(strings.ToLower(proposedAction), strings.ToLower(content)) {
			switch weight {
			case "critical":
				return "blocked", fmt.Sprintf("Violates critical decision: %s - %s", title, content), nil
			case "hard":
				return "warning", fmt.Sprintf("Violates hard decision: %s - %s", title, content), nil
			case "soft":
				return "suggestion", fmt.Sprintf("Consider: %s", title), nil
			}
		}
	}

	return "allowed", "", nil
}

func (g *Graph) ListDecisions() ([]Decision, error) {
	rows, err := g.db.Query(`
		SELECT id, title, content, weight, scope
		FROM nodes
		WHERE type = 'decision' AND status = 'active'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []Decision
	for rows.Next() {
		var d Decision
		if err := rows.Scan(&d.ID, &d.Title, &d.Rationale, &d.Weight, &d.Scope); err != nil {
			continue
		}
		decisions = append(decisions, d)
	}

	return decisions, nil
}

func (g *Graph) GetDriftScores(module string) ([]DriftScore, error) {
	query := "SELECT module, score, violations, sessions, updated_at FROM drift_scores WHERE 1=1"
	args := []interface{}{}

	if module != "" {
		query += " AND module LIKE ?"
		args = append(args, "%"+module+"%")
	}

	query += " ORDER BY score DESC"

	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []DriftScore
	for rows.Next() {
		var s DriftScore
		if err := rows.Scan(&s.Module, &s.Score, &s.Violations, &s.Sessions, &s.UpdatedAt); err != nil {
			continue
		}
		scores = append(scores, s)
	}

	return scores, nil
}

func (g *Graph) CalculateDrift(sessionID string) (float64, error) {
	var totalDecisions, violations int

	g.db.QueryRow(`
		SELECT COUNT(*) FROM nodes WHERE type = 'decision' AND status = 'active'
	`).Scan(&totalDecisions)

	g.db.QueryRow(`
		SELECT COUNT(*) FROM nodes WHERE type = 'violation' AND session_id = ?
	`, sessionID).Scan(&violations)

	if totalDecisions == 0 {
		return 0, nil
	}

	driftScore := float64(violations) / float64(totalDecisions)

	module := ""
	g.db.QueryRow("SELECT scope FROM nodes WHERE session_id = ? LIMIT 1", sessionID).Scan(&module)

	if module != "" && module != "global" {
		_, err := g.db.Exec(`
			INSERT OR REPLACE INTO drift_scores (id, module, score, violations, sessions, updated_at)
			VALUES (?, ?, ?, ?,
			        COALESCE((SELECT sessions FROM drift_scores WHERE module = ?), 0) + 1,
			        CURRENT_TIMESTAMP)
		`, "drift-"+uuid.New().String(), module, driftScore, violations, module)
		if err != nil {
			return driftScore, err
		}
	}

	return driftScore, nil
}

func (g *Graph) LogAudit(sessionID, toolCalled, actionType, target, riskLevel, result, metadata string) error {
	id := "log-" + uuid.New().String()

	_, err := g.db.Exec(`
		INSERT INTO audit_log (id, session_id, tool_called, action_type, target, risk_level, result, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, sessionID, toolCalled, actionType, target, riskLevel, result, metadata)

	return err
}

func (g *Graph) GetSessionAudit(sessionID, riskLevel string) ([]AuditLog, error) {
	query := "SELECT id, session_id, tool_called, action_type, target, risk_level, result, metadata, created_at FROM audit_log WHERE session_id = ?"
	args := []interface{}{sessionID}

	if riskLevel != "" {
		query += " AND risk_level >= ?"
		args = append(args, riskLevel)
	}

	query += " ORDER BY created_at DESC"

	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.SessionID, &l.ToolCalled, &l.ActionType, &l.Target, &l.RiskLevel, &l.Result, &l.Metadata, &l.Timestamp); err != nil {
			continue
		}
		logs = append(logs, l)
	}

	return logs, nil
}

func (g *Graph) CheckPackage(name, ecosystem, version string) (*PkgCheckResult, error) {
	cacheID := fmt.Sprintf("%s:%s:%s", ecosystem, name, version)

	var cached PkgCache
	err := g.db.QueryRow(`
		SELECT id, ecosystem, name, version, exists, trusted, cve_count, license, downloads, age_days, response, cached_at, expires_at
		FROM pkg_cache WHERE id = ? AND expires_at > datetime('now')
	`, cacheID).Scan(&cached.ID, &cached.Ecosystem, &cached.Name, &cached.Version, &cached.Exists, &cached.Trusted, &cached.CVECount, &cached.License, &cached.Downloads, &cached.AgeDays, &cached.Response, &cached.CachedAt, &cached.ExpiresAt)

	if err == nil {
		return &PkgCheckResult{
			Exists:    cached.Exists,
			Trusted:   cached.Trusted,
			CVECount:  cached.CVECount,
			License:   cached.License,
			Downloads: cached.Downloads,
			AgeDays:   cached.AgeDays,
		}, nil
	}

	result := &PkgCheckResult{
		Exists: true,
		Trusted: true,
	}

	expiresAt := time.Now().Add(time.Hour)
	_, err = g.db.Exec(`
		INSERT INTO pkg_cache (id, ecosystem, name, version, exists, trusted, cve_count, license, downloads, age_days, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cacheID, ecosystem, name, version, result.Exists, result.Trusted, result.CVECount, result.License, result.Downloads, result.AgeDays, expiresAt)

	return result, err
}

func (g *Graph) AuditDependencies(manifestPath string) ([]Issue, error) {
	return []Issue{}, nil
}

func (g *Graph) GetInsights() ([]Insight, error) {
	rows, err := g.db.Query(`
		SELECT type, title, COUNT(*) as count
		FROM nodes
		WHERE type IN ('bugfix', 'violation')
		GROUP BY type, title
		HAVING count >= 3
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []Insight
	for rows.Next() {
		var i Insight
		if err := rows.Scan(&i.Type, &i.Title, &i.Count); err != nil {
			continue
		}
		i.Details = fmt.Sprintf("This pattern has occurred %d times", i.Count)
		insights = append(insights, i)
	}

	return insights, nil
}

func (g *Graph) Trace(target string, depth int) ([]TraceEntry, error) {
	rows, err := g.db.Query(`
		SELECT DISTINCT s.id, s.goal, s.started_at, COALESCE(s.arch_violations, 0) as violations
		FROM sessions s
		LEFT JOIN nodes n ON n.session_id = s.id
		WHERE n.content LIKE ? OR n.title LIKE ? OR s.goal LIKE ?
		ORDER BY s.started_at DESC
		LIMIT 20
	`, "%"+target+"%", "%"+target+"%", "%"+target+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []TraceEntry
	for rows.Next() {
		var e TraceEntry
		if err := rows.Scan(&e.SessionID, &e.Action, &e.Timestamp, &e.Violations); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	return entries, nil
}

func (g *Graph) UpdateConfidence(nodeID string, confidence float64, reason string) error {
	_, err := g.db.Exec(`
		UPDATE nodes SET confidence = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, confidence, nodeID)
	return err
}

func (g *Graph) ExportChunks() error {
	chunksDir := filepath.Join(g.dataDir, "chunks")
	if err := os.MkdirAll(chunksDir, 0755); err != nil {
		return err
	}

	rows, err := g.db.Query("SELECT id, type, title, content, confidence, status, scope, weight, tags, session_id, created_at, updated_at FROM nodes")
	if err != nil {
		return err
	}
	defer rows.Close()

	chunk := []Node{}
	for rows.Next() {
		var n Node
		var tagsJSON string
		if err := rows.Scan(&n.ID, &n.Type, &n.Title, &n.Content, &n.Confidence, &n.Status, &n.Scope, &n.Weight, &tagsJSON, &n.SessionID, &n.CreatedAt, &n.UpdatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &n.Tags)
		chunk = append(chunk, n)
	}

	if len(chunk) > 0 {
		data, _ := json.Marshal(chunk)
		hash := sha256.Sum256(data)
		filename := hex.EncodeToString(hash[:]) + ".jsonl"
		return os.WriteFile(filepath.Join(chunksDir, filename), data, 0644)
	}

	return nil
}

func (g *Graph) ImportChunks() error {
	chunksDir := filepath.Join(g.dataDir, "chunks")
	files, err := os.ReadDir(chunksDir)
	if err != nil {
		return nil
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".jsonl" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(chunksDir, f.Name()))
		if err != nil {
			continue
		}

		var nodes []Node
		if err := json.Unmarshal(data, &nodes); err != nil {
			continue
		}

		for _, n := range nodes {
			g.SaveNode(&n)
		}
	}

	return nil
}

func (g *Graph) GetSyncStatus() (*SyncStatus, error) {
	chunksDir := filepath.Join(g.dataDir, "chunks")
	files, _ := os.ReadDir(chunksDir)

	status := &SyncStatus{
		Chunks:   len(files),
		LastSync: "never",
	}

	var lastSync time.Time
	g.db.QueryRow("SELECT MAX(created_at) FROM audit_log").Scan(&lastSync)
	if !lastSync.IsZero() {
		status.LastSync = lastSync.Format(time.RFC3339)
	}

	return status, nil
}

func (g *Graph) GetStats() (*Stats, error) {
	stats := &Stats{}

	g.db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&stats.TotalNodes)
	g.db.QueryRow("SELECT COUNT(*) FROM edges").Scan(&stats.TotalEdges)
	g.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&stats.TotalSessions)
	g.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE status = 'active'").Scan(&stats.ActiveSessions)
	g.db.QueryRow("SELECT COUNT(*) FROM nodes WHERE type = 'decision'").Scan(&stats.TotalDecisions)
	g.db.QueryRow("SELECT COUNT(*) FROM audit_log").Scan(&stats.AuditEntries)
	g.db.QueryRow("SELECT COUNT(*) FROM pkg_cache").Scan(&stats.CachedPackages)

	return stats, nil
}

func (g *Graph) ExportToJSON(filePath string) error {
	nodes, _, err := g.GetTimeline("", 100)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (g *Graph) ImportFromJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
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

func (g *Graph) AddAuditEntry(sessionID, toolCalled, actionType, target, riskLevel, result string) error {
	return g.LogAudit(sessionID, toolCalled, actionType, target, riskLevel, result, "")
}

func (g *Graph) SanitizePrivateTags(content string) string {
	replacer := strings.NewReplacer("<private>", "[REDACTED]", "</private>", "[/REDACTED]")
	return replacer.Replace(content)
}

func (g *Graph) SanitizeSecrets(content string) string {
	patterns := []struct {
		name  string
		regex string
	}{
		{"OpenAI", `sk-[A-Za-z0-9-_]{20,}`},
		{"Anthropic", `sk-ant-[A-Za-z0-9-_]{20,}`},
		{"GitHub", `ghp_[A-Za-z0-9]{36}`},
		{"AWS", `AKIA[A-Za-z0-9]{16}`},
		{"JWT", `eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+`},
	}

	result := content
	for _, p := range patterns {
		re := strings.NewReplacer(p.name+":", "", p.regex, "[REDACTED]")
		result = re.Replace(result)
	}

	return result
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
