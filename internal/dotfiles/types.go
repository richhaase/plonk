// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"

	"github.com/richhaase/plonk/internal/ignore"
	"github.com/richhaase/plonk/internal/resources"
)

// PathResolver handles all path resolution and generation operations
type PathResolver interface {
	ResolveDotfilePath(path string) (string, error)
	GetSourcePath(source string) string
	GetDestinationPath(destination string) (string, error)
	GenerateDestinationPath(resolvedPath string) (string, error)
	GenerateSourcePath(destination string) string
	GeneratePaths(resolvedPath string) (source, destination string, err error)
}

// PathValidator handles path validation and safety checks
type PathValidator interface {
	ValidatePath(path string) error
	ValidatePaths(source, destination string) error
	ShouldSkipPath(relPath string, info os.FileInfo, matcher *ignore.Matcher) bool
}

// DirectoryScanner handles directory operations and file discovery
type DirectoryScanner interface {
	ExpandDirectoryPaths(dirPath string) ([]DirectoryEntry, error)
	ExpandConfigDirectory(ignorePatterns []string) (map[string]string, error)
	ListDotfiles(dir string) ([]string, error)
	ExpandDirectory(sourceDir, destDir string) ([]DotfileInfo, error)
}

// ConfigHandler manages dotfile configuration operations
type ConfigHandler interface {
	GetConfiguredDotfiles() ([]resources.Item, error)
	GetActualDotfiles(ctx context.Context) ([]resources.Item, error)
}

// FileComparator handles file comparison operations
type FileComparator interface {
	CompareFiles(path1, path2 string) (bool, error)
	ComputeFileHash(path string) (string, error)
}

// DotfileInfo represents information about a dotfile (already exists in manager.go)
// Keep existing definition to avoid breaking changes

// DirectoryEntry represents a file found during directory expansion (already exists in manager.go)
// Keep existing definition to avoid breaking changes

// AddOptions configures dotfile addition operations (already exists in manager.go)
// Keep existing definition to avoid breaking changes

// RemoveOptions configures dotfile removal operations (already exists in manager.go)
// Keep existing definition to avoid breaking changes

// ApplyOptions configures dotfile apply operations (already exists in manager.go)
// Keep existing definition to avoid breaking changes
