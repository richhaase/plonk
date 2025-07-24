// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import "context"

// PackageManagerCapabilities describes which optional operations are supported by a package manager.
// This allows callers to check capabilities before attempting operations that may not be available.
type PackageManagerCapabilities interface {
	// SupportsSearch returns true if the package manager supports searching for packages.
	// If false, calling Search will return ErrOperationNotSupported.
	SupportsSearch() bool

	// Future optional operations can be added here:
	// SupportsUpgrade() bool
	// SupportsDependencyTree() bool
}

// PackageManager defines the standard interface for package managers.
// Package managers handle availability checking, listing, installing, and uninstalling packages.
// All methods accept a context for cancellation and timeout support.
type PackageManager interface {
	PackageManagerCapabilities

	// Core operations - these are always supported by all package managers
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	Install(ctx context.Context, name string) error
	Uninstall(ctx context.Context, name string) error
	IsInstalled(ctx context.Context, name string) (bool, error)
	GetInstalledVersion(ctx context.Context, name string) (string, error)
	Info(ctx context.Context, name string) (*PackageInfo, error)

	// Optional operations - check capabilities before calling
	Search(ctx context.Context, query string) ([]string, error)
}

// PackageInfo represents detailed information about a package
type PackageInfo struct {
	Name          string   `json:"name"`
	Version       string   `json:"version,omitempty"`
	Description   string   `json:"description,omitempty"`
	Homepage      string   `json:"homepage,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
	InstalledSize string   `json:"installed_size,omitempty"`
	Manager       string   `json:"manager"`
	Installed     bool     `json:"installed"`
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	Package string `json:"package"`
	Manager string `json:"manager"`
	Found   bool   `json:"found"`
}

// PackageConfigLoader defines how to load package configuration
type PackageConfigLoader interface {
	GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
}

// PackageConfigItem represents a package from configuration
type PackageConfigItem struct {
	Name string
}
