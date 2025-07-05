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
- NPM for packages not available via Homebrew`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(statusCmd)
}