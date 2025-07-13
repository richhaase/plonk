// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <files...>",
	Short: "Add dotfiles to plonk management",
	Long: `Add dotfiles to plonk configuration and import them.

This command adds dotfiles to your plonk configuration directory and manages them.
It will copy the dotfiles from their current locations to your plonk dotfiles
directory and preserve the original files in case you need to revert.

For directories, plonk will recursively add all files individually, respecting
ignore patterns configured in your plonk.yaml.

Path Resolution:
- Absolute paths: /home/user/.vimrc
- Tilde paths: ~/.vimrc
- Relative paths: First tries current directory, then home directory

Examples:
  plonk add ~/.zshrc                    # Add single file
  plonk add ~/.zshrc ~/.vimrc           # Add multiple files
  plonk add .zshrc .vimrc               # Finds files in home directory
  plonk add ~/.config/nvim/ ~/.tmux.conf # Add directory and file
  plonk add --dry-run ~/.zshrc ~/.vimrc # Preview what would be added`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")
	addCmd.Flags().BoolP("force", "f", false, "Force addition even if already managed")

	// Add file path completion
	addCmd.ValidArgsFunction = completeDotfilePaths
}

func runAdd(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Handle dotfiles using the existing dotfile addition logic
	return addDotfiles(cmd, args, dryRun)
}
