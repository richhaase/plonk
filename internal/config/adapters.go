// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"github.com/richhaase/plonk/internal/state"
)

// Compile-time interface compliance checks
var _ state.PackageConfigLoader = (*StatePackageConfigAdapter)(nil)
var _ state.DotfileConfigLoader = (*StateDotfileConfigAdapter)(nil)

// StatePackageConfigAdapter bridges the config package's ConfigAdapter to the state
// package's PackageConfigLoader interface. This adapter prevents circular dependencies
// between the config and state packages, allowing the state package to consume
// configuration data without directly importing the config package.
//
// Bridge: config.ConfigAdapter → state.PackageConfigLoader
type StatePackageConfigAdapter struct {
	configAdapter *ConfigAdapter
}

// NewStatePackageConfigAdapter creates a new adapter for state package interfaces
func NewStatePackageConfigAdapter(configAdapter *ConfigAdapter) *StatePackageConfigAdapter {
	return &StatePackageConfigAdapter{configAdapter: configAdapter}
}

// GetPackagesForManager implements state.PackageConfigLoader interface
func (s *StatePackageConfigAdapter) GetPackagesForManager(managerName string) ([]state.PackageConfigItem, error) {
	items, err := s.configAdapter.GetPackagesForManager(managerName)
	if err != nil {
		return nil, err
	}

	// Convert config.PackageConfigItem to state.PackageConfigItem
	stateItems := make([]state.PackageConfigItem, len(items))
	for i, item := range items {
		stateItems[i] = state.PackageConfigItem{Name: item.Name}
	}

	return stateItems, nil
}

// StateDotfileConfigAdapter bridges the config package's ConfigAdapter to the state
// package's DotfileConfigLoader interface. This adapter prevents circular dependencies
// between the config and state packages, allowing the state package to consume
// dotfile configuration without directly importing the config package.
//
// Bridge: config.ConfigAdapter → state.DotfileConfigLoader
type StateDotfileConfigAdapter struct {
	configAdapter *ConfigAdapter
}

// NewStateDotfileConfigAdapter creates a new adapter for state dotfile interfaces
func NewStateDotfileConfigAdapter(configAdapter *ConfigAdapter) *StateDotfileConfigAdapter {
	return &StateDotfileConfigAdapter{configAdapter: configAdapter}
}

// GetDotfileTargets implements state.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetDotfileTargets() map[string]string {
	return s.configAdapter.GetDotfileTargets()
}

// GetIgnorePatterns implements state.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetIgnorePatterns() []string {
	// Get the resolved config and call its GetIgnorePatterns method
	resolvedConfig := s.configAdapter.config.Resolve()
	return resolvedConfig.GetIgnorePatterns()
}

// GetExpandDirectories implements state.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetExpandDirectories() []string {
	// Get the resolved config and call its GetExpandDirectories method
	resolvedConfig := s.configAdapter.config.Resolve()
	return resolvedConfig.GetExpandDirectories()
}
