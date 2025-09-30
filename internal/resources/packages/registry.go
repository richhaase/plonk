// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"fmt"
	"sort"
	"sync"
)

// DefaultManager is the fallback manager when none is configured
const DefaultManager = "brew"

// ManagerFactory defines a function that creates a package manager instance
type ManagerFactory func() PackageManager

// ManagerFactoryV2 defines a function that creates a package manager with an injected executor
type ManagerFactoryV2 func(CommandExecutor) PackageManager

// managerEntry holds both v1 and v2 factories for a package manager
type managerEntry struct {
	v1 ManagerFactory
	v2 ManagerFactoryV2
}

// ManagerRegistry manages package manager creation and availability checking
type ManagerRegistry struct {
	mu       sync.RWMutex
	managers map[string]*managerEntry
}

// defaultRegistry is the global registry instance
var defaultRegistry = &ManagerRegistry{
	managers: make(map[string]*managerEntry),
}

// RegisterManager registers a package manager with the global registry.
// This is typically called from init() functions in package manager implementations.
func RegisterManager(name string, factory ManagerFactory) {
	defaultRegistry.Register(name, factory)
}

// RegisterManagerV2 registers a V2 package manager with the global registry.
// V2 managers accept an executor for dependency injection.
func RegisterManagerV2(name string, factory ManagerFactoryV2) {
	defaultRegistry.RegisterV2(name, factory)
}

// Register adds a V1 manager to the registry
func (r *ManagerRegistry) Register(name string, factory ManagerFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := r.managers[name]
	if entry == nil {
		entry = &managerEntry{}
		r.managers[name] = entry
	}
	entry.v1 = factory
}

// RegisterV2 adds a V2 manager to the registry
func (r *ManagerRegistry) RegisterV2(name string, factory ManagerFactoryV2) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := r.managers[name]
	if entry == nil {
		entry = &managerEntry{}
		r.managers[name] = entry
	}
	entry.v2 = factory
}

// NewManagerRegistry creates a new manager registry that uses the global registrations
func NewManagerRegistry() *ManagerRegistry {
	return defaultRegistry
}

// GetManager returns a package manager instance by name using the default executor
func (r *ManagerRegistry) GetManager(name string) (PackageManager, error) {
	return r.GetManagerWithExecutor(name, nil)
}

// GetManagerWithExecutor returns a package manager instance with an injected executor.
// If exec is nil, uses the default executor. Prefers V2 factories over V1.
func (r *ManagerRegistry) GetManagerWithExecutor(name string, exec CommandExecutor) (PackageManager, error) {
	r.mu.RLock()
	entry, exists := r.managers[name]
	r.mu.RUnlock()

	if !exists || entry == nil {
		return nil, fmt.Errorf("unsupported package manager: %s", name)
	}

	if exec == nil {
		exec = defaultExecutor
	}

	// Prefer V2 factory if present
	if entry.v2 != nil {
		return entry.v2(exec), nil
	}

	// Fall back to V1 factory
	if entry.v1 != nil {
		return entry.v1(), nil
	}

	return nil, fmt.Errorf("no factory registered for package manager: %s", name)
}

// GetAllManagerNames returns all registered manager names sorted alphabetically
func (r *ManagerRegistry) GetAllManagerNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.managers))
	for name := range r.managers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// HasManager checks if a manager is supported by the registry
func (r *ManagerRegistry) HasManager(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.managers[name]
	return exists
}

// ManagerInfo holds information about a package manager
type ManagerInfo struct {
	Name      string
	Available bool
	Error     error
}
