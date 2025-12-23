// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
)

// Reconcile performs dotfile reconciliation (backward compatibility wrapper)
func Reconcile(ctx context.Context, homeDir, configDir string) (resources.Result, error) {
	cfg := config.LoadWithDefaults(configDir)
	return ReconcileWithConfig(ctx, homeDir, configDir, cfg)
}

// ReconcileWithConfig reconciles dotfiles using injected config
func ReconcileWithConfig(ctx context.Context, homeDir, configDir string, cfg *config.Config) (resources.Result, error) {
	manager := NewManagerWithConfig(homeDir, configDir, cfg)

	// Get desired state from config
	desired, err := manager.GetConfiguredDotfiles()
	if err != nil {
		return resources.Result{}, err
	}

	// Get actual state from filesystem
	actual, err := manager.GetActualDotfiles(ctx)
	if err != nil {
		return resources.Result{}, err
	}

	// Reconcile desired vs actual using the generic reconciliation algorithm
	reconciled := resources.ReconcileItems(desired, actual)

	// Group by state for the result
	managed, missing, untracked := resources.GroupItemsByState(reconciled)

	return resources.Result{
		Domain:    "dotfile",
		Managed:   managed,
		Missing:   missing,
		Untracked: untracked,
	}, nil
}
