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

func TestYAMLLockService_ReadAndWrite(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create service
	service := NewYAMLLockService(tmpDir)

	// Test reading non-existent file returns empty lock
	lock, err := service.Read()
	if err != nil {
		t.Fatalf("Read empty lock failed: %v", err)
	}
	if lock.Version != LockFileVersion {
		t.Errorf("Expected version %d, got %d", LockFileVersion, lock.Version)
	}
	if len(lock.Resources) != 0 {
		t.Errorf("Expected empty resources, got %d", len(lock.Resources))
	}

	// Create test lock file
	testLock := &Lock{
		Version: LockFileVersion,
		Resources: []ResourceEntry{
			{
				Type:        "package",
				ID:          "brew:git",
				InstalledAt: time.Now().Format(time.RFC3339),
				Metadata: map[string]interface{}{
					"manager": "brew",
					"name":    "git",
					"version": "2.43.0",
				},
			},
			{
				Type:        "package",
				ID:          "npm:typescript",
				InstalledAt: time.Now().Format(time.RFC3339),
				Metadata: map[string]interface{}{
					"manager": "npm",
					"name":    "typescript",
					"version": "5.3.3",
				},
			},
			{
				Type:        "package",
				ID:          "go:gopls",
				InstalledAt: time.Now().Format(time.RFC3339),
				Metadata: map[string]interface{}{
					"manager":     "go",
					"name":        "gopls",
					"version":     "v0.14.2",
					"source_path": "golang.org/x/tools/cmd/gopls",
				},
			},
		},
	}

	// Write lock file
	if err := service.Write(testLock); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists
	lockPath := filepath.Join(tmpDir, LockFileName)
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("Lock file not created: %v", err)
	}

	// Read and verify
	loaded, err := service.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if loaded.Version != testLock.Version {
		t.Errorf("Version mismatch: expected %d, got %d", testLock.Version, loaded.Version)
	}

	if len(loaded.Resources) != 3 {
		t.Errorf("Resources mismatch: expected 3, got %d", len(loaded.Resources))
	}

	// Verify Go package has source path
	var foundGopls bool
	for _, resource := range loaded.Resources {
		if resource.ID == "go:gopls" {
			foundGopls = true
			if sourcePath, ok := resource.Metadata["source_path"].(string); !ok || sourcePath != "golang.org/x/tools/cmd/gopls" {
				t.Errorf("Expected source_path 'golang.org/x/tools/cmd/gopls', got %v", resource.Metadata["source_path"])
			}
		}
	}
	if !foundGopls {
		t.Error("gopls package not found in loaded resources")
	}
}

func TestYAMLLockService_AddPackage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	service := NewYAMLLockService(tmpDir)

	// Add first package with metadata
	metadata := map[string]interface{}{
		"manager": "brew",
		"name":    "git",
		"version": "2.43.0",
	}
	if err := service.AddPackage("brew", "git", "2.43.0", metadata); err != nil {
		t.Fatalf("AddPackage failed: %v", err)
	}

	// Verify package was added
	if !service.HasPackage("brew", "git") {
		t.Error("Package not found after adding")
	}

	// Add Go package with source path
	goMetadata := map[string]interface{}{
		"manager":     "go",
		"name":        "gopls",
		"version":     "v0.14.2",
		"source_path": "golang.org/x/tools/cmd/gopls",
	}
	if err := service.AddPackage("go", "gopls", "v0.14.2", goMetadata); err != nil {
		t.Fatalf("AddPackage for Go failed: %v", err)
	}

	// Update existing package
	updatedMetadata := map[string]interface{}{
		"manager": "brew",
		"name":    "git",
		"version": "2.44.0",
	}
	if err := service.AddPackage("brew", "git", "2.44.0", updatedMetadata); err != nil {
		t.Fatalf("Update package failed: %v", err)
	}

	// Read and verify
	lock, err := service.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Should have 2 packages (git was updated, not duplicated)
	if len(lock.Resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(lock.Resources))
	}

	// Verify git was updated
	for _, resource := range lock.Resources {
		if resource.ID == "brew:git" {
			if version, ok := resource.Metadata["version"].(string); !ok || version != "2.44.0" {
				t.Errorf("Expected version 2.44.0, got %v", resource.Metadata["version"])
			}
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
	metadata1 := map[string]interface{}{
		"manager": "brew",
		"name":    "git",
		"version": "2.43.0",
	}
	service.AddPackage("brew", "git", "2.43.0", metadata1)

	metadata2 := map[string]interface{}{
		"manager": "brew",
		"name":    "vim",
		"version": "9.0",
	}
	service.AddPackage("brew", "vim", "9.0", metadata2)

	// Remove git
	if err := service.RemovePackage("brew", "git"); err != nil {
		t.Fatalf("RemovePackage failed: %v", err)
	}

	// Verify git was removed
	if service.HasPackage("brew", "git") {
		t.Error("Package still exists after removal")
	}

	// Verify vim still exists
	if !service.HasPackage("brew", "vim") {
		t.Error("Other package was affected by removal")
	}

	// Try removing non-existent package (should not error)
	if err := service.RemovePackage("brew", "notexist"); err != nil {
		t.Errorf("RemovePackage on non-existent package should not error: %v", err)
	}
}

func TestYAMLLockService_GetPackages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	service := NewYAMLLockService(tmpDir)

	// Add packages to different managers
	brewMetadata := map[string]interface{}{
		"manager": "brew",
		"name":    "git",
		"version": "2.43.0",
	}
	service.AddPackage("brew", "git", "2.43.0", brewMetadata)

	npmMetadata := map[string]interface{}{
		"manager": "npm",
		"name":    "typescript",
		"version": "5.3.3",
	}
	service.AddPackage("npm", "typescript", "5.3.3", npmMetadata)

	// Get brew packages
	brewPackages, err := service.GetPackages("brew")
	if err != nil {
		t.Fatalf("GetPackages failed: %v", err)
	}
	if len(brewPackages) != 1 {
		t.Errorf("Expected 1 brew package, got %d", len(brewPackages))
	}

	// Get npm packages
	npmPackages, err := service.GetPackages("npm")
	if err != nil {
		t.Fatalf("GetPackages failed: %v", err)
	}
	if len(npmPackages) != 1 {
		t.Errorf("Expected 1 npm package, got %d", len(npmPackages))
	}

	// Get packages from non-existent manager
	goPackages, err := service.GetPackages("go")
	if err != nil {
		t.Fatalf("GetPackages failed: %v", err)
	}
	if len(goPackages) != 0 {
		t.Errorf("Expected 0 go packages, got %d", len(goPackages))
	}
}

