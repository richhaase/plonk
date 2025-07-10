// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"plonk/internal/errors"

	"github.com/spf13/cobra"
)

var (
	manager string
)

var pkgAddCmd = &cobra.Command{
	Use:   "add [package]",
	Short: "Add package(s) to plonk configuration and install them",
	Long: `Add one or more packages to your plonk.yaml configuration and install them.

With package name:
  plonk pkg add htop              # Add htop using default manager
  plonk pkg add git --manager homebrew  # Add git specifically to homebrew
  plonk pkg add lodash --manager npm     # Add lodash to npm global packages

Without arguments:
  plonk pkg add                   # Add all untracked packages
  plonk pkg add --dry-run         # Preview what would be added`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPkgAdd,
}

func init() {
	pkgCmd.AddCommand(pkgAddCmd)
	pkgAddCmd.Flags().StringVar(&manager, "manager", "", "Package manager to use (homebrew|npm)")
	pkgAddCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")
}

func runPkgAdd(cmd *cobra.Command, args []string) error {
	return errors.NewError(errors.ErrUnsupported, errors.DomainCommands, "pkg-add", "Package commands are being refactored to use lock file - temporarily disabled")
}
