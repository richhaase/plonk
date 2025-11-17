// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/ignore"
)

// DirectoryScannerImpl implements DirectoryScanner interface
type DirectoryScannerImpl struct {
	homeDir       string
	configDir     string
	pathValidator PathValidator
	pathResolver  PathResolver
}

// NewDirectoryScanner creates a new directory scanner
func NewDirectoryScanner(homeDir, configDir string, validator PathValidator, resolver PathResolver) *DirectoryScannerImpl {
	return &DirectoryScannerImpl{
		homeDir:       homeDir,
		configDir:     configDir,
		pathValidator: validator,
		pathResolver:  resolver,
	}
}

// ExpandDirectoryPaths walks a directory and returns individual file paths
func (ds *DirectoryScannerImpl) ExpandDirectoryPaths(dirPath string) ([]DirectoryEntry, error) {
	var entries []DirectoryEntry

	resolvedDirPath, err := ds.pathResolver.ResolveDotfilePath(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory path: %w", err)
	}

	info, err := os.Stat(resolvedDirPath)
	if err != nil {
		return nil, fmt.Errorf("directory does not exist: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	err = filepath.Walk(resolvedDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(resolvedDirPath, path)
		if err != nil {
			return err
		}

		entries = append(entries, DirectoryEntry{
			RelativePath: relPath,
			FullPath:     path,
			ParentDir:    dirPath,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return entries, nil
}

// ExpandConfigDirectory walks the config directory and returns all files suitable for dotfile management
func (ds *DirectoryScannerImpl) ExpandConfigDirectory(ignorePatterns []string) (map[string]string, error) {
	result := make(map[string]string)
	matcher := ignore.NewMatcher(ignorePatterns)

	err := filepath.Walk(ds.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		relPath, err := filepath.Rel(ds.configDir, path)
		if err != nil {
			return nil
		}

		// Always skip plonk config file
		if relPath == "plonk.yaml" {
			return nil
		}

		// Skip files based on ignore patterns
		if ds.pathValidator.ShouldSkipPath(relPath, info, matcher) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories themselves (we'll get the files inside)
		if info.IsDir() {
			return nil
		}

		// Add to results with proper mapping
		source := relPath
		target := SourceToTarget(source)
		result[source] = target

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk config directory: %w", err)
	}

	return result, nil
}

// ListDotfiles finds all dotfiles in the specified directory
func (ds *DirectoryScannerImpl) ListDotfiles(dir string) ([]string, error) {
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

// ExpandDirectory walks a directory and returns individual file entries as DotfileInfo
func (ds *DirectoryScannerImpl) ExpandDirectory(sourceDir, destDir string) ([]DotfileInfo, error) {
	var items []DotfileInfo
	sourcePath := filepath.Join(ds.configDir, sourceDir)

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		source := filepath.Join(sourceDir, relPath)
		destination := filepath.Join(destDir, relPath)
		name := ds.destinationToName(destination)

		items = append(items, DotfileInfo{
			Name:        name,
			Source:      source,
			Destination: destination,
			IsDirectory: false,
			ParentDir:   sourceDir,
			Metadata: map[string]interface{}{
				"source":      source,
				"destination": destination,
				"parent_dir":  sourceDir,
			},
		})

		return nil
	})

	return items, err
}

// destinationToName converts a destination path to a standardized name
func (ds *DirectoryScannerImpl) destinationToName(destination string) string {
	if strings.HasPrefix(destination, "~/") {
		return destination[2:]
	}
	return destination
}
