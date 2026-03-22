package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Git sync memories",
	Run: func(cmd *cobra.Command, args []string) {
		g, cleanup := getGraph()
		defer cleanup()

		if err := g.ExportChunks(); err != nil {
			log.Fatal("Export failed", "error", err)
		}
		if err := g.ImportChunks(); err != nil {
			log.Fatal("Import failed", "error", err)
		}

		fmt.Println("Memories synchronized with .fenrir/chunks/")
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
