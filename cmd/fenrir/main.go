package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/TU_ORG/fenrir/internal/graph"
	"github.com/TU_ORG/fenrir/internal/mcp"
	"github.com/TU_ORG/fenrir/internal/server"
	"github.com/TU_ORG/fenrir/internal/tui"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "0.1.0"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "fenrir",
		Short: "Fenrir - AI Governance & Memory Layer",
		Long: `Fenrir gives your AI coding agent a memory.

Agent-agnostic. Single binary. Zero dependencies.
Works with Claude Code, OpenCode, Cursor, Windsurf, Gemini CLI, Antigravity, and any MCP client.`,
	}

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("data-dir", "d", "", "Data directory (default: ~/.fenrir)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level: debug|info|warn|error")

	viper.BindPFlag("data_dir", rootCmd.PersistentFlags().Lookup("data-dir"))
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(driftCmd)
	rootCmd.AddCommand(insightsCmd)
	rootCmd.AddCommand(traceCmd)
	rootCmd.AddCommand(sessionCmd)
	rootCmd.AddCommand(archCmd)
	rootCmd.AddCommand(pkgCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(versionCmd)

	cobra.CheckErr(rootCmd.Execute())
}

func initConfig() {
	viper.SetDefault("data_dir", getDefaultDataDir())
	viper.SetDefault("log_level", "info")
}

func getDefaultDataDir() string {
	home, _ := os.UserHomeDir()
	return home + "/.fenrir"
}

func getGraph() (*graph.Graph, func()) {
	dataDir := viper.GetString("data_dir")
	if dataDir == "" {
		dataDir = getDefaultDataDir()
	}

	g, err := graph.New(dataDir)
	if err != nil {
		log.Fatal("Failed to initialize database", "error", err)
	}

	return g, func() { g.Close() }
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Fenrir in the current project",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		if err := g.Init(); err != nil {
			log.Fatal("Failed to initialize schema", "error", err)
		}

		projectName, _ := cmd.Flags().GetString("project")
		if projectName == "" {
			cwd, _ := os.Getwd()
			projectName = cwd
		}

		os.MkdirAll(".fenrir", 0755)
		writeProjectConfig(projectName)

		log.Info("Fenrir initialized", "project", projectName)
	},
}

func init() {
	initCmd.Flags().String("project", "", "Project name")
}

var setupCmd = &cobra.Command{
	Use:   "setup [agent]",
	Short: "Setup Fenrir for an AI agent",
	Long: `Setup Fenrir integration for various AI coding agents.

Supported agents:
  - opencode     OpenCode
  - claude-code  Claude Code
  - cursor       Cursor
  - windsurf     Windsurf
  - antigravity  Antigravity
  - gemini-cli   Gemini CLI
  - vscode       VS Code (Copilot)
  - codex        Codex
  - generic      Generic MCP client

Examples:
  fenrir setup opencode
  fenrir setup claude-code
  fenrir setup cursor`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agent := args[0]
		installer := getInstaller(agent)
		if err := installer.Install(); err != nil {
			log.Fatal("Setup failed", "agent", agent, "error", err)
		}
		log.Info("Setup complete", "agent", agent)
		fmt.Println("\nRestart your AI agent to start using Fenrir!")
	},
}

type Installer interface {
	Install() error
	Name() string
}

func getInstaller(agent string) Installer {
	switch agent {
	case "opencode":
		return &OpenCodeInstaller{}
	case "claude-code":
		return &ClaudeCodeInstaller{}
	case "cursor":
		return &CursorInstaller{}
	case "windsurf":
		return &WindsurfInstaller{}
	case "antigravity":
		return &AntigravityInstaller{}
	case "gemini-cli":
		return &GeminiCLIInstaller{}
	case "vscode":
		return &VSCodeInstaller{}
	case "codex":
		return &CodexInstaller{}
	default:
		return &GenericInstaller{}
	}
}

type OpenCodeInstaller struct{}

func (i *OpenCodeInstaller) Name() string { return "OpenCode" }

func (i *OpenCodeInstaller) Install() error {
	fenrirPath, err := os.Executable()
	if err != nil {
		fenrirPath = "fenrir"
	}

	configDir, _ := os.UserConfigDir()
	configPath := configDir + "/opencode/opencode.json"
	os.MkdirAll(configDir+"/opencode", 0755)

	config := fmt.Sprintf(`{
  "mcpServers": {
    "fenrir": {
      "command": "%s",
      "args": ["mcp"]
    }
  }
}`, fenrirPath)

	return os.WriteFile(configPath, []byte(config), 0644)
}

type ClaudeCodeInstaller struct{}

