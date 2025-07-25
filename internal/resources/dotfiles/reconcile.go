// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"

	"github.com/richhaase/plonk/internal/resources"
)

// Reconcile performs dotfile reconciliation (backward compatibility)
func Reconcile(ctx context.Context, homeDir, configDir string) (resources.Result, error) {
	// Create dotfile resource
	manager := NewManager(homeDir, configDir)
	dotfileResource := NewDotfileResource(manager)

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
	return resources.Result{
		Domain:    "dotfile",
		Managed:   managed,
		Missing:   missing,
		Untracked: untracked,
	}, nil
}
