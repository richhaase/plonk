package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"plonk/pkg/managers"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all package managers and configuration drift",
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
