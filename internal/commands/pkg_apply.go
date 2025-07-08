// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"plonk/internal/managers"

	"github.com/spf13/cobra"
)

var (
	dryRun bool
	prune  bool
)

var pkgApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply package configuration by installing missing packages",
	Long: `Install all packages that are defined in plonk.yaml but not currently installed.

This command will:
1. Read your plonk configuration
2. Check which packages are missing from each manager
3. Install missing packages using the appropriate manager
4. Report the results

Use --dry-run to see what would be installed without making changes.

Examples:
  plonk pkg apply           # Install all missing packages
  plonk pkg apply --dry-run # Show what would be installed`,
	RunE: runPkgApply,
}

func init() {
	pkgCmd.AddCommand(pkgApplyCmd)
	pkgApplyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be installed without making changes")
}

func runPkgApply(cmd *cobra.Command, args []string) error {
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

	// Initialize package managers
	packageManagers := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"npm":      managers.NewNpmManager(),
	}

	// Initialize config loader and reconciler
	loader := managers.NewPlonkConfigLoader(configDir)
	reconciler := managers.NewStateReconciler(loader, packageManagers)

	// Prepare output structure
	var outputData ApplyOutput
	outputData.DryRun = dryRun

	// Process each manager
	for managerName, manager := range packageManagers {
		if !manager.IsAvailable() {
			if format == OutputTable {
				fmt.Printf("# %s: Not available, skipping\n", managerName)
			}
			continue
		}

		// Get missing packages for this manager
		result, err := reconciler.ReconcileManager(managerName)
		if err != nil {
			if format == OutputTable {
				fmt.Printf("# %s: Error reconciling state: %v\n", managerName, err)
			}
			continue
		}

		// Process missing packages
		var managerResult ManagerApplyResult
		managerResult.Name = managerName
		managerResult.MissingCount = len(result.Missing)

		for _, pkg := range result.Missing {
			packageResult := PackageApplyResult{
				Name:   pkg.Name,
				Status: "pending",
			}

			if dryRun {
				packageResult.Status = "would-install"
				if format == OutputTable {
					fmt.Printf("Would install: %s (%s)\n", pkg.Name, managerName)
				}
			} else {
				// Actually install the package
				err := manager.Install(pkg.Name)
				if err != nil {
					packageResult.Status = "failed"
					packageResult.Error = err.Error()
					if format == OutputTable {
						fmt.Printf("Failed to install %s: %v\n", pkg.Name, err)
					}
				} else {
					packageResult.Status = "installed"
					if format == OutputTable {
						fmt.Printf("Installed: %s (%s)\n", pkg.Name, managerName)
					}
				}
			}

			managerResult.Packages = append(managerResult.Packages, packageResult)
		}

		if len(result.Missing) == 0 {
			if format == OutputTable {
				fmt.Printf("# %s: All packages up to date\n", managerName)
			}
		}

		outputData.Managers = append(outputData.Managers, managerResult)
	}

	// Calculate summary
	for _, mgr := range outputData.Managers {
		outputData.TotalMissing += mgr.MissingCount
		for _, pkg := range mgr.Packages {
			switch pkg.Status {
			case "installed":
				outputData.TotalInstalled++
			case "failed":
				outputData.TotalFailed++
			case "would-install":
				outputData.TotalWouldInstall++
			}
		}
	}

	// Output summary for table format
	if format == OutputTable {
		fmt.Println()
		if dryRun {
			fmt.Printf("Summary: %d packages would be installed\n", outputData.TotalWouldInstall)
		} else {
			fmt.Printf("Summary: %d installed, %d failed\n", outputData.TotalInstalled, outputData.TotalFailed)
		}
	}

	return RenderOutput(outputData, format)
}

// ApplyOutput represents the output structure for pkg apply
type ApplyOutput struct {
	DryRun            bool                  `json:"dry_run" yaml:"dry_run"`
	TotalMissing      int                   `json:"total_missing" yaml:"total_missing"`
	TotalInstalled    int                   `json:"total_installed" yaml:"total_installed"`
	TotalFailed       int                   `json:"total_failed" yaml:"total_failed"`
	TotalWouldInstall int                   `json:"total_would_install" yaml:"total_would_install"`
	Managers          []ManagerApplyResult  `json:"managers" yaml:"managers"`
}

// ManagerApplyResult represents the result for a single package manager
type ManagerApplyResult struct {
	Name         string                `json:"name" yaml:"name"`
	MissingCount int                   `json:"missing_count" yaml:"missing_count"`
	Packages     []PackageApplyResult  `json:"packages" yaml:"packages"`
}

// PackageApplyResult represents the result for a single package installation
type PackageApplyResult struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"` // pending, installed, failed, would-install
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// TableOutput generates human-friendly table output for apply command
func (a ApplyOutput) TableOutput() string {
	// Table output is handled in the command logic for better real-time feedback
	// This method is required by the interface but not used for table format
	return ""
}

// StructuredData returns the structured data for serialization
func (a ApplyOutput) StructuredData() any {
	return a
}