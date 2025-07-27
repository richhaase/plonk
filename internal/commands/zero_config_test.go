// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

// TestConfigResolutionInCommands tests that commands properly use resolved configuration
func TestConfigResolutionInCommands(t *testing.T) {
	t.Run("status command works with zero config", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Ensure no config file exists
		os.Remove(filepath.Join(tmpDir, "plonk.yaml"))

		// Load config (should get zero config)
		cfg, err := config.Load(tmpDir)
		if err != nil {
			t.Fatalf("Expected LoadConfig to work with missing file, got: %v", err)
		}

		// Verify config works
		if cfg.DefaultManager != "homebrew" {
			t.Errorf("Expected default manager 'homebrew', got '%s'", cfg.DefaultManager)
		}

		if cfg.OperationTimeout != 300 {
			t.Errorf("Expected operation timeout 300, got %d", cfg.OperationTimeout)
		}
	})

	t.Run("config with partial overrides resolves correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "plonk.yaml")

		// Create config with only some settings
		configContent := `default_manager: cargo
operation_timeout: 900
# Note: package_timeout and dotfile_timeout not specified - should use defaults
`

		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Load and resolve config
		cfg, err := config.Load(tmpDir)
		if err != nil {
			t.Fatalf("Expected LoadConfig to work, got: %v", err)
		}

		// Check overridden values
		if cfg.DefaultManager != "cargo" {
			t.Errorf("Expected overridden default manager 'cargo', got '%s'", cfg.DefaultManager)
		}

		if cfg.OperationTimeout != 900 {
			t.Errorf("Expected overridden operation timeout 900, got %d", cfg.OperationTimeout)
		}

		// Check default values for unspecified settings
		if cfg.PackageTimeout != 180 {
			t.Errorf("Expected default package timeout 180, got %d", cfg.PackageTimeout)
		}

		if cfg.DotfileTimeout != 60 {
			t.Errorf("Expected default dotfile timeout 60, got %d", cfg.DotfileTimeout)
		}
	})
}
