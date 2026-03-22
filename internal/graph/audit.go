package graph

import (
	"github.com/google/uuid"
)

func (g *Graph) AddAuditEntry(sessionID, toolCalled, actionType, target, riskLevel, result string) error {
	id := "audit-" + uuid.New().String()
	_, err := g.db.Exec(`
		INSERT INTO audit_log (id, session_id, tool_called, action_type, target, risk_level, result)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, sessionID, toolCalled, actionType, target, riskLevel, result)
	return err
}

func (g *Graph) GetSessionAudit(sessionID, minRisk string) ([]AuditLog, error) {
	sql := `SELECT id, session_id, tool_called, action_type, target, risk_level, result, created_at FROM audit_log WHERE session_id = ?`
	args := []interface{}{sessionID}
	
	if minRisk != "" {
		sql += " AND risk_level = ?"
		args = append(args, minRisk)
	}
	sql += " ORDER BY created_at ASC"

	rows, err := g.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		err := rows.Scan(&l.ID, &l.SessionID, &l.ToolCalled, &l.ActionType, &l.Target, &l.RiskLevel, &l.Result, &l.Timestamp)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
