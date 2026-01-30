// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
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
	homeDir := config.GetHomeDir()
	configDir := config.GetDefaultConfigDirectory()

	// Load configuration
	cfg := config.LoadWithDefaults(configDir)

	// Create DotfileManager and reconcile directly
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)
	statuses, err := dm.Reconcile()
	if err != nil {
		return err
	}

	// Separate by state for output.Result
	var managedItems, missingItems []output.Item
	for _, s := range statuses {
		item := output.Item{
			Name: "." + s.Name,
			Path: s.Target,
			Metadata: map[string]interface{}{
				"source":      s.Source,
				"destination": s.Target,
			},
		}

		switch s.State {
		case dotfiles.SyncStateManaged:
			item.State = output.StateManaged
			managedItems = append(managedItems, item)
		case dotfiles.SyncStateMissing:
			item.State = output.StateMissing
			missingItems = append(missingItems, item)
		case dotfiles.SyncStateDrifted:
			item.State = output.StateDegraded
			managedItems = append(managedItems, item)
		}
	}

	// Convert to output format
	outputResult := output.Result{
		Domain:  "dotfile",
		Managed: managedItems,
		Missing: missingItems,
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
