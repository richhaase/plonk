// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

// InstallOptions configures package installation operations
type InstallOptions struct {
	Manager string
	DryRun  bool
}

// UninstallOptions configures package uninstallation operations
type UninstallOptions struct {
	Manager string
	DryRun  bool
}

// InstallPackages orchestrates the installation of multiple packages
func InstallPackages(ctx context.Context, configDir string, packages []string, opts InstallOptions) ([]InstallResult, error) {
	// Thin wrapper: resolve defaults and delegate to InstallPackagesWith
	cfg := config.LoadWithDefaults(configDir)
	lockService := lock.NewYAMLLockService(configDir)
	registry := GetRegistry()
	return InstallPackagesWith(ctx, cfg, lockService, registry, packages, opts)
}

// InstallPackagesWith orchestrates installation with explicit dependencies
func InstallPackagesWith(ctx context.Context, cfg *config.Config, lockService lock.LockService, registry *ManagerRegistry, packages []string, opts InstallOptions) ([]InstallResult, error) {
	// Get manager - use default if not specified
	manager := opts.Manager
	if manager == "" {
		if cfg != nil && cfg.DefaultManager != "" {
			manager = cfg.DefaultManager
		} else {
			manager = DefaultManager // fallback default
		}
	}

	var results []InstallResult
	for _, packageName := range packages {
		if ctx.Err() != nil {
			break
		}
		result := installSinglePackage(ctx, cfg, lockService, packageName, manager, opts.DryRun, registry)
		results = append(results, result)
	}
	return results, nil
}

// UninstallPackages orchestrates the uninstallation of multiple packages
func UninstallPackages(ctx context.Context, configDir string, packages []string, opts UninstallOptions) ([]UninstallResult, error) {
	// Thin wrapper: resolve defaults and delegate to UninstallPackagesWith
	cfg := config.LoadWithDefaults(configDir)
	lockService := lock.NewYAMLLockService(configDir)
	registry := GetRegistry()
	return UninstallPackagesWith(ctx, cfg, lockService, registry, packages, opts)
}

// UninstallPackagesWith orchestrates uninstallation with explicit dependencies
func UninstallPackagesWith(ctx context.Context, cfg *config.Config, lockService lock.LockService, registry *ManagerRegistry, packages []string, opts UninstallOptions) ([]UninstallResult, error) {
	// Load config for default manager
	defaultManager := DefaultManager
	if cfg != nil && cfg.DefaultManager != "" {
		defaultManager = cfg.DefaultManager
	}

	var results []UninstallResult
	for _, packageName := range packages {
		if ctx.Err() != nil {
			break
		}

		// Determine which manager to use
		manager := opts.Manager
		if manager == "" {
			// Check if package exists in lock file first
			locations := lockService.FindPackage(packageName)
			if len(locations) > 0 {
				if mgr, ok := locations[0].Metadata["manager"].(string); ok {
					manager = mgr
				} else {
					manager = defaultManager
				}
			} else {
				manager = defaultManager
			}
		}

		result := uninstallSinglePackage(ctx, lockService, packageName, manager, opts.DryRun, registry)
		results = append(results, result)
	}
	return results, nil
}

