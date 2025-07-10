// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"plonk/internal/config"
	"plonk/internal/errors"
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
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "pkg-remove", "output-format", "invalid output format")
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Load existing configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Initialize result structure
	result := EnhancedRemoveOutput{
		Package: packageName,
		Actions: []string{},
	}

	// Find and remove package from configuration
	managerName, found := findAndRemovePackageFromConfig(cfg, packageName)
	result.Manager = managerName

	if !found {
		result.WasInConfig = false
		result.Actions = append(result.Actions, fmt.Sprintf("%s not found in configuration", packageName))
		return RenderOutput(result, format)
	}

	result.WasInConfig = true

	// Save updated configuration
	err = saveConfig(cfg, configDir)
	if err != nil {
		result.Error = fmt.Sprintf("failed to save configuration: %v", err)
		return RenderOutput(result, format)
	}

	result.ConfigRemoved = true
	result.Actions = append(result.Actions, fmt.Sprintf("Removed %s from %s configuration", packageName, managerName))

	// Optionally uninstall the package
	if uninstall {
		packageManagers := map[string]managers.PackageManager{
			"homebrew": managers.NewHomebrewManager(),
			"npm":      managers.NewNpmManager(),
		}

		mgr := packageManagers[managerName]
		ctx := context.Background()
		available, err := mgr.IsAvailable(ctx)
		if err != nil {
			result.Error = fmt.Sprintf("failed to check if %s manager is available: %v", managerName, err)
			return RenderOutput(result, format)
		}
		if !available {
			result.Error = fmt.Sprintf("manager '%s' is not available", managerName)
			return RenderOutput(result, format)
		}

		installed, err := mgr.IsInstalled(ctx, packageName)
		if err != nil {
			result.Error = fmt.Sprintf("failed to check if package is installed: %v", err)
			return RenderOutput(result, format)
		}

		result.WasInstalled = installed

		if !installed {
			result.Actions = append(result.Actions, fmt.Sprintf("%s was not installed", packageName))
		} else {
			err = mgr.Uninstall(ctx, packageName)
			if err != nil {
				result.Error = fmt.Sprintf("failed to uninstall package: %v", err)
				return RenderOutput(result, format)
			}

			result.Uninstalled = true
			result.Actions = append(result.Actions, fmt.Sprintf("Successfully uninstalled %s from system", packageName))
		}
	}

	return RenderOutput(result, format)
}

// findAndRemovePackageFromConfig finds and removes a package from the configuration
// Returns the manager name and whether the package was found
func findAndRemovePackageFromConfig(cfg *config.Config, packageName string) (string, bool) {
	// Check homebrew packages
	for i, pkg := range cfg.Homebrew {
		if pkg.Name == packageName {
			cfg.Homebrew = append(cfg.Homebrew[:i], cfg.Homebrew[i+1:]...)
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
