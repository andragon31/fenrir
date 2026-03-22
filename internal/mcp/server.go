package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/andragon31/fenrir/internal/graph"
	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	graph  *graph.Graph
	logger *log.Logger
	server *server.MCPServer
}

func NewServer(g *graph.Graph, logger *log.Logger) *Server {
	srv := server.NewMCPServer("fenrir", "0.1.0")

	s := &Server{
		graph:  g,
		logger: logger,
		server: srv,
	}

	s.registerAllTools()

	return s
}

func (s *Server) registerAllTools() {
	s.registerMemoryTools()
	s.registerValidatorTools()
	s.registerEnforcerTools()
	s.registerShieldTools()
	s.registerIntelligenceTools()
}

func (s *Server) registerMemoryTools() {
	memSave := mcp.NewTool("mem_save",
		mcp.WithDescription("Save a structured observation to the project knowledge graph"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Short descriptive title")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Type of observation"), mcp.Enum("bugfix", "decision", "pattern", "failed_attempt", "discovery", "config", "refactor")),
		mcp.WithString("topic_key", mcp.Description("Stable key to update evolving topics (e.g. auth-logic, database-schema)")),
		mcp.WithString("what", mcp.Required(), mcp.Description("What was done or discovered")),
		mcp.WithString("why", mcp.Required(), mcp.Description("Why it was necessary")),
		mcp.WithString("where", mcp.Description("Files or modules affected")),
		mcp.WithString("learned", mcp.Description("What to remember for next time")),
		mcp.WithArray("relates_to", mcp.Description("IDs of related nodes"), mcp.WithStringItems()),
		mcp.WithNumber("confidence", mcp.Description("Confidence score 0.0-1.0")),
	)
	s.server.AddTool(memSave, s.handleMemSave)

	memFind := mcp.NewTool("mem_find",
		mcp.WithDescription("Search the knowledge graph using full-text search"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithString("type", mcp.Description("Filter by node type")),
		mcp.WithString("scope", mcp.Description("Filter by module or file path")),
		mcp.WithNumber("limit", mcp.Description("Maximum results")),
	)
	s.server.AddTool(memFind, s.handleMemFind)

	memContext := mcp.NewTool("mem_context",
		mcp.WithDescription("Get relevant context from previous sessions including predictive alerts"),
		mcp.WithString("module", mcp.Description("Current working module path")),
		mcp.WithNumber("limit", mcp.Description("Maximum results")),
		mcp.WithBoolean("include_predictions", mcp.Description("Include predictive alerts")),
	)
	s.server.AddTool(memContext, s.handleMemContext)

	memTimeline := mcp.NewTool("mem_timeline",
		mcp.WithDescription("Get chronological history of a node and its relationships"),
		mcp.WithString("node_id", mcp.Required(), mcp.Description("Node ID")),
		mcp.WithNumber("depth", mcp.Description("Graph traversal depth")),
	)
	s.server.AddTool(memTimeline, s.handleMemTimeline)

	memSessionStart := mcp.NewTool("mem_session_start",
		mcp.WithDescription("Register session start and load predictive context"),
		mcp.WithString("goal", mcp.Required(), mcp.Description("What you intend to accomplish")),
		mcp.WithString("module", mcp.Description("Primary module you'll be working in")),
	)
	s.server.AddTool(memSessionStart, s.handleMemSessionStart)

	memSessionEnd := mcp.NewTool("mem_session_end",
		mcp.WithDescription("Close session and generate Session DNA. MANDATORY before ending any session."),
		mcp.WithString("goal", mcp.Required(), mcp.Description("Session goal")),
		mcp.WithString("discoveries", mcp.Description("Key discoveries made")),
		mcp.WithString("accomplished", mcp.Required(), mcp.Description("What was accomplished")),
		mcp.WithArray("files_modified", mcp.Description("Files modified"), mcp.WithStringItems()),
		mcp.WithString("open_questions", mcp.Description("Unanswered questions")),
	)
	s.server.AddTool(memSessionEnd, s.handleMemSessionEnd)

	memDNA := mcp.NewTool("mem_dna",
		mcp.WithDescription("View Session DNA of past sessions"),
		mcp.WithString("session_id", mcp.Description("Specific session ID")),
		mcp.WithNumber("limit", mcp.Description("Number of sessions to show")),
	)
	s.server.AddTool(memDNA, s.handleMemDNA)
}

