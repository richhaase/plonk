// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/richhaase/plonk/internal/resources"
	"golang.org/x/sync/errgroup"
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

		// Skip version checking during reconciliation for performance and to avoid hangs
		// Version information is not needed for determining what packages are installed
		// TODO: If version info is needed later, implement batch version checking or make it optional

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
	var (
		mu    sync.Mutex
		items []resources.Item
	)

	// Limit concurrency to min(GOMAXPROCS, 4)
	maxWorkers := runtime.GOMAXPROCS(0)
	if maxWorkers > 4 {
		maxWorkers = 4
	}

	// Create errgroup with limited concurrency
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxWorkers)

	// Get packages from all available managers in parallel
	for managerName := range m.registry.managers {
		managerName := managerName // capture loop variable

		g.Go(func() error {
			// Check context cancellation
			select {
			case <-gctx.Done():
				return gctx.Err()
			default:
			}

			manager, err := m.registry.GetManager(managerName)
			if err != nil {
				return nil
			}

			// Create a temporary resource to get actual state
			resource := NewPackageResource(manager)
			actualItems := resource.Actual(gctx)

			// Set the manager name on each item
			for i := range actualItems {
				actualItems[i].Manager = managerName
			}

			// Thread-safe append
			mu.Lock()
			items = append(items, actualItems...)
			mu.Unlock()

			return nil
		})
	}

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		return []resources.Item{}
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
