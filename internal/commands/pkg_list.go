// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/managers"
	"plonk/internal/state"

	"github.com/spf13/cobra"
)

var pkgListCmd = &cobra.Command{
	Use:   "list [filter]",
	Short: "List packages across all managers",
	Long: `List packages from Homebrew and NPM managers.

Available filters:
  (no filter)  List all installed packages
  managed      List packages managed by plonk configuration
  untracked    List installed packages not in plonk configuration  
  missing      List packages in configuration but not installed

Examples:
  plonk pkg list           # List all installed packages
  plonk pkg list managed   # List only packages in plonk.yaml
  plonk pkg list untracked # List packages not tracked by plonk`,
	RunE: runPkgList,
	Args: cobra.MaximumNArgs(1),
}

func init() {
	pkgCmd.AddCommand(pkgListCmd)
}

func runPkgList(cmd *cobra.Command, args []string) error {
	// Determine filter type
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
		if filter != "managed" && filter != "untracked" && filter != "missing" && filter != "all" {
			return fmt.Errorf("invalid filter '%s'. Use: managed, untracked, missing, or no filter for all", filter)
		}
	}

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
	ctx := context.Background()
	packageProvider := state.NewMultiManagerPackageProvider()
	configAdapter := config.NewConfigAdapter(cfg)
	packageConfigAdapter := config.NewStatePackageConfigAdapter(configAdapter)
	
	// Add Homebrew manager
	homebrewManager := managers.NewHomebrewManager()
	available, err := homebrewManager.IsAvailable(ctx)
	if err != nil {
		return fmt.Errorf("failed to check homebrew availability: %w", err)
	}
	if available {
		managerAdapter := state.NewManagerAdapter(homebrewManager)
		packageProvider.AddManager("homebrew", managerAdapter, packageConfigAdapter)
	}
	
	// Add NPM manager
	npmManager := managers.NewNpmManager()
	available, err = npmManager.IsAvailable(ctx)
	if err != nil {
		return fmt.Errorf("failed to check npm availability: %w", err)
	}
	if available {
		managerAdapter := state.NewManagerAdapter(npmManager)
		packageProvider.AddManager("npm", managerAdapter, packageConfigAdapter)
	}
	
	reconciler.RegisterProvider("package", packageProvider)

	// Reconcile package domain
	result, err := reconciler.ReconcileProvider(ctx, "package")
	if err != nil {
		return fmt.Errorf("failed to reconcile package state: %w", err)
	}

	// Filter items based on the requested filter
	var filteredItems []state.Item
	switch filter {
	case "all":
		filteredItems = append(filteredItems, result.Managed...)
		filteredItems = append(filteredItems, result.Untracked...)
	case "managed":
		filteredItems = result.Managed
	case "untracked":
		filteredItems = result.Untracked
	case "missing":
		filteredItems = result.Missing
	}

	// Group items by manager for output
	managerGroups := make(map[string][]state.Item)
	for _, item := range filteredItems {
		manager := item.Manager
		if manager == "" {
			manager = "unknown"
		}
		managerGroups[manager] = append(managerGroups[manager], item)
	}

	// Prepare output structure
	outputData := PackageListOutput{
		Filter:   filter,
		Managers: make([]ManagerOutput, 0, len(managerGroups)),
	}

	// Convert to legacy output format for compatibility
	for managerName, items := range managerGroups {
		// Convert manager name for display
		displayName := managerName
		switch managerName {
		case "homebrew":
			displayName = "Homebrew"
		case "npm":
			displayName = "NPM"
		}

		managerOutput := ManagerOutput{
			Name:     displayName,
			Count:    len(items),
			Packages: make([]PackageOutput, len(items)),
		}

		for i, item := range items {
			managerOutput.Packages[i] = PackageOutput{
				Name:  item.Name,
				State: item.State.String(),
			}
		}

		outputData.Managers = append(outputData.Managers, managerOutput)
	}

	return RenderOutput(outputData, format)
}