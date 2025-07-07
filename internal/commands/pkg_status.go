// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"plonk/pkg/managers"

	"github.com/spf13/cobra"
)

var pkgStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show package status across all managers",
	Long: `Display the status of packages across Homebrew, ASDF, and NPM managers.

Shows:
- How many packages are installed
- How many are managed by plonk configuration  
- How many are untracked (installed but not in config)
- How many are missing (in config but not installed)`,
	RunE: runPkgStatus,
	Args: cobra.NoArgs,
}

func init() {
	pkgCmd.AddCommand(pkgStatusCmd)
}

func runPkgStatus(cmd *cobra.Command, args []string) error {
	// Get plonk directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	plonkDir := filepath.Join(homeDir, ".config", "plonk")

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

	fmt.Println("Package Status")
	fmt.Println("==============")
	fmt.Println()

	for _, mgr := range packageManagers {
		fmt.Printf("## %s\n", mgr.name)
		
		if !mgr.manager.IsAvailable() {
			fmt.Printf("❌ Not available\n\n")
			continue
		}
		
		fmt.Printf("✅ Available\n")
		
		// Set config directory for state-aware methods
		mgr.manager.SetConfigDir(plonkDir)
		
		// Get package counts using state-aware methods
		installed, err := mgr.manager.ListInstalledPackages()
		if err != nil {
			fmt.Printf("⚠️  Error listing installed packages: %v\n\n", err)
			continue
		}
		
		managed, err := mgr.manager.ListManagedPackages()
		if err != nil {
			fmt.Printf("⚠️  Error listing managed packages: %v\n\n", err)
			continue
		}
		
		untracked, err := mgr.manager.ListUntrackedPackages()
		if err != nil {
			fmt.Printf("⚠️  Error listing untracked packages: %v\n\n", err)
			continue
		}
		
		missing, err := mgr.manager.ListMissingPackages()
		if err != nil {
			fmt.Printf("⚠️  Error listing missing packages: %v\n\n", err)
			continue
		}
		
		// Display package status
		if len(installed) == 0 && len(missing) == 0 {
			fmt.Printf("📦 No packages\n")
		} else {
			parts := []string{}
			
			if len(installed) > 0 {
				parts = append(parts, fmt.Sprintf("%d installed", len(installed)))
			}
			
			if len(managed) > 0 {
				parts = append(parts, fmt.Sprintf("%d managed", len(managed)))
			}
			
			if len(untracked) > 0 {
				parts = append(parts, fmt.Sprintf("%d untracked", len(untracked)))
			}
			
			if len(missing) > 0 {
				parts = append(parts, fmt.Sprintf("%d missing", len(missing)))
			}
			
			if len(parts) > 0 {
				fmt.Printf("📦 %s\n", parts[0])
				if len(parts) > 1 {
					for i := 1; i < len(parts); i++ {
						fmt.Printf("   %s\n", parts[i])
					}
				}
			}
		}
		
		fmt.Println()
	}
	
	return nil
}