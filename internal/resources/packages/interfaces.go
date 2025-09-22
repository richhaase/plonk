// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import "context"

// PackageManager defines the standard interface for package packages.
// Package managers handle availability checking, listing, installing, and uninstalling packages.
// All methods accept a context for cancellation and timeout support.
type PackageManager interface {

	// Core operations - these are always supported by all package packages
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	Install(ctx context.Context, name string) error
	Uninstall(ctx context.Context, name string) error
	IsInstalled(ctx context.Context, name string) (bool, error)
	InstalledVersion(ctx context.Context, name string) (string, error)
	Info(ctx context.Context, name string) (*PackageInfo, error)

	// Search operations - may return empty results if unsupported
	Search(ctx context.Context, query string) ([]string, error)

	// Health checking - provides diagnostic information about the package manager
	CheckHealth(ctx context.Context) (*HealthCheck, error)

	// Self-installation - automatically installs the package manager if not available
	SelfInstall(ctx context.Context) error

	// Upgrade upgrades one or more packages to their latest versions
	// If packages slice is empty, upgrades all installed packages for this manager
	Upgrade(ctx context.Context, packages []string) error

	// Dependencies returns package managers this manager depends on for self-installation
	// Returns empty slice if fully independent
	// Each string should match the manager name used in the registry
	Dependencies() []string
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
