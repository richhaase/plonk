// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var dotfilesCmd = &cobra.Command{
	Use:     "dotfiles",
	Aliases: []string{"d"},
	Short:   "Display dotfile status",
	Long: `Display the status of all plonk-managed dotfiles.

Shows:
- All managed dotfiles
- Missing dotfiles that need to be deployed
- Drifted dotfiles (modified after deployment)

Examples:
  plonk dotfiles    # Show all managed dotfiles
  plonk d           # Short alias`,
	RunE:         runDotfiles,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(dotfilesCmd)
}

func runDotfiles(cmd *cobra.Command, args []string) error {
	// Get directories
	homeDir, err := config.GetHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configDir := config.GetDefaultConfigDirectory()

	// Load configuration
	cfg := config.LoadWithDefaults(configDir)

	// Create DotfileManager and reconcile directly
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)
	statuses, err := dm.Reconcile()
	if err != nil {
		return err
	}

	// Separate by state and convert to output format
	managed, missing, errors := convertDotfileStatusToOutput(statuses)

	outputResult := output.Result{
		Domain:  "dotfile",
		Managed: managed,
		Missing: missing,
		Errors:  errors,
	}

	// Prepare output
	outputData := output.DotfilesStatusOutput{
		Result:    outputResult,
		ConfigDir: configDir,
		HomeDir:   homeDir,
	}

	// Create formatter and render
	formatter := output.NewDotfilesStatusFormatter(outputData)
	output.RenderOutput(formatter)
	return nil
}
