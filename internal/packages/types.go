// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

// Operation result types for packages domain

// InstallStatus represents the status of a package install operation
type InstallStatus string

const (
	InstallStatusAdded    InstallStatus = "added"
	InstallStatusWouldAdd InstallStatus = "would-add"
	InstallStatusSkipped  InstallStatus = "skipped"
	InstallStatusFailed   InstallStatus = "failed"
)

// String returns the string representation of InstallStatus
func (s InstallStatus) String() string {
	return string(s)
}

// InstallResult represents the result of installing a package
type InstallResult struct {
	// Name is the package name
	Name string

	// Manager is the package manager used (e.g., "brew", "apt", "npm")
	Manager string

	// Status is the result of the install operation
	Status InstallStatus

	// AlreadyManaged indicates whether the package was already managed by plonk
	AlreadyManaged bool

	// Error contains any error that occurred during the operation
	Error error
}

// UninstallStatus represents the status of a package uninstall operation
type UninstallStatus string

const (
	UninstallStatusRemoved     UninstallStatus = "removed"
	UninstallStatusWouldRemove UninstallStatus = "would-remove"
	UninstallStatusFailed      UninstallStatus = "failed"
)

// String returns the string representation of UninstallStatus
func (s UninstallStatus) String() string {
	return string(s)
}

// UninstallResult represents the result of uninstalling a package
type UninstallResult struct {
	// Name is the package name
	Name string

	// Manager is the package manager used (e.g., "brew", "apt", "npm")
	Manager string

	// Status is the result of the uninstall operation
	Status UninstallStatus

	// Error contains any error that occurred during the operation
	Error error
}
