// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

// Lock file constants
const (
	// LockFileName is the name of the lock file
	LockFileName = "plonk.lock"

	// LockFileVersion is the lock file format version
	LockFileVersion = 2
)

// Lock represents the structure of plonk.lock
type Lock struct {
	Version   int             `yaml:"version"`
	Resources []ResourceEntry `yaml:"resources"`
}

// ResourceEntry represents a generic resource in the lock file
type ResourceEntry struct {
	Type        string                 `yaml:"type"`                   // "package", "dotfile", etc.
	ID          string                 `yaml:"id"`                     // Resource-specific identifier (e.g., "go:gopls")
	Metadata    map[string]interface{} `yaml:"metadata"`               // Resource-specific data
	InstalledAt string                 `yaml:"installed_at,omitempty"` // ISO8601 timestamp
}
