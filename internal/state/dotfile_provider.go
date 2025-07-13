// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/dotfiles"
)

// DotfileConfigLoader defines how to load dotfile configuration
type DotfileConfigLoader interface {
	GetDotfileTargets() map[string]string // source -> destination mapping
	GetIgnorePatterns() []string          // ignore patterns for file filtering
	GetExpandDirectories() []string       // directories to expand in dot list
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

	// Get ignore patterns
	ignorePatterns := d.configLoader.GetIgnorePatterns()

	// Create skip context for home directory scanning (skip config dir only)
	skipCtx := &SkipContext{
		ConfigDir:         d.configDir,
		FilesOnlyMode:     false, // Allow directories in general GetActualItems
		SkipConfigDir:     true,  // Skip config directory when scanning home
		AllowConfigAccess: false, // Don't allow config access in this context
	}

	// Get dotfiles from home directory
	dotfiles, err := d.manager.ListDotfiles(d.homeDir)
	if err != nil {
		return nil, err
	}

	for _, dotfile := range dotfiles {
		fullPath := filepath.Join(d.homeDir, dotfile)

		// Check if file should be ignored
		info, err := os.Stat(fullPath)
		if err != nil {
			continue // Skip files we can't stat
		}

		if shouldSkipDotfile(dotfile, info, ignorePatterns, skipCtx) {
			continue // Skip ignored files
		}

		// For directories, check if we should expand them
		expandDirs := d.configLoader.GetExpandDirectories()
		if info.IsDir() && shouldExpandDirectory(dotfile, expandDirs) {
			// Expand directory to show individual files (limited depth)
			err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					// Skip directories we can't access but continue processing
					return nil
				}

				// Skip the root directory itself
				if path == fullPath {
					return nil
				}

				// Limit depth to avoid huge expansions
				relPath, err := filepath.Rel(fullPath, path)
				if err != nil {
					return nil
				}
				if strings.Count(relPath, string(os.PathSeparator)) > 2 {
					return nil
				}

				// Calculate relative path from home directory
				homeRelPath, err := filepath.Rel(d.homeDir, path)
				if err != nil {
					return nil
				}

				// Check if file should be ignored
				if shouldSkipDotfile(homeRelPath, info, ignorePatterns, skipCtx) {
					return nil
				}

				// Only add if not already in items (avoid duplicates)
				found := false
				for _, item := range items {
					if item.Name == homeRelPath {
						found = true
						break
					}
				}

				if !found {
					items = append(items, ActualItem{
						Name: homeRelPath,
						Path: path,
						Metadata: map[string]interface{}{
							"path": path,
						},
					})
				}

				return nil
			})
			if err != nil {
				// If we can't walk the directory, fall back to showing it as a single item
				items = append(items, ActualItem{
					Name: dotfile,
					Path: fullPath,
					Metadata: map[string]interface{}{
						"path": fullPath,
					},
				})
			}
		} else {
			// Single file or unexpanded directory
			items = append(items, ActualItem{
				Name: dotfile,
				Path: fullPath,
				Metadata: map[string]interface{}{
					"path": fullPath,
				},
			})
		}
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

				// Check if file should be ignored
				if shouldSkipDotfile(relPath, info, ignorePatterns, skipCtx) {
					return nil
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

			// Check if file should be ignored
			info, err := os.Stat(destPath)
			if err != nil {
				continue // Skip files we can't stat
			}
			if shouldSkipDotfile(relPath, info, ignorePatterns, skipCtx) {
				continue // Skip ignored files
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

// SkipContext provides context for filtering decisions
type SkipContext struct {
	ConfigDir         string
	FilesOnlyMode     bool
	SkipConfigDir     bool // When true, skip the config directory entirely (for home scanning)
	AllowConfigAccess bool // When true, allow access to config directory (for config reading)
}

// shouldSkipDotfile determines if a file/directory should be skipped based on ignore patterns and context
func shouldSkipDotfile(relPath string, info os.FileInfo, ignorePatterns []string, ctx *SkipContext) bool {
	// Always skip plonk config file
	if relPath == "plonk.yaml" {
		return true
	}

	// Skip plonk config directory when scanning home directory
	if ctx != nil && ctx.SkipConfigDir && ctx.ConfigDir != "" {
		// Extract the relative path pattern from the config directory
		// e.g., "/home/user/.config/plonk" -> ".config/plonk"
		if strings.Contains(ctx.ConfigDir, "/.config/plonk") {
			configPattern := ".config/plonk"
			if relPath == configPattern || strings.HasPrefix(relPath, configPattern+"/") {
				return true
			}
		}
		// Also handle other config directory patterns
		configBasename := filepath.Base(ctx.ConfigDir)
		if strings.HasSuffix(relPath, configBasename) || strings.Contains(relPath, configBasename+"/") {
			return true
		}
	}

	// Skip directories when in files-only mode
	if ctx != nil && ctx.FilesOnlyMode && info.IsDir() {
		return true
	}

	// Check against configured ignore patterns
	for _, pattern := range ignorePatterns {
		// Check exact match for file/directory name
		if pattern == info.Name() || pattern == relPath {
			return true
		}
		// Check glob pattern match
		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
	}

	return false
}

// shouldExpandDirectory determines if a directory should be expanded to show its contents
func shouldExpandDirectory(dirname string, expandDirs []string) bool {
	for _, dir := range expandDirs {
		if dirname == dir {
			return true
		}
	}
	return false
}
