// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import "github.com/richhaase/plonk/internal/state"

// Compile-time interface compliance check
var _ state.PackageConfigLoader = (*LockFileAdapter)(nil)

// LockFileAdapter bridges the lock package's LockService to the state package's
// PackageConfigLoader interface. This adapter enables the state package to treat
// the lock file as another source of package configuration, alongside the main
// config file. It prevents circular dependencies between the lock and state packages.
//
// Bridge: lock.LockService → state.PackageConfigLoader
type LockFileAdapter struct {
	lockService LockService
}

// NewLockFileAdapter creates a new lock file adapter
func NewLockFileAdapter(lockService LockService) *LockFileAdapter {
	return &LockFileAdapter{
		lockService: lockService,
	}
}

// GetPackagesForManager returns packages for a specific manager from the lock file
func (a *LockFileAdapter) GetPackagesForManager(managerName string) ([]state.PackageConfigItem, error) {
	packages, err := a.lockService.GetPackages(managerName)
	if err != nil {
		return nil, err
	}

	// Convert lock file entries to config items
	items := make([]state.PackageConfigItem, len(packages))
	for i, pkg := range packages {
		items[i] = state.PackageConfigItem{
			Name: pkg.Name,
		}
	}

	return items, nil
}
