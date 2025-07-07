// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"plonk/internal/managers"

	"github.com/spf13/cobra"
)

var pkgStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show package status summary across all managers",
	Long: `Display a summary of package status across Homebrew, ASDF, and NPM managers.

Shows counts of:
- Managed packages (in config and correctly installed)
- Missing packages (in config but not installed or wrong version)  
- Untracked packages (installed but not in config)

For detailed package lists, use:
  plonk pkg list managed
  plonk pkg list missing
  plonk pkg list untracked`,
	RunE: runPkgStatus,
	Args:  cobra.NoArgs,
}

func init() {
	pkgCmd.AddCommand(pkgStatusCmd)
}

func runPkgStatus(cmd *cobra.Command, args []string) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get config directory
	configDir, err := managers.DefaultConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Initialize reconciliation components
	loader := managers.NewPlonkConfigLoader(configDir)
	checkers := map[string]managers.VersionChecker{
		"homebrew": &managers.HomebrewVersionChecker{},
		"asdf":     &managers.AsdfVersionChecker{},
		"npm":      &managers.NpmVersionChecker{},
	}
	managerMap := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"asdf":     managers.NewAsdfManager(),
		"npm":      managers.NewNpmManager(),
	}

	reconciler := managers.NewStateReconciler(loader, managerMap, checkers)

	// Get overall status
	result, err := reconciler.ReconcileAll()
	if err != nil {
		return fmt.Errorf("failed to reconcile package state: %w", err)
	}

	// Prepare output structure
	outputData := PackageStatusOutput{
		Summary: StatusSummary{
			Managed:   len(result.Managed),
			Missing:   len(result.Missing),
			Untracked: len(result.Untracked),
		},
	}

	// Get breakdown by manager
	for _, managerName := range []string{"homebrew", "asdf", "npm"} {
		managerResult, err := reconciler.ReconcileManager(managerName)
		if err != nil {
			continue
		}

		// Only include managers that have activity
		totalPackages := len(managerResult.Managed) + len(managerResult.Missing)
		if totalPackages == 0 {
			continue
		}

		displayName := managerName
		switch managerName {
		case "homebrew":
			displayName = "Homebrew"
		case "asdf":
			displayName = "ASDF"
		case "npm":
			displayName = "NPM"
		}

		outputData.Details = append(outputData.Details, ManagerStatus{
			Name:    displayName,
			Managed: len(managerResult.Managed),
			Missing: len(managerResult.Missing),
		})
	}

	return RenderOutput(outputData, format)
}