func TestYAMLLockService_FindPackage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	service := NewYAMLLockService(tmpDir)

	// Add a package with the same name in different managers
	brewMetadata := map[string]interface{}{
		"manager": "brew",
		"name":    "jq",
		"version": "1.7",
	}
	service.AddPackage("brew", "jq", "1.7", brewMetadata)

	npmMetadata := map[string]interface{}{
		"manager": "npm",
		"name":    "jq",
		"version": "2.0.0",
	}
	service.AddPackage("npm", "jq", "2.0.0", npmMetadata)

	// Find package
	locations := service.FindPackage("jq")
	if len(locations) != 2 {
		t.Errorf("Expected 2 locations for 'jq', got %d", len(locations))
	}

	// Verify both managers are found
	foundBrew := false
	foundNpm := false
	for _, loc := range locations {
		if manager, ok := loc.Metadata["manager"].(string); ok {
			if manager == "brew" {
				foundBrew = true
			} else if manager == "npm" {
				foundNpm = true
			}
		}
	}
	if !foundBrew || !foundNpm {
		t.Error("Not all package locations were found")
	}

	// Find non-existent package
	notFound := service.FindPackage("notexist")
	if len(notFound) != 0 {
		t.Errorf("Expected 0 locations for non-existent package, got %d", len(notFound))
	}
}

func TestYAMLLockService_UnsupportedVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a lock file with unsupported version
	lockPath := filepath.Join(tmpDir, LockFileName)
	content := `version: 1
packages:
  brew:
    - name: git
      version: 2.43.0
      installed_at: 2024-01-01T10:00:00Z`

	if err := os.WriteFile(lockPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	service := NewYAMLLockService(tmpDir)
	_, err = service.Read()
	if err == nil {
		t.Error("Expected error for unsupported version")
	}
	if !strings.Contains(err.Error(), "unsupported lock file version") {
		t.Errorf("Expected unsupported version error, got: %v", err)
	}
}

func TestYAMLLockService_NPMScopedPackage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	service := NewYAMLLockService(tmpDir)

	// Add scoped npm package
	metadata := map[string]interface{}{
		"manager":   "npm",
		"name":      "arborist",
		"version":   "6.2.0",
		"scope":     "@npmcli",
		"full_name": "@npmcli/arborist",
	}
	if err := service.AddPackage("npm", "arborist", "6.2.0", metadata); err != nil {
		t.Fatalf("AddPackage failed: %v", err)
	}

	// Read and verify
	lock, err := service.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Find the package
	var found bool
	for _, resource := range lock.Resources {
		if resource.ID == "npm:arborist" {
			found = true
			if scope, ok := resource.Metadata["scope"].(string); !ok || scope != "@npmcli" {
				t.Errorf("Expected scope '@npmcli', got %v", resource.Metadata["scope"])
			}
			if fullName, ok := resource.Metadata["full_name"].(string); !ok || fullName != "@npmcli/arborist" {
				t.Errorf("Expected full_name '@npmcli/arborist', got %v", resource.Metadata["full_name"])
			}
		}
	}
	if !found {
		t.Error("Scoped npm package not found")
	}
}
