// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/resources"
)

// PackageResource adapts package managers to the Resource interface
type PackageResource struct {
	manager PackageManager
	desired []resources.Item
}

// NewPackageResource creates a new package resource adapter
func NewPackageResource(manager PackageManager) *PackageResource {
	return &PackageResource{manager: manager}
}

// ID returns a unique identifier for this resource
func (p *PackageResource) ID() string {
	// Determine manager name based on type
	switch p.manager.(type) {
	case *HomebrewManager:
		return "packages:homebrew"
	case *NpmManager:
		return "packages:npm"
	case *CargoManager:
		return "packages:cargo"
	case *PipManager:
		return "packages:pip"
	case *GemManager:
		return "packages:gem"
	case *GoInstallManager:
		return "packages:go"
	default:
		return "packages:unknown"
	}
}

// Desired returns the desired state (packages that should be installed)
func (p *PackageResource) Desired() []resources.Item {
	return p.desired
}

// SetDesired sets the desired packages for this manager
func (p *PackageResource) SetDesired(items []resources.Item) {
	p.desired = items
}

// Actual returns the actual state (packages currently installed)
func (p *PackageResource) Actual(ctx context.Context) []resources.Item {
	// Check if manager is available
	available, err := p.manager.IsAvailable(ctx)
	if err != nil || !available {
		return []resources.Item{}
	}

	// Get installed packages
	installed, err := p.manager.ListInstalled(ctx)
	if err != nil {
		return []resources.Item{}
	}

	// Convert to resources.Item
	items := make([]resources.Item, 0, len(installed))
	for _, pkgName := range installed {
		item := resources.Item{
			Name:   pkgName,
			State:  resources.StateUntracked, // Will be updated during reconciliation
			Domain: "package",
		}

		// Get version if possible
		version, err := p.manager.InstalledVersion(ctx, pkgName)
		if err == nil && version != "" {
			if item.Metadata == nil {
				item.Metadata = make(map[string]interface{})
			}
			item.Metadata["version"] = version
		}

		items = append(items, item)
	}

	return items
}

// Apply performs the necessary action to move an item to its desired state
func (p *PackageResource) Apply(ctx context.Context, item resources.Item) error {
	switch item.State {
	case resources.StateMissing:
		// Package should be installed
		return p.manager.Install(ctx, item.Name)
	case resources.StateUntracked:
		// Package should be uninstalled
		return p.manager.Uninstall(ctx, item.Name)
	case resources.StateManaged:
		// Package is already in desired state, nothing to do
		return nil
	default:
		return fmt.Errorf("unknown item state: %v", item.State)
	}
}

// MultiPackageResource manages multiple package managers as a single resource
type MultiPackageResource struct {
	resources map[string]*PackageResource
	registry  *ManagerRegistry
}

// NewMultiPackageResource creates a resource that manages all package managers
func NewMultiPackageResource() *MultiPackageResource {
	return &MultiPackageResource{
		resources: make(map[string]*PackageResource),
		registry:  NewManagerRegistry(),
	}
}

// ID returns a unique identifier for this resource
func (m *MultiPackageResource) ID() string {
	return "packages:all"
}

// Desired returns all desired packages across all managers
func (m *MultiPackageResource) Desired() []resources.Item {
	var items []resources.Item
	for _, resource := range m.resources {
		items = append(items, resource.Desired()...)
	}
	return items
}

// SetDesired distributes desired items to appropriate package resources
func (m *MultiPackageResource) SetDesired(items []resources.Item) {
	// Clear existing desired state
	m.resources = make(map[string]*PackageResource)

	// Group items by manager
	byManager := make(map[string][]resources.Item)
	for _, item := range items {
		if item.Manager != "" {
			byManager[item.Manager] = append(byManager[item.Manager], item)
		}
	}

	// Create resources for each manager with desired items
	for managerName, managerItems := range byManager {
		manager, err := m.registry.GetManager(managerName)
		if err != nil {
			continue // Skip invalid managers
		}
		resource := NewPackageResource(manager)
		resource.SetDesired(managerItems)
		m.resources[managerName] = resource
	}
}

// Actual returns all actual packages across all available managers
func (m *MultiPackageResource) Actual(ctx context.Context) []resources.Item {
	var items []resources.Item

	// Get packages from all available managers
	for managerName := range m.registry.managers {
		manager, err := m.registry.GetManager(managerName)
		if err != nil {
			continue
		}

		// Create a temporary resource to get actual state
		resource := NewPackageResource(manager)
		actualItems := resource.Actual(ctx)

		// Set the manager name on each item
		for i := range actualItems {
			actualItems[i].Manager = managerName
		}

		items = append(items, actualItems...)
	}

	return items
}

// Apply performs the necessary action to move an item to its desired state
func (m *MultiPackageResource) Apply(ctx context.Context, item resources.Item) error {
	if item.Manager == "" {
		return fmt.Errorf("package item missing manager information")
	}

	manager, err := m.registry.GetManager(item.Manager)
	if err != nil {
		return fmt.Errorf("getting manager %s: %w", item.Manager, err)
	}

	resource := NewPackageResource(manager)
	return resource.Apply(ctx, item)
}
