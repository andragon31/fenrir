package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

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
