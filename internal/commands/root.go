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

Use the 'repo' command for complete setup from a repository.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// IsDryRun checks if dry-run mode is enabled (either global or local flag)
func IsDryRun(cmd *cobra.Command) bool {
	// Handle nil command (e.g., in tests)
	if cmd == nil {
		return false
	}

	// Check local flag first
	if localFlag := cmd.Flags().Lookup("dry-run"); localFlag != nil {
		if dryRun, err := cmd.Flags().GetBool("dry-run"); err == nil && dryRun {
			return true
		}
	}

	// Check global flag
	if globalFlag := cmd.PersistentFlags().Lookup("dry-run"); globalFlag != nil {
		if dryRun, err := cmd.PersistentFlags().GetBool("dry-run"); err == nil && dryRun {
			return true
		}
	}

	// Check parent command's persistent flags
	if cmd.Parent() != nil {
		if dryRun, err := cmd.Parent().PersistentFlags().GetBool("dry-run"); err == nil && dryRun {
			return true
		}
	}

	return false
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without making any changes")

	// Add subcommands here
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(pkgCmd)
	rootCmd.AddCommand(importCmd)
}
