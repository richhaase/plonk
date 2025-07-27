// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

// TestZeroConfigCommands tests that commands work without configuration files
func TestZeroConfigCommands(t *testing.T) {
	// Create isolated test environment
	tmpDir := t.TempDir()

	// Set environment to use test directory
	originalHome := os.Getenv("HOME")
	originalPlonkDir := os.Getenv("PLONK_DIR")

	os.Setenv("PLONK_DIR", tmpDir)
	defer func() {
		os.Setenv("HOME", originalHome)
		if originalPlonkDir != "" {
			os.Setenv("PLONK_DIR", originalPlonkDir)
		} else {
			os.Unsetenv("PLONK_DIR")
		}
	}()

	t.Run("setup command creates config from zero state", func(t *testing.T) {
		// Create a fresh temp directory for this test
		testDir := t.TempDir()

		// Set PLONK_DIR to the fresh directory
		originalPlonkDir := os.Getenv("PLONK_DIR")
		os.Setenv("PLONK_DIR", testDir)
		defer func() {
			if originalPlonkDir != "" {
				os.Setenv("PLONK_DIR", originalPlonkDir)
			} else {
				os.Unsetenv("PLONK_DIR")
			}
		}()

		// Ensure no config exists
		configPath := filepath.Join(testDir, "plonk.yaml")
		os.Remove(configPath)

		// Reset setup command state
		setupYes = true // Non-interactive mode for tests

		// Run setup command
		err := runSetup(setupCmd, []string{})
		if err != nil {
			t.Errorf("Expected setup to succeed with no existing config, got: %v", err)
		}

		// Verify config file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Expected config file to be created")
		}

		// Verify config content
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "default_manager: homebrew") {
			t.Error("Expected config to contain default manager setting")
		}

		if !strings.Contains(contentStr, "# Plonk Configuration File") {
			t.Error("Expected config to contain header comment")
		}
	})

	t.Run("setup command exits early when config exists", func(t *testing.T) {
		// Create a directory with existing config
		testDir := t.TempDir()
		configPath := filepath.Join(testDir, "plonk.yaml")

		// Create existing config
		err := os.WriteFile(configPath, []byte("default_manager: npm\n"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Set PLONK_DIR to the directory with existing config
		originalPlonkDir := os.Getenv("PLONK_DIR")
		os.Setenv("PLONK_DIR", testDir)
		defer func() {
			if originalPlonkDir != "" {
				os.Setenv("PLONK_DIR", originalPlonkDir)
			} else {
				os.Unsetenv("PLONK_DIR")
			}
		}()

		// Reset setup command state
		setupYes = true

		// Run setup command again - should exit gracefully
		err = runSetup(setupCmd, []string{})
		if err != nil {
			t.Errorf("Expected setup to exit gracefully when config already exists, got: %v", err)
		}

		// Note: setup exits early with a message, no error
	})
}

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
