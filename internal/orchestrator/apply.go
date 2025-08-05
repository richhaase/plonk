// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// PackageApplyResult represents the result of package apply operations
// Type aliases to consolidated types in output package
type PackageApplyResult = output.PackageResults
type ManagerApplyResult = output.ManagerResults
type PackageOperationApplyResult = output.PackageOperation
type DotfileApplyResult = output.DotfileResults
type DotfileActionApplyResult = output.DotfileOperation
type DotfileSummaryApplyResult = output.DotfileSummary

// Legacy apply functions - keeping for backward compatibility during transition
// These will be removed in a future phase

// ApplyPackages applies package configuration and returns the result
func ApplyPackages(ctx context.Context, configDir string, cfg *config.Config, dryRun bool) (PackageApplyResult, error) {
	// Reconcile package domain to find missing packages
	result, err := packages.Reconcile(ctx, configDir)
	if err != nil {
		return PackageApplyResult{}, err
	}

	// Create multi-package resource for applying changes
	packageResource := packages.NewMultiPackageResource()

	// Group missing packages by manager
	missingByManager := make(map[string][]resources.Item)
	for _, item := range result.Missing {
		if item.Manager != "" {
			missingByManager[item.Manager] = append(missingByManager[item.Manager], item)
		}
	}

	var managerResults []ManagerApplyResult
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
		var packageResults []PackageOperationApplyResult
		installedCount := 0
		failedCount := 0
		wouldInstallCount := 0

		for _, item := range missingItems {
			packageIndex++
			// Show progress for each package
			output.ProgressUpdate(packageIndex, totalMissing, "Installing", item.Name)

			if dryRun {
				packageResults = append(packageResults, PackageOperationApplyResult{
					Name:   item.Name,
					Status: "would-install",
				})
				wouldInstallCount++
			} else {
				// Use resource Apply method
				err := packageResource.Apply(ctx, item)
				if err != nil {
					packageResults = append(packageResults, PackageOperationApplyResult{
						Name:   item.Name,
						Status: "failed",
						Error:  err.Error(),
					})
					failedCount++
				} else {
					packageResults = append(packageResults, PackageOperationApplyResult{
						Name:   item.Name,
						Status: "installed",
					})
					installedCount++
				}
			}
		}

		managerResults = append(managerResults, ManagerApplyResult{
			Name:         managerName,
			MissingCount: len(missingItems),
			Packages:     packageResults,
		})

		totalInstalled += installedCount
		totalFailed += failedCount
		totalWouldInstall += wouldInstallCount
	}

	return PackageApplyResult{
		DryRun:            dryRun,
		TotalMissing:      totalMissing,
		TotalInstalled:    totalInstalled,
		TotalFailed:       totalFailed,
		TotalWouldInstall: totalWouldInstall,
		Managers:          managerResults,
	}, nil
}

// ApplyDotfiles applies dotfile configuration and returns the result
func ApplyDotfiles(ctx context.Context, configDir, homeDir string, cfg *config.Config, dryRun bool) (DotfileApplyResult, error) {
	// Create dotfile resource
	manager := dotfiles.NewManager(homeDir, configDir)
	dotfileResource := dotfiles.NewDotfileResource(manager)

	// Get configured dotfiles and set as desired
	configuredItems, err := manager.GetConfiguredDotfiles()
	if err != nil {
		return DotfileApplyResult{}, err
	}
	dotfileResource.SetDesired(configuredItems)

	// Reconcile to find what needs to be done
	reconciled, err := resources.ReconcileResource(ctx, dotfileResource)
	if err != nil {
		return DotfileApplyResult{}, err
	}

	var actions []DotfileActionApplyResult
	summary := DotfileSummaryApplyResult{}

	// Count missing and drifted dotfiles
	applyCount := 0
	for _, item := range reconciled {
		if item.State == resources.StateMissing || item.State == resources.StateDegraded {
			applyCount++
		}
	}

	// Show overall progress for dotfiles
	if applyCount > 0 {
		output.StageUpdate(fmt.Sprintf("Applying dotfiles (%d missing/drifted)...", applyCount))
	}

	// Process missing and drifted dotfiles (need to be created/restored)
	dotfileIndex := 0
	for _, item := range reconciled {
		if item.State == resources.StateMissing || item.State == resources.StateDegraded {
			dotfileIndex++
			// Show progress for each dotfile
			output.ProgressUpdate(dotfileIndex, applyCount, "Deploying", item.Name)

			if !dryRun {
				// Apply the change using the resource interface
				err := dotfileResource.Apply(ctx, item)

				action := DotfileActionApplyResult{
					Source:      item.Path,
					Destination: item.Name,
					Action:      "copy",
					Status:      "added",
				}

				if err != nil {
					action.Action = "error"
					action.Status = "failed"
					action.Error = err.Error()
					summary.Failed++
				} else {
					summary.Added++
				}
				actions = append(actions, action)
			} else {
				// Dry run
				actions = append(actions, DotfileActionApplyResult{
					Source:      item.Path,
					Destination: item.Name,
					Action:      "would-copy",
					Status:      "would-add",
				})
				summary.Added++
			}
		} else if item.State == resources.StateManaged {
			// Already managed files are unchanged
			summary.Unchanged++
		}
	}

	return DotfileApplyResult{
		DryRun:     dryRun,
		TotalFiles: len(configuredItems),
		Actions:    actions,
		Summary:    summary,
	}, nil
}
