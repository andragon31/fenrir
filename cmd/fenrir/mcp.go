package main

import (
	"os"

	"github.com/andragon31/fenrir/internal/mcp"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

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
