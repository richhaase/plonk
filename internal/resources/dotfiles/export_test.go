// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"github.com/richhaase/plonk/internal/config"
)

// NewManager creates a new dotfile manager with all components.
// This is a test helper - production code should use NewManagerWithConfig.
func NewManager(homeDir, configDir string) *Manager {
	pathResolver := NewPathResolver(homeDir, configDir)
	pathValidator := NewPathValidator(homeDir, configDir)
	directoryScanner := NewDirectoryScanner(homeDir, configDir, pathValidator, pathResolver)
	fileComparator := NewFileComparator()
	configHandler := newConfigHandler(homeDir, configDir, pathResolver, directoryScanner, fileComparator)
	fileOperations := NewFileOperations(pathResolver)

	return &Manager{
		homeDir:          homeDir,
		configDir:        configDir,
		pathResolver:     pathResolver,
		pathValidator:    pathValidator,
		directoryScanner: directoryScanner,
		configHandler:    configHandler,
		fileComparator:   fileComparator,
		fileOperations:   fileOperations,
	}
}

// newConfigHandler creates a new config handler (loads config internally).
// This is a test helper - production code should use NewConfigHandlerWithConfig.
func newConfigHandler(homeDir, configDir string, resolver PathResolver, scanner DirectoryScanner, comparator FileComparator) *ConfigHandlerImpl {
	return &ConfigHandlerImpl{
		homeDir:          homeDir,
		configDir:        configDir,
		pathResolver:     resolver,
		directoryScanner: scanner,
		fileComparator:   comparator,
		cfg:              config.LoadWithDefaults(configDir),
	}
}

// NewFileOperationsWithWriter allows injecting a custom FileWriter (for testing).
func NewFileOperationsWithWriter(pathResolver PathResolver, writer FileWriter) *FileOperations {
	if writer == nil {
		writer = NewAtomicFileWriter()
	}
	return &FileOperations{
		pathResolver: pathResolver,
		writer:       writer,
	}
}
