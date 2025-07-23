// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestPhase1_NewConfigCompatibility verifies the new config can work alongside the old
func TestPhase1_NewConfigCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	// Write a test config that works with both systems
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

	// Load with OLD system
	oldCfg, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Old LoadConfig failed: %v", err)
	}

	// Load with NEW system
	newCfg, err := LoadNew(tempDir)
	if err != nil {
		t.Fatalf("New LoadNew failed: %v", err)
	}

	// Compare values
	t.Run("DefaultManager", func(t *testing.T) {
		oldVal := oldCfg.getDefaultManager("homebrew")
		newVal := newCfg.GetDefaultManager()
		if oldVal != newVal {
			t.Errorf("DefaultManager mismatch: old=%s, new=%s", oldVal, newVal)
		}
	})

	t.Run("Timeouts", func(t *testing.T) {
		// Compare resolved values
		oldResolved := oldCfg.Resolve()

		if oldResolved.OperationTimeout != newCfg.OperationTimeout {
			t.Errorf("OperationTimeout mismatch: old=%d, new=%d",
				oldResolved.OperationTimeout, newCfg.OperationTimeout)
		}
		if oldResolved.PackageTimeout != newCfg.PackageTimeout {
			t.Errorf("PackageTimeout mismatch: old=%d, new=%d",
				oldResolved.PackageTimeout, newCfg.PackageTimeout)
		}
		if oldResolved.DotfileTimeout != newCfg.DotfileTimeout {
			t.Errorf("DotfileTimeout mismatch: old=%d, new=%d",
				oldResolved.DotfileTimeout, newCfg.DotfileTimeout)
		}
	})

	t.Run("Arrays", func(t *testing.T) {
		oldResolved := oldCfg.Resolve()

		if !reflect.DeepEqual(oldResolved.ExpandDirectories, newCfg.ExpandDirectories) {
			t.Errorf("ExpandDirectories mismatch: old=%v, new=%v",
				oldResolved.ExpandDirectories, newCfg.ExpandDirectories)
		}
		if !reflect.DeepEqual(oldResolved.IgnorePatterns, newCfg.IgnorePatterns) {
			t.Errorf("IgnorePatterns mismatch: old=%v, new=%v",
				oldResolved.IgnorePatterns, newCfg.IgnorePatterns)
		}
	})
}

// TestPhase1_ZeroConfigBehavior verifies both systems handle missing files the same way
func TestPhase1_ZeroConfigBehavior(t *testing.T) {
	tempDir := t.TempDir()

	// OLD system behavior
	oldCfg := LoadConfigWithDefaults(tempDir)
	oldResolved := oldCfg.Resolve()

	// NEW system behavior
	newCfg := LoadNewWithDefaults(tempDir)

	// Both should return defaults
	if oldResolved.DefaultManager != newCfg.DefaultManager {
		t.Errorf("Zero-config DefaultManager mismatch: old=%s, new=%s",
			oldResolved.DefaultManager, newCfg.DefaultManager)
	}

	// Check all defaults match
	if oldResolved.OperationTimeout != newCfg.OperationTimeout {
		t.Errorf("Zero-config OperationTimeout mismatch: old=%d, new=%d",
			oldResolved.OperationTimeout, newCfg.OperationTimeout)
	}
}

// TestPhase1_GetterCompatibility verifies getter methods work the same
func TestPhase1_GetterCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	configContent := `
default_manager: cargo
operation_timeout: 400
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load both
	oldCfg, _ := LoadConfig(tempDir)
	oldResolved := oldCfg.Resolve()
	newCfg, _ := LoadNew(tempDir)

	// Test getters
	if oldResolved.GetDefaultManager() != newCfg.GetDefaultManager() {
		t.Error("GetDefaultManager mismatch")
	}
	if oldResolved.GetOperationTimeout() != newCfg.GetOperationTimeout() {
		t.Error("GetOperationTimeout mismatch")
	}
	if oldResolved.GetPackageTimeout() != newCfg.GetPackageTimeout() {
		t.Error("GetPackageTimeout mismatch")
	}
	if oldResolved.GetDotfileTimeout() != newCfg.GetDotfileTimeout() {
		t.Error("GetDotfileTimeout mismatch")
	}
}

// TestPhase1_ValidationCompatibility ensures both validate the same way
func TestPhase1_ValidationCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		shouldErr bool
	}{
		{
			name:      "invalid manager",
			content:   `default_manager: invalid_manager`,
			shouldErr: true,
		},
		{
			name:      "negative timeout",
			content:   `operation_timeout: -1`,
			shouldErr: true,
		},
		{
			name:      "timeout too large",
			content:   `package_timeout: 1801`,
			shouldErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "plonk.yaml")

			if err := os.WriteFile(configPath, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}

			// Test OLD system
			_, oldErr := LoadConfig(tempDir)

			// Test NEW system
			_, newErr := LoadNew(tempDir)

			// Both should error or both should succeed
			if (oldErr != nil) != (newErr != nil) {
				t.Errorf("Validation mismatch: old error=%v, new error=%v", oldErr, newErr)
			}
		})
	}
}
