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
	Long: `List packages from Homebrew, ASDF, and NPM managers.

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

	// Initialize package managers
	packageManagers := []struct {
		name    string
		manager managers.PackageManager
	}{
		{"Homebrew", managers.NewHomebrewManager()},
		{"ASDF", managers.NewAsdfManager()},
		{"NPM", managers.NewNpmManager()},
	}

	hasAnyPackages := false


	for _, mgr := range packageManagers {
		if !mgr.manager.IsAvailable() {
			continue
		}

		var packages []string
		var err error

		// Handle different filters
		switch filter {
		case "all":
			packages, err = mgr.manager.ListInstalled()
			if err != nil {
				fmt.Printf("# %s: Error listing packages: %v\n", mgr.name, err)
				continue
			}
		case "managed", "untracked", "missing":
			// Use state reconciliation for these filters
			configDir, configErr := managers.DefaultConfigDir()
			if configErr != nil {
				fmt.Printf("# %s: Error getting config directory: %v\n", mgr.name, configErr)
				continue
			}
			
			loader := managers.NewPlonkConfigLoader(configDir)
			checkers := map[string]managers.VersionChecker{
				"homebrew": &managers.HomebrewVersionChecker{},
				"asdf":     &managers.AsdfVersionChecker{},
				"npm":      &managers.NpmVersionChecker{},
			}
			// Convert manager display name to config name
			managerKey := mgr.name
			switch mgr.name {
			case "Homebrew":
				managerKey = "homebrew"
			case "ASDF":
				managerKey = "asdf"
			case "NPM":
				managerKey = "npm"
			}
			
			managerMap := map[string]managers.PackageManager{
				managerKey: mgr.manager,
			}
			
			reconciler := managers.NewStateReconciler(loader, managerMap, checkers)
			result, reconcileErr := reconciler.ReconcileManager(managerKey)
			if reconcileErr != nil {
				fmt.Printf("# %s: Error reconciling state: %v\n", mgr.name, reconcileErr)
				continue
			}
			
			// Extract packages based on filter
			var statePackages []managers.Package
			switch filter {
			case "managed":
				statePackages = result.Managed
			case "untracked":
				statePackages = result.Untracked
			case "missing":
				statePackages = result.Missing
			}
			
			// Convert to string slice for display
			packages = make([]string, len(statePackages))
			for i, pkg := range statePackages {
				packages[i] = pkg.Name
			}
		}

		if err != nil {
			fmt.Printf("# %s: Error listing packages: %v\n", mgr.name, err)
			continue
		}

		if len(packages) == 0 {
			continue
		}

		// Display packages for this manager
		if !hasAnyPackages {
			hasAnyPackages = true
		}

		fmt.Printf("# %s (%d packages)\n", mgr.name, len(packages))
		for _, pkg := range packages {
			fmt.Printf("%s\n", pkg)
		}
		fmt.Println()
	}

	if !hasAnyPackages {
		fmt.Println("No packages found")
	}

	return nil
}