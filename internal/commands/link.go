// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <files...>",
	Short: "Explicitly link dotfiles",
	Long: `Explicitly add dotfiles to plonk configuration and import them.

This command forces all arguments to be treated as dotfiles, regardless
of their format. Use this when the automatic detection in 'plonk add'
doesn't work correctly, or when you want to be explicit.

This command will:
- Copy the dotfiles from their current locations to your plonk dotfiles directory
- Add them to your plonk.yaml configuration
- Preserve the original files in case you need to revert

For directories, plonk will recursively add all files individually, respecting
ignore patterns configured in your plonk.yaml.

Path Resolution:
- Absolute paths: /home/user/.vimrc
- Tilde paths: ~/.vimrc
- Relative paths: First tries current directory, then home directory

Examples:
  plonk link ~/.zshrc                    # Add single file
  plonk link ~/.zshrc ~/.vimrc           # Add multiple files
  plonk link .zshrc .vimrc               # Finds files in home directory
  plonk link ~/.config/nvim/ ~/.tmux.conf # Add directory and file
  plonk link --dry-run ~/.zshrc ~/.vimrc # Preview what would be added
  plonk link config                      # Force 'config' to be treated as dotfile`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLink,
}

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")
	linkCmd.Flags().BoolP("force", "f", false, "Force addition even if already managed")

	// Add file path completion
	linkCmd.ValidArgsFunction = completeDotfilePaths
}

func runLink(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Handle dotfiles using the existing dotfile addition logic
	return addDotfiles(cmd, args, dryRun)
}
