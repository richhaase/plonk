// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources"
)

// ReconcileWithConfig performs package reconciliation with injected config.
// Returns a resources.Result for backward compatibility with status commands.
func ReconcileWithConfig(ctx context.Context, configDir string, cfg *config.Config) (resources.Result, error) {
	// Load lock file
	lockService := lock.NewYAMLLockService(configDir)
	lockData, err := lockService.Read()
	if err != nil {
		return resources.Result{}, err
	}

	// Convert lock.ResourceEntry to LockResource for reconciliation
	var lockResources []LockResource
	for _, r := range lockData.Resources {
		lockResources = append(lockResources, LockResource{
			Type:     r.Type,
			Metadata: r.Metadata,
		})
	}

	// Use simple reconciliation
	result := ReconcileFromLock(ctx, lockResources, GetRegistry())

	// Convert ReconcileResult to resources.Result for backward compatibility
	return toResourcesResult(result), nil
}

// toResourcesResult converts a ReconcileResult to resources.Result for backward compatibility.
func toResourcesResult(r ReconcileResult) resources.Result {
	managed := make([]resources.Item, 0, len(r.Managed))
	for _, pkg := range r.Managed {
		managed = append(managed, resources.Item{
			Name:    pkg.Name,
			Manager: pkg.Manager,
			Domain:  "package",
			State:   resources.StateManaged,
		})
	}

	missing := make([]resources.Item, 0, len(r.Missing))
	for _, pkg := range r.Missing {
		missing = append(missing, resources.Item{
			Name:    pkg.Name,
			Manager: pkg.Manager,
			Domain:  "package",
			State:   resources.StateMissing,
		})
	}

	untracked := make([]resources.Item, 0, len(r.Untracked))
	for _, pkg := range r.Untracked {
		untracked = append(untracked, resources.Item{
			Name:    pkg.Name,
			Manager: pkg.Manager,
			Domain:  "package",
			State:   resources.StateUntracked,
		})
	}

	return resources.Result{
		Domain:    "package",
		Managed:   managed,
		Missing:   missing,
		Untracked: untracked,
	}
}
