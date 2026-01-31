// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"gopkg.in/yaml.v3"
)

// v2 types - kept only for migration support
type lockV2 struct {
	Version   int              `yaml:"version"`
	Resources []resourceEntry `yaml:"resources"`
}

type resourceEntry struct {
	Type     string                 `yaml:"type"`
	Metadata map[string]interface{} `yaml:"metadata"`
}

// LockV3 represents the simplified v3 lock format
type LockV3 struct {
	Version  int                 `yaml:"version"`
	Packages map[string][]string `yaml:"packages,omitempty"` // manager -> []package
}

// NewLockV3 creates an empty v3 lock
func NewLockV3() *LockV3 {
	return &LockV3{
		Version:  3,
		Packages: make(map[string][]string),
	}
}

// AddPackage adds a package under its manager (maintains sorted order)
func (l *LockV3) AddPackage(manager, pkg string) {
	if l.Packages == nil {
		l.Packages = make(map[string][]string)
	}

	// Check if already exists
	if slices.Contains(l.Packages[manager], pkg) {
		return
	}

	l.Packages[manager] = append(l.Packages[manager], pkg)
	sort.Strings(l.Packages[manager])
}

// RemovePackage removes a package from its manager
func (l *LockV3) RemovePackage(manager, pkg string) {
	if l.Packages == nil {
		return
	}

	pkgs := l.Packages[manager]
	for i, existing := range pkgs {
		if existing == pkg {
			l.Packages[manager] = append(pkgs[:i], pkgs[i+1:]...)
			break
		}
	}

	// Remove manager key if empty
	if len(l.Packages[manager]) == 0 {
		delete(l.Packages, manager)
	}
}

// HasPackage checks if a package is tracked
func (l *LockV3) HasPackage(manager, pkg string) bool {
	return slices.Contains(l.Packages[manager], pkg)
}

// GetPackages returns all packages for a manager
func (l *LockV3) GetPackages(manager string) []string {
	return l.Packages[manager]
}

// GetAllPackages returns all manager:package pairs
func (l *LockV3) GetAllPackages() []string {
	var result []string
	for manager, pkgs := range l.Packages {
		for _, pkg := range pkgs {
			result = append(result, manager+":"+pkg)
		}
	}
	sort.Strings(result)
	return result
}

// LockV3Service handles v3 lock file operations
type LockV3Service struct {
	lockPath string
}

// NewLockV3Service creates a new v3 lock service
func NewLockV3Service(configDir string) *LockV3Service {
	return &LockV3Service{
		lockPath: filepath.Join(configDir, LockFileName),
	}
}

// Read reads the lock file, auto-migrating v2 if needed
func (s *LockV3Service) Read() (*LockV3, error) {
	// If lock file doesn't exist, return empty lock
	if _, err := os.Stat(s.lockPath); os.IsNotExist(err) {
		return NewLockV3(), nil
	}

	data, err := os.ReadFile(s.lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	// Try to detect version
	var versionCheck struct {
		Version int `yaml:"version"`
	}
	if err := yaml.Unmarshal(data, &versionCheck); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	// Handle v2 migration
	if versionCheck.Version == 2 {
		return s.migrateV2(data)
	}

	// Parse v3
	var lock LockV3
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	if lock.Version != 3 {
		return nil, fmt.Errorf("unsupported lock version %d", lock.Version)
	}

	return &lock, nil
}

// Write saves the lock file atomically using temp file + rename
func (s *LockV3Service) Write(lock *LockV3) error {
	if lock == nil {
		return fmt.Errorf("cannot write nil lock")
	}

	// Force version 3
	lock.Version = 3

	data, err := yaml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal lock: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpPath := s.lockPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp lock file: %w", err)
	}

	if err := os.Rename(tmpPath, s.lockPath); err != nil {
		// Clean up temp file on failure
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename lock file: %w", err)
	}

	return nil
}

// migrateV2 converts a v2 lock to v3 format and persists it
func (s *LockV3Service) migrateV2(data []byte) (*LockV3, error) {
	var old lockV2
	if err := yaml.Unmarshal(data, &old); err != nil {
		return nil, fmt.Errorf("failed to parse v2 lock: %w", err)
	}

	v3 := NewLockV3()

	for _, resource := range old.Resources {
		if resource.Type != "package" {
			continue
		}

		// Extract manager and name from metadata
		manager, _ := resource.Metadata["manager"].(string)
		name, _ := resource.Metadata["name"].(string)

		if manager != "" && name != "" {
			v3.AddPackage(manager, name)
		} else {
			// Log warning for packages that can't be migrated
			log.Printf("Warning: skipping v2 package during migration (missing manager=%q or name=%q)", manager, name)
		}
	}

	// Persist the migrated v3 format to disk
	if err := s.Write(v3); err != nil {
		return nil, fmt.Errorf("failed to persist v2 migration: %w", err)
	}

	return v3, nil
}

// GetLockPath returns the path to the lock file
func (s *LockV3Service) GetLockPath() string {
	return s.lockPath
}
