// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"os"
	"path/filepath"
)

// DotfileDiscoverer discovers existing dotfiles in the home directory.
type DotfileDiscoverer struct{}

// NewDotfileDiscoverer creates a new dotfile discoverer.
func NewDotfileDiscoverer() *DotfileDiscoverer {
	return &DotfileDiscoverer{}
}

// DiscoverDotfiles returns a list of managed dotfiles that exist in the home directory.
// Only checks for dotfiles that plonk can manage: .zshrc, .gitconfig, .zshenv
func (d *DotfileDiscoverer) DiscoverDotfiles() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Dotfiles that plonk can manage
	managedDotfiles := []string{".zshrc", ".gitconfig", ".zshenv"}
	var foundDotfiles []string

	for _, dotfile := range managedDotfiles {
		filePath := filepath.Join(homeDir, dotfile)
		if _, err := os.Stat(filePath); err == nil {
			// File exists
			foundDotfiles = append(foundDotfiles, dotfile)
		}
	}

	return foundDotfiles, nil
}
