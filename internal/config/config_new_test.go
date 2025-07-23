// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadNew_MissingFile(t *testing.T) {
	// Test zero-config behavior - missing file should return defaults
	tempDir := t.TempDir()

	cfg, err := LoadNew(tempDir)
	if err != nil {
		t.Fatalf("LoadNew with missing file should not error, got: %v", err)
	}

	// Check all defaults are applied
	if cfg.DefaultManager != "homebrew" {
		t.Errorf("Expected default manager 'homebrew', got %s", cfg.DefaultManager)
	}
	if cfg.OperationTimeout != 300 {
		t.Errorf("Expected operation timeout 300, got %d", cfg.OperationTimeout)
	}
	if cfg.PackageTimeout != 180 {
		t.Errorf("Expected package timeout 180, got %d", cfg.PackageTimeout)
	}
	if cfg.DotfileTimeout != 60 {
		t.Errorf("Expected dotfile timeout 60, got %d", cfg.DotfileTimeout)
	}
	if len(cfg.ExpandDirectories) != 7 {
		t.Errorf("Expected 7 expand directories, got %d", len(cfg.ExpandDirectories))
	}
	if len(cfg.IgnorePatterns) != 6 {
		t.Errorf("Expected 6 ignore patterns, got %d", len(cfg.IgnorePatterns))
	}
}

func TestLoadNew_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	// Write a valid config
	configContent := `
default_manager: npm
operation_timeout: 600
package_timeout: 300
dotfile_timeout: 120
expand_directories:
  - .vim
  - .emacs.d
ignore_patterns:
  - "*.log"
  - "*.cache"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadNew(tempDir)
	if err != nil {
		t.Fatalf("LoadNew failed: %v", err)
	}

	// Verify loaded values
	if cfg.DefaultManager != "npm" {
		t.Errorf("Expected default manager 'npm', got %s", cfg.DefaultManager)
	}
	if cfg.OperationTimeout != 600 {
		t.Errorf("Expected operation timeout 600, got %d", cfg.OperationTimeout)
	}
	if cfg.PackageTimeout != 300 {
		t.Errorf("Expected package timeout 300, got %d", cfg.PackageTimeout)
	}
	if cfg.DotfileTimeout != 120 {
		t.Errorf("Expected dotfile timeout 120, got %d", cfg.DotfileTimeout)
	}

	expectedDirs := []string{".vim", ".emacs.d"}
	if !reflect.DeepEqual(cfg.ExpandDirectories, expectedDirs) {
		t.Errorf("Expected expand directories %v, got %v", expectedDirs, cfg.ExpandDirectories)
	}

	expectedPatterns := []string{"*.log", "*.cache"}
	if !reflect.DeepEqual(cfg.IgnorePatterns, expectedPatterns) {
		t.Errorf("Expected ignore patterns %v, got %v", expectedPatterns, cfg.IgnorePatterns)
	}
}

func TestLoadNew_PartialConfig(t *testing.T) {
	// Test that unspecified fields get defaults
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	// Write a partial config
	configContent := `
default_manager: cargo
operation_timeout: 400
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadNew(tempDir)
	if err != nil {
		t.Fatalf("LoadNew failed: %v", err)
	}

	// Check specified values
	if cfg.DefaultManager != "cargo" {
		t.Errorf("Expected default manager 'cargo', got %s", cfg.DefaultManager)
	}
	if cfg.OperationTimeout != 400 {
		t.Errorf("Expected operation timeout 400, got %d", cfg.OperationTimeout)
	}

	// Check defaults for unspecified values
	if cfg.PackageTimeout != 180 {
		t.Errorf("Expected default package timeout 180, got %d", cfg.PackageTimeout)
	}
	if cfg.DotfileTimeout != 60 {
		t.Errorf("Expected default dotfile timeout 60, got %d", cfg.DotfileTimeout)
	}
	if len(cfg.ExpandDirectories) != 7 {
		t.Errorf("Expected default expand directories, got %d items", len(cfg.ExpandDirectories))
	}
	if len(cfg.IgnorePatterns) != 6 {
		t.Errorf("Expected default ignore patterns, got %d items", len(cfg.IgnorePatterns))
	}
}

func TestLoadNew_InvalidManager(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	// Write config with invalid manager
	configContent := `
default_manager: invalid_manager
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadNew(tempDir)
	if err == nil {
		t.Error("Expected validation error for invalid manager")
	}
}

func TestLoadNew_InvalidTimeout(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "negative operation timeout",
			content: `
operation_timeout: -1
`,
		},
		{
			name: "operation timeout too large",
			content: `
operation_timeout: 3601
`,
		},
		{
			name: "package timeout too large",
			content: `
package_timeout: 1801
`,
		},
		{
			name: "dotfile timeout too large",
			content: `
dotfile_timeout: 601
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "plonk.yaml")

			if err := os.WriteFile(configPath, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := LoadNew(tempDir)
			if err == nil {
				t.Error("Expected validation error for invalid timeout")
			}
		})
	}
}

