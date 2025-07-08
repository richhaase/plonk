// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"os"
	"path/filepath"
	"strings"
)

// DotfileConfigLoader defines how to load dotfile configuration
type DotfileConfigLoader interface {
	GetDotfileTargets() map[string]string // source -> destination mapping
}

// DotfileProvider implements the Provider interface for dotfile management
type DotfileProvider struct {
	homeDir      string
	configLoader DotfileConfigLoader
}

// NewDotfileProvider creates a new dotfile provider
func NewDotfileProvider(homeDir string, configLoader DotfileConfigLoader) *DotfileProvider {
	return &DotfileProvider{
		homeDir:      homeDir,
		configLoader: configLoader,
	}
}

// Domain returns the domain name for dotfiles
func (d *DotfileProvider) Domain() string {
	return "dotfile"
}

// GetConfiguredItems returns dotfiles defined in configuration
func (d *DotfileProvider) GetConfiguredItems() ([]ConfigItem, error) {
	targets := d.configLoader.GetDotfileTargets()
	
	items := make([]ConfigItem, 0, len(targets))
	for source, destination := range targets {
		// Convert destination to relative path for consistent naming
		name := d.destinationToName(destination)
		
		items = append(items, ConfigItem{
			Name: name,
			Metadata: map[string]interface{}{
				"source":      source,
				"destination": destination,
			},
		})
	}
	
	return items, nil
}

// GetActualItems returns dotfiles currently present in the home directory
func (d *DotfileProvider) GetActualItems() ([]ActualItem, error) {
	dotfiles, err := d.listDotfiles(d.homeDir)
	if err != nil {
		return nil, err
	}
	
	items := make([]ActualItem, len(dotfiles))
	for i, dotfile := range dotfiles {
		fullPath := filepath.Join(d.homeDir, dotfile)
		items[i] = ActualItem{
			Name: dotfile,
			Path: fullPath,
			Metadata: map[string]interface{}{
				"path": fullPath,
			},
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
				item.Path = d.expandPath(dest)
			}
		}
	}
	
	return item
}

// listDotfiles finds all dotfiles in the specified directory
func (d *DotfileProvider) listDotfiles(dir string) ([]string, error) {
	var dotfiles []string
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") && entry.Name() != "." && entry.Name() != ".." {
			dotfiles = append(dotfiles, entry.Name())
		}
	}
	
	return dotfiles, nil
}

// destinationToName converts a destination path to a standardized name
func (d *DotfileProvider) destinationToName(destination string) string {
	// Remove ~/ prefix if present
	if strings.HasPrefix(destination, "~/") {
		return destination[2:]
	}
	return destination
}

// expandPath expands ~ to home directory
func (d *DotfileProvider) expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(d.homeDir, path[2:])
	}
	return path
}