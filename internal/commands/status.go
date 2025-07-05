package commands

import (
	"fmt"

	"github.com/rdh/plonk/pkg/managers"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all package managers",
	Long: `Display the availability and installed packages for shell environment management:
- Homebrew (primary package installation)
- ASDF (programming language tools and versions)
- NPM (packages not available via Homebrew, like claude-code)`,
	RunE: runStatus,
}

// Removed --all flag - use 'plonk pkg list' for detailed package listings

func runStatus(cmd *cobra.Command, args []string) error {
	executor := managers.NewRealCommandExecutor()
	
	// Initialize package managers for shell environment management
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
			fmt.Printf("📦 No packages installed\n")
		} else {
			fmt.Printf("📦 %d packages installed\n", len(packages))
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