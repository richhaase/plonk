// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/spf13/cobra"
)

var (
	uninstall bool
)

var pkgRemoveCmd = &cobra.Command{
	Use:   "remove <package>",
	Short: "Remove a package from plonk configuration",
	Long: `Remove a package from your plonk.lock file.

By default, this only removes the package from your lock file,
leaving the actual package installed on your system.

Use the --uninstall flag to also uninstall the package from your system.

Examples:
  plonk pkg remove htop                 # Remove from lock file only
  plonk pkg remove htop --uninstall     # Remove from lock file and uninstall
  plonk pkg remove htop --dry-run       # Preview what would be removed`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgRemove,
}

func init() {
	pkgCmd.AddCommand(pkgRemoveCmd)
	pkgRemoveCmd.Flags().BoolVar(&uninstall, "uninstall", false, "Also uninstall the package from the system")
	pkgRemoveCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")
}

func runPkgRemove(cmd *cobra.Command, args []string) error {
	packageName := args[0]
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "pkg-remove", "output-format", "invalid output format")
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)

	// Initialize result structure
	result := EnhancedRemoveOutput{
		Package: packageName,
		Actions: []string{},
	}

	// Find package in lock file
	managerName, found := findPackageInLockFile(lockService, packageName)
	result.Manager = managerName

	if !found {
		result.WasInConfig = false
		result.Actions = append(result.Actions, fmt.Sprintf("%s not found in lock file", packageName))
		return RenderOutput(result, format)
	}

	result.WasInConfig = true

	// Remove package from lock file
	if !dryRun {
		err = lockService.RemovePackage(managerName, packageName)
		if err != nil {
			result.Error = fmt.Sprintf("failed to remove package from lock file: %v", err)
			return RenderOutput(result, format)
		}
		result.ConfigRemoved = true
		result.Actions = append(result.Actions, fmt.Sprintf("Removed %s from %s lock file", packageName, managerName))
	} else {
		result.Actions = append(result.Actions, fmt.Sprintf("Would remove %s from %s lock file", packageName, managerName))
	}

	// Optionally uninstall the package
	if uninstall {
		packageManagers := map[string]managers.PackageManager{
			"homebrew": managers.NewHomebrewManager(),
			"npm":      managers.NewNpmManager(),
			"cargo":    managers.NewCargoManager(),
		}

		mgr := packageManagers[managerName]
		if mgr == nil {
			result.Error = fmt.Sprintf("unsupported manager: %s", managerName)
			return RenderOutput(result, format)
		}

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
			if !dryRun {
				err = mgr.Uninstall(ctx, packageName)
				if err != nil {
					result.Error = fmt.Sprintf("failed to uninstall package: %v", err)
					return RenderOutput(result, format)
				}
				result.Uninstalled = true
				result.Actions = append(result.Actions, fmt.Sprintf("Successfully uninstalled %s from system", packageName))
			} else {
				result.Actions = append(result.Actions, fmt.Sprintf("Would uninstall %s from system", packageName))
			}
		}
	}

	return RenderOutput(result, format)
}

// findPackageInLockFile finds a package in the lock file
// Returns the manager name and whether the package was found
func findPackageInLockFile(lockService lock.LockService, packageName string) (string, bool) {
	// Check all supported managers
	managers := []string{"homebrew", "npm", "cargo"}

	for _, managerName := range managers {
		if lockService.HasPackage(managerName, packageName) {
			return managerName, true
		}
	}

	return "", false
}
