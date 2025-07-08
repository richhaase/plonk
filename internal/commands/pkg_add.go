// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/managers"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	manager string
)

var pkgAddCmd = &cobra.Command{
	Use:   "add <package>",
	Short: "Add a package to plonk configuration and install it",
	Long: `Add a package to your plonk.yaml configuration and install it immediately.

The package will be added to the appropriate manager section in your configuration
and then installed using that manager.

You can specify which manager to use with the --manager flag, otherwise it will
use the default manager from your configuration.

Examples:
  plonk pkg add htop              # Add htop using default manager
  plonk pkg add git --manager homebrew  # Add git specifically to homebrew
  plonk pkg add lodash --manager npm     # Add lodash to npm global packages`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgAdd,
}

func init() {
	pkgCmd.AddCommand(pkgAddCmd)
	pkgAddCmd.Flags().StringVar(&manager, "manager", "", "Package manager to use (homebrew|npm)")
}

func runPkgAdd(cmd *cobra.Command, args []string) error {
	packageName := args[0]

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "plonk")

	// Load existing configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine which manager to use
	targetManager := manager
	if targetManager == "" {
		targetManager = cfg.Settings.DefaultManager
	}

	// Validate manager
	if targetManager != "homebrew" && targetManager != "npm" {
		return fmt.Errorf("unsupported manager '%s'. Use: homebrew, npm", targetManager)
	}

	// Check if package is already in config
	if isPackageInConfig(cfg, packageName, targetManager) {
		if format == OutputTable {
			fmt.Printf("Package '%s' is already in %s configuration\n", packageName, targetManager)
		}
		return nil
	}

	// Add package to configuration
	err = addPackageToConfig(cfg, packageName, targetManager)
	if err != nil {
		return fmt.Errorf("failed to add package to config: %w", err)
	}

	// Save updated configuration
	err = saveConfig(cfg, configDir)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Install the package
	packageManagers := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"npm":      managers.NewNpmManager(),
	}

	mgr := packageManagers[targetManager]
	ctx := context.Background()
	available, err := mgr.IsAvailable(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if manager '%s' is available: %w", targetManager, err)
	}
	if !available {
		return fmt.Errorf("manager '%s' is not available", targetManager)
	}

	// Check if already installed
	installed, err := mgr.IsInstalled(ctx, packageName)
	if err != nil {
		return fmt.Errorf("failed to check if package is installed: %w", err)
	}
	if installed {
		if format == OutputTable {
			fmt.Printf("Package '%s' is already installed in %s\n", packageName, targetManager)
			fmt.Printf("Added to configuration: %s\n", packageName)
		}
	} else {
		// Install the package
		if format == OutputTable {
			fmt.Printf("Installing %s using %s...\n", packageName, targetManager)
		}
		
		err = mgr.Install(ctx, packageName)
		if err != nil {
			return fmt.Errorf("failed to install package: %w", err)
		}

		if format == OutputTable {
			fmt.Printf("Successfully installed and added to configuration: %s\n", packageName)
		}
	}

	// Prepare structured output
	result := AddOutput{
		Package: packageName,
		Manager: targetManager,
		Action:  "added",
	}

	return RenderOutput(result, format)
}

// isPackageInConfig checks if a package is already in the configuration
func isPackageInConfig(cfg *config.Config, packageName, targetManager string) bool {
	switch targetManager {
	case "homebrew":
		for _, brew := range cfg.Homebrew.Brews {
			if brew.Name == packageName {
				return true
			}
		}
		for _, cask := range cfg.Homebrew.Casks {
			if cask.Name == packageName {
				return true
			}
		}
	case "npm":
		for _, pkg := range cfg.NPM {
			if pkg.Name == packageName || pkg.Package == packageName {
				return true
			}
		}
	}
	return false
}

// addPackageToConfig adds a package to the appropriate section of the configuration
func addPackageToConfig(cfg *config.Config, packageName, targetManager string) error {
	switch targetManager {
	case "homebrew":
		// Add to brews section (we could add logic to detect casks vs brews later)
		newPackage := config.HomebrewPackage{
			Name: packageName,
		}
		cfg.Homebrew.Brews = append(cfg.Homebrew.Brews, newPackage)
	case "npm":
		newPackage := config.NPMPackage{
			Name: packageName,
		}
		cfg.NPM = append(cfg.NPM, newPackage)
	default:
		return fmt.Errorf("unsupported manager: %s", targetManager)
	}
	return nil
}

// saveConfig saves the configuration back to plonk.yaml
func saveConfig(cfg *config.Config, configDir string) error {
	configPath := filepath.Join(configDir, "plonk.yaml")
	
	// Create config directory if it doesn't exist
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal configuration to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddOutput represents the output structure for pkg add command
type AddOutput struct {
	Package string `json:"package" yaml:"package"`
	Manager string `json:"manager" yaml:"manager"`
	Action  string `json:"action" yaml:"action"`
}

// TableOutput generates human-friendly table output for add command
func (a AddOutput) TableOutput() string {
	return "" // Table output is handled in the command logic
}

// StructuredData returns the structured data for serialization
func (a AddOutput) StructuredData() any {
	return a
}