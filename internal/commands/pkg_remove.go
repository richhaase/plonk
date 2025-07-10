// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"plonk/internal/errors"

	"github.com/spf13/cobra"
)

var (
	uninstall bool
)

var pkgRemoveCmd = &cobra.Command{
	Use:   "remove <package>",
	Short: "Remove a package from plonk configuration",
	Long: `Remove a package from your plonk.yaml configuration.

By default, this only removes the package from your configuration file,
leaving the actual package installed on your system.

Use the --uninstall flag to also uninstall the package from your system.

Examples:
  plonk pkg remove htop                 # Remove from config only
  plonk pkg remove htop --uninstall     # Remove from config and uninstall
  plonk pkg remove htop --dry-run       # Preview what would be removed`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgRemove,
}

func init() {
	pkgCmd.AddCommand(pkgRemoveCmd)
	pkgRemoveCmd.Flags().BoolVar(&uninstall, "uninstall", false, "Also uninstall the package from the system")
	pkgRemoveCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")
}

func runPkgRemove(cmd *cobra.Command, args []string) error {
	return errors.NewError(errors.ErrUnsupported, errors.DomainCommands, "pkg-remove", "Package commands are being refactored to use lock file - temporarily disabled")
}
