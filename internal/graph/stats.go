package graph

type Stats struct {
	TotalNodes       int `json:"total_nodes"`
	TotalEdges       int `json:"total_edges"`
	TotalSessions    int `json:"total_sessions"`
	ActiveSessions   int `json:"active_sessions"`
	TotalDecisions   int `json:"total_decisions"`
	AuditEntries     int `json:"audit_entries"`
	CachedPackages   int `json:"cached_packages"`
}

func (g *Graph) GetStats() (*Stats, error) {
	s := &Stats{}
	g.db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&s.TotalNodes)
	g.db.QueryRow("SELECT COUNT(*) FROM edges").Scan(&s.TotalEdges)
	g.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&s.TotalSessions)
	g.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE status = 'active'").Scan(&s.ActiveSessions)
	g.db.QueryRow("SELECT COUNT(*) FROM nodes WHERE type = 'decision'").Scan(&s.TotalDecisions)
	g.db.QueryRow("SELECT COUNT(*) FROM audit_log").Scan(&s.AuditEntries)
	g.db.QueryRow("SELECT COUNT(*) FROM pkg_cache").Scan(&s.CachedPackages)
	return s, nil
}
