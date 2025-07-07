// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"plonk/pkg/managers"
)

// AsdfDiscoverer discovers globally configured ASDF tools and versions.
type AsdfDiscoverer struct {
	asdfManager *managers.AsdfManager
}

// NewAsdfDiscoverer creates a new ASDF tool discoverer.
func NewAsdfDiscoverer(executor managers.CommandExecutor) *AsdfDiscoverer {
	return &AsdfDiscoverer{
		asdfManager: managers.NewAsdfManager(executor),
	}
}

// DiscoverPackages returns a list of globally configured ASDF tools with versions.
// Delegates to the existing AsdfManager.ListGlobalTools() which reads ~/.tool-versions.
func (a *AsdfDiscoverer) DiscoverPackages() ([]string, error) {
	return a.asdfManager.ListGlobalTools()
}
