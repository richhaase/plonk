// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package paths

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
)

// PathValidator provides security and policy validation for dotfile paths
type PathValidator struct {
	homeDir        string
	configDir      string
	ignorePatterns []string
}

// NewPathValidator creates a new path validator
func NewPathValidator(homeDir, configDir string, ignorePatterns []string) *PathValidator {
	return &PathValidator{
		homeDir:        homeDir,
		configDir:      configDir,
		ignorePatterns: ignorePatterns,
	}
}

// NewPathValidatorFromDefaults creates a path validator with default ignore patterns
func NewPathValidatorFromDefaults(homeDir, configDir string) *PathValidator {
	defaults := config.GetDefaults()
	return NewPathValidator(homeDir, configDir, defaults.IgnorePatterns)
}

// ValidateSecure performs security validation on a path
func (v *PathValidator) ValidateSecure(path string) error {
	// Check for directory traversal attempts
	if strings.Contains(path, "..") {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate",
			"path contains directory traversal: "+path)
	}

	// Check for null bytes (potential security issue)
	if strings.Contains(path, "\x00") {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate",
			"path contains null bytes: "+path)
	}

	// Resolve to absolute path for boundary checking
	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "validate",
			"failed to resolve absolute path")
	}

	// Ensure path is within home directory
	if !strings.HasPrefix(absPath, v.homeDir) {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate",
			"path is outside home directory")
	}

	// Prevent access to plonk config directory
	if strings.HasPrefix(absPath, v.configDir) {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate",
			"cannot manage plonk configuration directory")
	}

	return nil
}

// ShouldSkipPath determines if a path should be skipped based on ignore patterns
func (v *PathValidator) ShouldSkipPath(relPath string, info os.FileInfo) bool {
	// Always skip plonk config file
	if relPath == "plonk.yaml" || relPath == "plonk.lock" {
		return true
	}

	// Check against ignore patterns
	for _, pattern := range v.ignorePatterns {
		// Handle directory patterns (ending with /)
		if strings.HasSuffix(pattern, "/") {
			dirPattern := strings.TrimSuffix(pattern, "/")
			if info.IsDir() && strings.Contains(relPath, dirPattern) {
				return true
			}
			// Also skip files within ignored directories
			if strings.Contains(relPath, dirPattern+"/") {
				return true
			}
		} else {
			// Handle file patterns
			matched, err := filepath.Match(pattern, filepath.Base(relPath))
			if err == nil && matched {
				return true
			}
			// Also check full path for patterns like "*.backup"
			matched, err = filepath.Match(pattern, relPath)
			if err == nil && matched {
				return true
			}
		}
	}

	return false
}

// ValidateDirectory validates that a directory can be safely expanded
func (v *PathValidator) ValidateDirectory(dirPath string) error {
	// First do security validation
	if err := v.ValidateSecure(dirPath); err != nil {
		return err
	}

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "validate",
				"directory does not exist: "+dirPath)
		}
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "validate",
			"failed to access directory")
	}

	// Ensure it's actually a directory
	if !info.IsDir() {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate",
			"path is not a directory: "+dirPath)
	}

	return nil
}

// ValidateFile validates that a file can be safely managed
func (v *PathValidator) ValidateFile(filePath string) error {
	// First do security validation
	if err := v.ValidateSecure(filePath); err != nil {
		return err
	}

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "validate",
				"file does not exist: "+filePath)
		}
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "validate",
			"failed to access file")
	}

	// Ensure it's actually a file
	if info.IsDir() {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate",
			"path is a directory, not a file: "+filePath)
	}

	return nil
}

// GetIgnorePatterns returns the current ignore patterns
func (v *PathValidator) GetIgnorePatterns() []string {
	return v.ignorePatterns
}

// AddIgnorePattern adds a new ignore pattern
func (v *PathValidator) AddIgnorePattern(pattern string) {
	v.ignorePatterns = append(v.ignorePatterns, pattern)
}
