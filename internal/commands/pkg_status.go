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

	// Display summary
	fmt.Println("Package Status")
	fmt.Println("==============")
	fmt.Println()

	if len(result.Managed) > 0 {
		fmt.Printf("✅ %d managed packages\n", len(result.Managed))
	} else {
		fmt.Println("📦 No managed packages")
	}

	if len(result.Missing) > 0 {
		fmt.Printf("❌ %d missing packages\n", len(result.Missing))
	}

	if len(result.Untracked) > 0 {
		fmt.Printf("🔍 %d untracked packages\n", len(result.Untracked))
	}

	// Show breakdown by manager if there are any issues
	if len(result.Missing) > 0 || len(result.Managed) > 0 {
		fmt.Println()
		fmt.Println("Details:")

		for _, managerName := range []string{"homebrew", "asdf", "npm"} {
			managerResult, err := reconciler.ReconcileManager(managerName)
			if err != nil {
				continue
			}

			// Only show managers that have activity
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

			fmt.Printf("  %s: ", displayName)
			parts := []string{}
			if len(managerResult.Managed) > 0 {
				parts = append(parts, fmt.Sprintf("%d managed", len(managerResult.Managed)))
			}
			if len(managerResult.Missing) > 0 {
				parts = append(parts, fmt.Sprintf("%d missing", len(managerResult.Missing)))
			}

			for i, part := range parts {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(part)
			}
			fmt.Println()
		}
	}

	return nil
}