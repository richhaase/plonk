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
	// Create dotfile resource
	manager := NewManagerWithConfig(homeDir, configDir, cfg)
	dotfileResource := NewDotfileResource(manager, false)

	// Get configured dotfiles and set as desired
	configured, err := manager.GetConfiguredDotfiles()
	if err != nil {
		return resources.Result{}, err
	}
	dotfileResource.SetDesired(configured)

	// Reconcile using the resource interface
	reconciled, err := resources.ReconcileResource(ctx, dotfileResource)
	if err != nil {
		return resources.Result{}, err
	}

	// Convert to Result format
	managed, missing, untracked := resources.GroupItemsByState(reconciled)

	// GroupItemsByState already includes StateDegraded items in the managed list
	// No need to append them again
	return resources.Result{
		Domain:    "dotfile",
		Managed:   managed,
		Missing:   missing,
		Untracked: untracked,
	}, nil
}
