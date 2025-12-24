// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
)

// Result contains the results of dotfile reconciliation
type Result struct {
	Domain    string
	Managed   []DotfileItem
	Missing   []DotfileItem
	Untracked []DotfileItem
}

// ReconcileWithConfig reconciles dotfiles using injected config
func ReconcileWithConfig(ctx context.Context, homeDir, configDir string, cfg *config.Config) (Result, error) {
	manager := NewManagerWithConfig(homeDir, configDir, cfg)

	// Get desired state from config
	desired, err := manager.GetConfiguredDotfiles()
	if err != nil {
		return Result{}, err
	}

	// Get actual state from filesystem
	actual, err := manager.GetActualDotfiles(ctx)
	if err != nil {
		return Result{}, err
	}

	// Reconcile desired vs actual using dotfile-specific reconciliation
	reconciled := ReconcileItems(desired, actual)

	// Group by state for the result
	managed, missing, untracked := GroupItemsByState(reconciled)

	return Result{
		Domain:    "dotfile",
		Managed:   managed,
		Missing:   missing,
		Untracked: untracked,
	}, nil
}
