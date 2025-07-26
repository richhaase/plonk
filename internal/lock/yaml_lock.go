// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"gopkg.in/yaml.v3"
)

// Use constants from the centralized constants package

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

// Load reads and parses the lock file
func (s *YAMLLockService) Load() (*LockFile, error) {
	// If lock file doesn't exist, return empty lock file
	if _, err := os.Stat(s.lockPath); os.IsNotExist(err) {
		return &LockFile{
			Version:  LockFileVersion,
			Packages: make(map[string][]PackageEntry),
		}, nil
	}

	data, err := os.ReadFile(s.lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file %s: %w", s.lockPath, err)
	}

	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file %s: %w", s.lockPath, err)
	}

	// Initialize packages map if nil
	if lock.Packages == nil {
		lock.Packages = make(map[string][]PackageEntry)
	}

	return &lock, nil
}

// Read reads and parses the lock file into LockData structure with version detection
func (s *YAMLLockService) Read() (*LockData, error) {
	// If lock file doesn't exist, return empty lock data
	if _, err := os.Stat(s.lockPath); os.IsNotExist(err) {
		return &LockData{
			Version:   CurrentVersion,
			Packages:  make(map[string][]Package),
			Resources: []ResourceEntry{},
		}, nil
	}

	data, err := os.ReadFile(s.lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file %s: %w", s.lockPath, err)
	}

	// Read raw YAML to determine version
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse lock file %s: %w", s.lockPath, err)
	}

	version := 1
	if v, ok := raw["version"].(int); ok {
		version = v
	}

	switch version {
	case 1:
		return s.readV1(raw)
	case 2:
		return s.readV2(raw)
	default:
		return nil, fmt.Errorf("unsupported lock version: %d", version)
	}
}

// readV1 reads a v1 lock file and converts it to LockData
func (s *YAMLLockService) readV1(raw map[string]interface{}) (*LockData, error) {
	var v1Lock LockFile
	data, err := yaml.Marshal(raw)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &v1Lock); err != nil {
		return nil, err
	}

	// Convert v1 to LockData format
	lockData := &LockData{
		Version:   1,
		Packages:  make(map[string][]Package),
		Resources: []ResourceEntry{},
	}

	// Convert PackageEntry to Package
	for manager, packages := range v1Lock.Packages {
		lockData.Packages[manager] = make([]Package, len(packages))
		for i, pkg := range packages {
			lockData.Packages[manager][i] = Package{
				Name:        pkg.Name,
				Version:     pkg.Version,
				InstalledAt: pkg.InstalledAt.Format(time.RFC3339),
			}
		}
	}

	return lockData, nil
}

// readV2 reads a v2 lock file into LockData
func (s *YAMLLockService) readV2(raw map[string]interface{}) (*LockData, error) {
	var v2Lock LockV2
	data, err := yaml.Marshal(raw)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &v2Lock); err != nil {
		return nil, err
	}

	return &LockData{
		Version:   v2Lock.Version,
		Packages:  v2Lock.Packages,
		Resources: v2Lock.Resources,
	}, nil
}

// Write writes the lock data to disk (always as v2)
func (s *YAMLLockService) Write(data *LockData) error {
	if data == nil {
		return errors.New("cannot write nil lock data")
	}

	// Always write as v2
	v2Data := s.migrateToV2(data)

	// Log migration if it occurred
	if data.Version < CurrentVersion {
		log.Printf("Migrated lock file from v%d to v%d", data.Version, CurrentVersion)
	}

	return s.writeV2(v2Data)
}

// migrateToV2 converts LockData to v2 format
func (s *YAMLLockService) migrateToV2(data *LockData) *LockV2 {
	v2Data := &LockV2{
		Version:   CurrentVersion,
		Packages:  data.Packages,
		Resources: data.Resources,
	}

	// If packages exist but no resources, convert packages to resources
	if len(data.Packages) > 0 && len(data.Resources) == 0 {
		for manager, packages := range data.Packages {
			for _, pkg := range packages {
				resource := ResourceEntry{
					Type:        "package",
					ID:          fmt.Sprintf("%s:%s", manager, pkg.Name),
					State:       "managed",
					InstalledAt: pkg.InstalledAt,
					Metadata: map[string]interface{}{
						"manager": manager,
						"name":    pkg.Name,
						"version": pkg.Version,
					},
				}
				v2Data.Resources = append(v2Data.Resources, resource)
			}
		}
	}

	return v2Data
}

// writeV2 writes a v2 lock file to disk
func (s *YAMLLockService) writeV2(v2Data *LockV2) error {
	// Marshal to YAML
	data, err := yaml.Marshal(v2Data)
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

// Save writes the lock file to disk
func (s *YAMLLockService) Save(lock *LockFile) error {
	if lock == nil {
		return errors.New("cannot save nil lock file")
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

// PackageLocation represents where a package is installed
type PackageLocation struct {
	Manager string
	Entry   PackageEntry
}

// FindPackage returns all locations where a package is installed
func (s *YAMLLockService) FindPackage(name string) []PackageLocation {
	lock, err := s.Load()
	if err != nil {
		return nil
	}

	var locations []PackageLocation

	for manager, packages := range lock.Packages {
		for _, pkg := range packages {
			if pkg.Name == name {
				locations = append(locations, PackageLocation{
					Manager: manager,
					Entry:   pkg,
				})
			}
		}
	}

	return locations
}

// GetLockPath returns the path to the lock file
func (s *YAMLLockService) GetLockPath() string {
	return s.lockPath
}
