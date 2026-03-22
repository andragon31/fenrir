package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

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
		os.WriteFile("FENRIR.md", []byte(fenrirInstructions), 0644)
		
		// Create default policies.json
		defaultPolicies := `{
  "rules": [
    {
      "id": "SEC-001",
      "name": "No Secrets",
      "pattern": "sk-[A-Za-z0-9-_]{20,}|AKIA[A-Za-z0-9]{16}",
      "severity": "critical",
      "message": "Potential secret leakage detected."
    },
    {
      "id": "ARCH-001",
      "name": "Module Privacy",
      "pattern": "import .*? from '.*?/internal/.*?/priv_.*?'",
      "severity": "hard",
      "message": "Do not import from private sub-modules."
    }
  ]
}
`
		if _, err := os.Stat(".fenrir/policies.json"); os.IsNotExist(err) {
			os.WriteFile(".fenrir/policies.json", []byte(defaultPolicies), 0644)
		}

		log.Info("Fenrir initialized", "project", projectName)
	},
}

func init() {
	initCmd.Flags().String("project", "", "Project name")
}
