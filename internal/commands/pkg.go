package commands

import (
	"fmt"
	"strings"

	"plonk/pkg/managers"

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

var pkgSearchCmd = &cobra.Command{
	Use:   "search <query> [manager]",
	Short: "Search for packages",
	Long: `Search for packages across package managers.
	
Examples:
  plonk pkg search git          # Search for git packages across all managers
  plonk pkg search node brew    # Search for node packages in Homebrew only
  plonk pkg search python asdf  # Search for python packages in ASDF only`,
	RunE: runPkgSearch,
	Args: cobra.MinimumNArgs(1),
}

var pkgInfoCmd = &cobra.Command{
	Use:   "info <package> [manager]",
	Short: "Show package information",
	Long: `Show detailed information about a package.
	
Examples:
  plonk pkg info git            # Show git package info from all managers
  plonk pkg info node brew      # Show node package info from Homebrew
  plonk pkg info python asdf    # Show python plugin info from ASDF`,
	RunE: runPkgInfo,
	Args: cobra.MinimumNArgs(1),
}

var pkgUpdateCmd = &cobra.Command{
	Use:   "update [package] [manager]",
	Short: "Update packages",
	Long: `Update packages across package managers.
	
Examples:
  plonk pkg update              # Update all packages across all managers
  plonk pkg update git          # Update git package across all managers
  plonk pkg update git brew     # Update git package in Homebrew only`,
	RunE: runPkgUpdate,
}

func init() {
	pkgCmd.AddCommand(pkgListCmd)
	pkgCmd.AddCommand(pkgSearchCmd)
	pkgCmd.AddCommand(pkgInfoCmd)
	pkgCmd.AddCommand(pkgUpdateCmd)
}

func runPkgList(cmd *cobra.Command, args []string) error {
	executor := managers.NewRealCommandExecutor()

	// Initialize all package managers
	allManagers := map[string]managers.PackageManagerInfo{
		"brew": {
			Name:    "Homebrew",
			Manager: managers.NewHomebrewManager(executor),
		},
		"asdf": {
			Name:    "ASDF",
			Manager: managers.NewAsdfManager(executor),
		},
		"npm": {
			Name:    "NPM",
			Manager: managers.NewNpmManager(executor),
		},
	}

	// Determine which managers to show
	var managersToShow []managers.PackageManagerInfo
	if len(args) == 0 {
		// Show all managers
		managersToShow = []managers.PackageManagerInfo{
			allManagers["brew"],
			allManagers["asdf"],
			allManagers["npm"],
		}
	} else {
		// Show specific manager
		managerKey := strings.ToLower(args[0])
		if mgr, exists := allManagers[managerKey]; exists {
			managersToShow = []managers.PackageManagerInfo{mgr}
		} else {
			return fmt.Errorf("unknown package manager '%s'. Available: brew, asdf, npm", args[0])
		}
	}

	// List packages for each manager
	for i, mgr := range managersToShow {
		if len(managersToShow) > 1 {
			fmt.Printf("## %s\n", mgr.Name)
		}

		if !mgr.Manager.IsAvailable() {
			fmt.Printf("❌ %s is not available\n", mgr.Name)
			if len(managersToShow) > 1 {
				fmt.Println()
			}
			continue
		}

		packages, err := mgr.Manager.ListInstalled()
		if err != nil {
			fmt.Printf("⚠️  Error listing %s packages: %v\n", mgr.Name, err)
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

func runPkgSearch(cmd *cobra.Command, args []string) error {
	executor := managers.NewRealCommandExecutor()
	query := args[0]

	// Initialize all package managers
	allManagers := map[string]managers.ExtendedPackageManager{
		"brew": managers.NewHomebrewManager(executor),
		"asdf": managers.NewAsdfManager(executor),
		"npm":  managers.NewNpmManager(executor),
	}

	// Determine which managers to search
	var managersToSearch []struct {
		name    string
		manager managers.ExtendedPackageManager
	}

	if len(args) >= 2 {
		// Search specific manager
		managerKey := strings.ToLower(args[1])
		if mgr, exists := allManagers[managerKey]; exists {
			managersToSearch = append(managersToSearch, struct {
				name    string
				manager managers.ExtendedPackageManager
			}{managerKey, mgr})
		} else {
			return fmt.Errorf("unknown package manager '%s'. Available: brew, asdf, npm", args[1])
		}
	} else {
		// Search all managers
		for name, mgr := range allManagers {
			managersToSearch = append(managersToSearch, struct {
				name    string
				manager managers.ExtendedPackageManager
			}{name, mgr})
		}
	}

	// Search packages for each manager
	found := false
	for i, entry := range managersToSearch {
		if !entry.manager.IsAvailable() {
			continue
		}

		results, err := entry.manager.Search(query)
		if err != nil {
			fmt.Printf("⚠️  Error searching %s: %v\n", entry.name, err)
			continue
		}

		if len(results) > 0 {
			found = true
			if len(managersToSearch) > 1 {
				fmt.Printf("## %s\n", strings.Title(entry.name))
			}

			for _, pkg := range results {
				fmt.Printf("  %s\n", pkg)
			}

			// Add spacing between managers (except for the last one)
			if len(managersToSearch) > 1 && i < len(managersToSearch)-1 {
				fmt.Println()
			}
		}
	}

	if !found {
		fmt.Printf("No packages found matching '%s'\n", query)
	}

	return nil
}

func runPkgInfo(cmd *cobra.Command, args []string) error {
	executor := managers.NewRealCommandExecutor()
	packageName := args[0]

	// Initialize all package managers
	allManagers := map[string]managers.ExtendedPackageManager{
		"brew": managers.NewHomebrewManager(executor),
		"asdf": managers.NewAsdfManager(executor),
		"npm":  managers.NewNpmManager(executor),
	}

	// Determine which managers to query
	var managersToQuery []struct {
		name    string
		manager managers.ExtendedPackageManager
	}

	if len(args) >= 2 {
		// Query specific manager
		managerKey := strings.ToLower(args[1])
		if mgr, exists := allManagers[managerKey]; exists {
			managersToQuery = append(managersToQuery, struct {
				name    string
				manager managers.ExtendedPackageManager
			}{managerKey, mgr})
		} else {
			return fmt.Errorf("unknown package manager '%s'. Available: brew, asdf, npm", args[1])
		}
	} else {
		// Query all managers
		for name, mgr := range allManagers {
			managersToQuery = append(managersToQuery, struct {
				name    string
				manager managers.ExtendedPackageManager
			}{name, mgr})
		}
	}

	// Get info from each manager
	found := false
	for i, entry := range managersToQuery {
		if !entry.manager.IsAvailable() {
			continue
		}

		info, err := entry.manager.Info(packageName)
		if err != nil {
			continue // Package might not exist in this manager
		}

		if info != "" {
			found = true
			if len(managersToQuery) > 1 {
				fmt.Printf("## %s\n", strings.Title(entry.name))
			}

			fmt.Printf("%s\n", info)

			// Add spacing between managers (except for the last one)
			if len(managersToQuery) > 1 && i < len(managersToQuery)-1 {
				fmt.Println()
			}
		}
	}

	if !found {
		fmt.Printf("No information found for package '%s'\n", packageName)
	}

	return nil
}

func runPkgUpdate(cmd *cobra.Command, args []string) error {
	executor := managers.NewRealCommandExecutor()

	// Initialize all package managers
	allManagers := map[string]managers.ExtendedPackageManager{
		"brew": managers.NewHomebrewManager(executor),
		"asdf": managers.NewAsdfManager(executor),
		"npm":  managers.NewNpmManager(executor),
	}

	var packageName string
	var managerFilter string

	// Parse arguments
	if len(args) >= 1 {
		packageName = args[0]
	}
	if len(args) >= 2 {
		managerFilter = strings.ToLower(args[1])
	}

	// Determine which managers to update
	var managersToUpdate []struct {
		name    string
		manager managers.ExtendedPackageManager
	}

	if managerFilter != "" {
		if mgr, exists := allManagers[managerFilter]; exists {
			managersToUpdate = append(managersToUpdate, struct {
				name    string
				manager managers.ExtendedPackageManager
			}{managerFilter, mgr})
		} else {
			return fmt.Errorf("unknown package manager '%s'. Available: brew, asdf, npm", managerFilter)
		}
	} else {
		// Update all managers
		for name, mgr := range allManagers {
			managersToUpdate = append(managersToUpdate, struct {
				name    string
				manager managers.ExtendedPackageManager
			}{name, mgr})
		}
	}

	// Update packages for each manager
	for _, entry := range managersToUpdate {
		if !entry.manager.IsAvailable() {
			fmt.Printf("❌ %s is not available\n", strings.Title(entry.name))
			continue
		}

		var err error
		if packageName != "" {
			// Update specific package
			fmt.Printf("🔄 Updating %s in %s...\n", packageName, strings.Title(entry.name))
			err = entry.manager.Update(packageName)
		} else {
			// Update all packages
			fmt.Printf("🔄 Updating all packages in %s...\n", strings.Title(entry.name))
			err = entry.manager.UpdateAll()
		}

		if err != nil {
			fmt.Printf("⚠️  Error updating %s: %v\n", strings.Title(entry.name), err)
		} else {
			fmt.Printf("✅ Successfully updated %s\n", strings.Title(entry.name))
		}
	}

	return nil
}
