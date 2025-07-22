// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/interfaces"
)

// Compile-time interface compliance checks
var _ DotfileConfigLoader = (*ConfigAdapter)(nil)
var _ PackageConfigLoader = (*ConfigAdapter)(nil)
var _ ManagerInterface = (*ManagerAdapter)(nil)

// ConfigAdapter bridges generic configuration implementations to the state package's
// specific interfaces (DotfileConfigLoader and PackageConfigLoader). This adapter
// allows the state package to work with any configuration source that implements
// the minimal ConfigInterface, without creating dependencies on specific config packages.
//
// Bridge: Generic ConfigInterface → state.DotfileConfigLoader + state.PackageConfigLoader
type ConfigAdapter struct {
	config ConfigInterface
}

// ConfigInterface defines the methods needed from the config package
type ConfigInterface interface {
	GetDotfileTargets() map[string]string
	GetHomebrewBrews() []string
	GetHomebrewCasks() []string
	GetNPMPackages() []string
	GetIgnorePatterns() []string
	GetExpandDirectories() []string
}

// NewConfigAdapter creates a new config adapter
func NewConfigAdapter(config ConfigInterface) *ConfigAdapter {
	return &ConfigAdapter{config: config}
}

// GetDotfileTargets implements DotfileConfigLoader
func (c *ConfigAdapter) GetDotfileTargets() map[string]string {
	return c.config.GetDotfileTargets()
}

// GetIgnorePatterns implements DotfileConfigLoader
func (c *ConfigAdapter) GetIgnorePatterns() []string {
	return c.config.GetIgnorePatterns()
}

// GetExpandDirectories implements DotfileConfigLoader
func (c *ConfigAdapter) GetExpandDirectories() []string {
	return c.config.GetExpandDirectories()
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

// ManagerAdapter provides backward compatibility for code that expects the state
// package's ManagerInterface. Since ManagerInterface is now an alias for
// interfaces.PackageManager after consolidation, this adapter serves as a
// pass-through wrapper to maintain existing API contracts while the codebase
// transitions to use the unified interface directly.
//
// Bridge: interfaces.PackageManager → state.ManagerInterface (alias)
//
// TODO: This adapter can be removed once all code directly uses interfaces.PackageManager
type ManagerAdapter struct {
	manager ManagerInterface
}

// ManagerInterface defines the methods needed from package managers
// This is an alias for the unified PackageManager interface to maintain backward compatibility.
type ManagerInterface = interfaces.PackageManager

// NewManagerAdapter creates a new manager adapter
func NewManagerAdapter(manager ManagerInterface) *ManagerAdapter {
	return &ManagerAdapter{manager: manager}
}

// Forward all methods to the underlying manager (since they have the same interface)

// IsAvailable implements PackageManager
func (m *ManagerAdapter) IsAvailable(ctx context.Context) (bool, error) {
	return m.manager.IsAvailable(ctx)
}

// ListInstalled implements PackageManager
func (m *ManagerAdapter) ListInstalled(ctx context.Context) ([]string, error) {
	return m.manager.ListInstalled(ctx)
}

// Install implements PackageManager
func (m *ManagerAdapter) Install(ctx context.Context, name string) error {
	return m.manager.Install(ctx, name)
}

// Uninstall implements PackageManager
func (m *ManagerAdapter) Uninstall(ctx context.Context, name string) error {
	return m.manager.Uninstall(ctx, name)
}

// IsInstalled implements PackageManager
func (m *ManagerAdapter) IsInstalled(ctx context.Context, name string) (bool, error) {
	return m.manager.IsInstalled(ctx, name)
}

// Search implements PackageManager
func (m *ManagerAdapter) Search(ctx context.Context, query string) ([]string, error) {
	return m.manager.Search(ctx, query)
}

// Info implements PackageManager
func (m *ManagerAdapter) Info(ctx context.Context, name string) (*interfaces.PackageInfo, error) {
	return m.manager.Info(ctx, name)
}

// GetInstalledVersion implements PackageManager
func (m *ManagerAdapter) GetInstalledVersion(ctx context.Context, name string) (string, error) {
	return m.manager.GetInstalledVersion(ctx, name)
}

// SupportsSearch implements PackageManagerCapabilities
func (m *ManagerAdapter) SupportsSearch() bool {
	return m.manager.SupportsSearch()
}
