package main

import (
	"context"
	"fmt"
	"os"

	"github.com/andragon31/fenrir/internal/server"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

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