func TestLoadNew_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	// Write invalid YAML
	configContent := `
default_manager: [this is not valid yaml
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadNew(tempDir)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoadNewWithDefaults(t *testing.T) {
	// Test that LoadNewWithDefaults always returns a config
	tempDir := t.TempDir()

	// Case 1: Missing file
	cfg := LoadNewWithDefaults(tempDir)
	if cfg == nil {
		t.Fatal("LoadNewWithDefaults should never return nil")
	}
	if cfg.DefaultManager != "homebrew" {
		t.Error("Should return defaults for missing file")
	}

	// Case 2: Invalid file
	configPath := filepath.Join(tempDir, "plonk.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: [yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg = LoadNewWithDefaults(tempDir)
	if cfg == nil {
		t.Fatal("LoadNewWithDefaults should never return nil")
	}
	if cfg.DefaultManager != "homebrew" {
		t.Error("Should return defaults for invalid file")
	}
}

func TestNewConfig_Getters(t *testing.T) {
	cfg := &NewConfig{
		DefaultManager:    "npm",
		OperationTimeout:  500,
		PackageTimeout:    200,
		DotfileTimeout:    100,
		ExpandDirectories: []string{".config", ".vim"},
		IgnorePatterns:    []string{"*.tmp", "*.log"},
	}

	// Test all getter methods
	if cfg.GetDefaultManager() != "npm" {
		t.Error("GetDefaultManager returned wrong value")
	}
	if cfg.GetOperationTimeout() != 500 {
		t.Error("GetOperationTimeout returned wrong value")
	}
	if cfg.GetPackageTimeout() != 200 {
		t.Error("GetPackageTimeout returned wrong value")
	}
	if cfg.GetDotfileTimeout() != 100 {
		t.Error("GetDotfileTimeout returned wrong value")
	}

	dirs := cfg.GetExpandDirectories()
	if len(dirs) != 2 || dirs[0] != ".config" {
		t.Error("GetExpandDirectories returned wrong value")
	}

	patterns := cfg.GetIgnorePatterns()
	if len(patterns) != 2 || patterns[0] != "*.tmp" {
		t.Error("GetIgnorePatterns returned wrong value")
	}
}

func TestNewConfig_Resolve(t *testing.T) {
	cfg := &NewConfig{
		DefaultManager: "pip",
	}

	// Resolve should return self
	resolved := cfg.Resolve()
	if resolved != cfg {
		t.Error("Resolve should return self")
	}
}

func TestLoadNewFromPath_PermissionError(t *testing.T) {
	// Skip on Windows where file permissions work differently
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	// Create file with no read permission
	if err := os.WriteFile(configPath, []byte("test: data"), 0000); err != nil {
		t.Fatal(err)
	}

	_, err := LoadNewFromPath(configPath)
	if err == nil {
		t.Error("Expected error for unreadable file")
	}
}

// TestNewConfig_DefaultsMatch is temporarily commented out for Phase 0 isolation.
// This test will be re-enabled in Phase 1 to verify compatibility with the old system.
/*
func TestNewConfig_DefaultsMatch(t *testing.T) {
	// Ensure our defaults match the current system's defaults
	expected := GetDefaults()
	actual := defaultConfig

	if actual.DefaultManager != expected.DefaultManager {
		t.Errorf("Default manager mismatch: got %s, want %s", actual.DefaultManager, expected.DefaultManager)
	}
	if actual.OperationTimeout != expected.OperationTimeout {
		t.Errorf("Operation timeout mismatch: got %d, want %d", actual.OperationTimeout, expected.OperationTimeout)
	}
	if actual.PackageTimeout != expected.PackageTimeout {
		t.Errorf("Package timeout mismatch: got %d, want %d", actual.PackageTimeout, expected.PackageTimeout)
	}
	if actual.DotfileTimeout != expected.DotfileTimeout {
		t.Errorf("Dotfile timeout mismatch: got %d, want %d", actual.DotfileTimeout, expected.DotfileTimeout)
	}
	if !reflect.DeepEqual(actual.ExpandDirectories, expected.ExpandDirectories) {
		t.Errorf("Expand directories mismatch: got %v, want %v", actual.ExpandDirectories, expected.ExpandDirectories)
	}
	if !reflect.DeepEqual(actual.IgnorePatterns, expected.IgnorePatterns) {
		t.Errorf("Ignore patterns mismatch: got %v, want %v", actual.IgnorePatterns, expected.IgnorePatterns)
	}
}
*/
