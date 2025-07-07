// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"path/filepath"

	"plonk/internal/directories"
	"plonk/pkg/config"
	"plonk/pkg/importer"
	"plonk/pkg/managers"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Generate plonk.yaml from existing shell environment",
	Long: `Import existing shell environment configuration to create plonk.yaml.
Discovers installed packages from:
- Homebrew (brew list)
- ASDF (asdf list)  
- NPM (npm list -g)

Copies dotfiles:
- .zshrc, .gitconfig, .zshenv

Generates a complete plonk.yaml configuration file.`,
	RunE: importCmdRun,
}

func importCmdRun(cmd *cobra.Command, args []string) error {
	dryRun := IsDryRun(cmd)
	return runImportWithOptions(args, dryRun)
}

func runImport(cmd *cobra.Command, args []string) error {
	return runImportWithOptions(args, false)
}

func runImportWithOptions(args []string, dryRun bool) error {
	if dryRun {
		fmt.Println("Dry-run mode: Showing what would be discovered and imported")
		fmt.Println()
	} else {
		fmt.Println("🔍 Discovering existing shell environment...")
		fmt.Println()
	}

	// Initialize command executor
	executor := managers.NewRealCommandExecutor()

	// Initialize discoverers
	homebrewDiscoverer := importer.NewHomebrewDiscoverer(executor)
	asdfDiscoverer := importer.NewAsdfDiscoverer(executor)
	npmDiscoverer := importer.NewNpmDiscoverer(executor)
	dotfileDiscoverer := importer.NewDotfileDiscoverer()

	// Collect discovery results
	results := config.DiscoveryResults{}

	// Discover Homebrew packages
	fmt.Print("📦 Discovering Homebrew packages... ")
	homebrewPkgs, err := homebrewDiscoverer.DiscoverPackages()
	if err != nil {
		fmt.Println("❌ Error:", err)
	} else {
		results.HomebrewPackages = homebrewPkgs
		fmt.Printf("✅ Found %d packages\n", len(homebrewPkgs))
	}

	// Discover ASDF tools
	fmt.Print("🔧 Discovering ASDF tools... ")
	asdfTools, err := asdfDiscoverer.DiscoverPackages()
	if err != nil {
		fmt.Println("❌ Error:", err)
	} else {
		results.AsdfTools = asdfTools
		fmt.Printf("✅ Found %d tools\n", len(asdfTools))
	}

	// Discover NPM packages
	fmt.Print("📦 Discovering NPM packages... ")
	npmPkgs, err := npmDiscoverer.DiscoverPackages()
	if err != nil {
		fmt.Println("❌ Error:", err)
	} else {
		results.NpmPackages = npmPkgs
		fmt.Printf("✅ Found %d packages\n", len(npmPkgs))
	}

	// Discover dotfiles
	fmt.Print("📄 Discovering dotfiles... ")
	dotfiles, err := dotfileDiscoverer.DiscoverDotfiles()
	if err != nil {
		fmt.Println("❌ Error:", err)
		return fmt.Errorf("failed to discover dotfiles: %w", err)
	}
	results.Dotfiles = dotfiles
	fmt.Printf("✅ Found %d dotfiles\n", len(dotfiles))

	fmt.Println()

	// Generate config from results
	generatedConfig := config.GenerateConfig(results)

	configPath := filepath.Join(directories.Default.RepoDir(), "plonk.yaml")

	if dryRun {
		fmt.Println()
		fmt.Printf("📋 Summary of what would be imported:\n")
		fmt.Printf("   • Homebrew packages: %d\n", len(results.HomebrewPackages))
		fmt.Printf("   • ASDF tools: %d\n", len(results.AsdfTools))
		fmt.Printf("   • NPM packages: %d\n", len(results.NpmPackages))
		fmt.Printf("   • Dotfiles: %d\n", len(results.Dotfiles))
		fmt.Println()
		fmt.Printf("💾 Would save configuration to: %s\n", configPath)
		fmt.Printf("📁 Would copy dotfiles to repo directory\n")
		fmt.Printf("⚙️  Would generate ZSH and Git configurations\n")
		fmt.Println()
		fmt.Println("Dry-run complete. No files were created or modified.")
		return nil
	}

	// Save config to plonk.yaml
	fmt.Printf("💾 Saving configuration to %s... ", configPath)

	if err := config.SaveConfig(generatedConfig, configPath); err != nil {
		fmt.Println("❌ Error")
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Println("✅ Done")

	fmt.Println()
	fmt.Println("✨ Import complete! Your existing environment has been captured in plonk.yaml")
	fmt.Println()
	fmt.Println("📋 Summary:")
	fmt.Printf("   • Homebrew packages: %d\n", len(results.HomebrewPackages))
	fmt.Printf("   • ASDF tools: %d\n", len(results.AsdfTools))
	fmt.Printf("   • NPM packages: %d\n", len(results.NpmPackages))
	fmt.Printf("   • Dotfiles: %d\n", len(results.Dotfiles))
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   1. Review the generated plonk.yaml")
	fmt.Println("   2. Run 'plonk status' to check configuration")
	fmt.Println("   3. Use 'plonk apply' to apply configurations")

	return nil
}
