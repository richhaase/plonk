// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"plonk/internal/config"
	"plonk/internal/lock"
	"plonk/internal/managers"
	"plonk/internal/state"

	"github.com/spf13/cobra"
)

var pkgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List packages across all managers",
	Long: `List packages from Homebrew, NPM, and Cargo managers.

By default, shows managed and missing packages with a count of untracked packages.
Use --verbose to see all packages including the full list of untracked packages.

Examples:
  plonk pkg list                    # Show managed + missing + untracked count
  plonk pkg list --verbose          # Show all packages including untracked
  plonk pkg list --manager homebrew # Show only Homebrew packages
  plonk pkg list --manager npm      # Show only NPM packages
  plonk pkg list --manager cargo    # Show only Cargo packages`,
	RunE: runPkgList,
	Args: cobra.NoArgs,
}

var (
	pkgListVerbose bool
	pkgListManager string
)

func init() {
	pkgListCmd.Flags().BoolVar(&pkgListVerbose, "verbose", false, "Show all packages including untracked")
	pkgListCmd.Flags().StringVar(&pkgListManager, "manager", "", "Filter by package manager (homebrew, npm, cargo)")
	pkgCmd.AddCommand(pkgListCmd)
}

func runPkgList(cmd *cobra.Command, args []string) error {
	// Validate manager filter if provided
	if pkgListManager != "" && pkgListManager != "homebrew" && pkgListManager != "npm" && pkgListManager != "cargo" {
		return fmt.Errorf("invalid manager '%s'. Use: homebrew, npm, cargo", pkgListManager)
	}

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register package provider (multi-manager)
	ctx := context.Background()
	packageProvider := state.NewMultiManagerPackageProvider()
	// Use lock adapter directly - it implements PackageConfigLoader interface
	packageConfigLoader := lockAdapter

	// Add managers based on filter
	if pkgListManager == "" || pkgListManager == "homebrew" {
		homebrewManager := managers.NewHomebrewManager()
		available, err := homebrewManager.IsAvailable(ctx)
		if err != nil {
			return fmt.Errorf("failed to check homebrew availability: %w", err)
		}
		if available {
			managerAdapter := state.NewManagerAdapter(homebrewManager)
			packageProvider.AddManager("homebrew", managerAdapter, packageConfigLoader)
		}
	}

	if pkgListManager == "" || pkgListManager == "npm" {
		npmManager := managers.NewNpmManager()
		available, err := npmManager.IsAvailable(ctx)
		if err != nil {
			return fmt.Errorf("failed to check npm availability: %w", err)
		}
		if available {
			managerAdapter := state.NewManagerAdapter(npmManager)
			packageProvider.AddManager("npm", managerAdapter, packageConfigLoader)
		}
	}

	if pkgListManager == "" || pkgListManager == "cargo" {
		cargoManager := managers.NewCargoManager()
		available, err := cargoManager.IsAvailable(ctx)
		if err != nil {
			return fmt.Errorf("failed to check cargo availability: %w", err)
		}
		if available {
			managerAdapter := state.NewManagerAdapter(cargoManager)
			packageProvider.AddManager("cargo", managerAdapter, packageConfigLoader)
		}
	}

	reconciler.RegisterProvider("package", packageProvider)

	// Reconcile package domain
	result, err := reconciler.ReconcileProvider(ctx, "package")
	if err != nil {
		return fmt.Errorf("failed to reconcile package state: %w", err)
	}

	// Collect all items for processing
	allItems := make([]state.Item, 0)
	allItems = append(allItems, result.Managed...)
	allItems = append(allItems, result.Missing...)
	allItems = append(allItems, result.Untracked...)

	// Convert to enhanced package output format
	var enhancedItems []EnhancedPackageOutput
	for _, item := range allItems {
		enhancedItems = append(enhancedItems, EnhancedPackageOutput{
			Name:    item.Name,
			State:   item.State.String(),
			Manager: item.Manager,
		})
	}

	// Sort items: by state first (managed, missing, untracked), then alphabetically
	sort.Slice(enhancedItems, func(i, j int) bool {
		// Define state priority
		stateOrder := map[string]int{
			"managed":   0,
			"missing":   1,
			"untracked": 2,
		}

		stateI := stateOrder[enhancedItems[i].State]
		stateJ := stateOrder[enhancedItems[j].State]

		if stateI != stateJ {
			return stateI < stateJ
		}

		// Same state, sort alphabetically
		return strings.ToLower(enhancedItems[i].Name) < strings.ToLower(enhancedItems[j].Name)
	})

	// Calculate counts
	managedCount := len(result.Managed)
	missingCount := len(result.Missing)
	untrackedCount := len(result.Untracked)
	totalCount := managedCount + missingCount + untrackedCount

	// Create enhanced manager outputs for structured data
	managerGroups := make(map[string]*EnhancedManagerOutput)
	for _, item := range enhancedItems {
		manager := item.Manager
		if manager == "" {
			manager = "unknown"
		}

		if managerGroups[manager] == nil {
			displayName := manager
			switch manager {
			case "homebrew":
				displayName = "Homebrew"
			case "npm":
				displayName = "NPM"
			case "cargo":
				displayName = "Cargo"
			}

			managerGroups[manager] = &EnhancedManagerOutput{
				Name:           displayName,
				ManagedCount:   0,
				MissingCount:   0,
				UntrackedCount: 0,
				Packages:       []EnhancedPackageOutput{},
			}
		}

		managerGroups[manager].Packages = append(managerGroups[manager].Packages, item)

		switch item.State {
		case "managed":
			managerGroups[manager].ManagedCount++
		case "missing":
			managerGroups[manager].MissingCount++
		case "untracked":
			managerGroups[manager].UntrackedCount++
		}
	}

	// Convert to slice for output
	managers := make([]EnhancedManagerOutput, 0, len(managerGroups))
	for _, mgr := range managerGroups {
		managers = append(managers, *mgr)
	}

	// Sort managers by name
	sort.Slice(managers, func(i, j int) bool {
		return strings.ToLower(managers[i].Name) < strings.ToLower(managers[j].Name)
	})

	// Prepare output structure
	outputData := PackageListOutput{
		ManagedCount:   managedCount,
		MissingCount:   missingCount,
		UntrackedCount: untrackedCount,
		TotalCount:     totalCount,
		Managers:       managers,
		Verbose:        pkgListVerbose,
		Items:          enhancedItems,
	}

	return RenderOutput(outputData, format)
}
