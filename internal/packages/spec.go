// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"strings"
)

// PackageSpec represents a parsed package specification
type PackageSpec struct {
	Name         string // Package name (required)
	Manager      string // Package manager (can be empty)
	OriginalSpec string // Original input for error messages
}

// ParsePackageSpec parses a package specification string
// Format: [manager:]package
// Examples: "git", "brew:wget", "npm:@types/node"
func ParsePackageSpec(spec string) (*PackageSpec, error) {
	if spec == "" {
		return nil, fmt.Errorf("package specification cannot be empty")
	}

	parts := strings.SplitN(spec, ":", 2)
	ps := &PackageSpec{OriginalSpec: spec}

	if len(parts) == 2 {
		ps.Manager = parts[0]
		ps.Name = parts[1]

		// Check for empty manager prefix (":package")
		if ps.Manager == "" {
			return nil, fmt.Errorf("manager prefix cannot be empty")
		}
	} else {
		ps.Name = spec
	}

	// Check for empty package name
	if ps.Name == "" {
		return nil, fmt.Errorf("package name cannot be empty")
	}

	return ps, nil
}

// ValidateManager checks if the manager is valid (if specified)
func (ps *PackageSpec) ValidateManager() error {
	if ps.Manager == "" {
		return nil // Empty manager is valid (will be resolved later)
	}

	registry := GetRegistry()
	if !registry.HasManager(ps.Manager) {
		return fmt.Errorf("unknown package manager %q", ps.Manager)
	}

	return nil
}

// ResolveManager sets the manager based on config if not already set
func (ps *PackageSpec) ResolveManager(defaultManager string) {
	if ps.Manager == "" && defaultManager != "" {
		ps.Manager = defaultManager
	}
}

// RequireManager ensures a manager is set, using the default if needed
func (ps *PackageSpec) RequireManager(defaultManager string) error {
	ps.ResolveManager(defaultManager)

	if ps.Manager == "" {
		return fmt.Errorf("no package manager specified and no default configured")
	}

	return ps.ValidateManager()
}

// String returns the canonical string representation
func (ps *PackageSpec) String() string {
	if ps.Manager != "" {
		return fmt.Sprintf("%s:%s", ps.Manager, ps.Name)
	}
	return ps.Name
}

// Key returns a unique key for reconciliation (manager:name)
func (ps *PackageSpec) Key() string {
	return ps.Manager + ":" + ps.Name
}

// ReconcileResult holds the result of package reconciliation
type ReconcileResult struct {
	Managed   []PackageSpec // In lock file AND installed
	Missing   []PackageSpec // In lock file BUT NOT installed
	Untracked []PackageSpec // Installed BUT NOT in lock file
}

// GetDesiredFromLock extracts desired packages from lock file resources.
// Each resource with type "package" is converted to a PackageSpec.
func GetDesiredFromLock(resources []LockResource) []PackageSpec {
	var specs []PackageSpec
	for _, r := range resources {
		if r.Type != "package" {
			continue
		}
		manager, _ := r.Metadata["manager"].(string)
		name, _ := r.Metadata["name"].(string)
		if manager != "" && name != "" {
			specs = append(specs, PackageSpec{
				Name:    name,
				Manager: manager,
			})
		}
	}
	return specs
}

// LockResource represents a resource entry from the lock file.
// This is a simplified view of lock.ResourceEntry for package reconciliation.
type LockResource struct {
	Type     string
	Metadata map[string]interface{}
}

// GetActualInstalled queries all available package managers for installed packages.
// Returns a list of PackageSpecs representing currently installed packages.
func GetActualInstalled(ctx context.Context, registry *ManagerRegistry) []PackageSpec {
	if registry == nil {
		registry = GetRegistry()
	}

	var specs []PackageSpec
	for _, managerName := range registry.GetAllManagerNames() {
		manager, err := registry.GetManager(managerName)
		if err != nil {
			continue
		}

		available, err := manager.IsAvailable(ctx)
		if err != nil || !available {
			continue
		}

		installed, err := manager.ListInstalled(ctx)
		if err != nil {
			continue
		}

		for _, pkgName := range installed {
			specs = append(specs, PackageSpec{
				Name:    pkgName,
				Manager: managerName,
			})
		}
	}

	return specs
}

// ReconcileFromLock loads the lock file and compares against installed packages.
// This is the main entry point for package reconciliation.
func ReconcileFromLock(ctx context.Context, lockResources []LockResource, registry *ManagerRegistry) ReconcileResult {
	desired := GetDesiredFromLock(lockResources)
	actual := GetActualInstalled(ctx, registry)
	return Reconcile(desired, actual)
}

// Reconcile compares desired packages (from lock file) against actual installed packages.
// Returns which packages are managed, missing, or untracked.
func Reconcile(desired, actual []PackageSpec) ReconcileResult {
	// Build lookup sets
	desiredSet := make(map[string]PackageSpec)
	for _, pkg := range desired {
		desiredSet[pkg.Key()] = pkg
	}

	actualSet := make(map[string]PackageSpec)
	for _, pkg := range actual {
		actualSet[pkg.Key()] = pkg
	}

	var result ReconcileResult

	// Find managed and missing (iterate over desired)
	for key, pkg := range desiredSet {
		if _, exists := actualSet[key]; exists {
			result.Managed = append(result.Managed, pkg)
		} else {
			result.Missing = append(result.Missing, pkg)
		}
	}

	// Find untracked (iterate over actual)
	for key, pkg := range actualSet {
		if _, exists := desiredSet[key]; !exists {
			result.Untracked = append(result.Untracked, pkg)
		}
	}

	return result
}
