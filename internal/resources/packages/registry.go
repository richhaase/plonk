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

// SupportedManagers returns all registered manager names.
// Deprecated: Use NewManagerRegistry().GetAllManagerNames() instead.
// This is kept for backward compatibility.
var SupportedManagers = []string{
	"brew",
	"cargo",
	"gem",
	"go",
	"npm",
	"pip",
}

// ManagerFactory defines a function that creates a package manager instance
type ManagerFactory func() PackageManager

// ManagerRegistry manages package manager creation and availability checking
type ManagerRegistry struct {
	mu       sync.RWMutex
	managers map[string]ManagerFactory
}

// defaultRegistry is the global registry instance
var defaultRegistry = &ManagerRegistry{
	managers: make(map[string]ManagerFactory),
}

// RegisterManager registers a package manager with the global registry.
// This is typically called from init() functions in package manager implementations.
func RegisterManager(name string, factory ManagerFactory) {
	defaultRegistry.Register(name, factory)
}

// Register adds a manager to the registry
func (r *ManagerRegistry) Register(name string, factory ManagerFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.managers[name] = factory
}

// NewManagerRegistry creates a new manager registry that uses the global registrations
func NewManagerRegistry() *ManagerRegistry {
	return defaultRegistry
}

// GetManager returns a package manager instance by name
func (r *ManagerRegistry) GetManager(name string) (PackageManager, error) {
	r.mu.RLock()
	factory, exists := r.managers[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported package manager: %s", name)
	}
	return factory(), nil
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