func (i *ClaudeCodeInstaller) Name() string { return "Claude Code" }

func (i *ClaudeCodeInstaller) Install() error {
	fenrirPath, err := os.Executable()
	if err != nil {
		fenrirPath = "fenrir"
	}

	os.MkdirAll(".claude", 0755)

	settings := `{
  "mcpServers": {
    "fenrir": {
      "command": "` + fenrirPath + `",
      "args": ["mcp"]
    }
  }
}`
	if err := os.WriteFile(".claude/settings.json", []byte(settings), 0644); err != nil {
		return err
	}

	fenrirMD := `# Fenrir Protocol

You have access to Fenrir for project memory.

## MANDATORY:
- START: Call mem_session_start with your goal
- BEFORE installing any package: Call pkg_check
- AFTER any bugfix/decision/discovery: Call mem_save
- END: Call mem_session_end

## After compaction:
Immediately call mem_context to recover session state.`

	return os.WriteFile("FENRIR.md", []byte(fenrirMD), 0644)
}

type CursorInstaller struct{}

func (i *CursorInstaller) Name() string { return "Cursor" }

func (i *CursorInstaller) Install() error {
	fenrirPath, err := os.Executable()
	if err != nil {
		fenrirPath = "fenrir"
	}

	os.MkdirAll(".cursor", 0755)

	config := fmt.Sprintf(`{
  "mcpServers": {
    "fenrir": {
      "command": "%s",
      "args": ["mcp"]
    }
  }
}`, fenrirPath)

	return os.WriteFile(".cursor/mcp.json", []byte(config), 0644)
}

type WindsurfInstaller struct{}

func (i *WindsurfInstaller) Name() string { return "Windsurf" }

func (i *WindsurfInstaller) Install() error {
	fenrirPath, err := os.Executable()
	if err != nil {
		fenrirPath = "fenrir"
	}

	os.MkdirAll(".windsurf", 0755)

	config := fmt.Sprintf(`{
  "mcpServers": {
    "fenrir": {
      "command": "%s",
      "args": ["mcp"]
    }
  }
}`, fenrirPath)

	return os.WriteFile(".windsurf/mcp.json", []byte(config), 0644)
}

type AntigravityInstaller struct{}

func (i *AntigravityInstaller) Name() string { return "Antigravity" }

func (i *AntigravityInstaller) Install() error {
	fenrirPath, err := os.Executable()
	if err != nil {
		fenrirPath = "fenrir"
	}

	os.MkdirAll(".antigravity", 0755)

	config := fmt.Sprintf(`{
  "mcpServers": {
    "fenrir": {
      "command": "%s",
      "args": ["mcp"]
    }
  }
}`, fenrirPath)

	return os.WriteFile(".antigravity/mcp.json", []byte(config), 0644)
}

type GeminiCLIInstaller struct{}

func (i *GeminiCLIInstaller) Name() string { return "Gemini CLI" }

func (i *GeminiCLIInstaller) Install() error {
	fenrirPath, err := os.Executable()
	if err != nil {
		fenrirPath = "fenrir"
	}

	os.MkdirAll(".gemini", 0755)

	config := fmt.Sprintf(`{
  "mcpServers": {
    "fenrir": {
      "command": "%s",
      "args": ["mcp"]
    }
  }
}`, fenrirPath)

	if err := os.WriteFile(".gemini/settings.json", []byte(config), 0644); err != nil {
		return err
	}

	fenrirMD := `# Fenrir Protocol

You have access to Fenrir for project memory.

## MANDATORY:
- START: Call mem_session_start with your goal
- END: Call mem_session_end`

	return os.WriteFile("FENRIR.md", []byte(fenrirMD), 0644)
}

type VSCodeInstaller struct{}

func (i *VSCodeInstaller) Name() string { return "VS Code" }

func (i *VSCodeInstaller) Install() error {
	fenrirPath, err := os.Executable()
	if err != nil {
		fenrirPath = "fenrir"
	}

	fmt.Println("VS Code MCP setup:")
	fmt.Println("1. Open VS Code")
	fmt.Println("2. Run: code --add-mcp '{\"name\":\"fenrir\",\"command\":\"" + fenrirPath + "\",\"args\":[\"mcp\"]}'")
	fmt.Println()
	return nil
}

type CodexInstaller struct{}

func (i *CodexInstaller) Name() string { return "Codex" }

func (i *CodexInstaller) Install() error {
	fenrirPath, err := os.Executable()
	if err != nil {
		fenrirPath = "fenrir"
	}

	fmt.Println("Codex MCP setup:")
	fmt.Printf("Add to your Codex config:\n\n")
	fmt.Printf(`{
  "mcpServers": {
    "fenrir": {
      "command": "%s",
      "args": ["mcp"]
    }
  }
}`, fenrirPath)
	fmt.Println()
	return nil
}

