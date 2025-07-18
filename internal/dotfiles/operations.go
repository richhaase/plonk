// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package dotfiles provides core dotfile management operations including
// file discovery, path resolution, directory expansion, and file operations.
package dotfiles

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
)

// Manager handles dotfile operations and path management
type Manager struct {
	homeDir   string
	configDir string
}

// NewManager creates a new dotfile manager
func NewManager(homeDir, configDir string) *Manager {
	return &Manager{
		homeDir:   homeDir,
		configDir: configDir,
	}
}

// DotfileInfo represents information about a dotfile
type DotfileInfo struct {
	Name        string
	Source      string // Path in config directory
	Destination string // Path in home directory
	IsDirectory bool
	ParentDir   string // For files expanded from directories
	Metadata    map[string]interface{}
}

// ListDotfiles finds all dotfiles in the specified directory
func (m *Manager) ListDotfiles(dir string) ([]string, error) {
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

// ExpandDirectory walks a directory and returns individual file entries
func (m *Manager) ExpandDirectory(sourceDir, destDir string) ([]DotfileInfo, error) {
	var items []DotfileInfo
	sourcePath := filepath.Join(m.configDir, sourceDir)

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path from source directory
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		// Build source and destination paths
		source := filepath.Join(sourceDir, relPath)
		destination := filepath.Join(destDir, relPath)
		name := m.DestinationToName(destination)

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

// DestinationToName converts a destination path to a standardized name
func (m *Manager) DestinationToName(destination string) string {
	// Remove ~/ prefix if present
	if strings.HasPrefix(destination, "~/") {
		return destination[2:]
	}
	return destination
}

// ExpandPath expands ~ to home directory
func (m *Manager) ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(m.homeDir, path[2:])
	}
	return path
}

// GetSourcePath returns the full source path for a dotfile
func (m *Manager) GetSourcePath(source string) string {
	return filepath.Join(m.configDir, source)
}

// GetDestinationPath returns the full destination path for a dotfile
func (m *Manager) GetDestinationPath(destination string) string {
	return m.ExpandPath(destination)
}

// FileExists checks if a file exists at the given path
func (m *Manager) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func (m *Manager) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CreateDotfileInfo creates a DotfileInfo from source and destination paths
func (m *Manager) CreateDotfileInfo(source, destination string) DotfileInfo {
	sourcePath := m.GetSourcePath(source)
	isDir := m.IsDirectory(sourcePath)

	return DotfileInfo{
		Name:        m.DestinationToName(destination),
		Source:      source,
		Destination: destination,
		IsDirectory: isDir,
		Metadata: map[string]interface{}{
			"source":      source,
			"destination": destination,
		},
	}
}

// ValidatePaths validates that source and destination paths are valid
func (m *Manager) ValidatePaths(source, destination string) error {
	// Check if source exists in config directory
	sourcePath := m.GetSourcePath(source)
	if !m.FileExists(sourcePath) {
		return errors.DotfileError(errors.ErrFileNotFound, "validate",
			"source file does not exist").
			WithItem(source).
			WithMetadata("source_path", sourcePath)
	}

	// Validate destination path format
	if !strings.HasPrefix(destination, "~/") && !filepath.IsAbs(destination) {
		return errors.DotfileError(errors.ErrPathValidation, "validate",
			"destination must start with ~/ or be absolute").
			WithItem(destination).
			WithMetadata("destination", destination)
	}

	return nil
}
