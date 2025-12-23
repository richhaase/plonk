// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
)

// ReconcileWithConfig performs package reconciliation with injected config.
// Returns a ReconcileResult containing domain-specific PackageSpec types.
func ReconcileWithConfig(ctx context.Context, configDir string, cfg *config.Config) (ReconcileResult, error) {
	// Load lock file
	lockService := lock.NewYAMLLockService(configDir)
	lockData, err := lockService.Read()
	if err != nil {
		return ReconcileResult{}, err
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

	return result, nil
}

