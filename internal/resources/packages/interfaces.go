// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import "context"

// PackageManager defines the core interface that all package managers must implement.
// Focuses on state checking and idempotent operations only.
type PackageManager interface {
	// IsAvailable checks if the package manager is available on the system
	IsAvailable(ctx context.Context) (bool, error)

	// ListInstalled lists all packages installed by this manager (for status display)
	ListInstalled(ctx context.Context) ([]string, error)

	// IsInstalled checks if a specific package is installed
	IsInstalled(ctx context.Context, name string) (bool, error)

	// Install installs a package (idempotent)
	Install(ctx context.Context, name string) error

	// Uninstall removes a package (idempotent)
	Uninstall(ctx context.Context, name string) error
}

// PackageUpgrader is an optional capability interface for upgrading packages.
// Implement this interface to enable upgrade functionality.
type PackageUpgrader interface {
	PackageManager
	// Upgrade upgrades one or more packages to their latest versions
	// If packages slice is empty, upgrades all installed packages for this manager
	Upgrade(ctx context.Context, packages []string) error
}

// PackageConfigLoader defines how to load package configuration
type PackageConfigLoader interface {
	GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
}

// PackageConfigItem represents a package from configuration
type PackageConfigItem struct {
	Name string
}
