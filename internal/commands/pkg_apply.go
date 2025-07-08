// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/managers"
	"plonk/internal/state"

	"github.com/spf13/cobra"
)

var (
	pkgApplyDryRun bool
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
	pkgApplyCmd.Flags().BoolVar(&pkgApplyDryRun, "dry-run", false, "Show what would be installed without making changes")
}

func runPkgApply(cmd *cobra.Command, args []string) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "plonk")

	// Load configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register package provider (multi-manager)
	packageProvider := createPackageProvider(cfg)
	reconciler.RegisterProvider("package", packageProvider)

	// Reconcile package domain to find missing packages
	result, err := reconciler.ReconcileProvider("package")
	if err != nil {
		return fmt.Errorf("failed to reconcile package state: %w", err)
	}

	// Group missing packages by manager
	missingByManager := make(map[string][]state.Item)
	for _, item := range result.Missing {
		manager := item.Manager
		if manager == "" {
			manager = "unknown"
		}
		missingByManager[manager] = append(missingByManager[manager], item)
	}

	// Prepare output structure
	outputData := ApplyOutput{
		DryRun:         pkgApplyDryRun,
		TotalMissing:   len(result.Missing),
		Managers:       make([]ManagerApplyResult, 0, len(missingByManager)),
	}

	// Handle case where no packages are missing
	if len(result.Missing) == 0 {
		if format == OutputTable {
			fmt.Println("# All packages up to date")
		}
		return RenderOutput(outputData, format)
	}

	// Process each manager that has missing packages
	managerInstances := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"npm":      managers.NewNpmManager(),
	}

	for managerName, missingItems := range missingByManager {
		managerInstance, exists := managerInstances[managerName]
		if !exists {
			if format == OutputTable {
				fmt.Printf("# %s: Unknown manager, skipping\n", managerName)
			}
			continue
		}

		if !managerInstance.IsAvailable() {
			if format == OutputTable {
				fmt.Printf("# %s: Not available, skipping\n", managerName)
			}
			continue
		}

		// Convert manager name for display
		displayName := managerName
		switch managerName {
		case "homebrew":
			displayName = "Homebrew"
		case "npm":
			displayName = "NPM"
		}

		// Process missing packages for this manager
		managerResult := ManagerApplyResult{
			Name:         displayName,
			MissingCount: len(missingItems),
			Packages:     make([]PackageApplyResult, 0, len(missingItems)),
		}

		for _, item := range missingItems {
			packageResult := PackageApplyResult{
				Name:   item.Name,
				Status: "pending",
			}

			if pkgApplyDryRun {
				packageResult.Status = "would-install"
				if format == OutputTable {
					fmt.Printf("Would install: %s (%s)\n", item.Name, displayName)
				}
				outputData.TotalWouldInstall++
			} else {
				// Actually install the package
				err := managerInstance.Install(item.Name)
				if err != nil {
					packageResult.Status = "failed"
					packageResult.Error = err.Error()
					if format == OutputTable {
						fmt.Printf("Failed to install %s: %v\n", item.Name, err)
					}
					outputData.TotalFailed++
				} else {
					packageResult.Status = "installed"
					if format == OutputTable {
						fmt.Printf("Installed: %s (%s)\n", item.Name, displayName)
					}
					outputData.TotalInstalled++
				}
			}

			managerResult.Packages = append(managerResult.Packages, packageResult)
		}

		outputData.Managers = append(outputData.Managers, managerResult)
	}

	// Output summary for table format
	if format == OutputTable {
		fmt.Println()
		if pkgApplyDryRun {
			fmt.Printf("Summary: %d packages would be installed\n", outputData.TotalWouldInstall)
		} else {
			fmt.Printf("Summary: %d installed, %d failed\n", outputData.TotalInstalled, outputData.TotalFailed)
		}
	}

	return RenderOutput(outputData, format)
}

// ApplyOutput represents the output structure for pkg apply command
type ApplyOutput struct {
	DryRun             bool                 `json:"dry_run" yaml:"dry_run"`
	TotalMissing       int                  `json:"total_missing" yaml:"total_missing"`
	TotalInstalled     int                  `json:"total_installed" yaml:"total_installed"`
	TotalFailed        int                  `json:"total_failed" yaml:"total_failed"`
	TotalWouldInstall  int                  `json:"total_would_install" yaml:"total_would_install"`
	Managers           []ManagerApplyResult `json:"managers" yaml:"managers"`
}

// ManagerApplyResult represents the result for a specific manager
type ManagerApplyResult struct {
	Name         string               `json:"name" yaml:"name"`
	MissingCount int                  `json:"missing_count" yaml:"missing_count"`
	Packages     []PackageApplyResult `json:"packages" yaml:"packages"`
}

// PackageApplyResult represents the result for a specific package
type PackageApplyResult struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// TableOutput generates human-friendly table output for apply results
func (a ApplyOutput) TableOutput() string {
	// Table output is handled inline in the command
	// This method is required by the OutputData interface
	return ""
}

// StructuredData returns the structured data for serialization
func (a ApplyOutput) StructuredData() any {
	return a
}