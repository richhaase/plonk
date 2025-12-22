// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"fmt"
	"sort"
)

// DefaultManager is the fallback manager when none is configured
const DefaultManager = "brew"

// supportedManagers lists all available package managers
var supportedManagers = []string{"brew", "cargo", "go", "npm", "pnpm", "bun", "uv"}

// ManagerRegistry manages package manager creation
type ManagerRegistry struct{}

// defaultRegistry is the global registry instance
var defaultRegistry = &ManagerRegistry{}

// GetRegistry returns the shared manager registry instance
func GetRegistry() *ManagerRegistry {
	return defaultRegistry
}

// GetManager returns a package manager instance by name using the default executor
func (r *ManagerRegistry) GetManager(name string) (PackageManager, error) {
	return r.GetManagerWithExecutor(name, nil)
}

// GetManagerWithExecutor returns a package manager instance with an injected executor.
func (r *ManagerRegistry) GetManagerWithExecutor(name string, exec CommandExecutor) (PackageManager, error) {
	if exec == nil {
		exec = defaultExecutor
	}

	switch name {
	case "brew":
		return NewBrewManager(exec), nil
	case "cargo":
		return NewCargoManager(exec), nil
	case "go":
		return NewGoManager(exec), nil
	case "npm":
		return NewNPMManager(ProviderNPM, exec), nil
	case "pnpm":
		return NewNPMManager(ProviderPNPM, exec), nil
	case "bun":
		return NewNPMManager(ProviderBun, exec), nil
	case "uv":
		return NewUVManager(exec), nil
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", name)
	}
}

// GetAllManagerNames returns all registered manager names sorted alphabetically
func (r *ManagerRegistry) GetAllManagerNames() []string {
	names := make([]string, len(supportedManagers))
	copy(names, supportedManagers)
	sort.Strings(names)
	return names
}

// HasManager checks if a manager is supported by the registry
func (r *ManagerRegistry) HasManager(name string) bool {
	for _, m := range supportedManagers {
		if m == name {
			return true
		}
	}
	return false
}

// ManagerInfo holds information about a package manager
type ManagerInfo struct {
	Name      string
	Available bool
	Error     error
}
