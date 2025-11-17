// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/ignore"
)

// Filter handles ignore pattern filtering for dotfiles
type Filter struct {
	configDir     string
	skipConfigDir bool
	matcher       *ignore.Matcher
}

// NewFilter creates a new filter with the given ignore patterns
func NewFilter(ignorePatterns []string, configDir string, skipConfigDir bool) *Filter {
	return &Filter{
		configDir:     configDir,
		skipConfigDir: skipConfigDir,
		matcher:       ignore.NewMatcher(ignorePatterns),
	}
}

// ShouldSkip determines if a file/directory should be skipped
func (f *Filter) ShouldSkip(relPath string, info os.FileInfo) bool {
	// Always skip plonk config file
	if relPath == "plonk.yaml" {
		return true
	}

	// Always skip .plonk/ directory (reserved for future plonk metadata)
	if relPath == ".plonk" || strings.HasPrefix(relPath, ".plonk/") {
		return true
	}

	// Skip plonk config directory when requested
	if f.skipConfigDir && f.configDir != "" {
		if f.isConfigPath(relPath) {
			return true
		}
	}

	// Check against configured ignore patterns using gitignore semantics
	if f.matcher != nil && f.matcher.ShouldIgnore(relPath, info.IsDir()) {
		return true
	}

	return false
}

// isConfigPath checks if a path is within the config directory
func (f *Filter) isConfigPath(relPath string) bool {
	// Extract the relative path pattern from the config directory
	// e.g., "/home/user/.config/plonk" -> ".config/plonk"
	if strings.Contains(f.configDir, "/.config/plonk") {
		configPattern := ".config/plonk"
		if relPath == configPattern || strings.HasPrefix(relPath, configPattern+"/") {
			return true
		}
	}

	// Also handle other config directory patterns
	configBasename := filepath.Base(f.configDir)
	if strings.HasSuffix(relPath, configBasename) || strings.Contains(relPath, configBasename+"/") {
		return true
	}

	return false
}
