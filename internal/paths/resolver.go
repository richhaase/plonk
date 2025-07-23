// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package paths provides centralized path resolution and validation utilities
// for dotfile management operations.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
)

// PathResolver handles dotfile path resolution, validation, and conversion
type PathResolver struct {
	homeDir   string
	configDir string
}

// NewPathResolver creates a new path resolver with the specified directories
func NewPathResolver(homeDir, configDir string) *PathResolver {
	return &PathResolver{
		homeDir:   homeDir,
		configDir: configDir,
	}
}

// NewPathResolverFromDefaults creates a path resolver using default directories
func NewPathResolverFromDefaults() (*PathResolver, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return nil, errors.NewError(errors.ErrInternal, errors.DomainDotfiles, "init",
			"HOME environment variable not set")
	}

	configDir := getDefaultConfigDirectory()
	return NewPathResolver(homeDir, configDir), nil
}

// getDefaultConfigDirectory returns the default config directory, checking PLONK_DIR environment variable first
func getDefaultConfigDirectory() string {
	// Check for PLONK_DIR environment variable
	if envDir := os.Getenv("PLONK_DIR"); envDir != "" {
		// Expand ~ if present
		if strings.HasPrefix(envDir, "~/") {
			return filepath.Join(os.Getenv("HOME"), envDir[2:])
		}
		return envDir
	}

	// Default location
	return filepath.Join(os.Getenv("HOME"), ".config", "plonk")
}

// ResolveDotfilePath resolves a dotfile path to an absolute path within the home directory
// Handles different path types:
// - ~/path: expands ~ to home directory
// - /absolute/path: validates it's within home directory
// - relative/path: tries current directory first, then home directory
func (p *PathResolver) ResolveDotfilePath(path string) (string, error) {
	var resolvedPath string

	// Handle different path types
	if strings.HasPrefix(path, "~/") {
		// Expand ~ to home directory
		resolvedPath = filepath.Join(p.homeDir, path[2:])
	} else if filepath.IsAbs(path) {
		// Already absolute path
		resolvedPath = path
	} else {
		// Relative path - try to resolve relative to current working directory first
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "resolve",
				"failed to resolve path")
		}

		// Check if file exists at the absolute path
		if _, err := os.Stat(absPath); err == nil {
			resolvedPath = absPath
		} else {
			// Fall back to home directory
			homeRelativePath := filepath.Join(p.homeDir, path)
			if _, err := os.Stat(homeRelativePath); err == nil {
				// File exists relative to home directory
				resolvedPath = homeRelativePath
			} else {
				// Neither location has the file, use the current working directory path
				// so the error message will be more intuitive
				resolvedPath = absPath
			}
		}
	}

	// Ensure it's within the home directory
	if !strings.HasPrefix(resolvedPath, p.homeDir) {
		return "", errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate",
			fmt.Sprintf("dotfile must be within home directory: %s", resolvedPath))
	}

	return resolvedPath, nil
}

// GenerateDestinationPath converts a resolved absolute path to a destination path (~/relative/path)
func (p *PathResolver) GenerateDestinationPath(resolvedPath string) (string, error) {
	// Get relative path from home directory
	relPath, err := filepath.Rel(p.homeDir, resolvedPath)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "generate",
			"failed to generate relative path")
	}

	// Generate destination (always relative to home with ~ prefix)
	destination := "~/" + relPath
	return destination, nil
}

// GenerateSourcePath converts a destination path to a source path using plonk's naming convention
func (p *PathResolver) GenerateSourcePath(destination string) string {
	return targetToSource(destination)
}

// targetToSource converts a target path to a source path following plonk conventions
func targetToSource(target string) string {
	// Remove ~/. prefix if present
	if len(target) > 3 && target[:3] == "~/." {
		return target[3:]
	}
	// Remove ~/ prefix if present (shouldn't happen with our convention)
	if len(target) > 2 && target[:2] == "~/" {
		return target[2:]
	}
	return target
}

// GeneratePaths generates both source and destination paths for a resolved dotfile path
func (p *PathResolver) GeneratePaths(resolvedPath string) (source, destination string, err error) {
	destination, err = p.GenerateDestinationPath(resolvedPath)
	if err != nil {
		return "", "", err
	}

	source = p.GenerateSourcePath(destination)
	return source, destination, nil
}

// ValidatePath validates that a path is safe and within allowed boundaries
func (p *PathResolver) ValidatePath(path string) error {
	// Check for directory traversal attempts
	if strings.Contains(path, "..") {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate",
			"path contains directory traversal: "+path)
	}

	// Resolve the path to check final location
	resolvedPath, err := p.ResolveDotfilePath(path)
	if err != nil {
		return err
	}

	// Ensure it's within home directory (already checked in ResolveDotfilePath, but double-check)
	if !strings.HasPrefix(resolvedPath, p.homeDir) {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate",
			fmt.Sprintf("path is outside home directory: %s", resolvedPath))
	}

	return nil
}

// ExpandDirectory walks a directory and returns individual file paths
// Returns both the relative paths within the directory and their full resolved paths
func (p *PathResolver) ExpandDirectory(dirPath string) ([]DirectoryEntry, error) {
	var entries []DirectoryEntry

	// Resolve the directory path first
	resolvedDirPath, err := p.ResolveDotfilePath(dirPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "expand",
			"failed to resolve directory path")
	}

	// Check if it's actually a directory
	info, err := os.Stat(resolvedDirPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFileNotFound, errors.DomainDotfiles, "expand",
			"directory does not exist")
	}

	if !info.IsDir() {
		return nil, errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "expand",
			"path is not a directory: "+dirPath)
	}

	err = filepath.Walk(resolvedDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, only process files
		if info.IsDir() {
			return nil
		}

		// Calculate relative path from the directory being expanded
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
		return nil, errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "expand",
			"failed to walk directory")
	}

	return entries, nil
}

// GetSourcePath returns the full path to a source file in the config directory
func (p *PathResolver) GetSourcePath(source string) string {
	return filepath.Join(p.configDir, source)
}

// GetDestinationPath converts a destination path to an absolute path in the home directory
func (p *PathResolver) GetDestinationPath(destination string) (string, error) {
	if strings.HasPrefix(destination, "~/") {
		return filepath.Join(p.homeDir, destination[2:]), nil
	}
	if filepath.IsAbs(destination) {
		return destination, nil
	}
	// Relative destination, assume it's relative to home
	return filepath.Join(p.homeDir, destination), nil
}

// DirectoryEntry represents a file found during directory expansion
type DirectoryEntry struct {
	RelativePath string // Path relative to the expanded directory
	FullPath     string // Full absolute path to the file
	ParentDir    string // Original directory path that was expanded
}
