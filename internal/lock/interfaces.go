// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

// LockService defines the interface for lock file operations
type LockService interface {
	// Read reads and parses the lock file
	Read() (*Lock, error)

	// Write writes the lock data to disk
	Write(lock *Lock) error

	// AddPackage adds a package to the lock file with metadata
	AddPackage(manager, name, version string, metadata map[string]interface{}) error

	// RemovePackage removes a package from the lock file
	RemovePackage(manager, name string) error

	// GetPackages returns all packages for a specific manager
	GetPackages(manager string) ([]ResourceEntry, error)

	// HasPackage checks if a package exists in the lock file
	HasPackage(manager, name string) bool

	// FindPackage returns all locations where a package is installed
	FindPackage(name string) []ResourceEntry
}
