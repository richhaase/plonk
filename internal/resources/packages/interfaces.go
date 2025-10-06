// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import "context"

// PackageManager defines the core interface that all package managers must implement.
// This includes basic package management operations. Additional capabilities can be
// implemented through optional capability interfaces.
type PackageManager interface {
	// IsAvailable checks if the package manager is available on the system
	IsAvailable(ctx context.Context) (bool, error)

	// ListInstalled lists all packages installed by this manager
	ListInstalled(ctx context.Context) ([]string, error)

	// Install installs a package
	Install(ctx context.Context, name string) error

	// Uninstall removes a package
	Uninstall(ctx context.Context, name string) error

	// IsInstalled checks if a specific package is installed
	IsInstalled(ctx context.Context, name string) (bool, error)

	// InstalledVersion returns the version of an installed package
	InstalledVersion(ctx context.Context, name string) (string, error)

	// Dependencies returns package managers this manager depends on for self-installation
	// Returns empty slice if fully independent
	Dependencies() []string
}

// PackageSearcher is an optional capability interface for searching packages.
// Implement this interface to enable search functionality for a package manager.
type PackageSearcher interface {
	PackageManager
	// Search searches for packages matching the query
	Search(ctx context.Context, query string) ([]string, error)
}

// PackageInfoProvider is an optional capability interface for getting package information.
// Implement this interface to provide detailed package metadata.
type PackageInfoProvider interface {
	PackageManager
	// Info returns detailed information about a package
	Info(ctx context.Context, name string) (*PackageInfo, error)
}

// PackageUpgrader is an optional capability interface for upgrading packages.
// Implement this interface to enable upgrade functionality.
type PackageUpgrader interface {
	PackageManager
	// Upgrade upgrades one or more packages to their latest versions
	// If packages slice is empty, upgrades all installed packages for this manager
	Upgrade(ctx context.Context, packages []string) error
}

// PackageHealthChecker is an optional capability interface for health checking.
// Implement this interface to provide diagnostic information about the package manager.
type PackageHealthChecker interface {
	PackageManager
	// CheckHealth provides diagnostic information about the package manager
	CheckHealth(ctx context.Context) (*HealthCheck, error)
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

// HealthCheck represents the health status of a package manager
type HealthCheck struct {
	Name        string   `json:"name" yaml:"name"`
	Category    string   `json:"category" yaml:"category"`
	Status      string   `json:"status" yaml:"status"`
	Message     string   `json:"message" yaml:"message"`
	Details     []string `json:"details,omitempty" yaml:"details,omitempty"`
	Issues      []string `json:"issues,omitempty" yaml:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}

// Capability checking functions

// SupportsSearch returns true if the manager implements PackageSearcher
func SupportsSearch(pm PackageManager) bool {
	_, ok := pm.(PackageSearcher)
	return ok
}

// SupportsInfo returns true if the manager implements PackageInfoProvider
func SupportsInfo(pm PackageManager) bool {
	_, ok := pm.(PackageInfoProvider)
	return ok
}

// SupportsUpgrade returns true if the manager implements PackageUpgrader
func SupportsUpgrade(pm PackageManager) bool {
	_, ok := pm.(PackageUpgrader)
	return ok
}

// SupportsHealthCheck returns true if the manager implements PackageHealthChecker
func SupportsHealthCheck(pm PackageManager) bool {
	_, ok := pm.(PackageHealthChecker)
	return ok
}
