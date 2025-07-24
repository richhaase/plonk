// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestYAMLLockService_SaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create service
	service := NewYAMLLockService(tmpDir)

	// Test loading non-existent file returns empty lock
	lock, err := service.Load()
	if err != nil {
		t.Fatalf("Load empty lock failed: %v", err)
	}
	if lock.Version != LockFileVersion {
		t.Errorf("Expected version %d, got %d", LockFileVersion, lock.Version)
	}
	if len(lock.Packages) != 0 {
		t.Errorf("Expected empty packages, got %d", len(lock.Packages))
	}

	// Create test lock file
	testLock := &LockFile{
		Version: LockFileVersion,
		Packages: map[string][]PackageEntry{
			"homebrew": {
				{Name: "git", Version: "2.43.0", InstalledAt: time.Now()},
				{Name: "vim", Version: "9.0", InstalledAt: time.Now()},
			},
			"npm": {
				{Name: "typescript", Version: "5.3.3", InstalledAt: time.Now()},
			},
		},
	}

	// Save lock file
	if err := service.Save(testLock); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	lockPath := filepath.Join(tmpDir, LockFileName)
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("Lock file not created: %v", err)
	}

	// Load and verify
	loaded, err := service.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != testLock.Version {
		t.Errorf("Version mismatch: expected %d, got %d", testLock.Version, loaded.Version)
	}

	if len(loaded.Packages["homebrew"]) != 2 {
		t.Errorf("Homebrew packages mismatch: expected 2, got %d", len(loaded.Packages["homebrew"]))
	}

	if len(loaded.Packages["npm"]) != 1 {
		t.Errorf("NPM packages mismatch: expected 1, got %d", len(loaded.Packages["npm"]))
	}
}

func TestYAMLLockService_AddPackage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	service := NewYAMLLockService(tmpDir)

	// Add first package
	if err := service.AddPackage("homebrew", "git", "2.43.0"); err != nil {
		t.Fatalf("AddPackage failed: %v", err)
	}

	// Verify package was added
	if !service.HasPackage("homebrew", "git") {
		t.Error("Package not found after adding")
	}

	// Add another package
	if err := service.AddPackage("homebrew", "vim", "9.0"); err != nil {
		t.Fatalf("AddPackage failed: %v", err)
	}

	// Update existing package
	if err := service.AddPackage("homebrew", "git", "2.44.0"); err != nil {
		t.Fatalf("Update package failed: %v", err)
	}

	// Verify update
	packages, err := service.GetPackages("homebrew")
	if err != nil {
		t.Fatalf("GetPackages failed: %v", err)
	}

	for _, pkg := range packages {
		if pkg.Name == "git" && pkg.Version != "2.44.0" {
			t.Errorf("Package version not updated: expected 2.44.0, got %s", pkg.Version)
		}
	}
}

func TestYAMLLockService_RemovePackage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	service := NewYAMLLockService(tmpDir)

	// Add packages
	service.AddPackage("homebrew", "git", "2.43.0")
	service.AddPackage("homebrew", "vim", "9.0")
	service.AddPackage("npm", "typescript", "5.3.3")

	// Remove package
	if err := service.RemovePackage("homebrew", "git"); err != nil {
		t.Fatalf("RemovePackage failed: %v", err)
	}

	// Verify removal
	if service.HasPackage("homebrew", "git") {
		t.Error("Package still exists after removal")
	}

	// Verify other packages remain
	if !service.HasPackage("homebrew", "vim") {
		t.Error("Other package was removed")
	}

	// Remove all packages from a manager
	service.RemovePackage("homebrew", "vim")

	// Verify manager entry is removed when empty
	packages, _ := service.GetPackages("homebrew")
	if len(packages) != 0 {
		t.Error("Manager entry not removed when empty")
	}

	// Remove non-existent package should not error
	if err := service.RemovePackage("homebrew", "nonexistent"); err != nil {
		t.Errorf("Removing non-existent package should not error: %v", err)
	}
}
