// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import (
	"os"
	"path/filepath"
	"time"

	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/errors"
	"gopkg.in/yaml.v3"
)

const (
	lockFileName   = "plonk.lock"
	currentVersion = 1
)

// YAMLLockService implements LockService using YAML storage
type YAMLLockService struct {
	lockPath string
}

// NewYAMLLockService creates a new YAML-based lock service
func NewYAMLLockService(configDir string) *YAMLLockService {
	return &YAMLLockService{
		lockPath: filepath.Join(configDir, lockFileName),
	}
}

// Load reads and parses the lock file
func (s *YAMLLockService) Load() (*LockFile, error) {
	// If lock file doesn't exist, return empty lock file
	if _, err := os.Stat(s.lockPath); os.IsNotExist(err) {
		return &LockFile{
			Version:  currentVersion,
			Packages: make(map[string][]PackageEntry),
		}, nil
	}

	data, err := os.ReadFile(s.lockPath)
	if err != nil {
		return nil, errors.ConfigError(errors.ErrFileIO, "Load", "Failed to read lock file").
			WithCause(err).
			WithMetadata("path", s.lockPath)
	}

	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, errors.ConfigError(errors.ErrConfigParseFailure, "Load", "Failed to parse lock file").
			WithCause(err).
			WithMetadata("path", s.lockPath)
	}

	// Initialize packages map if nil
	if lock.Packages == nil {
		lock.Packages = make(map[string][]PackageEntry)
	}

	return &lock, nil
}

// Save writes the lock file to disk
func (s *YAMLLockService) Save(lock *LockFile) error {
	if lock == nil {
		return errors.ConfigError(errors.ErrInvalidInput, "Save", "Cannot save nil lock file")
	}

	// Ensure version is set
	if lock.Version == 0 {
		lock.Version = currentVersion
	}

	// Marshal to YAML
	data, err := yaml.Marshal(lock)
	if err != nil {
		return errors.ConfigError(errors.ErrInternal, "Save", "Failed to marshal lock file").
			WithCause(err)
	}

	// Use atomic write to ensure safety
	writer := dotfiles.NewAtomicFileWriter()
	if err := writer.WriteFile(s.lockPath, data, 0644); err != nil {
		return errors.ConfigError(errors.ErrFileIO, "Save", "Failed to write lock file").
			WithCause(err).
			WithMetadata("path", s.lockPath)
	}

	return nil
}

// AddPackage adds a package to the lock file
func (s *YAMLLockService) AddPackage(manager, name, version string) error {
	lock, err := s.Load()
	if err != nil {
		return err
	}

	// Initialize manager slice if needed
	if lock.Packages[manager] == nil {
		lock.Packages[manager] = []PackageEntry{}
	}

	// Check if package already exists
	for i, pkg := range lock.Packages[manager] {
		if pkg.Name == name {
			// Update existing package
			lock.Packages[manager][i].Version = version
			lock.Packages[manager][i].InstalledAt = time.Now()
			return s.Save(lock)
		}
	}

	// Add new package
	lock.Packages[manager] = append(lock.Packages[manager], PackageEntry{
		Name:        name,
		Version:     version,
		InstalledAt: time.Now(),
	})

	return s.Save(lock)
}

// RemovePackage removes a package from the lock file
func (s *YAMLLockService) RemovePackage(manager, name string) error {
	lock, err := s.Load()
	if err != nil {
		return err
	}

	packages, exists := lock.Packages[manager]
	if !exists {
		return nil // Nothing to remove
	}

	// Filter out the package
	filtered := make([]PackageEntry, 0, len(packages))
	found := false
	for _, pkg := range packages {
		if pkg.Name != name {
			filtered = append(filtered, pkg)
		} else {
			found = true
		}
	}

	if !found {
		return nil // Package wasn't in lock file
	}

	// Update packages list
	if len(filtered) == 0 {
		delete(lock.Packages, manager)
	} else {
		lock.Packages[manager] = filtered
	}

	return s.Save(lock)
}

// GetPackages returns all packages for a specific manager
func (s *YAMLLockService) GetPackages(manager string) ([]PackageEntry, error) {
	lock, err := s.Load()
	if err != nil {
		return nil, err
	}

	packages, exists := lock.Packages[manager]
	if !exists {
		return []PackageEntry{}, nil
	}

	// Return a copy to prevent external modification
	result := make([]PackageEntry, len(packages))
	copy(result, packages)
	return result, nil
}

// HasPackage checks if a package exists in the lock file
func (s *YAMLLockService) HasPackage(manager, name string) bool {
	lock, err := s.Load()
	if err != nil {
		return false
	}

	packages, exists := lock.Packages[manager]
	if !exists {
		return false
	}

	for _, pkg := range packages {
		if pkg.Name == name {
			return true
		}
	}

	return false
}

// GetLockPath returns the path to the lock file
func (s *YAMLLockService) GetLockPath() string {
	return s.lockPath
}
