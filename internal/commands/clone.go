// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/richhaase/plonk/internal/clone"
	"github.com/spf13/cobra"
)

var cloneDryRun bool

var cloneCmd = &cobra.Command{
	GroupID: "sync",
	Use:     "clone <git-repo>",
	Short:   "Clone dotfiles repository and set up plonk",
	Long: `Clone a dotfiles repository into your plonk directory and run 'plonk apply'.

Install required package managers beforehand. Clone reports unavailable managers
and applies what it can. An existing plonk directory is not overwritten.

Git repository formats supported:
- GitHub shorthand: user/repo (defaults to HTTPS)
- HTTPS URL: https://github.com/user/repo.git
- SSH URL: git@github.com:user/repo.git
- Git protocol: git://github.com/user/repo.git

Examples:
  plonk clone user/dotfiles              # Clone and apply
  plonk clone richhaase/dotfiles         # Clone specific user's dotfiles`,
	Args:         cobra.ExactArgs(1),
	RunE:         runClone,
	SilenceUsage: true,
}

func init() {
	cloneCmd.Flags().BoolVarP(&cloneDryRun, "dry-run", "n", false, "Show what would be cloned without making changes")

	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	gitRepo := args[0]

	cloneConfig := clone.Config{
		DryRun: cloneDryRun,
	}

	return clone.CloneAndSetup(ctx, gitRepo, cloneConfig)
}
