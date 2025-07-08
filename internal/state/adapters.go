// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"fmt"
)

// ConfigAdapter adapts existing config types to the new state interfaces
type ConfigAdapter struct {
	config ConfigInterface
}

// ConfigInterface defines the methods needed from the config package
type ConfigInterface interface {
	GetDotfileTargets() map[string]string
	GetHomebrewBrews() []string
	GetHomebrewCasks() []string  
	GetNPMPackages() []string
}

// NewConfigAdapter creates a new config adapter
func NewConfigAdapter(config ConfigInterface) *ConfigAdapter {
	return &ConfigAdapter{config: config}
}

// GetDotfileTargets implements DotfileConfigLoader
func (c *ConfigAdapter) GetDotfileTargets() map[string]string {
	return c.config.GetDotfileTargets()
}

// GetPackagesForManager implements PackageConfigLoader
func (c *ConfigAdapter) GetPackagesForManager(managerName string) ([]PackageConfigItem, error) {
	var packageNames []string
	
	switch managerName {
	case "homebrew":
		// Combine brews and casks for homebrew
		brews := c.config.GetHomebrewBrews()
		casks := c.config.GetHomebrewCasks()
		packageNames = append(packageNames, brews...)
		packageNames = append(packageNames, casks...)
	case "npm":
		packageNames = c.config.GetNPMPackages()
	default:
		return nil, fmt.Errorf("unknown package manager: %s", managerName)
	}
	
	items := make([]PackageConfigItem, len(packageNames))
	for i, name := range packageNames {
		items[i] = PackageConfigItem{Name: name}
	}
	
	return items, nil
}

// ManagerAdapter adapts existing package manager types to the new state interface
type ManagerAdapter struct {
	manager ManagerInterface
}

// ManagerInterface defines the methods needed from package managers
type ManagerInterface interface {
	IsAvailable() bool
	ListInstalled() ([]string, error)
	Install(name string) error
	Uninstall(name string) error
	IsInstalled(name string) bool
}

// NewManagerAdapter creates a new manager adapter
func NewManagerAdapter(manager ManagerInterface) *ManagerAdapter {
	return &ManagerAdapter{manager: manager}
}

// IsAvailable implements PackageManager
func (m *ManagerAdapter) IsAvailable() bool {
	return m.manager.IsAvailable()
}

// ListInstalled implements PackageManager
func (m *ManagerAdapter) ListInstalled() ([]string, error) {
	return m.manager.ListInstalled()
}

// Install implements PackageManager
func (m *ManagerAdapter) Install(name string) error {
	return m.manager.Install(name)
}

// Uninstall implements PackageManager
func (m *ManagerAdapter) Uninstall(name string) error {
	return m.manager.Uninstall(name)
}

// IsInstalled implements PackageManager
func (m *ManagerAdapter) IsInstalled(name string) bool {
	return m.manager.IsInstalled(name)
}