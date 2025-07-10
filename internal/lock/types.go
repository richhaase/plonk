// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import "time"

// LockFile represents the structure of plonk.lock
type LockFile struct {
	Version  int                       `yaml:"version"`
	Packages map[string][]PackageEntry `yaml:"packages"`
}

// PackageEntry represents a single managed package
type PackageEntry struct {
	Name        string    `yaml:"name"`
	InstalledAt time.Time `yaml:"installed_at"`
	Version     string    `yaml:"version"`
}
