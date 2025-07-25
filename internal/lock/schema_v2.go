// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

const CurrentVersion = 2

// LockV2 represents the v2 structure of plonk.lock with generic resources
type LockV2 struct {
	Version   int                  `yaml:"version"`
	Packages  map[string][]Package `yaml:"packages,omitempty"`  // For compatibility
	Resources []ResourceEntry      `yaml:"resources,omitempty"` // New generic section
}

// ResourceEntry represents a generic resource in the lock file
type ResourceEntry struct {
	Type        string                 `yaml:"type"`                   // "package", "dotfile", "docker-compose"
	ID          string                 `yaml:"id"`                     // Resource-specific identifier
	State       string                 `yaml:"state"`                  // "managed", "missing", etc.
	Metadata    map[string]interface{} `yaml:"metadata"`               // Resource-specific data
	InstalledAt string                 `yaml:"installed_at,omitempty"` // ISO8601 timestamp
}

// Package represents a package entry for backward compatibility
type Package struct {
	Name        string `yaml:"name"`
	InstalledAt string `yaml:"installed_at"`
	Version     string `yaml:"version"`
}

// LockData represents the internal lock data structure used by the application
type LockData struct {
	Version   int
	Packages  map[string][]Package
	Resources []ResourceEntry
}
