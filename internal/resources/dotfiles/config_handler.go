// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
)

// ConfigHandlerImpl implements ConfigHandler interface
type ConfigHandlerImpl struct {
	homeDir          string
	configDir        string
	pathResolver     PathResolver
	directoryScanner DirectoryScanner
	fileComparator   FileComparator
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(homeDir, configDir string, resolver PathResolver, scanner DirectoryScanner, comparator FileComparator) *ConfigHandlerImpl {
	return &ConfigHandlerImpl{
		homeDir:          homeDir,
		configDir:        configDir,
		pathResolver:     resolver,
		directoryScanner: scanner,
		fileComparator:   comparator,
	}
}

// GetConfiguredDotfiles returns dotfiles defined in configuration
func (ch *ConfigHandlerImpl) GetConfiguredDotfiles() ([]resources.Item, error) {
	// Load config to get ignore patterns
	cfg := config.LoadWithDefaults(ch.configDir)

	targets, err := ch.directoryScanner.ExpandConfigDirectory(cfg.IgnorePatterns)
	if err != nil {
		return nil, fmt.Errorf("expanding config directory: %w", err)
	}

	items := make([]resources.Item, 0)

	for source, destination := range targets {
		// Check if source is a directory
		sourcePath := ch.pathResolver.GetSourcePath(source)
		info, err := os.Stat(sourcePath)
		if err != nil {
			// Source doesn't exist yet, treat as single file
			name := ch.destinationToName(destination)
			items = append(items, resources.Item{
				Name:   name,
				Domain: "dotfile",
				Path:   destination,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
					"compare_fn":  ch.createCompareFunc(source, destination),
				},
			})
			continue
		}

		if info.IsDir() {
			// For directories, just use the directory itself as one item
			name := ch.destinationToName(destination)
			items = append(items, resources.Item{
				Name:   name,
				Domain: "dotfile",
				Path:   destination,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
					"isDirectory": true,
					"compare_fn":  ch.createCompareFunc(source, destination),
				},
			})
		} else {
			// Single file
			name := ch.destinationToName(destination)
			items = append(items, resources.Item{
				Name:   name,
				Domain: "dotfile",
				Path:   destination,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
					"compare_fn":  ch.createCompareFunc(source, destination),
				},
			})
		}
	}

	return items, nil
}

// GetActualDotfiles returns dotfiles currently present in the home directory
func (ch *ConfigHandlerImpl) GetActualDotfiles(ctx context.Context) ([]resources.Item, error) {
	// Load config to get ignore patterns and expand directories
	cfg := config.LoadWithDefaults(ch.configDir)

	// Only use IgnorePatterns for managed dotfile scanning
	// UnmanagedFilters should NOT be applied here as they're only for unmanaged file discovery
	filter := NewFilter(cfg.IgnorePatterns, ch.configDir, true)

	// Create scanner
	scanner := NewScanner(ch.homeDir, filter)

	// Create expander
	expander := NewExpander(ch.homeDir, cfg.ExpandDirectories, scanner)

	var items []resources.Item

	// Scan home directory for dotfiles
	scanResults, err := scanner.ScanDotfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("scanning dotfiles: %w", err)
	}

	// Process scan results
	for _, result := range scanResults {
		// Check if directory should be expanded
		shouldExpand := false
		for _, dir := range cfg.ExpandDirectories {
			if result.Name == dir {
				shouldExpand = true
				break
			}
		}
		if result.Info.IsDir() && shouldExpand {
			// For expanded directories, scan inside them
			err := filepath.Walk(result.Path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip files we can't read
				}

				// Skip the directory itself
				if path == result.Path {
					return nil
				}

				// Skip directories (we want files)
				if info.IsDir() {
					return nil
				}

				// Get relative path from home
				relPath, err := filepath.Rel(ch.homeDir, path)
				if err != nil {
					return nil
				}

				// Apply filter
				if filter != nil && filter.ShouldSkip(relPath, info) {
					return nil
				}

				items = append(items, resources.Item{
					Name:   relPath,
					State:  resources.StateUntracked,
					Domain: "dotfile",
					Path:   path,
					Metadata: map[string]interface{}{
						"path": path,
					},
				})

				return nil
			})
			if err != nil {
				// If we can't walk the directory, just skip it
				continue
			}
		} else {
			// Single file or unexpanded directory
			expander.CheckDuplicate(result.Name)
			items = append(items, resources.Item{
				Name:     result.Name,
				State:    resources.StateUntracked,
				Domain:   "dotfile",
				Path:     result.Path,
				Metadata: result.Metadata,
			})
		}
	}

	return items, nil
}

// createCompareFunc creates a comparison function for a dotfile
func (ch *ConfigHandlerImpl) createCompareFunc(source, destination string) func() (bool, error) {
	return func() (bool, error) {
		sourcePath := ch.pathResolver.GetSourcePath(source)
		destPath, err := ch.pathResolver.GetDestinationPath(destination)
		if err != nil {
			return false, err
		}
		// If destination doesn't exist, they're not the same
		if !ch.fileExists(destPath) {
			return false, nil
		}
		return ch.fileComparator.CompareFiles(sourcePath, destPath)
	}
}

// destinationToName converts a destination path to a standardized name
func (ch *ConfigHandlerImpl) destinationToName(destination string) string {
	if strings.HasPrefix(destination, "~/") {
		return destination[2:]
	}
	return destination
}

// fileExists checks if a file exists at the given path
func (ch *ConfigHandlerImpl) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
