// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
)

// ApplyFilterOptions contains options for selective dotfile apply operations
type ApplyFilterOptions struct {
	DryRun bool
	// Filter is a set of normalized destination paths to apply.
	// If empty or nil, all dotfiles are applied.
	Filter map[string]bool
}

// ApplySelective applies only the dotfiles whose destination paths are in the filter set.
// The filter should contain normalized absolute paths (use filepath.Abs and filepath.Clean).
func ApplySelective(ctx context.Context, configDir, homeDir string, cfg *config.Config, opts ApplyFilterOptions) (output.DotfileResults, error) {
	return applyWithFilter(ctx, configDir, homeDir, cfg, opts.DryRun, opts.Filter)
}

// Apply applies dotfile configuration and returns the result
func Apply(ctx context.Context, configDir, homeDir string, cfg *config.Config, dryRun bool) (output.DotfileResults, error) {
	return applyWithFilter(ctx, configDir, homeDir, cfg, dryRun, nil)
}

// applyWithFilter is the internal implementation that supports optional filtering
func applyWithFilter(ctx context.Context, configDir, homeDir string, cfg *config.Config, dryRun bool, filter map[string]bool) (output.DotfileResults, error) {
	// Create dotfile resource
	manager := NewManagerWithConfig(homeDir, configDir, cfg)
	dotfileResource := NewDotfileResource(manager, dryRun)

	// Get configured dotfiles and set as desired
	configuredItems, err := manager.GetConfiguredDotfiles()
	if err != nil {
		return output.DotfileResults{}, err
	}
	dotfileResource.SetDesired(configuredItems)

	// Reconcile to find what needs to be done
	reconciled, err := resources.ReconcileResource(ctx, dotfileResource)
	if err != nil {
		return output.DotfileResults{}, err
	}

	// If we have a filter, only keep items that match
	if len(filter) > 0 {
		reconciled = filterItems(reconciled, filter, homeDir)
	}

	var actions []output.DotfileOperation
	summary := output.DotfileSummary{}

	// Count missing and drifted dotfiles
	applyCount := 0
	for _, item := range reconciled {
		if item.State == resources.StateMissing || item.State == resources.StateDegraded {
			applyCount++
		}
	}

	// Create spinner manager for all dotfile operations if we have files to apply
	var spinnerManager *output.SpinnerManager
	if applyCount > 0 {
		spinnerManager = output.NewSpinnerManager(applyCount)
	}

	// Process missing and drifted dotfiles (need to be created/restored)
	for _, item := range reconciled {
		if item.State == resources.StateMissing || item.State == resources.StateDegraded {
			// Start spinner for this dotfile
			var spinner *output.Spinner
			if spinnerManager != nil {
				spinner = spinnerManager.StartSpinner("Deploying", item.Name)
			}

			if !dryRun {
				// Apply the change using the resource interface
				err := dotfileResource.Apply(ctx, item)

				action := output.DotfileOperation{
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
					if spinner != nil {
						spinner.Error(fmt.Sprintf("Failed to deploy %s: %s", item.Name, err.Error()))
					}
				} else {
					summary.Added++
					if spinner != nil {
						spinner.Success(fmt.Sprintf("deployed %s", item.Name))
					}
				}
				actions = append(actions, action)
			} else {
				// Dry run
				actions = append(actions, output.DotfileOperation{
					Source:      item.Path,
					Destination: item.Name,
					Action:      "would-copy",
					Status:      "would-add",
				})
				summary.Added++
				if spinner != nil {
					spinner.Success(fmt.Sprintf("would-deploy %s", item.Name))
				}
			}
		} else if item.State == resources.StateManaged {
			// Already managed files are unchanged
			summary.Unchanged++
		}
	}

	// For selective apply, TotalFiles reflects filtered count
	totalFiles := len(configuredItems)
	if len(filter) > 0 {
		totalFiles = len(reconciled)
	}

	return output.DotfileResults{
		DryRun:     dryRun,
		TotalFiles: totalFiles,
		Actions:    actions,
		Summary:    summary,
	}, nil
}

// filterItems filters reconciled items to only include those matching the filter set
func filterItems(items []resources.Item, filter map[string]bool, homeDir string) []resources.Item {
	if len(filter) == 0 {
		return items
	}

	var filtered []resources.Item
	for _, item := range items {
		// Get the destination path from metadata
		dest, ok := item.Metadata["destination"].(string)
		if !ok {
			// Try using item.Path as fallback
			dest = item.Path
		}

		// Normalize the destination path
		normalizedDest := normalizeDestPath(dest, homeDir)

		// Check if this item is in the filter set
		if filter[normalizedDest] {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// normalizeDestPath normalizes a destination path for comparison
func normalizeDestPath(path, homeDir string) string {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path // Return as-is if we can't resolve
	}

	return filepath.Clean(absPath)
}
