// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/spf13/cobra"
)

var dotCmd = &cobra.Command{
	Use:   "dot",
	Short: "Dotfile management commands",
	Long: `Manage dotfiles and track their state relative to your plonk configuration.

View dotfile status, compare against configuration, and understand what
dotfiles are managed, missing, untracked, or modified.`,
}

func init() {
	rootCmd.AddCommand(dotCmd)
}