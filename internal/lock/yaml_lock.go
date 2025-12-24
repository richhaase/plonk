// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/richhaase/plonk/internal/dotfiles"
	"gopkg.in/yaml.v3"
)

// YAMLLockService implements LockService using YAML storage
type YAMLLockService struct {
	lockPath string
}

// NewYAMLLockService creates a new YAML-based lock service
func NewYAMLLockService(configDir string) *YAMLLockService {
	return &YAMLLockService{
		lockPath: filepath.Join(configDir, LockFileName),
	}
}

// Read reads and parses the lock file
func (s *YAMLLockService) Read() (*Lock, error) {
	// If lock file doesn't exist, return empty lock
	if _, err := os.Stat(s.lockPath); os.IsNotExist(err) {
		return &Lock{
			Version:   LockFileVersion,
			Resources: []ResourceEntry{},
		}, nil
	}

	data, err := os.ReadFile(s.lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file %s: %w", s.lockPath, err)
	}

	var lock Lock
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file %s: %w", s.lockPath, err)
	}

	// Check version
	if lock.Version != LockFileVersion {
		return nil, fmt.Errorf("unsupported lock file version %d (expected %d). Please remove %s and reinstall your packages",
			lock.Version, LockFileVersion, LockFileName)
	}

	// Initialize resources if nil
	if lock.Resources == nil {
		lock.Resources = []ResourceEntry{}
	}

	return &lock, nil
}

// Write writes the lock data to disk
func (s *YAMLLockService) Write(lock *Lock) error {
	if lock == nil {
		return errors.New("cannot write nil lock")
	}

	// Ensure version is set
	if lock.Version == 0 {
		lock.Version = LockFileVersion
	}

	// Marshal to YAML
	data, err := yaml.Marshal(lock)
	if err != nil {
		return err
	}

	// Use atomic write to ensure safety
	writer := dotfiles.NewAtomicFileWriter()
	if err := writer.WriteFile(s.lockPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file %s: %w", s.lockPath, err)
	}

	return nil
}

// AddPackage adds a package to the lock file with metadata
func (s *YAMLLockService) AddPackage(manager, name, version string, metadata map[string]interface{}) error {
	lock, err := s.Read()
	if err != nil {
		return err
	}

	// Create package ID
	packageID := fmt.Sprintf("%s:%s", manager, name)

	// Check if package already exists and update it
	for i, resource := range lock.Resources {
		if resource.Type == "package" && resource.ID == packageID {
			// Update existing package
			lock.Resources[i].InstalledAt = time.Now().Format(time.RFC3339)
			lock.Resources[i].Metadata = metadata
			return s.Write(lock)
		}
	}

	// Add new package
	resource := ResourceEntry{
		Type:        "package",
		ID:          packageID,
		InstalledAt: time.Now().Format(time.RFC3339),
		Metadata:    metadata,
	}

	lock.Resources = append(lock.Resources, resource)
	return s.Write(lock)
}

// RemovePackage removes a package from the lock file
func (s *YAMLLockService) RemovePackage(manager, name string) error {
	lock, err := s.Read()
	if err != nil {
		return err
	}

	// Create package ID
	packageID := fmt.Sprintf("%s:%s", manager, name)

	// Filter out the package
	filtered := make([]ResourceEntry, 0, len(lock.Resources))
	found := false
	for _, resource := range lock.Resources {
		if resource.Type == "package" && resource.ID == packageID {
			found = true
		} else {
			filtered = append(filtered, resource)
		}
	}

	if !found {
		return nil // Package wasn't in lock file
	}

	lock.Resources = filtered
	return s.Write(lock)
}

// GetPackages returns all packages for a specific manager
func (s *YAMLLockService) GetPackages(manager string) ([]ResourceEntry, error) {
	lock, err := s.Read()
	if err != nil {
		return nil, err
	}

	var packages []ResourceEntry
	for _, resource := range lock.Resources {
		if resource.Type == "package" {
			// Check if this package belongs to the requested manager
			if mgr, ok := resource.Metadata["manager"].(string); ok && mgr == manager {
				packages = append(packages, resource)
			}
		}
	}

	return packages, nil
}

// HasPackage checks if a package exists in the lock file
func (s *YAMLLockService) HasPackage(manager, name string) bool {
	lock, err := s.Read()
	if err != nil {
		return false
	}

	// Create package ID
	packageID := fmt.Sprintf("%s:%s", manager, name)

	for _, resource := range lock.Resources {
		if resource.Type == "package" && resource.ID == packageID {
			return true
		}
	}

	return false
}

// FindPackage returns all locations where a package is installed
func (s *YAMLLockService) FindPackage(name string) []ResourceEntry {
	lock, err := s.Read()
	if err != nil {
		return nil
	}

	var locations []ResourceEntry

	for _, resource := range lock.Resources {
		if resource.Type == "package" {
			// Check if the package name matches
			if pkgName, ok := resource.Metadata["name"].(string); ok && pkgName == name {
				locations = append(locations, resource)
			}
		}
	}

	return locations
}

// GetLockPath returns the path to the lock file
func (s *YAMLLockService) GetLockPath() string {
	return s.lockPath
}
