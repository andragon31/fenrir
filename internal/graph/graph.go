package graph

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const (
	MaxNodesWarning = 50000
	DBVersion       = 1
)

type Graph struct {
	db          *sql.DB
	dataDir     string
	nodeCount   int
	policyCache []Policy
	cacheMutex  sync.RWMutex
}

type Policy struct {
	Title   string
	Content string
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
			topic_key TEXT,
			confidence REAL DEFAULT 1.0,
			status TEXT DEFAULT 'active',
			scope TEXT DEFAULT 'global',
			weight TEXT DEFAULT 'soft',
			tags TEXT,
			session_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_nodes_topic_key ON nodes(topic_key) WHERE topic_key IS NOT NULL`,

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

	// Migration: Add topic_key to nodes if missing
	g.db.Exec(`ALTER TABLE nodes ADD COLUMN topic_key TEXT`) // Ignore error if exists

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
