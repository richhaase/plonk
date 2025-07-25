// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

// LockReader defines the interface for reading lock files
type LockReader interface {
	// Load reads and parses the lock file
	Load() (*LockFile, error)
	// Read reads and parses the lock file into LockData structure
	Read() (*LockData, error)
}

// LockWriter defines the interface for writing lock files
type LockWriter interface {
	// Save writes the lock file to disk
	Save(lock *LockFile) error
	// Write writes the lock data to disk (always as v2)
	Write(data *LockData) error
}

// LockService combines reading and writing with additional operations
type LockService interface {
	LockReader
	LockWriter

	// AddPackage adds a package to the lock file
	AddPackage(manager, name, version string) error

	// RemovePackage removes a package from the lock file
	RemovePackage(manager, name string) error

	// GetPackages returns all packages for a specific manager
	GetPackages(manager string) ([]PackageEntry, error)

	// HasPackage checks if a package exists in the lock file
	HasPackage(manager, name string) bool
}
