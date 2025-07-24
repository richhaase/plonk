// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"github.com/richhaase/plonk/internal/paths"
)

// ConfigBasedDotfileLoader implements DotfileConfigLoader using a config interface
type ConfigBasedDotfileLoader struct {
	ignorePatterns    []string
	expandDirectories []string
}

// NewConfigBasedDotfileLoader creates a new loader from config data
func NewConfigBasedDotfileLoader(ignorePatterns, expandDirectories []string) *ConfigBasedDotfileLoader {
	return &ConfigBasedDotfileLoader{
		ignorePatterns:    ignorePatterns,
		expandDirectories: expandDirectories,
	}
}

// GetDotfileTargets returns a map of source -> destination paths for dotfiles
func (c *ConfigBasedDotfileLoader) GetDotfileTargets() map[string]string {
	// Use PathResolver to expand config directory and generate paths
	resolver, err := paths.NewPathResolverFromDefaults()
	if err != nil {
		return make(map[string]string)
	}

	// Delegate to PathResolver for directory expansion and path mapping
	result, err := resolver.ExpandConfigDirectory(c.ignorePatterns)
	if err != nil {
		return make(map[string]string)
	}

	return result
}

// GetIgnorePatterns returns patterns to ignore for file filtering
func (c *ConfigBasedDotfileLoader) GetIgnorePatterns() []string {
	return c.ignorePatterns
}

// GetExpandDirectories returns directories to expand in dot list
func (c *ConfigBasedDotfileLoader) GetExpandDirectories() []string {
	return c.expandDirectories
}
