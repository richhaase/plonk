// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"plonk/pkg/managers"

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
	executor := managers.NewRealCommandExecutor()
	packageManagers := []struct {
		name    string
		manager managers.PackageManager
	}{
		{"Homebrew", managers.NewHomebrewManager(executor)},
		{"ASDF", managers.NewAsdfManager(executor)},
		{"NPM", managers.NewNpmManager(executor)},
	}

	hasAnyPackages := false


	for _, mgr := range packageManagers {
		if !mgr.manager.IsAvailable() {
			continue
		}

		var packages []string
		var err error

		// Use different methods based on filter
		switch filter {
		case "all":
			packages, err = mgr.manager.ListInstalled()
		case "managed":
			// TODO: Implement managed filter at higher level
			fmt.Printf("# %s (managed filter not yet implemented)\n", mgr.name)
			continue
		case "untracked":
			// TODO: Implement untracked filter at higher level
			fmt.Printf("# %s (untracked filter not yet implemented)\n", mgr.name)
			continue
		case "missing":
			// TODO: Implement missing filter at higher level
			fmt.Printf("# %s (missing filter not yet implemented)\n", mgr.name)
			continue
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