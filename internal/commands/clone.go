// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/clone"
	"github.com/spf13/cobra"
)

var (
	cloneYes     bool
	cloneNoApply bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone <git-repo>",
	Short: "Clone dotfiles repository and set up plonk",
	Long: `Clone an existing dotfiles repository and intelligently set up plonk.

This command:
- Clones the repository into your plonk directory
- Reads the plonk.lock file to detect required package managers
- Installs ONLY the package managers needed by your dotfiles
- Runs 'plonk apply' to configure your system (unless --no-apply is used)

The intelligent detection feature means you don't need to manually specify
which package managers to install - plonk will figure it out from your lock file.

Git repository formats supported:
- GitHub shorthand: user/repo (defaults to HTTPS)
- HTTPS URL: https://github.com/user/repo.git
- SSH URL: git@github.com:user/repo.git
- Git protocol: git://github.com/user/repo.git

Examples:
  plonk clone user/dotfiles              # Clone and auto-detect managers
  plonk clone richhaase/dotfiles         # Clone specific user's dotfiles
  plonk clone user/repo --no-apply       # Clone without running apply
  plonk clone --yes user/dotfiles        # Non-interactive mode`,
	Args:         cobra.ExactArgs(1),
	RunE:         runClone,
	SilenceUsage: true,
}

func init() {
	cloneCmd.Flags().BoolVar(&cloneYes, "yes", false, "Non-interactive mode - answer yes to all prompts")
	cloneCmd.Flags().BoolVar(&cloneNoApply, "no-apply", false, "Skip running 'plonk apply' after setup")

	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	gitRepo := args[0]

	cloneConfig := clone.Config{
		Interactive: !cloneYes,
		Verbose:     false, // Could add --verbose flag later
		NoApply:     cloneNoApply,
		// No SkipManagers for clone - it auto-detects
	}

	return clone.CloneAndSetup(ctx, gitRepo, cloneConfig)
}
