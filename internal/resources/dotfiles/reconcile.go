// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
)

// Reconcile performs dotfile reconciliation (backward compatibility)
func Reconcile(ctx context.Context, homeDir, configDir string) (resources.Result, error) {
	// Backward-compatible wrapper that loads config
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

	// Convert to Result format for backward compatibility
	managed, missing, untracked := resources.GroupItemsByState(reconciled)

	// Also collect drifted items (they have StateDegraded but aren't in managed)
	var drifted []resources.Item
	for _, item := range reconciled {
		if item.State == resources.StateDegraded {
			drifted = append(drifted, item)
		}
	}

	// For now, put drifted items in the managed list but they'll have StateDegraded
	// This preserves backward compatibility while allowing status to detect drift
	allManaged := append(managed, drifted...)

	return resources.Result{
		Domain:    "dotfile",
		Managed:   allManaged,
		Missing:   missing,
		Untracked: untracked,
	}, nil
}
