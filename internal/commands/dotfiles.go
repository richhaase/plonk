// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/dotfiles"
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
	homeDir := config.GetHomeDir()
	configDir := config.GetDefaultConfigDirectory()

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
		Managed:   convertDotfileItemsToOutput(dotfileResult.Managed),
		Missing:   convertDotfileItemsToOutput(dotfileResult.Missing),
		Untracked: convertDotfileItemsToOutput(dotfileResult.Untracked),
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

// convertDotfileItemsToOutput converts dotfiles.DotfileItem to output.Item
// This is used by the dotfiles command to convert domain-specific types to output types
func convertDotfileItemsToOutput(items []dotfiles.DotfileItem) []output.Item {
	converted := make([]output.Item, len(items))
	for i, item := range items {
		// Convert state
		var state output.ItemState
		switch item.State {
		case dotfiles.StateManaged:
			state = output.ItemState("managed")
		case dotfiles.StateMissing:
			state = output.ItemState("missing")
		case dotfiles.StateUntracked:
			state = output.ItemState("untracked")
		case dotfiles.StateDegraded:
			state = output.ItemState("degraded")
		}

		// Build metadata
		metadata := make(map[string]interface{})
		if item.Metadata != nil {
			for k, v := range item.Metadata {
				metadata[k] = v
			}
		}
		metadata["source"] = item.Source
		metadata["destination"] = item.Destination
		metadata["isTemplate"] = item.IsTemplate
		metadata["isDirectory"] = item.IsDirectory

		converted[i] = output.Item{
			Name:     item.Name,
			Path:     item.Destination,
			State:    state,
			Metadata: metadata,
		}
	}
	return converted
}
