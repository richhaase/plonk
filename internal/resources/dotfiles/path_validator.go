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

// PathValidatorImpl implements PathValidator interface
type PathValidatorImpl struct {
	homeDir   string
	configDir string
}

// NewPathValidator creates a new path validator
func NewPathValidator(homeDir, configDir string) *PathValidatorImpl {
	return &PathValidatorImpl{
		homeDir:   homeDir,
		configDir: configDir,
	}
}

// ValidatePath validates that a path is safe and within allowed boundaries
func (pv *PathValidatorImpl) ValidatePath(path string) error {
	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}

	// Clean and resolve the path
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return err
	}

	// Ensure path is within home directory
	if !strings.HasPrefix(absPath, pv.homeDir) {
		return fmt.Errorf("path is outside home directory: %s", absPath)
	}

	// Ensure we're not managing plonk's own config
	if strings.HasPrefix(absPath, pv.configDir) {
		return fmt.Errorf("cannot manage plonk configuration directory")
	}

	return nil
}

// ValidatePaths validates that source and destination paths are valid
func (pv *PathValidatorImpl) ValidatePaths(source, destination string) error {
	sourcePath := filepath.Join(pv.configDir, source)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file %s does not exist at %s", source, sourcePath)
	}

	if !strings.HasPrefix(destination, "~/") && !filepath.IsAbs(destination) {
		return fmt.Errorf("destination %s must start with ~/ or be absolute", destination)
	}

	return nil
}

// ShouldSkipPath determines if a path should be skipped based on ignore patterns
func (pv *PathValidatorImpl) ShouldSkipPath(relPath string, info os.FileInfo, matcher *ignore.Matcher) bool {
	// Always skip plonk config files
	if relPath == "plonk.yaml" || relPath == "plonk.lock" {
		return true
	}

	// Always skip .plonk/ directory (reserved for future plonk metadata)
	if relPath == ".plonk" || strings.HasPrefix(relPath, ".plonk/") {
		return true
	}

	if matcher != nil && matcher.ShouldIgnore(relPath, info.IsDir()) {
		return true
	}

	return false
}
