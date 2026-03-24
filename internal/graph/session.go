package graph

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID              string     `json:"id"`
	Goal            string     `json:"goal"`
	Status          string     `json:"status"`
	DriftDelta      float64    `json:"drift_delta"`
	ArchViolations  int        `json:"arch_violations"`
	PkgChecks       int        `json:"pkg_checks"`
	ReflectionDepth int        `json:"reflection_depth"`
	FilesModified   []string   `json:"files_modified"`
	Discoveries     string     `json:"discoveries"`
	Accomplished    string     `json:"accomplished"`
	OpenQuestions   string     `json:"open_questions"`
	Warnings        []string   `json:"warnings"`
	ToolCalls       int        `json:"tool_calls"`
	StartedAt       time.Time  `json:"started_at"`
	ClosedAt        *time.Time `json:"closed_at"`
}

type SessionDNA struct {
	ID              string   `json:"id"`
	Goal            string   `json:"goal"`
	Status          string   `json:"status"`
	DriftDelta      float64  `json:"drift_delta"`
	ArchViolations  int      `json:"arch_violations"`
	PkgChecks       int      `json:"pkg_checks"`
	ReflectionDepth int      `json:"reflection_depth"`
	FilesModified   []string `json:"files_modified"`
	Discoveries     string   `json:"discoveries"`
	Accomplished    string   `json:"accomplished"`
	OpenQuestions   string   `json:"open_questions"`
	Warnings        []string `json:"warnings"`
	ToolCalls       int      `json:"tool_calls"`
	StartedAt       string   `json:"started_at"`
	ClosedAt        string   `json:"closed_at"`
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

	recentObs, _ := g.Search("", "", module, 5, false)
	prompts, _ := g.GetPrompts(sessionID, 5)
	incidents, _ := g.GetIncidents("open", "", 5)
	specs, _ := g.ListSpecs("", "active")

	autoInjected := map[string]interface{}{
		"recent_observations":             recentObs,
		"recent_prompts":                  prompts,
		"open_incidents":                  incidents,
		"affected_specs":                  specs[:min(len(specs), 3)],
		"compaction_checkpoint_available": false,
	}

	lastCheckpoint, _ := g.GetCheckpoint(sessionID)
	if lastCheckpoint != nil {
		autoInjected["compaction_checkpoint_available"] = true
		autoInjected["last_checkpoint"] = lastCheckpoint.CreatedAt.Format(time.RFC3339)
	}

	ctx.AutoInjected = autoInjected

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

	// Update Drift Scores for modified files
	g.UpdateDriftScores(sessionID, filesModified)

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
