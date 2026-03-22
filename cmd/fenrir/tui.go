package main

import (
	"github.com/andragon31/fenrir/internal/tui"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		if err := tui.Run(g); err != nil {
			log.Fatal("TUI error", "error", err)
		}
	},
}
