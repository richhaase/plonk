// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"plonk/pkg/managers"
)

func TestStatusCommandExists(t *testing.T) {
	// Test that the status command is properly configured
	if statusCmd == nil {
		t.Error("Status command should not be nil")
	}

	if statusCmd.Use != "status" {
		t.Errorf("Expected command use to be 'status', got '%s'", statusCmd.Use)
	}

	if statusCmd.Short == "" {
		t.Error("Status command should have a short description")
	}

	if statusCmd.RunE == nil {
		t.Error("Status command should have a RunE function")
	}
}

func TestPackageManagerInterface(t *testing.T) {
	// Test that the PackageManager interface has the expected methods
	// This is a compile-time check - if the interface changes, this won't compile
	var _ managers.PackageManager = &testPackageManager{}
}

func TestStatusWithDrift(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}

	// Create config with dotfiles that will have drift
	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
`

	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create source file but not target (drift)
	err = os.WriteFile(filepath.Join(plonkDir, "zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test status with drift flag
	err = runStatusWithDrift()
	if err != nil {
		t.Fatalf("Status with drift failed: %v", err)
	}
}

func TestStatusNoDrift(t *testing.T) {
	// Setup temporary directory
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}

	// Create config with dotfiles
	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
`

	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create both source and matching target files (no drift)
	err = os.WriteFile(filepath.Join(plonkDir, "zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempHome, ".zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Test status with drift detection (should show no drift)
	err = runStatusWithDrift()
	if err != nil {
		t.Fatalf("Status with drift failed: %v", err)
	}
}

// testPackageManager implements PackageManager for testing
type testPackageManager struct{}

func (t *testPackageManager) IsAvailable() bool {
	return true
}

func (t *testPackageManager) ListInstalled() ([]string, error) {
	return []string{"test-package"}, nil
}
