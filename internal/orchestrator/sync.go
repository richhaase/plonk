// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/richhaase/plonk/internal/state"
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

	// Group missing packages by manager
	missingByManager := make(map[string][]state.Item)
	for _, item := range result.Missing {
		manager := item.Metadata["manager"].(string)
		missingByManager[manager] = append(missingByManager[manager], item)
	}

	var managerResults []ManagerSyncResult
	totalMissing := len(result.Missing)
	totalInstalled := 0
	totalFailed := 0
	totalWouldInstall := 0

	// Get manager registry
	registry := packages.NewManagerRegistry()

	// Process each manager's missing packages
	for managerName, packages := range missingByManager {
		manager, err := registry.GetManager(managerName)
		if err != nil {
			return PackageSyncResult{}, fmt.Errorf("failed to get package manager %s: %w", managerName, err)
		}

		available, err := manager.IsAvailable(ctx)
		if err != nil {
			return PackageSyncResult{}, fmt.Errorf("failed to check manager %s availability: %w", managerName, err)
		}

		var packageResults []PackageOperationSyncResult
		installedCount := 0
		failedCount := 0
		wouldInstallCount := 0

		for _, pkg := range packages {
			if dryRun {
				packageResults = append(packageResults, PackageOperationSyncResult{
					Name:   pkg.Name,
					Status: "would-install",
				})
				wouldInstallCount++
			} else if !available {
				packageResults = append(packageResults, PackageOperationSyncResult{
					Name:   pkg.Name,
					Status: "failed",
					Error:  "package manager not available",
				})
				failedCount++
			} else {
				// Install the package
				err := manager.Install(ctx, pkg.Name)
				if err != nil {
					packageResults = append(packageResults, PackageOperationSyncResult{
						Name:   pkg.Name,
						Status: "failed",
						Error:  err.Error(),
					})
					failedCount++
				} else {
					packageResults = append(packageResults, PackageOperationSyncResult{
						Name:   pkg.Name,
						Status: "installed",
					})
					installedCount++
				}
			}
		}

		managerResults = append(managerResults, ManagerSyncResult{
			Name:         managerName,
			MissingCount: len(packages),
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
	// Get configured dotfiles
	configuredItems, err := dotfiles.GetConfiguredDotfiles(homeDir, configDir)
	if err != nil {
		return DotfileSyncResult{}, fmt.Errorf("failed to get configured dotfiles: %w", err)
	}

	var actions []DotfileActionSyncResult
	summary := DotfileSummarySyncResult{}

	// Process each configured dotfile
	for _, item := range configuredItems {
		// Create dotfile manager
		manager := dotfiles.NewManager(homeDir, configDir)
		opts := dotfiles.ApplyOptions{
			DryRun: dryRun,
			Backup: backup,
		}

		result, err := manager.ProcessDotfileForApply(ctx,
			item.Metadata["source"].(string),
			item.Metadata["destination"].(string),
			opts)

		action := DotfileActionSyncResult{
			Source:      result.Source,
			Destination: result.Destination,
			Action:      result.Action,
			Status:      result.Status,
			Error:       result.Error,
		}

		if err != nil {
			action.Action = "error"
			action.Status = "failed"
			action.Error = err.Error()
			summary.Failed++
		} else {
			switch action.Status {
			case "added":
				summary.Added++
			case "updated":
				summary.Updated++
			case "unchanged":
				summary.Unchanged++
			case "failed":
				summary.Failed++
			}
		}

		actions = append(actions, action)
	}

	return DotfileSyncResult{
		DryRun:     dryRun,
		Backup:     backup,
		TotalFiles: len(configuredItems),
		Actions:    actions,
		Summary:    summary,
	}, nil
}
