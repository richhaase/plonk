// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Expander handles directory expansion for dotfiles
type Expander struct {
	homeDir        string
	expandDirs     []string
	maxDepth       int
	scanner        *Scanner
	duplicateCheck map[string]bool
}

// NewExpander creates a new expander
func NewExpander(homeDir string, expandDirs []string, scanner *Scanner) *Expander {
	return &Expander{
		homeDir:        homeDir,
		expandDirs:     expandDirs,
		maxDepth:       2, // Default max depth for expansion
		scanner:        scanner,
		duplicateCheck: make(map[string]bool),
	}
}

// ExpandDirectory expands a directory to individual file results
func (e *Expander) ExpandDirectory(ctx context.Context, dirPath string, relName string) ([]ScanResult, error) {
	// Use scanner to walk the directory
	results, err := e.scanner.ScanDirectory(ctx, dirPath, e.maxDepth)
	if err != nil {
		// If we can't scan, return the directory itself as a single result
		info, statErr := os.Stat(dirPath)
		if statErr != nil {
			return nil, err
		}
		return []ScanResult{{
			Name: relName,
			Path: dirPath,
			Info: info,
			Metadata: map[string]interface{}{
				"path": dirPath,
			},
		}}, nil
	}

	// Filter out the root directory itself and adjust names
	var expandedResults []ScanResult
	for _, result := range results {
		// Skip the root directory
		if result.Path == dirPath {
			continue
		}

		// Calculate the full relative path from home
		fullRelPath := filepath.Join(relName, result.Name)

		// Check for duplicates
		if e.duplicateCheck[fullRelPath] {
			continue
		}
		e.duplicateCheck[fullRelPath] = true

		expandedResults = append(expandedResults, ScanResult{
			Name: fullRelPath,
			Path: result.Path,
			Info: result.Info,
			Metadata: map[string]interface{}{
				"path": result.Path,
			},
		})
	}

	return expandedResults, nil
}

// CheckDuplicate checks if a path has already been processed
func (e *Expander) CheckDuplicate(name string) bool {
	if e.duplicateCheck[name] {
		return true
	}
	e.duplicateCheck[name] = true
	return false
}

// ExpandConfiguredDestination expands a configured destination path
func (e *Expander) ExpandConfiguredDestination(ctx context.Context, manager *Manager, destPath string) ([]ScanResult, error) {
	var results []ScanResult

	// Check if destination exists
	if !manager.FileExists(destPath) {
		return results, nil
	}

	// If it's a directory, walk it
	if manager.IsDirectory(destPath) {
		err := filepath.Walk(destPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files with errors
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Calculate relative path from home
			relPath, err := filepath.Rel(e.homeDir, path)
			if err != nil {
				return nil
			}

			// Check for duplicates
			if e.CheckDuplicate(relPath) {
				return nil
			}

			results = append(results, ScanResult{
				Name: relPath,
				Path: path,
				Info: info,
				Metadata: map[string]interface{}{
					"path": path,
				},
			})

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Single file
		info, err := os.Stat(destPath)
		if err != nil {
			return results, nil
		}

		relPath, err := filepath.Rel(e.homeDir, destPath)
		if err != nil {
			return results, nil
		}

		// Check for duplicates
		if !e.CheckDuplicate(relPath) {
			results = append(results, ScanResult{
				Name: relPath,
				Path: destPath,
				Info: info,
				Metadata: map[string]interface{}{
					"path": destPath,
				},
			})
		}
	}

	return results, nil
}

// Reset clears the duplicate check cache
func (e *Expander) Reset() {
	e.duplicateCheck = make(map[string]bool)
}

// SetMaxDepth sets the maximum depth for directory expansion
func (e *Expander) SetMaxDepth(depth int) {
	e.maxDepth = depth
}

// CalculateRelativePath calculates the relative path from home directory
func (e *Expander) CalculateRelativePath(path string) (string, error) {
	// If path is already relative, return it
	if !filepath.IsAbs(path) {
		return path, nil
	}

	// Calculate relative to home
	relPath, err := filepath.Rel(e.homeDir, path)
	if err != nil {
		return "", err
	}

	// If the path goes outside home (starts with ..), return the original
	if strings.HasPrefix(relPath, "..") {
		return filepath.Base(path), nil
	}

	return relPath, nil
}
