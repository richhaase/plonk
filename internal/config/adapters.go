// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"plonk/internal/state"
)

// StatePackageConfigAdapter adapts ConfigAdapter to work with state.PackageConfigLoader
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

// StateDotfileConfigAdapter adapts ConfigAdapter to work with state.DotfileConfigLoader
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
	// Get the underlying config and call its GetIgnorePatterns method
	return s.configAdapter.config.GetIgnorePatterns()
}

// GetExpandDirectories implements state.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetExpandDirectories() []string {
	// Get the underlying config and call its GetExpandDirectories method
	return s.configAdapter.config.GetExpandDirectories()
}
