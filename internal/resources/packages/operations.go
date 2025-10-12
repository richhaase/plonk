// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources"
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
func InstallPackages(ctx context.Context, configDir string, packages []string, opts InstallOptions) ([]resources.OperationResult, error) {
	// Thin wrapper: resolve defaults and delegate to InstallPackagesWith
	cfg := config.LoadWithDefaults(configDir)
	lockService := lock.NewYAMLLockService(configDir)
	registry := NewManagerRegistry()
	return InstallPackagesWith(ctx, cfg, lockService, registry, packages, opts)
}

// InstallPackagesWith orchestrates installation with explicit dependencies
func InstallPackagesWith(ctx context.Context, cfg *config.Config, lockService lock.LockService, registry *ManagerRegistry, packages []string, opts InstallOptions) ([]resources.OperationResult, error) {
	// Load manager configs from plonk.yaml before any operations
	if registry != nil {
		registry.LoadV2Configs(cfg)
	}

	// Get manager - use default if not specified
	manager := opts.Manager
	if manager == "" {
		if cfg != nil && cfg.DefaultManager != "" {
			manager = cfg.DefaultManager
		} else {
			manager = DefaultManager // fallback default
		}
	}

	var results []resources.OperationResult
	for _, packageName := range packages {
		if ctx.Err() != nil {
			break
		}
		result := installSinglePackage(ctx, lockService, packageName, manager, opts.DryRun, registry)
		results = append(results, result)
	}
	return results, nil
}

// UninstallPackages orchestrates the uninstallation of multiple packages
func UninstallPackages(ctx context.Context, configDir string, packages []string, opts UninstallOptions) ([]resources.OperationResult, error) {
	// Thin wrapper: resolve defaults and delegate to UninstallPackagesWith
	cfg := config.LoadWithDefaults(configDir)
	lockService := lock.NewYAMLLockService(configDir)
	registry := NewManagerRegistry()
	return UninstallPackagesWith(ctx, cfg, lockService, registry, packages, opts)
}

// UninstallPackagesWith orchestrates uninstallation with explicit dependencies
func UninstallPackagesWith(ctx context.Context, cfg *config.Config, lockService lock.LockService, registry *ManagerRegistry, packages []string, opts UninstallOptions) ([]resources.OperationResult, error) {
	// Load manager configs from plonk.yaml before any operations
	if registry != nil {
		registry.LoadV2Configs(cfg)
	}

	// Load config for default manager
	defaultManager := DefaultManager
	if cfg != nil && cfg.DefaultManager != "" {
		defaultManager = cfg.DefaultManager
	}

	var results []resources.OperationResult
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
func installSinglePackage(ctx context.Context, lockService lock.LockService, packageName, manager string, dryRun bool, registry *ManagerRegistry) resources.OperationResult {
	result := resources.OperationResult{
		Name:    packageName,
		Manager: manager,
	}

	// For Go packages, we need to check with the binary name
	checkPackageName := packageName
	if manager == "go" {
		checkPackageName = ExtractBinaryNameFromPath(packageName)
	}

	// Check if already managed
	if lockService.HasPackage(manager, checkPackageName) {
		result.Status = "skipped"
		result.AlreadyManaged = true
		return result
	}

	if dryRun {
		result.Status = "would-add"
		return result
	}

	// Get package manager instance
	pkgManager, err := getPackageManager(registry, manager)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("install %s: failed to get package manager: %w", packageName, err)
		return result
	}

	// Check if manager is available
	available, err := pkgManager.IsAvailable(ctx)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("install %s: failed to check %s availability: %w", packageName, manager, err)
		return result
	}
	if !available {
		result.Status = "failed"
		result.Error = fmt.Errorf("install %s: %s manager not available (%s)", packageName, manager, getManagerInstallSuggestion(manager))
		return result
	}

	// Install the package (relies on manager's idempotency)
	err = pkgManager.Install(ctx, packageName)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("install %s via %s: %w", packageName, manager, err)
		return result
	}

	// For Go packages, we need to determine the actual binary name
	lockPackageName := packageName
	if manager == "go" {
		// Extract binary name from module path
		lockPackageName = ExtractBinaryNameFromPath(packageName)
	}

	// Note: InstalledVersion() method has been removed from all managers
	// Version tracking is no longer supported
	version := ""

	// Create metadata for the package
	metadata := map[string]interface{}{
		"manager": manager,
		"name":    lockPackageName,
		"version": version,
	}

	// Add source path for Go packages
	if manager == "go" {
		metadata["source_path"] = packageName
	}

	// Handle npm scoped packages
	if manager == "npm" && strings.HasPrefix(packageName, "@") {
		// Extract scope from scoped package name
		parts := strings.SplitN(packageName, "/", 2)
		if len(parts) == 2 {
			metadata["scope"] = parts[0]
			metadata["full_name"] = packageName
		}
	}

	// Add to lock file
	err = lockService.AddPackage(manager, lockPackageName, version, metadata)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("install %s: failed to add to lock file (manager: %s, version: %s): %w", packageName, manager, version, err)
		return result
	}

	result.Status = "added"
	return result
}

