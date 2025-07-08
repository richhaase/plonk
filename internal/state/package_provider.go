// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"fmt"
)

// PackageManager defines the interface for package managers (from managers package)
type PackageManager interface {
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	Install(ctx context.Context, name string) error
	Uninstall(ctx context.Context, name string) error
	IsInstalled(ctx context.Context, name string) (bool, error)
}

// PackageConfigLoader defines how to load package configuration
type PackageConfigLoader interface {
	GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
}

// PackageConfigItem represents a package from configuration
type PackageConfigItem struct {
	Name string
}

// PackageProvider implements the Provider interface for package management
type PackageProvider struct {
	managerName   string
	manager       PackageManager
	configLoader  PackageConfigLoader
}

// NewPackageProvider creates a new package provider for a specific manager
func NewPackageProvider(managerName string, manager PackageManager, configLoader PackageConfigLoader) *PackageProvider {
	return &PackageProvider{
		managerName:  managerName,
		manager:      manager,
		configLoader: configLoader,
	}
}

// Domain returns the domain name for packages
func (p *PackageProvider) Domain() string {
	return "package"
}

// GetConfiguredItems returns packages defined in configuration
func (p *PackageProvider) GetConfiguredItems() ([]ConfigItem, error) {
	packages, err := p.configLoader.GetPackagesForManager(p.managerName)
	if err != nil {
		return nil, err
	}
	
	items := make([]ConfigItem, len(packages))
	for i, pkg := range packages {
		items[i] = ConfigItem{
			Name: pkg.Name,
			Metadata: map[string]interface{}{
				"manager": p.managerName,
			},
		}
	}
	
	return items, nil
}

// GetActualItems returns packages currently installed by this manager
func (p *PackageProvider) GetActualItems(ctx context.Context) ([]ActualItem, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	available, err := p.manager.IsAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check manager availability: %w", err)
	}
	if !available {
		return []ActualItem{}, nil
	}
	
	installed, err := p.manager.ListInstalled(ctx)
	if err != nil {
		return nil, err
	}
	
	items := make([]ActualItem, len(installed))
	for i, pkg := range installed {
		items[i] = ActualItem{
			Name: pkg,
			Metadata: map[string]interface{}{
				"manager": p.managerName,
			},
		}
	}
	
	return items, nil
}

// CreateItem creates an Item from package data
func (p *PackageProvider) CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item {
	item := Item{
		Name:    name,
		State:   state,
		Domain:  "package",
		Manager: p.managerName,
		Metadata: map[string]interface{}{
			"manager": p.managerName,
		},
	}
	
	// Add any additional metadata from configured or actual
	if configured != nil {
		for k, v := range configured.Metadata {
			item.Metadata[k] = v
		}
	}
	if actual != nil {
		for k, v := range actual.Metadata {
			item.Metadata[k] = v
		}
	}
	
	return item
}

// MultiManagerPackageProvider aggregates multiple package managers
type MultiManagerPackageProvider struct {
	providers map[string]*PackageProvider
}

// NewMultiManagerPackageProvider creates a provider that handles multiple package managers
func NewMultiManagerPackageProvider() *MultiManagerPackageProvider {
	return &MultiManagerPackageProvider{
		providers: make(map[string]*PackageProvider),
	}
}

// AddManager adds a package manager to the multi-manager provider
func (m *MultiManagerPackageProvider) AddManager(managerName string, manager PackageManager, configLoader PackageConfigLoader) {
	m.providers[managerName] = NewPackageProvider(managerName, manager, configLoader)
}

// Domain returns the domain name for packages
func (m *MultiManagerPackageProvider) Domain() string {
	return "package"
}

// GetConfiguredItems returns packages from all managers
func (m *MultiManagerPackageProvider) GetConfiguredItems() ([]ConfigItem, error) {
	var allItems []ConfigItem
	
	for _, provider := range m.providers {
		items, err := provider.GetConfiguredItems()
		if err != nil {
			return nil, fmt.Errorf("failed to get configured items from %s: %w", provider.managerName, err)
		}
		allItems = append(allItems, items...)
	}
	
	return allItems, nil
}

// GetActualItems returns installed packages from all managers
func (m *MultiManagerPackageProvider) GetActualItems(ctx context.Context) ([]ActualItem, error) {
	var allItems []ActualItem
	
	for _, provider := range m.providers {
		items, err := provider.GetActualItems(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get actual items from %s: %w", provider.managerName, err)
		}
		allItems = append(allItems, items...)
	}
	
	return allItems, nil
}

// CreateItem creates an Item from package data
func (m *MultiManagerPackageProvider) CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item {
	// Determine which manager this package belongs to
	managerName := "unknown"
	if configured != nil {
		if mgr, ok := configured.Metadata["manager"].(string); ok {
			managerName = mgr
		}
	}
	if actual != nil {
		if mgr, ok := actual.Metadata["manager"].(string); ok {
			managerName = mgr
		}
	}
	
	// Delegate to the specific provider if available
	if provider, exists := m.providers[managerName]; exists {
		return provider.CreateItem(name, state, configured, actual)
	}
	
	// Fallback generic item creation
	item := Item{
		Name:    name,
		State:   state,
		Domain:  "package",
		Manager: managerName,
		Metadata: map[string]interface{}{
			"manager": managerName,
		},
	}
	
	return item
}