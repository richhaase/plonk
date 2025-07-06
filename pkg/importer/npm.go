package importer

import (
	"plonk/pkg/managers"
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
