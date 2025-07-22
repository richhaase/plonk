// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"

	"github.com/richhaase/plonk/internal/constants"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/state"
)

// ManagerFactory defines a function that creates a package manager instance
type ManagerFactory func() PackageManager

// ManagerRegistry manages package manager creation and availability checking
type ManagerRegistry struct {
	managers map[string]ManagerFactory
}

// NewManagerRegistry creates a new manager registry with all supported package managers
func NewManagerRegistry() *ManagerRegistry {
	return &ManagerRegistry{
		managers: map[string]ManagerFactory{
			"homebrew": func() PackageManager { return NewHomebrewManager() },
			"npm":      func() PackageManager { return NewNpmManager() },
			"cargo":    func() PackageManager { return NewCargoManager() },
			"pip":      func() PackageManager { return NewPipManager() },
			"gem":      func() PackageManager { return NewGemManager() },
			"go":       func() PackageManager { return NewGoInstallManager() },
		},
	}
}

// GetManager returns a package manager instance by name
func (r *ManagerRegistry) GetManager(name string) (PackageManager, error) {
	factory, exists := r.managers[name]
	if !exists {
		return nil, errors.NewError(errors.ErrInvalidInput, errors.DomainPackages, "get-manager",
			"unsupported package manager: "+name).WithSuggestionMessage("Check available managers with 'plonk doctor' or change default with 'plonk config edit'")
	}
	return factory(), nil
}

// GetAvailableManagers returns a list of manager names that are currently available
func (r *ManagerRegistry) GetAvailableManagers(ctx context.Context) []string {
	var available []string
	for name, factory := range r.managers {
		manager := factory()
		if isAvailable, err := manager.IsAvailable(ctx); err == nil && isAvailable {
			available = append(available, name)
		}
	}
	return available
}

// GetAllManagerNames returns all supported manager names regardless of availability
func (r *ManagerRegistry) GetAllManagerNames() []string {
	// Return a copy to prevent external modification
	names := make([]string, len(constants.SupportedManagers))
	copy(names, constants.SupportedManagers)
	return names
}

// CreateMultiProvider creates a MultiManagerPackageProvider with all available managers
func (r *ManagerRegistry) CreateMultiProvider(ctx context.Context, configLoader state.PackageConfigLoader) (*state.MultiManagerPackageProvider, error) {
	packageProvider := state.NewMultiManagerPackageProvider()

	for name, factory := range r.managers {
		manager := factory()
		available, err := manager.IsAvailable(ctx)
		if err != nil {
			// Log the error but continue without this manager
			// Skip problematic managers gracefully
			continue
		}
		if available {
			managerAdapter := state.NewManagerAdapter(manager)
			packageProvider.AddManager(name, managerAdapter, configLoader)
		}
	}

	return packageProvider, nil
}

// ManagerInfo holds information about a package manager
type ManagerInfo struct {
	Name      string
	Available bool
	Error     error
}

// GetManagerInfo returns detailed information about all managers including availability status
func (r *ManagerRegistry) GetManagerInfo(ctx context.Context) []ManagerInfo {
	info := make([]ManagerInfo, 0, len(r.managers))
	for name, factory := range r.managers {
		manager := factory()
		available, err := manager.IsAvailable(ctx)
		info = append(info, ManagerInfo{
			Name:      name,
			Available: err == nil && available,
			Error:     err,
		})
	}
	return info
}
