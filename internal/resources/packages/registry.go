// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"fmt"
)

// SupportedManagers contains all package packages supported by plonk
var SupportedManagers = []string{
	"apt",
	"brew",
	"cargo",
	"gem",
	"go",
	"npm",
	"pip",
}

// DefaultManager is the fallback manager when none is configured
const DefaultManager = "brew"

// ManagerFactory defines a function that creates a package manager instance
type ManagerFactory func() PackageManager

// ManagerRegistry manages package manager creation and availability checking
type ManagerRegistry struct {
	managers map[string]ManagerFactory
}

// NewManagerRegistry creates a new manager registry with all supported package packages
func NewManagerRegistry() *ManagerRegistry {
	return &ManagerRegistry{
		managers: map[string]ManagerFactory{
			"apt":   func() PackageManager { return NewAptManager() },
			"brew":  func() PackageManager { return NewHomebrewManager() },
			"npm":   func() PackageManager { return NewNpmManager() },
			"cargo": func() PackageManager { return NewCargoManager() },
			"pip":   func() PackageManager { return NewPipManager() },
			"gem":   func() PackageManager { return NewGemManager() },
			"go":    func() PackageManager { return NewGoInstallManager() },
		},
	}
}

// GetManager returns a package manager instance by name
func (r *ManagerRegistry) GetManager(name string) (PackageManager, error) {
	factory, exists := r.managers[name]
	if !exists {
		return nil, fmt.Errorf("unsupported package manager: %s", name)
	}
	return factory(), nil
}

// GetAllManagerNames returns all supported manager names regardless of availability
func (r *ManagerRegistry) GetAllManagerNames() []string {
	// Return a copy to prevent external modification
	names := make([]string, len(SupportedManagers))
	copy(names, SupportedManagers)
	return names
}

// ManagerInfo holds information about a package manager
type ManagerInfo struct {
	Name      string
	Available bool
	Error     error
}
