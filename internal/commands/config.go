// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage plonk configuration",
	Long: `Manage plonk configuration files and settings.

Commands:
  show      Display current configuration
  validate  Validate configuration syntax
  edit      Edit configuration file
  init      Initialize new configuration`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}