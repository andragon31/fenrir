package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

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
