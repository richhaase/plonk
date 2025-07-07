package commands

import (
	"fmt"

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

	Outputln(cmd, "Package Manager Status")
	Outputln(cmd, "=====================")
	Outputln(cmd, "")

	for _, mgr := range packageManagers {
		Outputln(cmd, "## %s", mgr.Name)

		if !mgr.Manager.IsAvailable() {
			Outputln(cmd, "❌ Not available\n")
			continue
		}

		Outputln(cmd, "✅ Available")

		packages, err := mgr.Manager.ListInstalled()
		if err != nil {
			// Always show errors, even in quiet mode
			fmt.Printf("⚠️  Error listing packages: %v\n\n", err)
			continue
		}

		if !IsQuiet(cmd) {
			if len(packages) == 0 {
				fmt.Printf("📦 No packages installed\n")
			} else {
				fmt.Printf("📦 %d packages installed\n", len(packages))
			}
		}

		if IsVerbose(cmd) {
			// Show actual package list in verbose mode
			if len(packages) > 0 {
				fmt.Println("   Installed packages:")
				for _, pkg := range packages {
					fmt.Printf("   - %s\n", pkg)
				}
			}
		}

		if !IsQuiet(cmd) {
			fmt.Println()
		}
	}

	// Always show drift detection.
	return showDriftStatus(cmd)
}

// runStatusWithDrift runs status with drift detection (for testing).
func runStatusWithDrift() error {
	return runStatus(nil, []string{})
}

// showDriftStatus displays configuration drift information.
func showDriftStatus(cmd *cobra.Command) error {
	if !IsQuiet(cmd) {
		fmt.Println("Configuration Drift Detection")
		fmt.Println("============================")
		fmt.Println()
	}

	drift, err := detectConfigDrift()
	if err != nil {
		// Always show errors
		fmt.Printf("⚠️  Error detecting drift: %v\n", err)
		return nil // Don't fail status command for drift errors.
	}

	if !drift.HasDrift() {
		if !IsQuiet(cmd) {
			fmt.Println("✅ No configuration drift detected")
			fmt.Println("All configurations are in sync with your plonk.yaml")
			fmt.Println()
		}
		return nil
	}

	if !IsQuiet(cmd) {
		fmt.Println("🔄 Configuration drift detected:")
		fmt.Println()
	}

	// Show missing files.
	if len(drift.MissingFiles) > 0 {
		if !IsQuiet(cmd) {
			fmt.Printf("📄 Missing configuration files (%d):\n", len(drift.MissingFiles))
			for _, file := range drift.MissingFiles {
				fmt.Printf("   • %s\n", file)
			}
			fmt.Println()
		}
	}

	// Show modified files.
	if len(drift.ModifiedFiles) > 0 {
		if !IsQuiet(cmd) {
			fmt.Printf("📝 Modified configuration files (%d):\n", len(drift.ModifiedFiles))
			for _, file := range drift.ModifiedFiles {
				fmt.Printf("   • %s\n", file)
			}
			fmt.Println()
		}

		if IsVerbose(cmd) {
			// In verbose mode, show what's different
			fmt.Println("   Details of modifications:")
			for _, file := range drift.ModifiedFiles {
				fmt.Printf("   - %s (content differs from configuration)\n", file)
			}
		}
	}

	// Show missing packages.
	if len(drift.MissingPackages) > 0 {
		if !IsQuiet(cmd) {
			fmt.Printf("📦 Missing packages (%d):\n", len(drift.MissingPackages))
			for _, pkg := range drift.MissingPackages {
				fmt.Printf("   • %s\n", pkg)
			}
			fmt.Println()
		}
	}

	// Show extra packages.
	if len(drift.ExtraPackages) > 0 {
		if !IsQuiet(cmd) {
			fmt.Printf("➕ Extra packages (%d):\n", len(drift.ExtraPackages))
			for _, pkg := range drift.ExtraPackages {
				fmt.Printf("   • %s\n", pkg)
			}
			fmt.Println()
		}
	}

	if !IsQuiet(cmd) {
		fmt.Println("💡 To fix drift:")
		fmt.Println("   plonk install  # Install missing packages")
		fmt.Println("   plonk apply    # Apply missing configurations")
		fmt.Println()
	}

	return nil
}
