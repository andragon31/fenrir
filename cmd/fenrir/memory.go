package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

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
			indent := ""
			for i := 0; i < t.Depth; i++ {
				indent += "  "
			}
			fmt.Printf("%s%s (%s)\n", indent, t.Action, t.Timestamp)
		}
	},
}

func init() {
	contextCmd.Flags().String("module", "", "Module path")
	driftCmd.Flags().String("module", "", "Module path")
}
