// Package commands implements the CLI interface for Plonk using Cobra.
// It provides commands for managing shell environments including install,
// apply, status, clone, pull, restore, and backup operations.
//
// Each command handles specific aspects of shell environment management:
// - install: Install packages from configuration
// - apply: Deploy configuration files to target locations
// - status: Show package manager availability and drift
// - clone/pull: Git repository operations for configuration sharing
// - restore: Restore files from timestamped backups
// - backup: Create backups of configuration files
package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "plonk",
	Short: "A shell environment lifecycle manager",
	Long: `plonk is a CLI tool for managing shell environments across multiple machines.
It helps you manage package installations and environment switching using:
- Homebrew for primary package installation
- ASDF for programming language tools and versions
- NPM for packages not available via Homebrew

Convenience usage:
  plonk <repository>    # Complete setup from repository (clone + install + apply)`,
	RunE: rootCmdRun,
	Args: cobra.MaximumNArgs(1),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func rootCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// No arguments, show help
		return cmd.Help()
	}

	// Single argument should be a repository URL for convenience setup
	return runRepo(args)
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(pkgCmd)
	rootCmd.AddCommand(importCmd)
}
