// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
)

// Apply applies dotfile configuration and returns the result
func Apply(ctx context.Context, configDir, homeDir string, cfg *config.Config, dryRun bool) (output.DotfileResults, error) {
	// Create dotfile resource
	manager := NewManager(homeDir, configDir)
	dotfileResource := NewDotfileResource(manager)

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

	return output.DotfileResults{
		DryRun:     dryRun,
		TotalFiles: len(configuredItems),
		Actions:    actions,
		Summary:    summary,
	}, nil
}