// uninstallSinglePackage uninstalls a single package
func uninstallSinglePackage(ctx context.Context, lockService lock.LockService, packageName, manager string, dryRun bool, registry *ManagerRegistry) resources.OperationResult {
	result := resources.OperationResult{
		Name:    packageName,
		Manager: manager,
	}

	// For Go packages, we need to check with the binary name
	checkPackageName := packageName
	if manager == "go" {
		checkPackageName = ExtractBinaryNameFromPath(packageName)
	}

	// Check if package is managed
	isManaged := lockService.HasPackage(manager, checkPackageName)

	if dryRun {
		result.Status = "would-remove"
		return result
	}

	// Get package manager instance
	pkgManager, err := getPackageManager(registry, manager)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("uninstall %s: failed to get package manager: %w", packageName, err)
		return result
	}

	// Check if manager is available
	available, err := pkgManager.IsAvailable(ctx)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("uninstall %s: failed to check %s availability: %w", packageName, manager, err)
		return result
	}
	if !available {
		result.Status = "failed"
		result.Error = fmt.Errorf("uninstall %s: %s manager not available", packageName, manager)
		return result
	}

	// Uninstall the package
	uninstallErr := pkgManager.Uninstall(ctx, packageName)

	// If package is managed, remove from lock file
	if isManaged {
		lockErr := lockService.RemovePackage(manager, checkPackageName)
		if lockErr != nil {
			// If we removed from system but failed to update lock, still partial success
			if uninstallErr == nil {
				result.Status = "removed"
				result.Error = fmt.Errorf("package uninstalled but failed to update lock file: %w", lockErr)
			} else {
				result.Status = "failed"
				result.Error = fmt.Errorf("uninstall failed and couldn't update lock: %w", uninstallErr)
			}
			return result
		}

		// Successfully removed from lock file
		if uninstallErr != nil {
			// Removed from lock but system uninstall failed - this is still success per spec
			result.Status = "removed"
			result.Error = fmt.Errorf("removed from plonk management (system uninstall failed: %w)", uninstallErr)
		} else {
			// Both succeeded
			result.Status = "removed"
		}
	} else {
		// Package not managed - pure pass-through
		if uninstallErr != nil {
			result.Status = "failed"
			result.Error = fmt.Errorf("uninstall %s via %s: %w", packageName, manager, uninstallErr)
		} else {
			result.Status = "removed"
		}
	}

	return result
}

// getPackageManager returns the appropriate package manager instance
func getPackageManager(registry *ManagerRegistry, manager string) (PackageManager, error) {
	if registry == nil {
		registry = NewManagerRegistry()
	}
	return registry.GetManager(manager)
}

// getManagerInstallSuggestion returns installation suggestion for a manager
func getManagerInstallSuggestion(manager string) string {
	// All supported package managers are cross-platform
	suggestions := map[string]string{
		"brew":  "install with: /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"",
		"npm":   "install Node.js from https://nodejs.org/",
		"cargo": "install Rust from https://rustup.rs/",
		"gem":   "install Ruby from https://ruby-lang.org/",
		"go":    "install Go from https://golang.org/",
		"uv":    "install UV from https://docs.astral.sh/uv/",
	}
	if suggestion, ok := suggestions[manager]; ok {
		return suggestion
	}
	return "check installation instructions for " + manager
}
