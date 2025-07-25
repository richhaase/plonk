// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// PackageSyncResult represents the result of package sync operations
type PackageSyncResult struct {
	DryRun            bool                `json:"dry_run" yaml:"dry_run"`
	TotalMissing      int                 `json:"total_missing" yaml:"total_missing"`
	TotalInstalled    int                 `json:"total_installed" yaml:"total_installed"`
	TotalFailed       int                 `json:"total_failed" yaml:"total_failed"`
	TotalWouldInstall int                 `json:"total_would_install" yaml:"total_would_install"`
	Managers          []ManagerSyncResult `json:"managers" yaml:"managers"`
}

// ManagerSyncResult represents the result for a specific manager
type ManagerSyncResult struct {
	Name         string                       `json:"name" yaml:"name"`
	MissingCount int                          `json:"missing_count" yaml:"missing_count"`
	Packages     []PackageOperationSyncResult `json:"packages" yaml:"packages"`
}

// PackageOperationSyncResult represents the result for a specific package operation
type PackageOperationSyncResult struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileSyncResult represents the result of dotfile sync operations
type DotfileSyncResult struct {
	DryRun     bool                      `json:"dry_run" yaml:"dry_run"`
	Backup     bool                      `json:"backup" yaml:"backup"`
	TotalFiles int                       `json:"total_files" yaml:"total_files"`
	Actions    []DotfileActionSyncResult `json:"actions" yaml:"actions"`
	Summary    DotfileSummarySyncResult  `json:"summary" yaml:"summary"`
}

// DotfileActionSyncResult represents an action taken on a dotfile
type DotfileActionSyncResult struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Status      string `json:"status" yaml:"status"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileSummarySyncResult provides summary statistics
type DotfileSummarySyncResult struct {
	Added     int `json:"added" yaml:"added"`
	Updated   int `json:"updated" yaml:"updated"`
	Unchanged int `json:"unchanged" yaml:"unchanged"`
	Failed    int `json:"failed" yaml:"failed"`
}

// SyncPackages applies package configuration and returns the result
func SyncPackages(ctx context.Context, configDir string, cfg *config.Config, dryRun bool) (PackageSyncResult, error) {
	// Reconcile package domain to find missing packages
	result, err := ReconcilePackages(ctx, configDir)
	if err != nil {
		return PackageSyncResult{}, fmt.Errorf("failed to reconcile package state: %w", err)
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

	var managerResults []ManagerSyncResult
	totalMissing := len(result.Missing)
	totalInstalled := 0
	totalFailed := 0
	totalWouldInstall := 0

	// Process each manager's missing packages
	for managerName, missingItems := range missingByManager {
		var packageResults []PackageOperationSyncResult
		installedCount := 0
		failedCount := 0
		wouldInstallCount := 0

		for _, item := range missingItems {
			if dryRun {
				packageResults = append(packageResults, PackageOperationSyncResult{
					Name:   item.Name,
					Status: "would-install",
				})
				wouldInstallCount++
			} else {
				// Use resource Apply method
				err := packageResource.Apply(ctx, item)
				if err != nil {
					packageResults = append(packageResults, PackageOperationSyncResult{
						Name:   item.Name,
						Status: "failed",
						Error:  err.Error(),
					})
					failedCount++
				} else {
					packageResults = append(packageResults, PackageOperationSyncResult{
						Name:   item.Name,
						Status: "installed",
					})
					installedCount++
				}
			}
		}

		managerResults = append(managerResults, ManagerSyncResult{
			Name:         managerName,
			MissingCount: len(missingItems),
			Packages:     packageResults,
		})

		totalInstalled += installedCount
		totalFailed += failedCount
		totalWouldInstall += wouldInstallCount
	}

	return PackageSyncResult{
		DryRun:            dryRun,
		TotalMissing:      totalMissing,
		TotalInstalled:    totalInstalled,
		TotalFailed:       totalFailed,
		TotalWouldInstall: totalWouldInstall,
		Managers:          managerResults,
	}, nil
}

// SyncDotfiles applies dotfile configuration and returns the result
func SyncDotfiles(ctx context.Context, configDir, homeDir string, cfg *config.Config, dryRun, backup bool) (DotfileSyncResult, error) {
	// Create dotfile resource
	manager := dotfiles.NewManager(homeDir, configDir)
	dotfileResource := dotfiles.NewDotfileResource(manager)

	// Get configured dotfiles and set as desired
	configuredItems, err := manager.GetConfiguredDotfiles()
	if err != nil {
		return DotfileSyncResult{}, fmt.Errorf("failed to get configured dotfiles: %w", err)
	}
	dotfileResource.SetDesired(configuredItems)

	// Reconcile to find what needs to be done
	reconciled, err := ReconcileResource(ctx, dotfileResource)
	if err != nil {
		return DotfileSyncResult{}, fmt.Errorf("failed to reconcile dotfiles: %w", err)
	}

	var actions []DotfileActionSyncResult
	summary := DotfileSummarySyncResult{}

	// Process missing dotfiles (need to be created/linked)
	for _, item := range reconciled {
		if item.State == resources.StateMissing {
			if !dryRun {
				// Apply the change using the resource interface
				err := dotfileResource.Apply(ctx, item)

				action := DotfileActionSyncResult{
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
				actions = append(actions, DotfileActionSyncResult{
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

	return DotfileSyncResult{
		DryRun:     dryRun,
		Backup:     backup,
		TotalFiles: len(configuredItems),
		Actions:    actions,
		Summary:    summary,
	}, nil
}
