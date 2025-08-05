// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathResolverImpl implements PathResolver interface
type PathResolverImpl struct {
	homeDir   string
	configDir string
}

// NewPathResolver creates a new path resolver
func NewPathResolver(homeDir, configDir string) *PathResolverImpl {
	return &PathResolverImpl{
		homeDir:   homeDir,
		configDir: configDir,
	}
}

// ResolveDotfilePath resolves a dotfile path to an absolute path within the home directory
func (pr *PathResolverImpl) ResolveDotfilePath(path string) (string, error) {
	var resolvedPath string

	if strings.HasPrefix(path, "~/") {
		resolvedPath = filepath.Join(pr.homeDir, path[2:])
	} else if filepath.IsAbs(path) {
		resolvedPath = path
	} else {
		// Relative path - try current directory first, then home directory
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}

		if _, err := os.Stat(absPath); err == nil {
			resolvedPath = absPath
		} else {
			homeRelativePath := filepath.Join(pr.homeDir, path)
			if _, err := os.Stat(homeRelativePath); err == nil {
				resolvedPath = homeRelativePath
			} else {
				resolvedPath = absPath
			}
		}
	}

	// Ensure it's within the home directory
	if !strings.HasPrefix(resolvedPath, pr.homeDir) {
		return "", fmt.Errorf("dotfile must be within home directory: %s", resolvedPath)
	}

	return resolvedPath, nil
}

// GetSourcePath returns the full path to a source file in the config directory
func (pr *PathResolverImpl) GetSourcePath(source string) string {
	return filepath.Join(pr.configDir, source)
}

// GetDestinationPath converts a destination path to an absolute path in the home directory
func (pr *PathResolverImpl) GetDestinationPath(destination string) (string, error) {
	if strings.HasPrefix(destination, "~/") {
		return filepath.Join(pr.homeDir, destination[2:]), nil
	}
	if filepath.IsAbs(destination) {
		return destination, nil
	}
	// Relative destination, assume it's relative to home
	return filepath.Join(pr.homeDir, destination), nil
}

// GenerateDestinationPath converts a resolved absolute path to a destination path (~/relative/path)
func (pr *PathResolverImpl) GenerateDestinationPath(resolvedPath string) (string, error) {
	relPath, err := filepath.Rel(pr.homeDir, resolvedPath)
	if err != nil {
		return "", err
	}
	return "~/" + relPath, nil
}

// GenerateSourcePath converts a destination path to a source path using plonk's naming convention
func (pr *PathResolverImpl) GenerateSourcePath(destination string) string {
	return TargetToSource(destination)
}

// GeneratePaths generates both source and destination paths for a resolved dotfile path
func (pr *PathResolverImpl) GeneratePaths(resolvedPath string) (source, destination string, err error) {
	destination, err = pr.GenerateDestinationPath(resolvedPath)
	if err != nil {
		return "", "", err
	}
	source = pr.GenerateSourcePath(destination)
	return source, destination, nil
}
