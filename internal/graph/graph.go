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

		`CREATE TABLE IF NOT EXISTS drift_scores (
			id TEXT PRIMARY KEY,
			module TEXT NOT NULL,
			score REAL DEFAULT 0.0,
			violations INTEGER DEFAULT 0,
			sessions INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS incidents (
			id TEXT PRIMARY KEY,
			"type" TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			severity TEXT DEFAULT 'medium',
			status TEXT DEFAULT 'open',
			module TEXT,
			plan_id TEXT,
			resolved_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS conflicts (
			id TEXT PRIMARY KEY,
			node_a TEXT NOT NULL,
			node_b TEXT NOT NULL,
			description TEXT,
			resolved INTEGER DEFAULT 0,
			resolution TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS pending_rules (
			id TEXT PRIMARY KEY,
			rule_id TEXT NOT NULL,
			reason TEXT,
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS scan_runs (
			id TEXT PRIMARY KEY,
			"type" TEXT NOT NULL,
			status TEXT DEFAULT 'running',
			results TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		)`,

		`CREATE TABLE IF NOT EXISTS velocity_metrics (
			id TEXT PRIMARY KEY,
			metric_name TEXT NOT NULL,
			value REAL,
			module TEXT,
			session_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS pr_summaries (
			id TEXT PRIMARY KEY,
			plan_id TEXT,
			pr_number INTEGER,
			summary TEXT,
			quality_score REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS commit_registry (
			id TEXT PRIMARY KEY,
			commit_hash TEXT NOT NULL,
			plan_id TEXT,
			session_id TEXT,
			message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS specs (
			id TEXT PRIMARY KEY,
			capability TEXT NOT NULL,
			title TEXT NOT NULL,
			requirement TEXT NOT NULL,
			scenarios TEXT NOT NULL,
			status TEXT DEFAULT 'active',
			node_id TEXT REFERENCES nodes(id),
			imported_from TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS spec_deltas (
			id TEXT PRIMARY KEY,
			plan_id TEXT NOT NULL,
			spec_id TEXT REFERENCES specs(id),
			delta_type TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS session_checkpoints (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			trigger TEXT NOT NULL,
			summary TEXT,
			snapshot TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS prompts (
			id TEXT PRIMARY KEY,
			session_id TEXT REFERENCES sessions(id),
			text TEXT NOT NULL,
			module TEXT,
			node_id TEXT REFERENCES nodes(id),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
		"CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type)",
		"CREATE INDEX IF NOT EXISTS idx_nodes_session ON nodes(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status)",
		"CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status)",
		"CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity)",
		"CREATE INDEX IF NOT EXISTS idx_specs_capability ON specs(capability)",
		"CREATE INDEX IF NOT EXISTS idx_specs_status ON specs(status)",
		"CREATE INDEX IF NOT EXISTS idx_spec_deltas_plan ON spec_deltas(plan_id)",
		"CREATE INDEX IF NOT EXISTS idx_checkpoints_session ON session_checkpoints(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_prompts_session ON prompts(session_id)",
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
