// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package services

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/state"
)

// PackageApplyOptions configures package apply operations
type PackageApplyOptions struct {
	ConfigDir string
	Config    *config.Config
	DryRun    bool
}

// PackageApplyResult represents the result of package apply operations
type PackageApplyResult struct {
	DryRun            bool                 `json:"dry_run" yaml:"dry_run"`
	TotalMissing      int                  `json:"total_missing" yaml:"total_missing"`
	TotalInstalled    int                  `json:"total_installed" yaml:"total_installed"`
	TotalFailed       int                  `json:"total_failed" yaml:"total_failed"`
	TotalWouldInstall int                  `json:"total_would_install" yaml:"total_would_install"`
	Managers          []ManagerApplyResult `json:"managers" yaml:"managers"`
}

// ManagerApplyResult represents the result for a specific manager
type ManagerApplyResult struct {
	Name         string          `json:"name" yaml:"name"`
	MissingCount int             `json:"missing_count" yaml:"missing_count"`
	Packages     []PackageResult `json:"packages" yaml:"packages"`
}

// PackageResult represents the result for a specific package
type PackageResult struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// ApplyPackages applies package configuration and returns the result
func ApplyPackages(ctx context.Context, options PackageApplyOptions) (PackageApplyResult, error) {
	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Create lock file adapter
	lockService := lock.NewYAMLLockService(options.ConfigDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	// Create package provider using registry
	registry := managers.NewManagerRegistry()
	packageProvider, err := registry.CreateMultiProvider(ctx, lockAdapter)
	if err != nil {
		return PackageApplyResult{}, errors.Wrap(err, errors.ErrProviderNotFound, errors.DomainPackages, "apply",
			"failed to create package provider")
	}

	reconciler.RegisterProvider("package", packageProvider)

	// Reconcile package domain to find missing packages
	result, err := reconciler.ReconcileProvider(ctx, "package")
	if err != nil {
		return PackageApplyResult{}, errors.Wrap(err, errors.ErrReconciliation, errors.DomainPackages, "reconcile", "failed to reconcile package state")
	}

	// Group missing packages by manager
	missingByManager := make(map[string][]state.Item)
	for _, item := range result.Missing {
		manager := item.Metadata["manager"].(string)
		missingByManager[manager] = append(missingByManager[manager], item)
	}

	var managerResults []ManagerApplyResult
	totalMissing := len(result.Missing)
	totalInstalled := 0
	totalFailed := 0
	totalWouldInstall := 0

	// Process each manager's missing packages
	for managerName, packages := range missingByManager {
		manager, err := registry.GetManager(managerName)
		if err != nil {
			return PackageApplyResult{}, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "apply",
				"failed to get package manager")
		}

		available, err := manager.IsAvailable(ctx)
		if err != nil {
			return PackageApplyResult{}, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "apply",
				"failed to check manager availability")
		}

		var packageResults []PackageResult
		installedCount := 0
		failedCount := 0
		wouldInstallCount := 0

		for _, pkg := range packages {
			if options.DryRun {
				packageResults = append(packageResults, PackageResult{
					Name:   pkg.Name,
					Status: "would-install",
				})
				wouldInstallCount++
			} else if !available {
				packageResults = append(packageResults, PackageResult{
					Name:   pkg.Name,
					Status: "failed",
					Error:  "package manager not available",
				})
				failedCount++
			} else {
				// Install the package
				err := manager.Install(ctx, pkg.Name)
				if err != nil {
					packageResults = append(packageResults, PackageResult{
						Name:   pkg.Name,
						Status: "failed",
						Error:  err.Error(),
					})
					failedCount++
				} else {
					packageResults = append(packageResults, PackageResult{
						Name:   pkg.Name,
						Status: "installed",
					})
					installedCount++
				}
			}
		}

		managerResults = append(managerResults, ManagerApplyResult{
			Name:         managerName,
			MissingCount: len(packages),
			Packages:     packageResults,
		})

		totalInstalled += installedCount
		totalFailed += failedCount
		totalWouldInstall += wouldInstallCount
	}

	return PackageApplyResult{
		DryRun:            options.DryRun,
		TotalMissing:      totalMissing,
		TotalInstalled:    totalInstalled,
		TotalFailed:       totalFailed,
		TotalWouldInstall: totalWouldInstall,
		Managers:          managerResults,
	}, nil
}

// CreatePackageProvider creates a multi-manager package provider
func CreatePackageProvider(ctx context.Context, configDir string) (*state.MultiManagerPackageProvider, error) {
	lockService := lock.NewYAMLLockService(configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	registry := managers.NewManagerRegistry()
	return registry.CreateMultiProvider(ctx, lockAdapter)
}
