// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/spf13/cobra"
)

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Package management commands",
	Long: `Manage packages across Homebrew, ASDF, and NPM package managers.
	
View package status, install missing packages, and import existing packages
into your plonk configuration.`,
}

func init() {
	rootCmd.AddCommand(pkgCmd)
}