type GenericInstaller struct{}

func (i *GenericInstaller) Name() string { return "Generic MCP" }

func (i *GenericInstaller) Install() error {
	fenrirPath, err := os.Executable()
	if err != nil {
		fenrirPath = "fenrir"
	}

	fmt.Println("Generic MCP setup:")
	fmt.Printf("Add this to your MCP config:\n\n")
	fmt.Printf(`{
  "mcpServers": {
    "fenrir": {
      "command": "%s",
      "args": ["mcp"]
    }
  }
}`, fenrirPath)
	fmt.Println()
	return nil
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server (stdio mode)",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		if err := g.Init(); err != nil {
			log.Fatal("Failed to initialize", "error", err)
		}

		logger := log.New(os.Stderr)
		logger.SetLevel(log.InfoLevel)

		srv := mcp.NewServer(g, logger)
		if err := srv.RunStdio(); err != nil {
			log.Fatal("MCP server error", "error", err)
		}
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve [port]",
	Short: "Start HTTP API server",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port := 7438
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &port)
		}

		g, cleanup := getGraph()
		defer cleanup()

		logger := log.New(os.Stderr)
		srv := server.New(g, logger, port)
		if err := srv.Run(context.Background()); err != nil {
			log.Fatal("Server error", "error", err)
		}
	},
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		model := tui.NewModel(g)
		p := tui.NewProgram(model)
		if _, err := p.Run(); err != nil {
			log.Fatal("TUI error", "error", err)
		}
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search memories",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		results, err := g.Search(args[0], "", "", 20, false)
		if err != nil {
			log.Fatal("Search failed", "error", err)
		}

		if len(results) == 0 {
			fmt.Println("No results found")
			return
		}

		for _, r := range results {
			fmt.Printf("[%s] %s\n  %s\n\n", r.Type, r.Title, r.Content)
		}
	},
}

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Get session context",
	Run: func(cmd *cobra.Command, args []string) {
		module, _ := cmd.Flags().GetString("module")

		g, cleanup := getGraph()
		defer cleanup()

		ctx, err := g.GetContext(module, 20, true)
		if err != nil {
			log.Fatal("Context failed", "error", err)
		}

		fmt.Printf("Active Sessions: %d\n", ctx.ActiveSessions)
		if len(ctx.Predictions) > 0 {
			fmt.Println("\n=== Predictions ===")
			for _, p := range ctx.Predictions {
				fmt.Printf("[%s] %s\n", p.Severity, p.Message)
			}
		}
		if len(ctx.Observations) > 0 {
			fmt.Println("\n=== Recent ===")
			for _, o := range ctx.Observations {
				fmt.Printf("[%s] %s\n", o.Type, o.Title)
			}
		}
	},
}

func init() {
	contextCmd.Flags().String("module", "", "Module path")
}

var driftCmd = &cobra.Command{
	Use:   "drift",
	Short: "Show drift scores",
	Run: func(cmd *cobra.Command, args []string) {
		module, _ := cmd.Flags().GetString("module")

		g, cleanup := getGraph()
		defer cleanup()

		scores, err := g.GetDriftScores(module)
		if err != nil {
			log.Fatal("Drift failed", "error", err)
		}

		if len(scores) == 0 {
			fmt.Println("No drift data")
			return
		}

		for _, s := range scores {
			bar := driftBar(s.Score)
			fmt.Printf("%-20s %s %.2f\n", s.Module, bar, s.Score)
		}
	},
}

func init() {
	driftCmd.Flags().String("module", "", "Module path")
}

var insightsCmd = &cobra.Command{
	Use:   "insights",
	Short: "Show detected patterns",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		insights, err := g.GetInsights()
		if err != nil {
			log.Fatal("Insights failed", "error", err)
		}

		if len(insights) == 0 {
			fmt.Println("No insights yet")
			return
		}

		for _, i := range insights {
			fmt.Printf("[%s] %s (x%d)\n", i.Type, i.Title, i.Count)
		}
	},
}

var traceCmd = &cobra.Command{
	Use:   "trace <target>",
	Short: "Trace file/decision across sessions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		trace, err := g.Trace(args[0], 3)
		if err != nil {
			log.Fatal("Trace failed", "error", err)
		}

		for _, t := range trace {
			fmt.Printf("[%s] %s\n", t.SessionID, t.Action)
		}
	},
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		sessions, err := g.ListSessions(20)
		if err != nil {
			log.Fatal("List failed", "error", err)
		}

		for _, s := range sessions {
			fmt.Printf("[%s] %s - %s\n", s.ID[:8], s.Goal, s.Status)
		}
	},
}

var sessionShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show session DNA",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		dna, err := g.GetSessionDNA(args[0])
		if err != nil {
			log.Fatal("Show failed", "error", err)
		}

		fmt.Printf("Goal: %s\n", dna.Goal)
		fmt.Printf("Drift: %.2f | Violations: %d\n", dna.DriftDelta, dna.ArchViolations)
	},
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionShowCmd)
}

var archCmd = &cobra.Command{
	Use:   "arch",
	Short: "Manage architectural decisions",
}

var archListCmd = &cobra.Command{
	Use:   "list",
	Short: "List decisions",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		decisions, err := g.ListDecisions()
		if err != nil {
			log.Fatal("List failed", "error", err)
		}

		for _, d := range decisions {
			fmt.Printf("[%s] %s (%s)\n", d.Weight, d.Title, d.Rationale)
		}
	},
}

var archAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add decision",
	Run: func(cmd *cobra.Command, args []string) {
		decision, _ := cmd.Flags().GetString("decision")
		rationale, _ := cmd.Flags().GetString("rationale")
		weight, _ := cmd.Flags().GetString("weight")

		g, cleanup := getGraph()
		defer cleanup()

		id, err := g.SaveDecision(decision, rationale, weight, "global")
		if err != nil {
			log.Fatal("Save failed", "error", err)
		}

		fmt.Printf("Decision saved: %s\n", id)
	},
}

func init() {
	archAddCmd.Flags().String("decision", "", "Decision text")
	archAddCmd.Flags().String("rationale", "", "Why")
	archAddCmd.Flags().String("weight", "soft", "soft|hard|critical")
	archCmd.AddCommand(archListCmd)
	archCmd.AddCommand(archAddCmd)
}

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Package validation",
}

var pkgCheckCmd = &cobra.Command{
	Use:   "check <name>",
	Short: "Check package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eco, _ := cmd.Flags().GetString("eco")

		g, cleanup := getGraph()
		defer cleanup()

		result, err := g.CheckPackage(args[0], eco, "")
		if err != nil {
			log.Fatal("Check failed", "error", err)
		}

		fmt.Printf("Package: %s\n", args[0])
		fmt.Printf("Exists: %v | Trusted: %v | CVEs: %d\n", result.Exists, result.Trusted, result.CVECount)
		if result.Warning != "" {
			fmt.Printf("Warning: %s\n", result.Warning)
		}
	},
}

func init() {
	pkgCheckCmd.Flags().StringP("eco", "e", "npm", "npm|pypi|cargo|nuget")
	pkgCmd.AddCommand(pkgCheckCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Git sync memories",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		if err := g.ExportChunks(); err != nil {
			log.Fatal("Sync failed", "error", err)
		}

		fmt.Println("Memories exported to .fenrir/chunks/")
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		stats, err := g.GetStats()
		if err != nil {
			log.Fatal("Stats failed", "error", err)
		}

		fmt.Println("=== Fenrir Stats ===")
		fmt.Printf("Nodes: %d | Edges: %d\n", stats.TotalNodes, stats.TotalEdges)
		fmt.Printf("Sessions: %d (%d active)\n", stats.TotalSessions, stats.ActiveSessions)
		fmt.Printf("Decisions: %d | Audit: %d\n", stats.TotalDecisions, stats.AuditEntries)
	},
}

var exportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Export memories",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "fenrir-export.json"
		if len(args) > 0 {
			path = args[0]
		}

		g, cleanup := getGraph()
		defer cleanup()

		if err := g.ExportToJSON(path); err != nil {
			log.Fatal("Export failed", "error", err)
		}

		fmt.Printf("Exported to %s\n", path)
	},
}

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import memories",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		if err := g.ImportFromJSON(args[0]); err != nil {
			log.Fatal("Import failed", "error", err)
		}

		fmt.Println("Imported successfully")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Fenrir v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
	},
}

func writeProjectConfig(name string) error {
	config := fmt.Sprintf(`{
  "project": "%s",
  "validator": { "enabled": true, "ecosystems": ["npm", "pypi"] },
  "enforcer": { "enabled": true }
}`, name)
	return os.WriteFile(".fenrir/config.json", []byte(config), 0644)
}

func driftBar(score float64) string {
	filled := int(score * 20)
	if filled > 20 {
		filled = 20
	}
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := filled; i < 20; i++ {
		bar += "░"
	}
	return bar
}
