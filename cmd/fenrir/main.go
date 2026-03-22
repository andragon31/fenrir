package main

import (
	"embed"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed plugins/opencode/*
var openCodeFS embed.FS

var (
	version = "0.5.0"
)

const fenrirInstructions = `# Fenrir Protocol

You have access to Fenrir, an AI Governance & Memory Layer for project memory.

## Fenrir Persistent Memory — Protocol (MANDATORY)

### WHEN TO SAVE (MANDATORY after any of these):
1. Call **mem_save** IMMEDIATELY after:
- Bug fix completed
- Architecture or design decision made
- Non-obvious discovery or pattern established
- Configuration change or environment setup
- User preference or constraint learned

2. Use **topic_key** (e.g. "auth-logic", "db-schema", "api-design") to update evolving topics instead of creating new scattered observations.

### WHEN TO SEARCH:
- When asked about past work ("remember", "what did we do", "recordar", "qué hicimos")
- Before asking the user — check memory first with **mem_find** or **mem_context**.

### SESSION WORKFLOW:
1. **Start**: Call **mem_session_start** with your goal.
2. **Work**: Use **pkg_check** before adding dependencies.
3. **Capture**: Use **mem_save** to record progress.
4. **End**: Call **mem_session_end** with:
   - Goals achieved
   - Technical findings/discoveries
   - Decisons made
   - Next steps (what's pending)

## Fenrir Rules:
1. NEVER skip **mem_session_start**.
2. ALWAYS use **topic_key** for evolving topics.
3. If unsure about a package, use **pkg_license** and **pkg_audit**.
4. If you see a "Compaction" or "Context Reset" message, recover context with **mem_context**.
`

const fenrirProtocolMarkdown = `## Fenrir Protocol
You have access to Fenrir, an AI Governance & Memory Layer.

### MEMORY TOOLS:
#### Session Start (MANDATORY at session start)
Call: mem_session_start(goal="your goal", module="current module")

#### Before installing packages (MANDATORY)
Call: pkg_check(name="package name", version="optional version")

#### After making changes (MANDATORY)
Call: mem_save(title="...", type="...", what="...", why="...", where="...", learned="...")

#### Search memory
Call: mem_find(query="search terms")

#### Session End (MANDATORY before ending)
Call: mem_session_end()
`

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

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Fenrir v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		},
	}
	rootCmd.AddCommand(versionCmd)

	cobra.CheckErr(rootCmd.Execute())
}

func initConfig() {
	viper.SetDefault("data_dir", getDefaultDataDir())
	viper.SetDefault("log_level", "info")
}
