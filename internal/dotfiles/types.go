// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"

	"github.com/richhaase/plonk/internal/ignore"
)

// ItemState represents the reconciliation state of a dotfile
type ItemState int

const (
	StateManaged   ItemState = iota // In config AND present/installed
	StateMissing                    // In config BUT not present/installed
	StateUntracked                  // Present/installed BUT not in config
	StateDegraded                   // In config AND present BUT content differs (drifted)
)

// String returns a human-readable representation of the item state
func (s ItemState) String() string {
	switch s {
	case StateManaged:
		return "managed"
	case StateMissing:
		return "missing"
	case StateUntracked:
		return "untracked"
	case StateDegraded:
		return "drifted"
	default:
		return "unknown"
	}
}

// DotfileItem represents a dotfile with its state and metadata
type DotfileItem struct {
	// Name is the relative path from home directory (e.g., ".bashrc", ".config/nvim/init.vim")
	Name string

	// State is the reconciliation state of this dotfile
	State ItemState

	// Source is the path to the dotfile in the config directory (e.g., "dotfiles/bashrc")
	Source string

	// Destination is the target path where the dotfile should be deployed (e.g., "~/.bashrc")
	Destination string

	// IsTemplate indicates if this dotfile is a template that needs processing
	IsTemplate bool

	// IsDirectory indicates if this dotfile is a directory
	IsDirectory bool

	// CompareFunc is used for drift detection - compares source and destination
	// Returns true if identical (no drift), false if different (drift detected)
	CompareFunc func() (bool, error)

	// Error contains any error message associated with this dotfile
	Error string

	// Metadata contains additional dotfile-specific information
	Metadata map[string]interface{}
}

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
	GetConfiguredDotfiles() ([]DotfileItem, error)
	GetActualDotfiles(ctx context.Context) ([]DotfileItem, error)
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
