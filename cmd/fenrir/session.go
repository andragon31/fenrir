package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

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
