// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"
	"os"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/paths"
	"github.com/richhaase/plonk/internal/state"
)

// GetConfiguredDotfiles returns dotfiles defined in configuration
func GetConfiguredDotfiles(homeDir, configDir string) ([]state.Item, error) {
	// Load config to get ignore patterns
	cfg := config.LoadConfigWithDefaults(configDir)

	// Get dotfile targets from path resolver
	resolver, err := paths.NewPathResolverFromDefaults()
	if err != nil {
		return nil, fmt.Errorf("creating path resolver: %w", err)
	}

	targets, err := resolver.ExpandConfigDirectory(cfg.IgnorePatterns)
	if err != nil {
		return nil, fmt.Errorf("expanding config directory: %w", err)
	}

	manager := NewManager(homeDir, configDir)
	items := make([]state.Item, 0)

	for source, destination := range targets {
		// Check if source is a directory
		sourcePath := manager.GetSourcePath(source)
		info, err := os.Stat(sourcePath)
		if err != nil {
			// Source doesn't exist yet, treat as single file
			name := manager.DestinationToName(destination)
			items = append(items, state.Item{
				Name:   name,
				State:  state.StateMissing,
				Domain: "dotfile",
				Path:   destination,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
				},
			})
			continue
		}

		if info.IsDir() {
			// For directories, just use the directory itself as one item
			name := manager.DestinationToName(destination)
			items = append(items, state.Item{
				Name:   name,
				State:  state.StateManaged,
				Domain: "dotfile",
				Path:   destination,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
					"isDirectory": true,
				},
			})
		} else {
			// Single file
			name := manager.DestinationToName(destination)
			items = append(items, state.Item{
				Name:   name,
				State:  state.StateManaged,
				Domain: "dotfile",
				Path:   destination,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
				},
			})
		}
	}

	return items, nil
}

// GetActualDotfiles returns dotfiles currently present in the home directory
func GetActualDotfiles(ctx context.Context, homeDir, configDir string) ([]state.Item, error) {
	// Load config to get ignore patterns and expand directories
	cfg := config.LoadConfigWithDefaults(configDir)

	// Create filter with ignore patterns
	filter := NewFilter(cfg.IgnorePatterns, configDir, true)

	// Create scanner
	scanner := NewScanner(homeDir, filter)

	// Create expander
	expander := NewExpander(homeDir, cfg.ExpandDirectories, scanner)

	var items []state.Item

	// Scan home directory for dotfiles
	scanResults, err := scanner.ScanDotfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("scanning dotfiles: %w", err)
	}

	// Process scan results
	for _, result := range scanResults {
		// Check if directory should be expanded
		if result.Info.IsDir() && expander.ShouldExpandDirectory(result.Name) {
			// For expanded directories, treat as single item
			expander.CheckDuplicate(result.Name)
			items = append(items, state.Item{
				Name:   result.Name,
				State:  state.StateUntracked,
				Domain: "dotfile",
				Path:   result.Path,
				Metadata: map[string]interface{}{
					"isDirectory": true,
					"expanded":    true,
				},
			})
		} else {
			// Single file or unexpanded directory
			expander.CheckDuplicate(result.Name)
			items = append(items, state.Item{
				Name:     result.Name,
				State:    state.StateUntracked,
				Domain:   "dotfile",
				Path:     result.Path,
				Metadata: result.Metadata,
			})
		}
	}

	return items, nil
}
