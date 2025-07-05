package commands

import (
	"fmt"

	"github.com/rdh/plonk/pkg/managers"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all package managers",
	Long: `Display the availability and installed packages for all supported package managers:
- Homebrew (macOS packages)
- ASDF (programming language tools and versions)
- NPM (Node.js global packages)
- Pip (Python packages)
- Cargo (Rust packages)`,
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	executor := managers.NewRealCommandExecutor()
	
	// Initialize all package managers
	packageManagers := []PackageManagerInfo{
		{
			name:    "Homebrew",
			manager: managers.NewHomebrewManager(executor),
		},
		{
			name:    "ASDF",
			manager: managers.NewAsdfManager(executor),
		},
		{
			name:    "NPM",
			manager: managers.NewNpmManager(executor),
		},
		{
			name:    "Pip",
			manager: managers.NewPipManager(executor),
		},
		{
			name:    "Cargo",
			manager: managers.NewCargoManager(executor),
		},
	}
	
	fmt.Println("Package Manager Status")
	fmt.Println("=====================")
	fmt.Println()
	
	for _, mgr := range packageManagers {
		fmt.Printf("## %s\n", mgr.name)
		
		if !mgr.manager.IsAvailable() {
			fmt.Printf("❌ Not available\n\n")
			continue
		}
		
		fmt.Printf("✅ Available\n")
		
		packages, err := mgr.manager.ListInstalled()
		if err != nil {
			fmt.Printf("⚠️  Error listing packages: %v\n\n", err)
			continue
		}
		
		if len(packages) == 0 {
			fmt.Printf("📦 No packages installed\n\n")
			continue
		}
		
		fmt.Printf("📦 %d packages installed:\n", len(packages))
		
		// Show first few packages, with option to show all
		const maxDisplay = 5
		displayCount := len(packages)
		if displayCount > maxDisplay {
			displayCount = maxDisplay
		}
		
		for i := 0; i < displayCount; i++ {
			fmt.Printf("   - %s\n", packages[i])
		}
		
		if len(packages) > maxDisplay {
			fmt.Printf("   ... and %d more\n", len(packages)-maxDisplay)
		}
		
		fmt.Println()
	}
	
	return nil
}

// PackageManagerInfo holds a package manager and its display name
type PackageManagerInfo struct {
	name    string
	manager PackageManager
}

// PackageManager interface defines the common operations for all package managers
type PackageManager interface {
	IsAvailable() bool
	ListInstalled() ([]string, error)
}