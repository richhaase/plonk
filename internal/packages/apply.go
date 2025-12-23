// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"sort"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/output"
)

// Apply applies package configuration and returns the result
func Apply(ctx context.Context, configDir string, cfg *config.Config, dryRun bool) (output.PackageResults, error) {
	registry := GetRegistry()

	// Load lock file to get desired packages
	lockService := lock.NewYAMLLockService(configDir)
	lockData, err := lockService.Read()
	if err != nil {
		return output.PackageResults{}, err
	}

	// Convert lock.ResourceEntry to LockResource for reconciliation
	var lockResources []LockResource
	for _, r := range lockData.Resources {
		lockResources = append(lockResources, LockResource{
			Type:     r.Type,
			Metadata: r.Metadata,
		})
	}

	// Simple reconciliation: compare lock file vs installed packages
	result := ReconcileFromLock(ctx, lockResources, registry)

	// Group missing packages by manager
	missingByManager := make(map[string][]PackageSpec)
	for _, pkg := range result.Missing {
		missingByManager[pkg.Manager] = append(missingByManager[pkg.Manager], pkg)
	}

	var managerResults []output.ManagerResults
	totalMissing := len(result.Missing)
	totalInstalled := 0
	totalFailed := 0
	totalWouldInstall := 0

	// Create spinner manager for all package operations if we have missing packages
	var spinnerManager *output.SpinnerManager
	if totalMissing > 0 {
		spinnerManager = output.NewSpinnerManager(totalMissing)
	}

	// Deterministic iteration over managers
	managerNames := make([]string, 0, len(missingByManager))
	for m := range missingByManager {
		managerNames = append(managerNames, m)
	}
	sort.Strings(managerNames)

	// Process each manager's missing packages
	for _, managerName := range managerNames {
		missingPkgs := missingByManager[managerName]
		var packageResults []output.PackageOperation
		installedCount := 0
		failedCount := 0
		wouldInstallCount := 0

		// Get the manager instance
		manager, err := registry.GetManager(managerName)
		if err != nil {
			// Manager not available - mark all as failed
			for _, pkg := range missingPkgs {
				packageResults = append(packageResults, output.PackageOperation{
					Name:   pkg.Name,
					Status: "failed",
					Error:  fmt.Sprintf("manager %s not available: %v", managerName, err),
				})
				failedCount++
			}
			managerResults = append(managerResults, output.ManagerResults{
				Name:         managerName,
				MissingCount: len(missingPkgs),
				Packages:     packageResults,
			})
			totalFailed += failedCount
			continue
		}

		// Sort packages by name for deterministic output
		sort.Slice(missingPkgs, func(i, j int) bool { return missingPkgs[i].Name < missingPkgs[j].Name })

		for _, pkg := range missingPkgs {
			// Start spinner for this package
			var spinner *output.Spinner
			if spinnerManager != nil {
				spinner = spinnerManager.StartSpinner("Installing", fmt.Sprintf("%s (%s)", pkg.Name, managerName))
			}

			if dryRun {
				packageResults = append(packageResults, output.PackageOperation{
					Name:   pkg.Name,
					Status: "would-install",
				})
				wouldInstallCount++
				if spinner != nil {
					spinner.Success(fmt.Sprintf("would-install %s", pkg.Name))
				}
			} else {
				// Install directly via manager
				err := manager.Install(ctx, pkg.Name)
				if err != nil {
					packageResults = append(packageResults, output.PackageOperation{
						Name:   pkg.Name,
						Status: "failed",
						Error:  err.Error(),
					})
					failedCount++
					if spinner != nil {
						spinner.Error(fmt.Sprintf("Failed to install %s: %s", pkg.Name, err.Error()))
					}
				} else {
					packageResults = append(packageResults, output.PackageOperation{
						Name:   pkg.Name,
						Status: "installed",
					})
					installedCount++
					if spinner != nil {
						spinner.Success(fmt.Sprintf("installed %s", pkg.Name))
					}
				}
			}
		}

		managerResults = append(managerResults, output.ManagerResults{
			Name:         managerName,
			MissingCount: len(missingPkgs),
			Packages:     packageResults,
		})

		totalInstalled += installedCount
		totalFailed += failedCount
		totalWouldInstall += wouldInstallCount
	}

	return output.PackageResults{
		DryRun:            dryRun,
		TotalMissing:      totalMissing,
		TotalInstalled:    totalInstalled,
		TotalFailed:       totalFailed,
		TotalWouldInstall: totalWouldInstall,
		Managers:          managerResults,
	}, nil
}
