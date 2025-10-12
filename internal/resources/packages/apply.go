// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"sort"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
)

// Apply applies package configuration and returns the result
func Apply(ctx context.Context, configDir string, cfg *config.Config, dryRun bool) (output.PackageResults, error) {
	// Load v2 configs from plonk.yaml before any operations
	registry := NewManagerRegistry()
	if registry != nil {
		registry.LoadV2Configs(cfg)
	}

	// Reconcile package domain to find missing packages
	result, err := Reconcile(ctx, configDir)
	if err != nil {
		return output.PackageResults{}, err
	}

	// Create multi-package resource for applying changes
	packageResource := NewMultiPackageResource()

	// Group missing packages by manager
	missingByManager := make(map[string][]resources.Item)
	for _, item := range result.Missing {
		if item.Manager != "" {
			missingByManager[item.Manager] = append(missingByManager[item.Manager], item)
		}
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
		missingItems := missingByManager[managerName]
		var packageResults []output.PackageOperation
		installedCount := 0
		failedCount := 0
		wouldInstallCount := 0

		// Sort items by name for deterministic output
		sort.Slice(missingItems, func(i, j int) bool { return missingItems[i].Name < missingItems[j].Name })
		for _, item := range missingItems {
			// Start spinner for this package
			var spinner *output.Spinner
			if spinnerManager != nil {
				spinner = spinnerManager.StartSpinner("Installing", fmt.Sprintf("%s (%s)", item.Name, managerName))
			}

			if dryRun {
				packageResults = append(packageResults, output.PackageOperation{
					Name:   item.Name,
					Status: "would-install",
				})
				wouldInstallCount++
				if spinner != nil {
					spinner.Success(fmt.Sprintf("would-install %s", item.Name))
				}
			} else {
				// Use resource Apply method
				err := packageResource.Apply(ctx, item)
				if err != nil {
					packageResults = append(packageResults, output.PackageOperation{
						Name:   item.Name,
						Status: "failed",
						Error:  err.Error(),
					})
					failedCount++
					if spinner != nil {
						spinner.Error(fmt.Sprintf("Failed to install %s: %s", item.Name, err.Error()))
					}
				} else {
					packageResults = append(packageResults, output.PackageOperation{
						Name:   item.Name,
						Status: "installed",
					})
					installedCount++
					if spinner != nil {
						spinner.Success(fmt.Sprintf("installed %s", item.Name))
					}
				}
			}
		}

		managerResults = append(managerResults, output.ManagerResults{
			Name:         managerName,
			MissingCount: len(missingItems),
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
