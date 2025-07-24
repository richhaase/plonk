// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/state"
)

// PackageProvider manages package state for a single package manager
type PackageProvider struct {
	managerName  string
	manager      PackageManager
	configLoader PackageConfigLoader
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

// ManagerName returns the name of the package manager
func (p *PackageProvider) ManagerName() string {
	return p.managerName
}

// GetConfiguredItems returns packages defined in configuration
func (p *PackageProvider) GetConfiguredItems() ([]state.ConfigItem, error) {
	packages, err := p.configLoader.GetPackagesForManager(p.managerName)
	if err != nil {
		return nil, err
	}

	items := make([]state.ConfigItem, len(packages))
	for i, pkg := range packages {
		items[i] = state.ConfigItem{
			Name: pkg.Name,
			Metadata: map[string]interface{}{
				"manager": p.managerName,
			},
		}
	}

	return items, nil
}

// GetActualItems returns packages currently installed by this manager
func (p *PackageProvider) GetActualItems(ctx context.Context) ([]state.ActualItem, error) {
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
		return []state.ActualItem{}, nil
	}

	installed, err := p.manager.ListInstalled(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]state.ActualItem, len(installed))
	for i, pkg := range installed {
		items[i] = state.ActualItem{
			Name: pkg,
			Metadata: map[string]interface{}{
				"manager": p.managerName,
			},
		}
	}

	return items, nil
}

// CreateItem creates an Item from package data
func (p *PackageProvider) CreateItem(name string, itemState state.ItemState, configured *state.ConfigItem, actual *state.ActualItem) state.Item {
	item := state.Item{
		Name:    name,
		State:   itemState,
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
