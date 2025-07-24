// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/state"
)

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
func (m *MultiManagerPackageProvider) GetConfiguredItems() ([]state.ConfigItem, error) {
	var allItems []state.ConfigItem

	for _, provider := range m.providers {
		items, err := provider.GetConfiguredItems()
		if err != nil {
			return nil, fmt.Errorf("failed to get configured items from %s: %w", provider.ManagerName(), err)
		}
		allItems = append(allItems, items...)
	}

	return allItems, nil
}

// GetActualItems returns installed packages from all managers
func (m *MultiManagerPackageProvider) GetActualItems(ctx context.Context) ([]state.ActualItem, error) {
	var allItems []state.ActualItem

	for _, provider := range m.providers {
		items, err := provider.GetActualItems(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get actual items from %s: %w", provider.ManagerName(), err)
		}
		allItems = append(allItems, items...)
	}

	return allItems, nil
}

// CreateItem creates an Item from package data
func (m *MultiManagerPackageProvider) CreateItem(name string, itemState state.ItemState, configured *state.ConfigItem, actual *state.ActualItem) state.Item {
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
		return provider.CreateItem(name, itemState, configured, actual)
	}

	// Fallback generic item creation
	item := state.Item{
		Name:    name,
		State:   itemState,
		Domain:  "package",
		Manager: managerName,
		Metadata: map[string]interface{}{
			"manager": managerName,
		},
	}

	return item
}
