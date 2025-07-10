// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plonk/internal/config"
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

	t.Run("init command creates config from zero state", func(t *testing.T) {
		// Ensure no config exists
		configPath := filepath.Join(tmpDir, "plonk.yaml")
		os.Remove(configPath)

		// Reset init command state
		initForce = false

		// Run init command
		err := runInit(initCmd, []string{})
		if err != nil {
			t.Errorf("Expected init to succeed with no existing config, got: %v", err)
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

	t.Run("init command fails when config exists without force", func(t *testing.T) {
		// Config should exist from previous test
		// Reset init command state
		initForce = false

		// Run init command again
		err := runInit(initCmd, []string{})
		if err == nil {
			t.Error("Expected init to fail when config already exists")
		}

		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected 'already exists' error, got: %v", err)
		}
	})

	t.Run("init command with force overwrites existing config", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "plonk.yaml")

		// Write custom content
		customContent := "# Custom config\nsettings:\n  default_manager: npm\n"
		err := os.WriteFile(configPath, []byte(customContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Reset init command state with force
		initForce = true

		// Run init command with force
		err = runInit(initCmd, []string{})
		if err != nil {
			t.Errorf("Expected init --force to succeed, got: %v", err)
		}

		// Verify config was overwritten
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(string(content), "# Custom config") {
			t.Error("Expected original config to be overwritten")
		}

		if !strings.Contains(string(content), "# Plonk Configuration File") {
			t.Error("Expected new config template to be written")
		}
	})
}

// TestConfigResolutionInCommands tests that commands properly use resolved configuration
func TestConfigResolutionInCommands(t *testing.T) {
	t.Run("status command works with zero config", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Ensure no config file exists
		os.Remove(filepath.Join(tmpDir, "plonk.yaml"))

		// Load config (should get zero config)
		cfg, err := config.LoadConfig(tmpDir)
		if err != nil {
			t.Fatalf("Expected LoadConfig to work with missing file, got: %v", err)
		}

		// Verify resolution works
		resolved := cfg.Resolve()
		if resolved.GetDefaultManager() != "homebrew" {
			t.Errorf("Expected default manager 'homebrew', got '%s'", resolved.GetDefaultManager())
		}

		if resolved.GetOperationTimeout() != 300 {
			t.Errorf("Expected operation timeout 300, got %d", resolved.GetOperationTimeout())
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
		cfg, err := config.LoadConfig(tmpDir)
		if err != nil {
			t.Fatalf("Expected LoadConfig to work, got: %v", err)
		}

		resolved := cfg.Resolve()

		// Check overridden values
		if resolved.GetDefaultManager() != "cargo" {
			t.Errorf("Expected overridden default manager 'cargo', got '%s'", resolved.GetDefaultManager())
		}

		if resolved.GetOperationTimeout() != 900 {
			t.Errorf("Expected overridden operation timeout 900, got %d", resolved.GetOperationTimeout())
		}

		// Check default values for unspecified settings
		if resolved.GetPackageTimeout() != 180 {
			t.Errorf("Expected default package timeout 180, got %d", resolved.GetPackageTimeout())
		}

		if resolved.GetDotfileTimeout() != 60 {
			t.Errorf("Expected default dotfile timeout 60, got %d", resolved.GetDotfileTimeout())
		}
	})
}

// TestConfigAdapterZeroConfig tests config adapter behavior with zero config
func TestConfigAdapterZeroConfig(t *testing.T) {
	t.Run("config adapter handles empty config", func(t *testing.T) {
		cfg := &config.Config{} // Empty config
		adapter := config.NewConfigAdapter(cfg)

		// Test dotfile operations
		targets := adapter.GetDotfileTargets()
		if targets == nil {
			t.Error("Expected dotfile targets map to be non-nil")
		}

		// Test package operations (should all return empty - packages now in lock file)
		for _, manager := range []string{"homebrew", "npm", "cargo"} {
			packages, err := adapter.GetPackagesForManager(manager)
			if err != nil {
				t.Errorf("Expected no error for %s packages, got: %v", manager, err)
			}

			if len(packages) != 0 {
				t.Errorf("Expected empty packages for %s (packages now in lock file), got %d", manager, len(packages))
			}
		}
	})

	t.Run("state adapters work with zero config", func(t *testing.T) {
		cfg := &config.Config{} // Empty config
		configAdapter := config.NewConfigAdapter(cfg)

		// Test state package config adapter
		statePackageAdapter := config.NewStatePackageConfigAdapter(configAdapter)
		packages, err := statePackageAdapter.GetPackagesForManager("homebrew")
		if err != nil {
			t.Errorf("Expected no error from state package adapter, got: %v", err)
		}
		if len(packages) != 0 {
			t.Errorf("Expected empty packages from state adapter, got %d", len(packages))
		}

		// Test state dotfile config adapter
		stateDotfileAdapter := config.NewStateDotfileConfigAdapter(configAdapter)

		// Should get default ignore patterns
		patterns := stateDotfileAdapter.GetIgnorePatterns()
		if len(patterns) == 0 {
			t.Error("Expected default ignore patterns from state adapter")
		}

		// Should get default expand directories
		dirs := stateDotfileAdapter.GetExpandDirectories()
		if len(dirs) == 0 {
			t.Error("Expected default expand directories from state adapter")
		}

		// Should get dotfile targets (may be empty but shouldn't error)
		targets := stateDotfileAdapter.GetDotfileTargets()
		if targets == nil {
			t.Error("Expected dotfile targets map to be non-nil")
		}
	})
}

// TestZeroConfigValidation tests validation with zero config scenarios
func TestZeroConfigValidation(t *testing.T) {
	t.Run("empty config validates successfully", func(t *testing.T) {
		cfg := &config.Config{} // Empty config
		validator := config.NewSimpleValidator()

		result := validator.ValidateConfig(cfg)
		if !result.IsValid() {
			t.Errorf("Expected empty config to be valid, got errors: %v", result.Errors)
		}

		if len(result.Warnings) > 0 {
			// Warnings are okay for empty config
			t.Logf("Validation warnings for empty config: %v", result.Warnings)
		}
	})

	t.Run("config with nil fields validates successfully", func(t *testing.T) {
		cfg := &config.Config{
			// All fields nil by default
		}
		validator := config.NewSimpleValidator()

		result := validator.ValidateConfig(cfg)
		if !result.IsValid() {
			t.Errorf("Expected config with nil fields to be valid, got errors: %v", result.Errors)
		}
	})

	t.Run("config with empty fields validates successfully", func(t *testing.T) {
		cfg := &config.Config{} // Empty config (all fields nil)
		validator := config.NewSimpleValidator()

		result := validator.ValidateConfig(cfg)
		if !result.IsValid() {
			t.Errorf("Expected config with empty settings to be valid, got errors: %v", result.Errors)
		}
	})
}
