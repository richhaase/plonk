// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
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
	manager := NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)
	if err := applyDeployModes(manager, cfg); err != nil {
		return output.DotfileResults{DryRun: opts.DryRun}, err
	}

	// Get all statuses
	statuses, err := manager.Reconcile()
	if err != nil {
		return output.DotfileResults{DryRun: opts.DryRun}, err
	}

	// Filter if needed
	if len(opts.Filter) > 0 {
		var filtered []DotfileStatus
		for _, s := range statuses {
			if opts.Filter[normalizePath(s.Target)] {
				filtered = append(filtered, s)
			}
		}
		statuses = filtered
	}

	return applyStatuses(ctx, manager, statuses, opts.DryRun)
}

// Apply applies dotfile configuration and returns the result
func Apply(ctx context.Context, configDir, homeDir string, cfg *config.Config, dryRun bool) (output.DotfileResults, error) {
	return ApplySelective(ctx, configDir, homeDir, cfg, ApplyFilterOptions{DryRun: dryRun})
}

// applyDeployModes configures the manager with per-dotfile deploy modes from
// the config, if any.
func applyDeployModes(manager *DotfileManager, cfg *config.Config) error {
	modes, err := cfg.Dotfiles.DeployModes()
	if err != nil {
		return err
	}
	if len(modes) > 0 {
		manager.SetDeployModes(modes)
	}
	return nil
}

func normalizePath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}

// applyStatuses applies the given dotfile statuses and returns results
func applyStatuses(ctx context.Context, manager *DotfileManager, statuses []DotfileStatus, dryRun bool) (output.DotfileResults, error) {
	result := output.DotfileResults{
		DryRun:     dryRun,
		TotalFiles: len(statuses),
	}

	spinnerCount := 0
	for _, s := range statuses {
		if s.State == SyncStateMissing || s.State == SyncStateDrifted {
			spinnerCount++
		}
	}

	var spinnerManager *output.SpinnerManager
	if spinnerCount > 0 {
		spinnerManager = output.NewSpinnerManager(spinnerCount)
	}

	for _, s := range statuses {
		// Check for context cancellation/timeout between files
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("apply canceled: %w", err)
		}

		switch s.State {
		case SyncStateManaged:
			result.Summary.Unchanged++

		case SyncStateError:
			action := output.DotfileOperation{
				Source:      s.Source,
				Destination: s.Target,
				Action:      "error",
				Status:      "failed",
			}
			if s.Error != nil {
				action.Error = s.Error.Error()
			}
			result.Actions = append(result.Actions, action)
			result.Summary.Failed++

		case SyncStateMissing, SyncStateDrifted:
			stage, verb, completed := "Deploying", "deploy", "deployed"
			status, dryStatus := "added", "would-add"
			if s.State == SyncStateDrifted {
				stage, verb, completed = "Updating", "update", "updated"
				status, dryStatus = "updated", "would-update"
			}
			spinner := spinnerManager.StartSpinner(stage, s.Name)
			action := output.DotfileOperation{
				Source:      s.Source,
				Destination: s.Target,
				Action:      "copy",
				Status:      status,
			}

			var deployErr error
			if !dryRun {
				deployErr = manager.Deploy(s.Name)
			}
			if deployErr != nil {
				action.Action = "error"
				action.Status = "failed"
				action.Error = deployErr.Error()
				result.Summary.Failed++
				spinner.Error("Failed to " + verb + " " + s.Name + ": " + deployErr.Error())
			} else {
				if s.State == SyncStateMissing {
					result.Summary.Added++
				} else {
					result.Summary.Updated++
				}
				if dryRun {
					action.Action = "would-copy"
					action.Status = dryStatus
					spinner.Success("would-" + verb + " " + s.Name)
				} else {
					spinner.Success(completed + " " + s.Name)
				}
			}
			result.Actions = append(result.Actions, action)
		}
	}

	if result.Summary.Failed > 0 {
		return result, fmt.Errorf("failed to deploy %d file(s)", result.Summary.Failed)
	}

	return result, nil
}