func (s *Server) registerValidatorTools() {
	pkgCheck := mcp.NewTool("pkg_check",
		mcp.WithDescription("Validate package existence, trustworthiness and known CVEs before installing"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Package name")),
		mcp.WithString("ecosystem", mcp.Required(), mcp.Description("Package ecosystem"), mcp.Enum("npm", "pypi", "cargo", "nuget")),
		mcp.WithString("version", mcp.Description("Package version")),
	)
	s.server.AddTool(pkgCheck, s.handlePkgCheck)

	pkgLicense := mcp.NewTool("pkg_license",
		mcp.WithDescription("Check package license and compatibility with project policies"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Package name")),
		mcp.WithString("ecosystem", mcp.Required(), mcp.Description("Package ecosystem"), mcp.Enum("npm", "pypi", "cargo", "nuget")),
		mcp.WithString("version", mcp.Description("Package version")),
	)
	s.server.AddTool(pkgLicense, s.handlePkgLicense)

	pkgAudit := mcp.NewTool("pkg_audit",
		mcp.WithDescription("Full audit of all project dependencies for CVEs and license issues"),
		mcp.WithString("manifest_path", mcp.Description("Path to package.json, requirements.txt, etc.")),
	)
	s.server.AddTool(pkgAudit, s.handlePkgAudit)
}

func (s *Server) registerEnforcerTools() {
	archSave := mcp.NewTool("arch_save",
		mcp.WithDescription("Save an architectural decision to the knowledge graph"),
		mcp.WithString("decision", mcp.Required(), mcp.Description("The architectural decision")),
		mcp.WithString("rationale", mcp.Required(), mcp.Description("Why this decision was made")),
		mcp.WithString("scope", mcp.Description("Scope: global, module:<path>")),
		mcp.WithString("weight", mcp.Description("Decision weight"), mcp.Enum("soft", "hard", "critical")),
		mcp.WithArray("tags", mcp.Description("Tags"), mcp.WithStringItems()),
		mcp.WithString("supersedes", mcp.Description("ID of decision this replaces")),
	)
	s.server.AddTool(archSave, s.handleArchSave)

	archVerify := mcp.NewTool("arch_verify",
		mcp.WithDescription("Verify a proposed action against architectural decisions"),
		mcp.WithString("proposed_action", mcp.Required(), mcp.Description("The proposed action")),
		mcp.WithString("context", mcp.Description("File or module context")),
	)
	s.server.AddTool(archVerify, s.handleArchVerify)

	archDrift := mcp.NewTool("arch_drift",
		mcp.WithDescription("Get drift score for a module or file"),
		mcp.WithString("path", mcp.Description("Module or file path")),
	)
	s.server.AddTool(archDrift, s.handleArchDrift)

	policyCheck := mcp.NewTool("policy_check",
		mcp.WithDescription("Check an action against team policies"),
		mcp.WithString("action", mcp.Required(), mcp.Description("The action to check")),
		mcp.WithString("context", mcp.Description("Context for the check")),
	)
	s.server.AddTool(policyCheck, s.handlePolicyCheck)

	predict := mcp.NewTool("predict",
		mcp.WithDescription("Get predictive alerts for current session based on module history"),
		mcp.WithString("module", mcp.Description("Module to predict for")),
		mcp.WithString("context", mcp.Description("Additional context")),
	)
	s.server.AddTool(predict, s.handlePredict)
}

func (s *Server) registerShieldTools() {
	auditLog := mcp.NewTool("audit_log",
		mcp.WithDescription("Log an agent action to the audit trail"),
		mcp.WithString("tool_called", mcp.Required(), mcp.Description("Tool that was called")),
		mcp.WithString("action_type", mcp.Required(), mcp.Description("Action type"), mcp.Enum("read", "write", "execute", "network", "validate")),
		mcp.WithString("target", mcp.Description("Target of the action")),
		mcp.WithString("risk_level", mcp.Description("Risk level"), mcp.Enum("low", "medium", "high", "critical")),
		mcp.WithString("result", mcp.Description("Result"), mcp.Enum("success", "blocked", "warning", "error")),
	)
	s.server.AddTool(auditLog, s.handleAuditLog)

	sessionAudit := mcp.NewTool("session_audit",
		mcp.WithDescription("Get complete audit log for current or specified session"),
		mcp.WithString("session_id", mcp.Description("Session ID")),
		mcp.WithString("risk_level", mcp.Description("Filter by minimum risk level")),
	)
	s.server.AddTool(sessionAudit, s.handleSessionAudit)

	injectGuard := mcp.NewTool("inject_guard",
		mcp.WithDescription("Check content for prompt injection patterns"),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content to check")),
	)
	s.server.AddTool(injectGuard, s.handleInjectGuard)

	fenrirStats := mcp.NewTool("fenrir_stats",
		mcp.WithDescription("System statistics and knowledge graph health"),
	)
	s.server.AddTool(fenrirStats, s.handleFenrirStats)
}

func (s *Server) registerIntelligenceTools() {
	insights := mcp.NewTool("insights",
		mcp.WithDescription("Auto-detected patterns across sessions: recurring bugs, unstable dependencies, stable patterns"),
		mcp.WithString("type", mcp.Description("Type of insights"), mcp.Enum("bugs", "dependencies", "patterns", "all")),
	)
	s.server.AddTool(insights, s.handleInsights)

	trace := mcp.NewTool("trace",
		mcp.WithDescription("Full traceability of a file, decision or bug across sessions"),
		mcp.WithString("target", mcp.Required(), mcp.Description("File path, node ID or search term")),
		mcp.WithNumber("depth", mcp.Description("Trace depth")),
	)
	s.server.AddTool(trace, s.handleTrace)

	confidenceUpdate := mcp.NewTool("confidence_update",
		mcp.WithDescription("Update confidence score of a decision or observation"),
		mcp.WithString("node_id", mcp.Required(), mcp.Description("Node ID")),
		mcp.WithNumber("confidence", mcp.Required(), mcp.Description("New confidence score 0.0-1.0")),
		mcp.WithString("reason", mcp.Required(), mcp.Description("Reason for the update")),
	)
	s.server.AddTool(confidenceUpdate, s.handleConfidenceUpdate)
}

func (s *Server) RunStdio() error {
	return server.ServeStdio(s.server)
}

func (s *Server) handleMemSave(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	title := getString(args, "title")
	nodeType := getString(args, "type")
	topicKey := getStringOrDefault(args, "topic_key", "")
	what := getString(args, "what")
	why := getString(args, "why")
	where := getStringOrDefault(args, "where", "")
	learned := getStringOrDefault(args, "learned", "")
	relatesTo := getStringSlice(args, "relates_to")
	confidence := getFloatOrDefault(args, "confidence", 1.0)

	content := fmt.Sprintf("%s\n\n**Why:** %s", what, why)
	if where != "" {
		content = fmt.Sprintf("%s\n\n**Where:** %s", content, where)
	}
	if learned != "" {
		content = fmt.Sprintf("%s\n\n**Learned:** %s", content, learned)
	}

	node := &graph.Node{
		Type:       nodeType,
		Title:      title,
		TopicKey:   topicKey,
		Content:    s.graph.SanitizePrivateTags(s.graph.SanitizeSecrets(content)),
		Confidence: confidence,
		Scope:      where,
	}

	id, err := s.graph.SaveNode(node)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	edgesCreated := 0
	if relatesTo != nil {
		for _, relatedID := range relatesTo {
			s.graph.SaveEdge(&graph.Edge{
				FromID:   id,
				ToID:     relatedID,
				Relation: "related",
			})
			edgesCreated++
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"id": "%s", "created": true, "edges_created": %d}`, id, edgesCreated)), nil
}

func (s *Server) handleMemFind(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	query := getString(args, "query")
	nodeType := getStringOrDefault(args, "type", "")
	scope := getStringOrDefault(args, "scope", "")
	limit := getIntOrDefault(args, "limit", 10)

	nodes, err := s.graph.Search(query, nodeType, scope, limit, false)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, _ := json.Marshal(nodes)
	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) handleMemContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	module := getStringOrDefault(args, "module", "")
	limit := getIntOrDefault(args, "limit", 20)
	includePredictions := getBoolOrDefault(args, "include_predictions", true)

	memCtx, err := s.graph.GetContext(module, limit, includePredictions)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, _ := json.Marshal(memCtx)
	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) handleMemTimeline(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	nodeID := getString(args, "node_id")
	depth := getIntOrDefault(args, "depth", 2)

	nodes, edges, err := s.graph.GetTimeline(nodeID, depth)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	}
	data, _ := json.Marshal(result)

	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleMemSessionStart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	goal := getString(args, "goal")
	module := getStringOrDefault(args, "module", "")

	sessionID, sessionCtx, err := s.graph.StartSession(goal, module)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"session_id":   sessionID,
		"context":      sessionCtx,
		"predictions":  sessionCtx.Predictions,
		"drift_alerts": sessionCtx.DriftAlerts,
	}
	data, _ := json.Marshal(result)

	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleMemSessionEnd(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	goal := getString(args, "goal")
	accomplished := getString(args, "accomplished")
	discoveries := getStringOrDefault(args, "discoveries", "")
	filesModified := getStringSlice(args, "files_modified")
	openQuestions := getStringOrDefault(args, "open_questions", "")

	var sessionID string
	s.graph.DB().QueryRow("SELECT id FROM sessions WHERE status = 'active' ORDER BY started_at DESC LIMIT 1").Scan(&sessionID)

	dna, err := s.graph.EndSession(sessionID, goal, discoveries, accomplished, openQuestions, filesModified)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.graph.CalculateDrift(sessionID)

	result, _ := json.Marshal(dna)
	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) handleMemDNA(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	sessionID := getStringOrDefault(args, "session_id", "")
	limit := getIntOrDefault(args, "limit", 5)

	var result interface{}
	if sessionID != "" {
		dna, err := s.graph.GetSessionDNA(sessionID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		result = dna
	} else {
		sessions, err := s.graph.ListSessions(limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		result = sessions
	}

	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handlePkgCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	name := getString(args, "name")
	ecosystem := getString(args, "ecosystem")
	version := getStringOrDefault(args, "version", "")

	result, err := s.graph.CheckPackage(name, ecosystem, version)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handlePkgLicense(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	name := getString(args, "name")
	ecosystem := getString(args, "ecosystem")

	result, err := s.graph.CheckPackage(name, ecosystem, "")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"license": "%s", "compatible": true}`, result.License)), nil
}

func (s *Server) handlePkgAudit(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	manifestPath := getStringOrDefault(args, "manifest_path", "")

	issues, err := s.graph.AuditDependencies(manifestPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.Marshal(issues)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleArchSave(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	decision := getString(args, "decision")
	rationale := getString(args, "rationale")
	scope := getStringOrDefault(args, "scope", "global")
	weight := getStringOrDefault(args, "weight", "soft")
	supersedes := getStringOrDefault(args, "supersedes", "")

	id, err := s.graph.SaveDecision(decision, rationale, weight, scope)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if supersedes != "" {
		s.graph.SaveEdge(&graph.Edge{
			FromID:   id,
			ToID:     supersedes,
			Relation: "supersedes",
		})
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"id": "%s", "saved": true}`, id)), nil
}

func (s *Server) handleArchVerify(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	proposedAction := getString(args, "proposed_action")
	context := getStringOrDefault(args, "context", "")

	status, message, err := s.graph.VerifyAction(proposedAction, context)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"status": "%s", "message": "%s"}`, status, message)), nil
}

func (s *Server) handleArchDrift(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	path := getStringOrDefault(args, "path", "")

	scores, err := s.graph.GetDriftScores(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.Marshal(scores)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handlePolicyCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	action := getString(args, "action")
	context := getStringOrDefault(args, "context", "")

	status, message, _ := s.graph.VerifyAction(action, context)

	return mcp.NewToolResultText(fmt.Sprintf(`{"status": "%s", "message": "%s"}`, status, message)), nil
}

func (s *Server) handlePredict(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	module := getStringOrDefault(args, "module", "")

	predictions, err := s.graph.GetPredictions(module)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.Marshal(predictions)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleAuditLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	toolCalled := getString(args, "tool_called")
	actionType := getString(args, "action_type")
	target := getStringOrDefault(args, "target", "")
	riskLevel := getStringOrDefault(args, "risk_level", "low")
	result := getStringOrDefault(args, "result", "success")

	var sessionID string
	s.graph.DB().QueryRow("SELECT id FROM sessions WHERE status = 'active' ORDER BY started_at DESC LIMIT 1").Scan(&sessionID)

	err := s.graph.AddAuditEntry(sessionID, toolCalled, actionType, target, riskLevel, result)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(`{"logged": true}`), nil
}

func (s *Server) handleSessionAudit(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	sessionID := getStringOrDefault(args, "session_id", "")
	riskLevel := getStringOrDefault(args, "risk_level", "")

	if sessionID == "" {
		s.graph.DB().QueryRow("SELECT id FROM sessions WHERE status = 'active' ORDER BY started_at DESC LIMIT 1").Scan(&sessionID)
	}

	logs, err := s.graph.GetSessionAudit(sessionID, riskLevel)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.Marshal(logs)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleInjectGuard(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	content := getString(args, "content")

	suspicious := false
	patterns := []string{
		"ignore previous instructions",
		"disregard your instructions",
		"forget all previous rules",
		"you are now",
		"pretend you are",
		"system prompt",
		"reveal your",
		"new instructions",
	}

	contentLower := strings.ToLower(content)
	for _, pattern := range patterns {
		if strings.Contains(contentLower, pattern) {
			suspicious = true
			break
		}
	}

	var sessionID string
	s.graph.DB().QueryRow("SELECT id FROM sessions WHERE status = 'active' ORDER BY started_at DESC LIMIT 1").Scan(&sessionID)

	if suspicious {
		s.graph.AddAuditEntry(sessionID, "inject_guard", "validate", "", "high", "warning")
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"suspicious": %v}`, suspicious)), nil
}

func (s *Server) handleFenrirStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stats, err := s.graph.GetStats()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.Marshal(stats)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleInsights(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	insightsType := getStringOrDefault(args, "type", "all")

	insights, err := s.graph.GetInsights()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if insightsType != "all" {
		var filtered []graph.Insight
		for _, i := range insights {
			if i.Type == insightsType {
				filtered = append(filtered, i)
			}
		}
		insights = filtered
	}

	data, _ := json.Marshal(insights)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleTrace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	target := getString(args, "target")
	depth := getIntOrDefault(args, "depth", 3)

	trace, err := s.graph.Trace(target, depth)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.Marshal(trace)
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleConfidenceUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	nodeID := getString(args, "node_id")
	confidence := getFloat(args, "confidence")
	reason := getString(args, "reason")

	err := s.graph.UpdateConfidence(nodeID, confidence, reason)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"node_id": "%s", "confidence": %.2f, "updated": true}`, nodeID, confidence)), nil
}

func getString(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getStringOrDefault(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getStringSlice(args map[string]interface{}, key string) []string {
	if v, ok := args[key]; ok {
		if s, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(s))
			for _, item := range s {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return nil
}

func getFloat(args map[string]interface{}, key string) float64 {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func getFloatOrDefault(args map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return defaultVal
}

func getIntOrDefault(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return defaultVal
}

func getBoolOrDefault(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}
