// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"plonk/internal/managers"
)

// NpmDiscoverer discovers globally installed NPM packages.
type NpmDiscoverer struct {
	npmManager *managers.NpmManager
}

// NewNpmDiscoverer creates a new NPM package discoverer.
func NewNpmDiscoverer(executor managers.CommandExecutor) *NpmDiscoverer {
	return &NpmDiscoverer{
		npmManager: managers.NewNpmManager(executor),
	}
}

// DiscoverPackages returns a list of globally installed NPM packages.
// Delegates to the existing NpmManager which already handles scoped packages correctly.
func (n *NpmDiscoverer) DiscoverPackages() ([]string, error) {
	return n.npmManager.ListInstalled()
}
