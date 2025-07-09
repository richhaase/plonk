// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/dotfiles"
)

// DotfileConfigLoader defines how to load dotfile configuration
type DotfileConfigLoader interface {
	GetDotfileTargets() map[string]string // source -> destination mapping
}

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
				return nil, fmt.Errorf("failed to expand directory %s: %w", source, err)
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
	var items []ActualItem
	
	// Get dotfiles from home directory
	dotfiles, err := d.manager.ListDotfiles(d.homeDir)
	if err != nil {
		return nil, err
	}
	
	for _, dotfile := range dotfiles {
		fullPath := filepath.Join(d.homeDir, dotfile)
		items = append(items, ActualItem{
			Name: dotfile,
			Path: fullPath,
			Metadata: map[string]interface{}{
				"path": fullPath,
			},
		})
	}
	
	// Also check configured destinations to find files in subdirectories
	targets := d.configLoader.GetDotfileTargets()
	for _, destination := range targets {
		destPath := d.manager.ExpandPath(destination)
		
		// Check if destination exists
		if !d.manager.FileExists(destPath) {
			continue
		}
		
		// If it's a directory, walk it to find individual files
		if d.manager.IsDirectory(destPath) {
			err := filepath.Walk(destPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				
				// Skip directories
				if info.IsDir() {
					return nil
				}
				
				// Calculate relative path to create a proper name
				relPath, err := filepath.Rel(d.homeDir, path)
				if err != nil {
					return err
				}
				
				// Only add if not already in items (avoid duplicates)
				name := relPath
				found := false
				for _, item := range items {
					if item.Name == name {
						found = true
						break
					}
				}
				
				if !found {
					items = append(items, ActualItem{
						Name: name,
						Path: path,
						Metadata: map[string]interface{}{
							"path": path,
						},
					})
				}
				
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to walk directory %s: %w", destPath, err)
			}
		} else {
			// Single file - calculate relative path
			relPath, err := filepath.Rel(d.homeDir, destPath)
			if err != nil {
				return nil, fmt.Errorf("failed to get relative path for %s: %w", destPath, err)
			}
			
			// Only add if not already in items (avoid duplicates)
			name := relPath
			found := false
			for _, item := range items {
				if item.Name == name {
					found = true
					break
				}
			}
			
			if !found {
				items = append(items, ActualItem{
					Name: name,
					Path: destPath,
					Metadata: map[string]interface{}{
						"path": destPath,
					},
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
				item.Path = d.manager.ExpandPath(dest)
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