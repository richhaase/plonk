// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
)

// Apply applies package configuration and returns the result
func Apply(ctx context.Context, configDir string, cfg *config.Config, dryRun bool) (output.PackageResults, error) {
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

	// Show overall progress for packages
	if totalMissing > 0 {
		output.StageUpdate(fmt.Sprintf("Applying packages (%d missing)...", totalMissing))
	}

	// Process each manager's missing packages
	packageIndex := 0
	for managerName, missingItems := range missingByManager {
		var packageResults []output.PackageOperation
		installedCount := 0
		failedCount := 0
		wouldInstallCount := 0

		for _, item := range missingItems {
			packageIndex++
			// Show progress for each package
			output.ProgressUpdate(packageIndex, totalMissing, "Installing", item.Name)

			if dryRun {
				packageResults = append(packageResults, output.PackageOperation{
					Name:   item.Name,
					Status: "would-install",
				})
				wouldInstallCount++
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
				} else {
					packageResults = append(packageResults, output.PackageOperation{
						Name:   item.Name,
						Status: "installed",
					})
					installedCount++
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
