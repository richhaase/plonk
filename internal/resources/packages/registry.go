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

// ManagerRegistry manages package manager creation and availability checking
type ManagerRegistry struct {
	mu         sync.RWMutex
	v2Managers map[string]config.ManagerConfig
	enableV2   bool
}

// defaultRegistry is the global registry instance
var defaultRegistry = &ManagerRegistry{
	v2Managers: make(map[string]config.ManagerConfig),
	enableV2:   true, // Feature flag for v2
}

// GetRegistry returns the shared manager registry instance
func GetRegistry() *ManagerRegistry {
	return defaultRegistry
}

// LoadManagerConfigs loads manager configs from Config
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

// GetManager returns a package manager instance by name using the default executor
func (r *ManagerRegistry) GetManager(name string) (PackageManager, error) {
	return r.GetManagerWithExecutor(name, nil)
}

// GetManagerWithExecutor returns a package manager instance with an injected executor.
// If exec is nil, uses the default executor. Checks config-defined managers first.
func (r *ManagerRegistry) GetManagerWithExecutor(name string, exec CommandExecutor) (PackageManager, error) {
	if exec == nil {
		exec = defaultExecutor
	}

	// Only config-defined managers are supported
	r.mu.RLock()
	v2Config, hasV2 := r.v2Managers[name]
	enableV2 := r.enableV2
	r.mu.RUnlock()

	if enableV2 && hasV2 {
		return NewGenericManager(v2Config, exec), nil
	}

	return nil, fmt.Errorf("unsupported package manager: %s", name)
}

// GetAllManagerNames returns all registered manager names sorted alphabetically
func (r *ManagerRegistry) GetAllManagerNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.v2Managers))
	for name := range r.v2Managers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// HasManager checks if a manager is supported by the registry
func (r *ManagerRegistry) HasManager(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.v2Managers[name]
	return exists
}

// ManagerInfo holds information about a package manager
type ManagerInfo struct {
	Name      string
	Available bool
	Error     error
}
