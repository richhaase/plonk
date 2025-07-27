// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package lock

import (
	"os"
	"path/filepath"
	"strings"
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
			"brew": {
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

	if len(loaded.Packages["brew"]) != 2 {
		t.Errorf("Homebrew packages mismatch: expected 2, got %d", len(loaded.Packages["brew"]))
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
	if err := service.AddPackage("brew", "git", "2.43.0"); err != nil {
		t.Fatalf("AddPackage failed: %v", err)
	}

	// Verify package was added
	if !service.HasPackage("brew", "git") {
		t.Error("Package not found after adding")
	}

	// Add another package
	if err := service.AddPackage("brew", "vim", "9.0"); err != nil {
		t.Fatalf("AddPackage failed: %v", err)
	}

	// Update existing package
	if err := service.AddPackage("brew", "git", "2.44.0"); err != nil {
		t.Fatalf("Update package failed: %v", err)
	}

	// Verify update
	packages, err := service.GetPackages("brew")
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
	service.AddPackage("brew", "git", "2.43.0")
	service.AddPackage("brew", "vim", "9.0")
	service.AddPackage("npm", "typescript", "5.3.3")

	// Remove package
	if err := service.RemovePackage("brew", "git"); err != nil {
		t.Fatalf("RemovePackage failed: %v", err)
	}

	// Verify removal
	if service.HasPackage("brew", "git") {
		t.Error("Package still exists after removal")
	}

	// Verify other packages remain
	if !service.HasPackage("brew", "vim") {
		t.Error("Other package was removed")
	}

	// Remove all packages from a manager
	service.RemovePackage("brew", "vim")

	// Verify manager entry is removed when empty
	packages, _ := service.GetPackages("brew")
	if len(packages) != 0 {
		t.Error("Manager entry not removed when empty")
	}

	// Remove non-existent package should not error
	if err := service.RemovePackage("brew", "nonexistent"); err != nil {
		t.Errorf("Removing non-existent package should not error: %v", err)
	}
}

func TestYAMLLockService_ReadV1(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a v1 lock file manually
	v1Content := `version: 1
packages:
  brew:
    - name: git
      version: 2.43.0
      installed_at: 2025-01-01T12:00:00Z
    - name: vim
      version: 9.0
      installed_at: 2025-01-01T12:00:00Z
  npm:
    - name: typescript
      version: 5.3.3
      installed_at: 2025-01-01T12:00:00Z
`

	lockPath := filepath.Join(tmpDir, LockFileName)
	if err := os.WriteFile(lockPath, []byte(v1Content), 0644); err != nil {
		t.Fatal(err)
	}

	service := NewYAMLLockService(tmpDir)

	// Test reading v1 format
	lockData, err := service.Read()
	if err != nil {
		t.Fatalf("Read v1 failed: %v", err)
	}

	if lockData.Version != 1 {
		t.Errorf("Expected version 1, got %d", lockData.Version)
	}

	if len(lockData.Packages["brew"]) != 2 {
		t.Errorf("Expected 2 brew packages, got %d", len(lockData.Packages["brew"]))
	}

	if len(lockData.Packages["npm"]) != 1 {
		t.Errorf("Expected 1 npm package, got %d", len(lockData.Packages["npm"]))
	}

	// Verify package data
	gitPkg := lockData.Packages["brew"][0]
	if gitPkg.Name != "git" || gitPkg.Version != "2.43.0" {
		t.Errorf("Git package data incorrect: %+v", gitPkg)
	}
}

func TestYAMLLockService_ReadV2(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a v2 lock file manually
	v2Content := `version: 2
packages:
  brew:
    - name: git
      version: 2.43.0
      installed_at: "2025-01-01T12:00:00Z"
resources:
  - type: package
    id: "brew:git"
    state: managed
    installed_at: "2025-01-01T12:00:00Z"
    metadata:
      manager: brew
      name: git
      version: 2.43.0
  - type: dotfile
    id: ".vimrc"
    state: managed
    installed_at: "2025-01-01T12:00:00Z"
    metadata:
      source: ~/.dotfiles/.vimrc
      target: ~/.vimrc
`

	lockPath := filepath.Join(tmpDir, LockFileName)
	if err := os.WriteFile(lockPath, []byte(v2Content), 0644); err != nil {
		t.Fatal(err)
	}

	service := NewYAMLLockService(tmpDir)

	// Test reading v2 format
	lockData, err := service.Read()
	if err != nil {
		t.Fatalf("Read v2 failed: %v", err)
	}

	if lockData.Version != 2 {
		t.Errorf("Expected version 2, got %d", lockData.Version)
	}

	if len(lockData.Resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(lockData.Resources))
	}

	// Verify resource data
	pkgResource := lockData.Resources[0]
	if pkgResource.Type != "package" || pkgResource.ID != "brew:git" {
		t.Errorf("Package resource data incorrect: %+v", pkgResource)
	}

	dotfileResource := lockData.Resources[1]
	if dotfileResource.Type != "dotfile" || dotfileResource.ID != ".vimrc" {
		t.Errorf("Dotfile resource data incorrect: %+v", dotfileResource)
	}
}

func TestYAMLLockService_Migration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a v1 lock file
	v1Content := `version: 1
packages:
  brew:
    - name: git
      version: 2.43.0
      installed_at: 2025-01-01T12:00:00Z
`

	lockPath := filepath.Join(tmpDir, LockFileName)
	if err := os.WriteFile(lockPath, []byte(v1Content), 0644); err != nil {
		t.Fatal(err)
	}

	service := NewYAMLLockService(tmpDir)

	// Read v1 file
	lockData, err := service.Read()
	if err != nil {
		t.Fatalf("Read v1 failed: %v", err)
	}

	// Write it back (should migrate to v2)
	if err := service.Write(lockData); err != nil {
		t.Fatalf("Write migration failed: %v", err)
	}

	// Read the file content to verify it's v2
	content, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "version: 2") {
		t.Error("File was not migrated to v2")
	}

	// Verify reading the migrated file works
	migratedData, err := service.Read()
	if err != nil {
		t.Fatalf("Read migrated file failed: %v", err)
	}

	if migratedData.Version != 2 {
		t.Errorf("Expected migrated version 2, got %d", migratedData.Version)
	}

	// Should have resources from package conversion
	if len(migratedData.Resources) == 0 {
		t.Error("Expected resources after migration, got none")
	}
}

func TestYAMLLockService_RoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	service := NewYAMLLockService(tmpDir)

	// Create original data
	originalData := &LockData{
		Version: CurrentVersion,
		Packages: map[string][]Package{
			"brew": {{Name: "git", Version: "2.43.0", InstalledAt: "2025-01-01T12:00:00Z"}},
		},
		Resources: []ResourceEntry{
			{
				Type:        "dotfile",
				ID:          ".vimrc",
				State:       "managed",
				InstalledAt: "2025-01-01T12:00:00Z",
				Metadata:    map[string]interface{}{"source": "~/.dotfiles/.vimrc"},
			},
		},
	}

	// Write data
	if err := service.Write(originalData); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read it back
	readData, err := service.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Verify round-trip consistency
	if readData.Version != originalData.Version {
		t.Errorf("Version mismatch: expected %d, got %d", originalData.Version, readData.Version)
	}

	if len(readData.Packages["brew"]) != len(originalData.Packages["brew"]) {
		t.Error("Package count mismatch after round-trip")
	}

	if len(readData.Resources) != len(originalData.Resources) {
		t.Error("Resource count mismatch after round-trip")
	}
}
