package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Package validation",
}

var pkgCheckCmd = &cobra.Command{
	Use:   "check <name>",
	Short: "Check package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eco, _ := cmd.Flags().GetString("eco")
		version, _ := cmd.Flags().GetString("version")

		g, cleanup := getGraph()
		defer cleanup()

		result, err := g.CheckPackage(args[0], eco, version)
		if err != nil {
			log.Fatal("Check failed", "error", err)
		}

		fmt.Printf("Package: %s\n", args[0])
		fmt.Printf("Exists: %v | Trusted: %v | CVEs: %d\n", result.Exists, result.Trusted, result.CVECount)
		if result.Warning != "" {
			fmt.Printf("Warning: %s\n", result.Warning)
		}
	},
}

var pkgLicenseCmd = &cobra.Command{
	Use:   "license <name>",
	Short: "Check package license",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eco, _ := cmd.Flags().GetString("eco")

		g, cleanup := getGraph()
		defer cleanup()

		result, err := g.CheckPackage(args[0], eco, "")
		if err != nil {
			log.Fatal("Check failed", "error", err)
		}

		fmt.Printf("Package: %s\n", args[0])
		fmt.Printf("License: %s\n", result.License)
	},
}

func init() {
	pkgCheckCmd.Flags().String("eco", "npm", "Ecosystem (npm|pypi|go|cargo)")
	pkgCheckCmd.Flags().String("version", "", "Package version")
	pkgLicenseCmd.Flags().String("eco", "npm", "Ecosystem (npm|pypi|go|cargo)")
	pkgCmd.AddCommand(pkgCheckCmd)
	pkgCmd.AddCommand(pkgLicenseCmd)
}
