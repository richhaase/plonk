// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/resources"
)

// GetActualPackages returns packages currently installed on the system
func GetActualPackages(ctx context.Context) ([]resources.Item, error) {
	registry := NewManagerRegistry()
	items := make([]resources.Item, 0)

	// Get all available managers
	for _, managerName := range registry.GetAllManagerNames() {
		manager, err := registry.GetManager(managerName)
		if err != nil {
			continue // Skip unavailable managers
		}

		// Check if manager is actually available
		available, err := manager.IsAvailable(ctx)
		if err != nil || !available {
			continue
		}

		// Get installed packages for this manager
		installed, err := manager.ListInstalled(ctx)
		if err != nil {
			// Log error but continue with other managers
			continue
		}

		// Convert to resources.Item
		for _, pkgName := range installed {
			items = append(items, resources.Item{
				Name:    pkgName,
				State:   resources.StateUntracked, // Will be updated during reconciliation
				Domain:  "package",
				Manager: managerName,
			})
		}
	}

	return items, nil
}

// GetActualPackagesForManager returns packages installed by a specific manager
func GetActualPackagesForManager(ctx context.Context, managerName string) ([]resources.Item, error) {
	registry := NewManagerRegistry()
	manager, err := registry.GetManager(managerName)
	if err != nil {
		return nil, fmt.Errorf("getting manager %s: %w", managerName, err)
	}

	available, err := manager.IsAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking manager availability: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("manager %s is not available", managerName)
	}

	// Get installed packages
	installed, err := manager.ListInstalled(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing installed packages: %w", err)
	}

	// Convert to resources.Item
	items := make([]resources.Item, 0, len(installed))
	for _, pkgName := range installed {
		items = append(items, resources.Item{
			Name:    pkgName,
			State:   resources.StateUntracked, // Will be updated during reconciliation
			Domain:  "package",
			Manager: managerName,
		})
	}

	return items, nil
}
