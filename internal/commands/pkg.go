package commands

import (
	"fmt"
	"strings"

	"github.com/rdh/plonk/pkg/managers"
	"github.com/spf13/cobra"
)

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Package management commands",
	Long: `Manage packages across your shell environment using Homebrew, ASDF, and NPM.
	
Available package managers:
- brew: Homebrew packages
- asdf: ASDF programming language tools  
- npm: NPM global packages`,
}

var pkgListCmd = &cobra.Command{
	Use:   "list [manager]",
	Short: "List installed packages",
	Long: `List installed packages from one or all package managers.
	
Examples:
  plonk pkg list          # List packages from all managers
  plonk pkg list brew     # List only Homebrew packages
  plonk pkg list asdf     # List only ASDF tools
  plonk pkg list npm      # List only NPM packages`,
	RunE: runPkgList,
	Args: cobra.MaximumNArgs(1),
}

func init() {
	pkgCmd.AddCommand(pkgListCmd)
}

func runPkgList(cmd *cobra.Command, args []string) error {
	executor := managers.NewRealCommandExecutor()
	
	// Initialize all package managers
	allManagers := map[string]PackageManagerInfo{
		"brew": {
			name:    "Homebrew",
			manager: managers.NewHomebrewManager(executor),
		},
		"asdf": {
			name:    "ASDF",
			manager: managers.NewAsdfManager(executor),
		},
		"npm": {
			name:    "NPM",
			manager: managers.NewNpmManager(executor),
		},
	}
	
	// Determine which managers to show
	var managersToShow []PackageManagerInfo
	if len(args) == 0 {
		// Show all managers
		managersToShow = []PackageManagerInfo{
			allManagers["brew"],
			allManagers["asdf"], 
			allManagers["npm"],
		}
	} else {
		// Show specific manager
		managerKey := strings.ToLower(args[0])
		if mgr, exists := allManagers[managerKey]; exists {
			managersToShow = []PackageManagerInfo{mgr}
		} else {
			return fmt.Errorf("unknown package manager '%s'. Available: brew, asdf, npm", args[0])
		}
	}
	
	// List packages for each manager
	for i, mgr := range managersToShow {
		if len(managersToShow) > 1 {
			fmt.Printf("## %s\n", mgr.name)
		}
		
		if !mgr.manager.IsAvailable() {
			fmt.Printf("❌ %s is not available\n", mgr.name)
			if len(managersToShow) > 1 {
				fmt.Println()
			}
			continue
		}
		
		packages, err := mgr.manager.ListInstalled()
		if err != nil {
			fmt.Printf("⚠️  Error listing %s packages: %v\n", mgr.name, err)
			if len(managersToShow) > 1 {
				fmt.Println()
			}
			continue
		}
		
		if len(packages) == 0 {
			fmt.Printf("No packages installed\n")
		} else {
			fmt.Printf("%d packages installed:\n", len(packages))
			for _, pkg := range packages {
				fmt.Printf("  %s\n", pkg)
			}
		}
		
		// Add spacing between managers (except for the last one)
		if len(managersToShow) > 1 && i < len(managersToShow)-1 {
			fmt.Println()
		}
	}
	
	return nil
}