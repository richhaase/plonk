// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [manager]",
	Short: "List installed packages (alias for 'pkg list')",
	Long: `List installed packages from one or all package managers.
	
This is an alias for 'plonk pkg list'.

Examples:
  plonk ls          # List packages from all managers
  plonk ls brew     # List only Homebrew packages
  plonk ls asdf     # List only ASDF tools
  plonk ls npm      # List only NPM packages`,
	RunE: runPkgList,
	Args: cobra.MaximumNArgs(1),
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