// installSinglePackage installs a single package
func installSinglePackage(ctx context.Context, cfg *config.Config, lockService lock.LockService, packageName, manager string, dryRun bool, registry *ManagerRegistry) InstallResult {
	result := InstallResult{
		Name:    packageName,
		Manager: manager,
	}

	// Check if already managed
	if lockService.HasPackage(manager, packageName) {
		result.Status = InstallStatusSkipped
		result.AlreadyManaged = true
		return result
	}

	if dryRun {
		result.Status = InstallStatusWouldAdd
		return result
	}

	// Get package manager instance
	pkgManager, err := getPackageManager(registry, manager)
	if err != nil {
		result.Status = InstallStatusFailed
		result.Error = fmt.Errorf("install %s: failed to get package manager: %w", packageName, err)
		return result
	}

	// Check if manager is available
	available, err := pkgManager.IsAvailable(ctx)
	if err != nil {
		result.Status = InstallStatusFailed
		result.Error = fmt.Errorf("install %s: failed to check %s availability: %w", packageName, manager, err)
		return result
	}
	if !available {
		result.Status = InstallStatusFailed
		result.Error = fmt.Errorf("install %s: %s manager not available (%s)", packageName, manager, managerInstallHint(manager))
		return result
	}

	// Install the package (relies on manager's idempotency)
	err = pkgManager.Install(ctx, packageName)
	if err != nil {
		result.Status = InstallStatusFailed
		result.Error = fmt.Errorf("install %s via %s: %w", packageName, manager, err)
		return result
	}

	// Create metadata for the package
	metadata := map[string]interface{}{
		"manager": manager,
		"name":    packageName,
	}

	// Add to lock file
	err = lockService.AddPackage(manager, packageName, "", metadata)
	if err != nil {
		result.Status = InstallStatusFailed
		result.Error = fmt.Errorf("install %s: failed to add to lock file (manager: %s): %w", packageName, manager, err)
		return result
	}

	result.Status = InstallStatusAdded
	return result
}

// uninstallSinglePackage uninstalls a single package
func uninstallSinglePackage(ctx context.Context, lockService lock.LockService, packageName, manager string, dryRun bool, registry *ManagerRegistry) UninstallResult {
	result := UninstallResult{
		Name:    packageName,
		Manager: manager,
	}

	isManaged := lockService.HasPackage(manager, packageName)

	if dryRun {
		result.Status = UninstallStatusWouldRemove
		return result
	}

	// Get and validate package manager
	pkgManager, err := getPackageManager(registry, manager)
	if err != nil {
		result.Status = UninstallStatusFailed
		result.Error = fmt.Errorf("uninstall %s: failed to get package manager: %w", packageName, err)
		return result
	}

	available, err := pkgManager.IsAvailable(ctx)
	if err != nil {
		result.Status = UninstallStatusFailed
		result.Error = fmt.Errorf("uninstall %s: failed to check %s availability: %w", packageName, manager, err)
		return result
	}
	if !available {
		result.Status = UninstallStatusFailed
		result.Error = fmt.Errorf("uninstall %s: %s manager not available (%s)", packageName, manager, managerInstallHint(manager))
		return result
	}

	// Perform uninstall and update lock file if managed
	uninstallErr := pkgManager.Uninstall(ctx, packageName)
	var lockErr error
	if isManaged {
		lockErr = lockService.RemovePackage(manager, packageName)
	}

	// Resolve final outcome
	result.Status, result.Error = resolveUninstallOutcome(packageName, manager, isManaged, uninstallErr, lockErr)
	return result
}

// resolveUninstallOutcome determines the final status and error based on the combination
// of isManaged, uninstallErr, and lockErr. This encapsulates the complex decision logic.
func resolveUninstallOutcome(packageName, manager string, isManaged bool, uninstallErr, lockErr error) (UninstallStatus, error) {
	// Unmanaged package - simple pass-through to system
	if !isManaged {
		if uninstallErr != nil {
			return UninstallStatusFailed, fmt.Errorf("uninstall %s via %s: %w", packageName, manager, uninstallErr)
		}
		return UninstallStatusRemoved, nil
	}

	// Managed package - handle lock file update outcomes
	if lockErr != nil {
		if uninstallErr == nil {
			// System uninstall succeeded but lock update failed - partial success
			return UninstallStatusRemoved, fmt.Errorf("package uninstalled but failed to update lock file: %w", lockErr)
		}
		// Both failed
		return UninstallStatusFailed, fmt.Errorf("uninstall failed and couldn't update lock: %w", uninstallErr)
	}

	// Lock file updated successfully
	if uninstallErr != nil {
		// Removed from plonk but system uninstall failed - still success per spec
		return UninstallStatusRemoved, fmt.Errorf("removed from plonk management (system uninstall failed: %w)", uninstallErr)
	}

	// Both succeeded
	return UninstallStatusRemoved, nil
}

// getPackageManager returns the appropriate package manager instance
func getPackageManager(registry *ManagerRegistry, manager string) (PackageManager, error) {
	if registry == nil {
		registry = GetRegistry()
	}
	return registry.GetManager(manager)
}
