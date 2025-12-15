// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
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
  plonk dotfiles           # Show all managed dotfiles
  plonk dotfiles -o json   # Show as JSON
  plonk dotfiles -o yaml   # Show as YAML`,
	RunE:         runDotfiles,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(dotfilesCmd)
}

func runDotfiles(cmd *cobra.Command, args []string) error {
	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load configuration
	cfg := config.LoadWithDefaults(configDir)
	ctx := context.Background()

	// Reconcile dotfiles
	dotfileResult, err := dotfiles.ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return err
	}

	// Convert to output format
	outputResult := output.Result{
		Domain:    dotfileResult.Domain,
		Managed:   convertItemsToOutput(dotfileResult.Managed),
		Missing:   convertItemsToOutput(dotfileResult.Missing),
		Untracked: convertItemsToOutput(dotfileResult.Untracked),
	}

	// Prepare output
	outputData := output.DotfilesStatusOutput{
		Result:    outputResult,
		ConfigDir: configDir,
	}

	// Create formatter and render
	formatter := output.NewDotfilesStatusFormatter(outputData)
	output.RenderOutput(formatter)
	return nil
}
