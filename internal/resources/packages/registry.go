// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"fmt"
	"sort"
	"sync"

	"github.com/richhaase/plonk/internal/config"
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
	mu         sync.RWMutex
	managers   map[string]*managerEntry
	v2Managers map[string]config.ManagerConfig
	enableV2   bool
}

// defaultRegistry is the global registry instance
var defaultRegistry = &ManagerRegistry{
	managers:   make(map[string]*managerEntry),
	v2Managers: make(map[string]config.ManagerConfig),
	enableV2:   true, // Feature flag for v2
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

// LoadV2Configs loads v2 manager configs from Config
// It resets the registry and loads defaults first, then merges user configs
func (r *ManagerRegistry) LoadV2Configs(cfg *config.Config) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Reset and load defaults first
	r.v2Managers = make(map[string]config.ManagerConfig)
	for name, managerCfg := range config.GetDefaultManagers() {
		r.v2Managers[name] = managerCfg
	}

	// Then merge/override with user configs
	if cfg != nil && cfg.Managers != nil {
		for name, managerCfg := range cfg.Managers {
			r.v2Managers[name] = managerCfg
		}
	}
}

// EnableV2 enables v2 manager configs
func (r *ManagerRegistry) EnableV2(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enableV2 = enabled
}

// GetManager returns a package manager instance by name using the default executor
func (r *ManagerRegistry) GetManager(name string) (PackageManager, error) {
	return r.GetManagerWithExecutor(name, nil)
}

// GetManagerWithExecutor returns a package manager instance with an injected executor.
// If exec is nil, uses the default executor. Checks v2 configs first, then factories.
func (r *ManagerRegistry) GetManagerWithExecutor(name string, exec CommandExecutor) (PackageManager, error) {
	if exec == nil {
		exec = defaultExecutor
	}

	// Check v2 config first (if enabled)
	r.mu.RLock()
	v2Config, hasV2 := r.v2Managers[name]
	enableV2 := r.enableV2
	r.mu.RUnlock()

	if enableV2 && hasV2 {
		return NewGenericManager(v2Config, exec), nil
	}

	// Fall back to Go factory
	r.mu.RLock()
	entry, exists := r.managers[name]
	r.mu.RUnlock()

	if !exists || entry == nil {
		return nil, fmt.Errorf("unsupported package manager: %s", name)
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

	nameSet := make(map[string]bool)
	for name := range r.v2Managers {
		nameSet[name] = true
	}
	for name := range r.managers {
		nameSet[name] = true
	}

	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// HasManager checks if a manager is supported by the registry
func (r *ManagerRegistry) HasManager(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, exists := r.v2Managers[name]; exists {
		return true
	}
	_, exists := r.managers[name]
	return exists
}

// ManagerInfo holds information about a package manager
type ManagerInfo struct {
	Name      string
	Available bool
	Error     error
}
