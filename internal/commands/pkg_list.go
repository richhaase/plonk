// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"plonk/internal/managers"

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

	// Initialize package managers
	packageManagers := []struct {
		name    string
		manager managers.PackageManager
	}{
		{"Homebrew", managers.NewHomebrewManager()},
		{"NPM", managers.NewNpmManager()},
	}

	// Prepare output structure
	var outputData PackageListOutput
	outputData.Filter = filter


	for _, mgr := range packageManagers {
		if !mgr.manager.IsAvailable() {
			continue
		}

		var packages []string
		var statePackages []managers.Package
		var err error

		// Handle different filters
		switch filter {
		case "all":
			packages, err = mgr.manager.ListInstalled()
			if err != nil {
				if format == OutputTable {
					fmt.Printf("# %s: Error listing packages: %v\n", mgr.name, err)
				}
				continue
			}
		case "managed", "untracked", "missing":
			// Use state reconciliation for these filters
			configDir, configErr := managers.DefaultConfigDir()
			if configErr != nil {
				if format == OutputTable {
					fmt.Printf("# %s: Error getting config directory: %v\n", mgr.name, configErr)
				}
				continue
			}
			
			loader := managers.NewPlonkConfigLoader(configDir)
			// Convert manager display name to config name
			managerKey := mgr.name
			switch mgr.name {
			case "Homebrew":
				managerKey = "homebrew"
			case "NPM":
				managerKey = "npm"
			}
			
			managerMap := map[string]managers.PackageManager{
				managerKey: mgr.manager,
			}
			
			reconciler := managers.NewStateReconciler(loader, managerMap)
			result, reconcileErr := reconciler.ReconcileManager(managerKey)
			if reconcileErr != nil {
				if format == OutputTable {
					fmt.Printf("# %s: Error reconciling state: %v\n", mgr.name, reconcileErr)
				}
				continue
			}
			
			// Extract packages based on filter
			switch filter {
			case "managed":
				statePackages = result.Managed
			case "untracked":
				statePackages = result.Untracked
			case "missing":
				statePackages = result.Missing
			}
			
			// Convert to string slice for backwards compatibility
			packages = make([]string, len(statePackages))
			for i, pkg := range statePackages {
				packages[i] = pkg.Name
			}
		}

		if err != nil {
			if format == OutputTable {
				fmt.Printf("# %s: Error listing packages: %v\n", mgr.name, err)
			}
			continue
		}

		// Create manager output
		managerOutput := ManagerOutput{
			Name:     mgr.name,
			Count:    len(packages),
			Packages: make([]PackageOutput, len(packages)),
		}

		// Convert packages to output format
		if len(statePackages) > 0 {
			// Use rich package data for state-based filters
			for i, pkg := range statePackages {
				managerOutput.Packages[i] = PackageOutput{
					Name:  pkg.Name,
					State: stateToString(pkg.State),
				}
			}
		} else {
			// Use simple package names for "all" filter
			for i, pkg := range packages {
				managerOutput.Packages[i] = PackageOutput{
					Name: pkg,
				}
			}
		}

		outputData.Managers = append(outputData.Managers, managerOutput)
	}

	return RenderOutput(outputData, format)
}

// stateToString converts PackageState to string
func stateToString(state managers.PackageState) string {
	switch state {
	case managers.StateManaged:
		return "managed"
	case managers.StateMissing:
		return "missing"
	case managers.StateUntracked:
		return "untracked"
	default:
		return "unknown"
	}
}