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
	cfg              *config.Config
}

// NewConfigHandler creates a new config handler (loads config internally for backwards compatibility)
func NewConfigHandler(homeDir, configDir string, resolver PathResolver, scanner DirectoryScanner, comparator FileComparator) *ConfigHandlerImpl {
	return &ConfigHandlerImpl{
		homeDir:          homeDir,
		configDir:        configDir,
		pathResolver:     resolver,
		directoryScanner: scanner,
		fileComparator:   comparator,
		cfg:              config.LoadWithDefaults(configDir),
	}
}

// NewConfigHandlerWithConfig creates a new config handler with injected config
func NewConfigHandlerWithConfig(homeDir, configDir string, cfg *config.Config, resolver PathResolver, scanner DirectoryScanner, comparator FileComparator) *ConfigHandlerImpl {
	return &ConfigHandlerImpl{
		homeDir:          homeDir,
		configDir:        configDir,
		pathResolver:     resolver,
		directoryScanner: scanner,
		fileComparator:   comparator,
		cfg:              cfg,
	}
}

// GetConfiguredDotfiles returns dotfiles defined in configuration
func (ch *ConfigHandlerImpl) GetConfiguredDotfiles() ([]resources.Item, error) {
	// Use injected config for ignore patterns
	cfg := ch.cfg

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
	// Use injected config to get ignore patterns and expand directories
	cfg := ch.cfg

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

	// Ensure we include any configured targets that exist even if not expanded by default
	// This prevents false "missing" for nested paths (e.g., ~/.claude/*) when only top-level
	// directories are scanned.
	// Build a quick set of already discovered names to avoid duplicates
	existing := make(map[string]struct{}, len(items))
	for _, it := range items {
		existing[it.Name] = struct{}{}
	}

	// Discover configured sources/targets from the plonk config directory
	targets, err := ch.directoryScanner.ExpandConfigDirectory(cfg.IgnorePatterns)
	if err == nil {
		for _, destination := range targets {
			// destination is like "~/.something/path" -> convert to rel name used by scanner
			relName := destination
			if strings.HasPrefix(relName, "~/") {
				relName = relName[2:]
			}

			// Skip if we already discovered this entry via scanning
			if _, seen := existing[relName]; seen {
				continue
			}

			// Resolve absolute destination path and check existence
			destPath, err := ch.pathResolver.GetDestinationPath(destination)
			if err != nil {
				continue
			}
			info, err := os.Stat(destPath)
			if err != nil {
				// Does not exist or not accessible; ignore here (will be treated as missing elsewhere)
				continue
			}

			// Apply ignore filter (managed scan uses IgnorePatterns only)
			if filter != nil && filter.ShouldSkip(relName, info) {
				continue
			}

			// Add as an actual item so reconciliation can match it to desired
			items = append(items, resources.Item{
				Name:   relName,
				State:  resources.StateUntracked,
				Domain: "dotfile",
				Path:   destPath,
				Metadata: map[string]interface{}{
					"path": destPath,
				},
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
