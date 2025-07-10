// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"plonk/internal/config"
	"plonk/internal/errors"
	"plonk/internal/lock"
	"plonk/internal/managers"
	"plonk/internal/state"

	"github.com/spf13/cobra"
)

var (
	manager string
)

var pkgAddCmd = &cobra.Command{
	Use:   "add [package]",
	Short: "Add package(s) to plonk configuration and install them",
	Long: `Add one or more packages to your plonk.lock file and install them.

With package name:
  plonk pkg add htop              # Add htop using default manager
  plonk pkg add git --manager homebrew  # Add git specifically to homebrew
  plonk pkg add lodash --manager npm     # Add lodash to npm global packages
  plonk pkg add ripgrep --manager cargo  # Add ripgrep to cargo packages

Without arguments:
  plonk pkg add                   # Add all untracked packages
  plonk pkg add --dry-run         # Preview what would be added`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPkgAdd,
}

func init() {
	pkgCmd.AddCommand(pkgAddCmd)
	pkgAddCmd.Flags().StringVar(&manager, "manager", "", "Package manager to use (homebrew|npm|cargo)")
	pkgAddCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")
}

func runPkgAdd(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if len(args) == 0 {
		// No package specified - add all untracked packages
		return addAllUntrackedPackages(cmd, dryRun)
	}

	packageName := args[0]

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "pkg-add", "output-format", "invalid output format")
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Load config for default manager
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)

	// Determine which manager to use
	targetManager := manager
	if targetManager == "" {
		targetManager = cfg.Settings.DefaultManager
	}

	// Validate manager
	if targetManager != "homebrew" && targetManager != "npm" && targetManager != "cargo" {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainPackages, "validate", fmt.Sprintf("unsupported manager '%s'. Use: homebrew, npm, cargo", targetManager))
	}

	// Initialize result structure
	result := EnhancedAddOutput{
		Package: packageName,
		Manager: targetManager,
		Actions: []string{},
	}

	// Check if package is already in lock file
	alreadyInLock := lockService.HasPackage(targetManager, packageName)
	if alreadyInLock {
		result.AlreadyInConfig = true
		result.Actions = append(result.Actions, fmt.Sprintf("%s already managed by %s", packageName, targetManager))
	} else {
		// Get package manager for version detection
		packageManagers := map[string]managers.PackageManager{
			"homebrew": managers.NewHomebrewManager(),
			"npm":      managers.NewNpmManager(),
			"cargo":    managers.NewCargoManager(),
		}

		mgr := packageManagers[targetManager]
		ctx := context.Background()
		available, err := mgr.IsAvailable(ctx)
		if err != nil {
			result.Error = fmt.Sprintf("failed to check if %s manager is available: %v", targetManager, err)
			return RenderOutput(result, format)
		}
		if !available {
			result.Error = fmt.Sprintf("manager '%s' is not available", targetManager)
			return RenderOutput(result, format)
		}

		// Install package first (so we can get version info)
		installed, err := mgr.IsInstalled(ctx, packageName)
		if err != nil {
			result.Error = fmt.Sprintf("failed to check if package is installed: %v", err)
			return RenderOutput(result, format)
		}

		if !installed {
			if !dryRun {
				err = mgr.Install(ctx, packageName)
				if err != nil {
					result.Error = fmt.Sprintf("failed to install package: %v", err)
					return RenderOutput(result, format)
				}
				result.Installed = true
				result.Actions = append(result.Actions, fmt.Sprintf("Successfully installed %s", packageName))
			} else {
				result.Actions = append(result.Actions, fmt.Sprintf("Would install %s", packageName))
			}
		} else {
			result.AlreadyInstalled = true
			result.Actions = append(result.Actions, fmt.Sprintf("%s already installed", packageName))
		}

		// Add package to lock file (with basic version info)
		if !dryRun {
			// Use "latest" as version for now - we could enhance this later
			err = lockService.AddPackage(targetManager, packageName, "latest")
			if err != nil {
				result.Error = fmt.Sprintf("failed to add package to lock file: %v", err)
				return RenderOutput(result, format)
			}
			result.ConfigAdded = true
			result.Actions = append(result.Actions, fmt.Sprintf("Added %s to lock file", packageName))
		} else {
			result.Actions = append(result.Actions, fmt.Sprintf("Would add %s to lock file", packageName))
		}
	}

	return RenderOutput(result, format)
}

// addAllUntrackedPackages adds all untracked packages to the lock file
func addAllUntrackedPackages(cmd *cobra.Command, dryRun bool) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "pkg-add-all", "output-format", "invalid output format")
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	// Create reconciler to get untracked packages
	reconciler := state.NewReconciler()

	// Create package provider using lock file adapter
	ctx := context.Background()
	packageProvider := state.NewMultiManagerPackageProvider()

	// Add all available managers
	managers := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"npm":      managers.NewNpmManager(),
		"cargo":    managers.NewCargoManager(),
	}

	for managerName, mgr := range managers {
		available, err := mgr.IsAvailable(ctx)
		if err != nil {
			return fmt.Errorf("failed to check %s availability: %w", managerName, err)
		}
		if available {
			managerAdapter := state.NewManagerAdapter(mgr)
			packageProvider.AddManager(managerName, managerAdapter, lockAdapter)
		}
	}

	reconciler.RegisterProvider("package", packageProvider)

	// Reconcile to get package states
	result, err := reconciler.ReconcileProvider(ctx, "package")
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile package states")
	}

	untrackedPackages := result.Untracked

	if len(untrackedPackages) == 0 {
		if format == OutputTable {
			fmt.Println("No untracked packages found")
		}
		return nil
	}

	if dryRun {
		if format == OutputTable {
			fmt.Printf("Would add %d untracked packages:\n\n", len(untrackedPackages))
			for _, pkg := range untrackedPackages {
				fmt.Printf("  %s (%s)\n", pkg.Name, pkg.Manager)
			}
		}
		return nil
	}

	// Add packages to lock file
	addedCount := 0
	for _, pkg := range untrackedPackages {
		if !lockService.HasPackage(pkg.Manager, pkg.Name) {
			err = lockService.AddPackage(pkg.Manager, pkg.Name, "latest")
			if err != nil {
				return errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainCommands, "update", pkg.Name, "failed to add package to lock file")
			}
			addedCount++
		}
	}

	if addedCount == 0 {
		if format == OutputTable {
			fmt.Println("No packages were added (all were already managed)")
		}
		return nil
	}

	if format == OutputTable {
		fmt.Printf("Successfully added %d packages to lock file\n", addedCount)
	}

	// Prepare structured output
	addAllResult := AddAllOutput{
		Added:  addedCount,
		Total:  len(untrackedPackages),
		Action: "added-all",
	}

	return RenderOutput(addAllResult, format)
}
