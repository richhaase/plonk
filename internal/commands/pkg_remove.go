// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/managers"

	"github.com/spf13/cobra"
)

var (
	uninstall bool
)

var pkgRemoveCmd = &cobra.Command{
	Use:   "remove <package>",
	Short: "Remove a package from plonk configuration",
	Long: `Remove a package from your plonk.yaml configuration.

By default, this only removes the package from your configuration file,
leaving the actual package installed on your system.

Use the --uninstall flag to also uninstall the package from your system.

Examples:
  plonk pkg remove htop                 # Remove from config only
  plonk pkg remove htop --uninstall     # Remove from config and uninstall`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgRemove,
}

func init() {
	pkgCmd.AddCommand(pkgRemoveCmd)
	pkgRemoveCmd.Flags().BoolVar(&uninstall, "uninstall", false, "Also uninstall the package from the system")
}

func runPkgRemove(cmd *cobra.Command, args []string) error {
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

	// Find and remove package from configuration
	managerName, found := findAndRemovePackageFromConfig(cfg, packageName)
	if !found {
		if format == OutputTable {
			fmt.Printf("Package '%s' not found in configuration\n", packageName)
		}
		return nil
	}

	// Save updated configuration
	err = saveConfig(cfg, configDir)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	var action string
	var uninstallError error

	// Optionally uninstall the package
	if uninstall {
		packageManagers := map[string]managers.PackageManager{
			"homebrew": managers.NewHomebrewManager(),
			"npm":      managers.NewNpmManager(),
		}

		mgr := packageManagers[managerName]
		if !mgr.IsAvailable() {
			uninstallError = fmt.Errorf("manager '%s' is not available", managerName)
		} else if !mgr.IsInstalled(packageName) {
			if format == OutputTable {
				fmt.Printf("Package '%s' is not installed in %s\n", packageName, managerName)
			}
			action = "removed_from_config_only"
		} else {
			if format == OutputTable {
				fmt.Printf("Uninstalling %s from %s...\n", packageName, managerName)
			}
			
			err = mgr.Uninstall(packageName)
			if err != nil {
				uninstallError = err
				action = "removed_from_config_uninstall_failed"
			} else {
				if format == OutputTable {
					fmt.Printf("Successfully removed from configuration and uninstalled: %s\n", packageName)
				}
				action = "removed_and_uninstalled"
			}
		}
	} else {
		if format == OutputTable {
			fmt.Printf("Removed from configuration: %s\n", packageName)
		}
		action = "removed_from_config_only"
	}

	// Prepare structured output
	result := RemoveOutput{
		Package: packageName,
		Manager: managerName,
		Action:  action,
	}

	if uninstallError != nil {
		result.Error = uninstallError.Error()
		if format == OutputTable {
			fmt.Printf("Warning: Failed to uninstall package: %v\n", uninstallError)
		}
	}

	return RenderOutput(result, format)
}

// findAndRemovePackageFromConfig finds and removes a package from the configuration
// Returns the manager name and whether the package was found
func findAndRemovePackageFromConfig(cfg *config.Config, packageName string) (string, bool) {
	// Check homebrew brews
	for i, brew := range cfg.Homebrew.Brews {
		if brew.Name == packageName {
			cfg.Homebrew.Brews = append(cfg.Homebrew.Brews[:i], cfg.Homebrew.Brews[i+1:]...)
			return "homebrew", true
		}
	}

	// Check homebrew casks
	for i, cask := range cfg.Homebrew.Casks {
		if cask.Name == packageName {
			cfg.Homebrew.Casks = append(cfg.Homebrew.Casks[:i], cfg.Homebrew.Casks[i+1:]...)
			return "homebrew", true
		}
	}

	// Check npm packages
	for i, pkg := range cfg.NPM {
		if pkg.Name == packageName || pkg.Package == packageName {
			cfg.NPM = append(cfg.NPM[:i], cfg.NPM[i+1:]...)
			return "npm", true
		}
	}

	return "", false
}

// RemoveOutput represents the output structure for pkg remove command
type RemoveOutput struct {
	Package string `json:"package" yaml:"package"`
	Manager string `json:"manager" yaml:"manager"`
	Action  string `json:"action" yaml:"action"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}

// TableOutput generates human-friendly table output for remove command
func (r RemoveOutput) TableOutput() string {
	return "" // Table output is handled in the command logic
}

// StructuredData returns the structured data for serialization
func (r RemoveOutput) StructuredData() any {
	return r
}