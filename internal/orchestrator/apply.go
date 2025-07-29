// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/richhaase/plonk/internal/resources/packages"
)

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
	Name         string                        `json:"name" yaml:"name"`
	MissingCount int                           `json:"missing_count" yaml:"missing_count"`
	Packages     []PackageOperationApplyResult `json:"packages" yaml:"packages"`
}

// PackageOperationApplyResult represents the result for a specific package operation
type PackageOperationApplyResult struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileApplyResult represents the result of dotfile apply operations
type DotfileApplyResult struct {
	DryRun     bool                       `json:"dry_run" yaml:"dry_run"`
	TotalFiles int                        `json:"total_files" yaml:"total_files"`
	Actions    []DotfileActionApplyResult `json:"actions" yaml:"actions"`
	Summary    DotfileSummaryApplyResult  `json:"summary" yaml:"summary"`
}

// DotfileActionApplyResult represents an action taken on a dotfile
type DotfileActionApplyResult struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Status      string `json:"status" yaml:"status"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileSummaryApplyResult provides summary statistics
type DotfileSummaryApplyResult struct {
	Added     int `json:"added" yaml:"added"`
	Updated   int `json:"updated" yaml:"updated"`
	Unchanged int `json:"unchanged" yaml:"unchanged"`
	Failed    int `json:"failed" yaml:"failed"`
}

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

	// Process each manager's missing packages
	for managerName, missingItems := range missingByManager {
		var packageResults []PackageOperationApplyResult
		installedCount := 0
		failedCount := 0
		wouldInstallCount := 0

		for _, item := range missingItems {
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

	// Process missing dotfiles (need to be created/linked)
	for _, item := range reconciled {
		if item.State == resources.StateMissing {
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
