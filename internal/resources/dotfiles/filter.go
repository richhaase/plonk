// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"strings"
)

// Filter handles ignore pattern filtering for dotfiles
type Filter struct {
	ignorePatterns []string
	configDir      string
	skipConfigDir  bool
}

// NewFilter creates a new filter with the given ignore patterns
func NewFilter(ignorePatterns []string, configDir string, skipConfigDir bool) *Filter {
	return &Filter{
		ignorePatterns: ignorePatterns,
		configDir:      configDir,
		skipConfigDir:  skipConfigDir,
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

	// Check against configured ignore patterns
	for _, pattern := range f.ignorePatterns {
		if f.matchesPattern(pattern, relPath, info) {
			return true
		}
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

// matchesPattern checks if a path matches an ignore pattern
func (f *Filter) matchesPattern(pattern, relPath string, info os.FileInfo) bool {
	// Check exact match for file/directory name
	if pattern == info.Name() || pattern == relPath {
		return true
	}

	// Handle ** patterns by converting to simple string matching
	if strings.Contains(pattern, "**") {
		return f.matchesDoubleStarPattern(pattern, relPath)
	}

	// Check glob pattern match
	if matched, _ := filepath.Match(pattern, info.Name()); matched {
		return true
	}
	if matched, _ := filepath.Match(pattern, relPath); matched {
		return true
	}

	return false
}

// matchesDoubleStarPattern handles patterns with ** wildcards
func (f *Filter) matchesDoubleStarPattern(pattern, relPath string) bool {
	// Handle patterns like "**/something/**"
	if strings.HasPrefix(pattern, "**/") && strings.HasSuffix(pattern, "/**") {
		// Extract the middle part (e.g., "node_modules", "*-*-*-*-*", ".git", "*cache*")
		middle := pattern[3 : len(pattern)-3]

		// Split path into components
		pathParts := strings.Split(relPath, "/")

		// Check each path component
		for _, part := range pathParts {
			// For literal matches (like "node_modules", ".git")
			if middle == part {
				return true
			}
			// For glob patterns (like "*-*-*-*-*", "*cache*")
			if matched, _ := filepath.Match(middle, part); matched {
				return true
			}
			// For cache patterns, check case-insensitive
			if strings.Contains(middle, "cache") && strings.Contains(strings.ToLower(part), "cache") {
				return true
			}
		}
	}

	return false
}
