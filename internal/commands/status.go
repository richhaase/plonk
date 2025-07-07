// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"

	"plonk/internal/directories"
	"plonk/pkg/managers"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"check"},
	Short:   "Show status of all package managers and configuration drift",
	Long: `Display the availability and installed packages for shell environment management:
- Homebrew (primary package installation)
- ASDF (programming language tools and versions)
- NPM (packages not available via Homebrew, like claude-code)

Includes configuration drift detection comparing current system state with plonk configuration.`,
	RunE: runStatus,
}

func init() {
	// No flags needed - drift detection is always enabled.
}

// Removed --all flag - use 'plonk pkg list' for detailed package listings.

func runStatus(cmd *cobra.Command, args []string) error {
	executor := managers.NewRealCommandExecutor()

	// Initialize package managers for shell environment management.
	packageManagers := []managers.PackageManagerInfo{
		{
			Name:    "Homebrew",
			Manager: managers.NewHomebrewManager(executor),
		},
		{
			Name:    "ASDF",
			Manager: managers.NewAsdfManager(executor),
		},
		{
			Name:    "NPM",
			Manager: managers.NewNpmManager(executor),
		},
	}

	fmt.Println("Package Manager Status")
	fmt.Println("=====================")
	fmt.Println()

	for _, mgr := range packageManagers {
		fmt.Printf("## %s\n", mgr.Name)

		if !mgr.Manager.IsAvailable() {
			fmt.Printf("❌ Not available\n\n")
			continue
		}

		fmt.Printf("✅ Available\n")

		packages, err := mgr.Manager.ListInstalled()
		if err != nil {
			fmt.Printf("⚠️  Error listing packages: %v\n\n", err)
			continue
		}

		if len(packages) == 0 {
			fmt.Printf("📦 No packages installed\n")
		} else {
			fmt.Printf("📦 %d packages installed\n", len(packages))
		}

		fmt.Println()
	}

	// Show dotfiles management status
	if err := showDotfilesStatus(); err != nil {
		fmt.Printf("⚠️  Error showing dotfiles status: %v\n", err)
	}

	// Always show drift detection.
	return showDriftStatus()
}

// runStatusWithDrift runs status with drift detection (for testing).
func runStatusWithDrift() error {
	return runStatus(nil, []string{})
}

// showDriftStatus displays configuration drift information.
func showDriftStatus() error {
	fmt.Println("Configuration Drift Detection")
	fmt.Println("============================")
	fmt.Println()

	drift, err := detectConfigDrift()
	if err != nil {
		fmt.Printf("⚠️  Error detecting drift: %v\n", err)
		return nil // Don't fail status command for drift errors.
	}

	if !drift.HasDrift() {
		fmt.Println("✅ No configuration drift detected")
		fmt.Println("All configurations are in sync with your plonk.yaml")
		fmt.Println()
		return nil
	}

	fmt.Println("🔄 Configuration drift detected:")
	fmt.Println()

	// Show missing files.
	if len(drift.MissingFiles) > 0 {
		fmt.Printf("📄 Missing configuration files (%d):\n", len(drift.MissingFiles))
		for _, file := range drift.MissingFiles {
			fmt.Printf("   • %s\n", file)
		}
		fmt.Println()
	}

	// Show modified files.
	if len(drift.ModifiedFiles) > 0 {
		fmt.Printf("📝 Modified configuration files (%d):\n", len(drift.ModifiedFiles))
		for _, file := range drift.ModifiedFiles {
			fmt.Printf("   • %s\n", file)
		}
		fmt.Println()
	}

	// Show missing packages.
	if len(drift.MissingPackages) > 0 {
		fmt.Printf("📦 Missing packages (%d):\n", len(drift.MissingPackages))
		for _, pkg := range drift.MissingPackages {
			fmt.Printf("   • %s\n", pkg)
		}
		fmt.Println()
	}

	// Show extra packages.
	if len(drift.ExtraPackages) > 0 {
		fmt.Printf("➕ Extra packages (%d):\n", len(drift.ExtraPackages))
		for _, pkg := range drift.ExtraPackages {
			fmt.Printf("   • %s\n", pkg)
		}
		fmt.Println()
	}

	fmt.Println("💡 To fix drift:")
	fmt.Println("   plonk install  # Install missing packages")
	fmt.Println("   plonk apply    # Apply missing configurations")
	fmt.Println()

	return nil
}

// showDotfilesStatus displays dotfiles management information.
func showDotfilesStatus() error {
	fmt.Println("Dotfiles Management Status")
	fmt.Println("=========================")
	fmt.Println()

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Get plonk directory
	plonkDir := directories.Default.PlonkDir()

	// Create dotfiles manager
	dotfilesManager := managers.NewDotfilesManager(homeDir, plonkDir)

	// Get dotfiles information
	managed, err := dotfilesManager.ListManaged()
	if err != nil {
		return fmt.Errorf("failed to list managed dotfiles: %w", err)
	}

	untracked, err := dotfilesManager.ListUntracked()
	if err != nil {
		return fmt.Errorf("failed to list untracked dotfiles: %w", err)
	}

	missing, err := dotfilesManager.ListMissing()
	if err != nil {
		return fmt.Errorf("failed to list missing dotfiles: %w", err)
	}

	modified, err := dotfilesManager.ListModified()
	if err != nil {
		return fmt.Errorf("failed to list modified dotfiles: %w", err)
	}

	// Display summary
	if len(managed) > 0 {
		fmt.Printf("📄 %d dotfiles managed", len(managed))
		if len(managed) <= 5 {
			fmt.Print(" (")
			for i, dotfile := range managed {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(dotfile.Name)
			}
			fmt.Print(")")
		}
		fmt.Println()
	} else {
		fmt.Println("📄 No dotfiles currently managed")
	}

	if len(untracked) > 0 {
		fmt.Printf("🔍 %d dotfiles untracked", len(untracked))
		if len(untracked) <= 5 {
			fmt.Print(" (")
			for i, dotfile := range untracked {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(dotfile.Name)
			}
			fmt.Print(")")
		}
		fmt.Println()
	}

	if len(missing) > 0 {
		fmt.Printf("❌ %d dotfiles missing", len(missing))
		if len(missing) <= 5 {
			fmt.Print(" (")
			for i, dotfile := range missing {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(dotfile.Name)
			}
			fmt.Print(")")
		}
		fmt.Println()
	}

	if len(modified) > 0 {
		fmt.Printf("📝 %d dotfiles modified", len(modified))
		if len(modified) <= 5 {
			fmt.Print(" (")
			for i, dotfile := range modified {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(dotfile.Name)
			}
			fmt.Print(")")
		}
		fmt.Println()
	}

	fmt.Println()
	return nil
}
