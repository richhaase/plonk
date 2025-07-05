package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "plonk",
	Short: "A shell environment lifecycle manager",
	Long: `plonk is a CLI tool for managing shell environments across multiple machines.
It helps you manage package installations, configurations, and environment switching
across different package managers like Homebrew, ASDF, NPM, Pip, and Cargo.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(statusCmd)
}