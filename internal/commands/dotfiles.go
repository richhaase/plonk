// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/spf13/cobra"
)

var dotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "List dotfiles specifically",
	Long: `Show detailed information about dotfiles only.

This command provides dotfile-specific listing with detailed state information.
Use this when you want to focus specifically on dotfiles without packages.

Examples:
  plonk dotfiles                # Show all dotfiles with state
  plonk dotfiles -v             # Show detailed dotfile information
  plonk dotfiles --missing      # Show only missing dotfiles
  plonk dotfiles --managed      # Show only managed dotfiles
  plonk dotfiles --untracked    # Show only untracked dotfiles`,
	RunE: runDotfiles,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(dotfilesCmd)

	// State filter flags
	dotfilesCmd.Flags().Bool("managed", false, "Show managed dotfiles only")
	dotfilesCmd.Flags().Bool("missing", false, "Show missing dotfiles only")
	dotfilesCmd.Flags().Bool("untracked", false, "Show untracked dotfiles only")
	dotfilesCmd.MarkFlagsMutuallyExclusive("managed", "missing", "untracked")

	// Detail flags
	dotfilesCmd.Flags().BoolP("verbose", "v", false, "Show detailed information")
}

func runDotfiles(cmd *cobra.Command, args []string) error {
	// Delegate to the existing dotfile list implementation
	// This reuses the dot list command logic but with a cleaner interface
	return runDotList(cmd, args)
}
