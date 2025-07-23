// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"fmt"
	"os"

	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/interfaces"
)

// DotfileConfigLoader defines how to load dotfile configuration
// DotfileConfigLoader is now an alias to the unified interfaces package
type DotfileConfigLoader = interfaces.DotfileConfigLoader

// DotfileProvider implements the Provider interface for dotfile management
type DotfileProvider struct {
	homeDir      string
	configDir    string
	configLoader DotfileConfigLoader
	manager      *dotfiles.Manager
}

// NewDotfileProvider creates a new dotfile provider
func NewDotfileProvider(homeDir string, configDir string, configLoader DotfileConfigLoader) *DotfileProvider {
	return &DotfileProvider{
		homeDir:      homeDir,
		configDir:    configDir,
		configLoader: configLoader,
		manager:      dotfiles.NewManager(homeDir, configDir),
	}
}

// Domain returns the domain name for dotfiles
func (d *DotfileProvider) Domain() string {
	return "dotfile"
}

// GetManager returns the dotfile manager for direct access
func (d *DotfileProvider) GetManager() *dotfiles.Manager {
	return d.manager
}

// GetConfiguredItems returns dotfiles defined in configuration
func (d *DotfileProvider) GetConfiguredItems() ([]ConfigItem, error) {
	targets := d.configLoader.GetDotfileTargets()

	items := make([]ConfigItem, 0)
	for source, destination := range targets {
		// Check if source is a directory
		sourcePath := d.manager.GetSourcePath(source)
		info, err := os.Stat(sourcePath)
		if err != nil {
			// Source doesn't exist yet, treat as single file
			name := d.manager.DestinationToName(destination)
			items = append(items, ConfigItem{
				Name: name,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
				},
			})
			continue
		}

		if info.IsDir() {
			// Expand directory to individual files
			dirItems, err := d.expandConfigDirectory(source, destination)
			if err != nil {
				return nil, errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "scan",
					fmt.Sprintf("failed to expand directory %s", source))
			}
			items = append(items, dirItems...)
		} else {
			// Single file
			name := d.manager.DestinationToName(destination)
			items = append(items, ConfigItem{
				Name: name,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
				},
			})
		}
	}

	return items, nil
}

// GetActualItems returns dotfiles currently present in the home directory
func (d *DotfileProvider) GetActualItems(ctx context.Context) ([]ActualItem, error) {
	// Create filter with ignore patterns
	filter := dotfiles.NewFilter(
		d.configLoader.GetIgnorePatterns(),
		d.configDir,
		true, // Skip config directory when scanning home
	)

	// Create scanner
	scanner := dotfiles.NewScanner(d.homeDir, filter)

	// Create expander
	expander := dotfiles.NewExpander(
		d.homeDir,
		d.configLoader.GetExpandDirectories(),
		scanner,
	)

	var items []ActualItem

	// Step 1: Scan home directory for dotfiles
	scanResults, err := scanner.ScanDotfiles(ctx)
	if err != nil {
		return nil, err
	}

	// Process scan results
	for _, result := range scanResults {
		// Check if directory should be expanded
		if result.Info.IsDir() && expander.ShouldExpandDirectory(result.Name) {
			// Expand directory
			expandedResults, err := expander.ExpandDirectory(ctx, result.Path, result.Name)
			if err != nil {
				// Fall back to showing directory as single item
				items = append(items, ActualItem{
					Name:     result.Name,
					Path:     result.Path,
					Metadata: result.Metadata,
				})
				continue
			}

			// Add expanded results
			for _, expanded := range expandedResults {
				items = append(items, ActualItem{
					Name:     expanded.Name,
					Path:     expanded.Path,
					Metadata: expanded.Metadata,
				})
			}
		} else {
			// Single file or unexpanded directory
			// Mark as seen in expander to avoid duplicates later
			expander.CheckDuplicate(result.Name)
			items = append(items, ActualItem{
				Name:     result.Name,
				Path:     result.Path,
				Metadata: result.Metadata,
			})
		}
	}

	// Step 2: Check configured destinations
	targets := d.configLoader.GetDotfileTargets()
	for _, destination := range targets {
		destPath, err := d.manager.GetDestinationPath(destination)
		if err != nil {
			// Skip files with invalid destination paths
			continue
		}

		// Skip if destination doesn't exist
		if !d.manager.FileExists(destPath) {
			continue
		}

		// For single files, check if we've already seen them
		if !d.manager.IsDirectory(destPath) {
			relPath, err := expander.CalculateRelativePath(destPath)
			if err != nil {
				continue
			}

			// Skip if already processed
			if expander.CheckDuplicate(relPath) {
				continue
			}

			info, err := os.Stat(destPath)
			if err != nil {
				continue
			}

			// Apply filter
			if filter.ShouldSkip(relPath, info) {
				continue
			}

			items = append(items, ActualItem{
				Name: relPath,
				Path: destPath,
				Metadata: map[string]interface{}{
					"path": destPath,
				},
			})
		} else {
			// For directories, expand them
			destResults, err := expander.ExpandConfiguredDestination(ctx, d.manager, destPath)
			if err != nil {
				continue
			}

			// Add results (duplicates are already filtered by expander)
			for _, result := range destResults {
				// Apply filter
				if filter.ShouldSkip(result.Name, result.Info) {
					continue
				}

				items = append(items, ActualItem{
					Name:     result.Name,
					Path:     result.Path,
					Metadata: result.Metadata,
				})
			}
		}
	}

	return items, nil
}

// CreateItem creates an Item from dotfile data
func (d *DotfileProvider) CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item {
	item := Item{
		Name:   name,
		State:  state,
		Domain: "dotfile",
	}

	// Set path from actual or infer from configured
	if actual != nil {
		item.Path = actual.Path
		item.Metadata = make(map[string]interface{})
		for k, v := range actual.Metadata {
			item.Metadata[k] = v
		}
	} else {
		item.Metadata = make(map[string]interface{})
	}

	// Add configured metadata
	if configured != nil {
		for k, v := range configured.Metadata {
			item.Metadata[k] = v
		}

		// If no path set and we have destination, use that
		if item.Path == "" {
			if dest, ok := configured.Metadata["destination"].(string); ok {
				if path, err := d.manager.GetDestinationPath(dest); err == nil {
					item.Path = path
				}
			}
		}
	}

	return item
}

// expandConfigDirectory walks a directory and creates individual ConfigItems for each file
func (d *DotfileProvider) expandConfigDirectory(sourceDir, destDir string) ([]ConfigItem, error) {
	dotfileInfos, err := d.manager.ExpandDirectory(sourceDir, destDir)
	if err != nil {
		return nil, err
	}

	items := make([]ConfigItem, len(dotfileInfos))
	for i, info := range dotfileInfos {
		items[i] = ConfigItem{
			Name:     info.Name,
			Metadata: info.Metadata,
		}
	}

	return items, nil
}
