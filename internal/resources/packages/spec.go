// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
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

	registry := NewManagerRegistry()
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
