// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoader_LoadOrDefault(t *testing.T) {
	t.Run("missing config returns empty config", func(t *testing.T) {
		tmpDir := t.TempDir()
		loader := NewConfigLoader(tmpDir)

		cfg := loader.LoadOrDefault()
		if cfg == nil {
			t.Error("LoadOrDefault should never return nil")
		}

		// Should have nil fields (empty config)
		if cfg.DefaultManager != nil {
			t.Error("Expected empty config with nil DefaultManager")
		}
	})

	t.Run("valid config is loaded", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "plonk.yaml")

		// Write valid config
		validConfig := `default_manager: npm
operation_timeout: 60
ignore_patterns:
  - "*.tmp"
  - ".cache/"
`
		if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
			t.Fatal(err)
		}

		loader := NewConfigLoader(tmpDir)
		cfg := loader.LoadOrDefault()

		if cfg.DefaultManager == nil || *cfg.DefaultManager != "npm" {
			t.Error("Expected config to be loaded with npm as default manager")
		}
		if cfg.OperationTimeout == nil || *cfg.OperationTimeout != 60 {
			t.Error("Expected operation timeout to be 60")
		}
		if len(cfg.IgnorePatterns) != 2 {
			t.Error("Expected 2 ignore patterns")
		}
	})

	t.Run("invalid config returns empty config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "plonk.yaml")

		// Write invalid YAML
		invalidConfig := `default_manager: [this is invalid yaml`
		if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
			t.Fatal(err)
		}

		loader := NewConfigLoader(tmpDir)
		cfg := loader.LoadOrDefault()

		// Should return empty config on error
		if cfg == nil {
			t.Error("LoadOrDefault should never return nil")
		}
		if cfg.DefaultManager != nil {
			t.Error("Expected empty config on parse error")
		}
	})
}

func TestConfigLoader_Exists(t *testing.T) {
	t.Run("returns false for missing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		loader := NewConfigLoader(tmpDir)

		if loader.Exists() {
			t.Error("Expected Exists to return false for missing config")
		}
	})

	t.Run("returns true for existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "plonk.yaml")

		if err := os.WriteFile(configPath, []byte("# empty config"), 0644); err != nil {
			t.Fatal(err)
		}

		loader := NewConfigLoader(tmpDir)
		if !loader.Exists() {
			t.Error("Expected Exists to return true for existing config")
		}
	})
}

func TestConfigLoader_Validate(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "plonk.yaml")

		validConfig := `default_manager: homebrew
operation_timeout: 120
`
		if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
			t.Fatal(err)
		}

		loader := NewConfigLoader(tmpDir)
		result, err := loader.Validate()
		if err != nil {
			t.Fatalf("Validate returned error: %v", err)
		}

		if !result.IsValid() {
			t.Errorf("Expected valid config, got errors: %v", result.Errors)
		}
	})

	t.Run("invalid config fails validation", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "plonk.yaml")

		invalidConfig := `default_manager: invalid-manager
operation_timeout: -5
`
		if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
			t.Fatal(err)
		}

		loader := NewConfigLoader(tmpDir)
		result, err := loader.Validate()
		if err != nil {
			t.Fatalf("Validate returned error: %v", err)
		}

		if result.IsValid() {
			t.Error("Expected validation to fail for invalid config")
		}
		if len(result.Errors) == 0 {
			t.Error("Expected validation errors")
		}
	})
}

func TestConfigManager_Save(t *testing.T) {
	t.Run("saves config successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		manager := NewConfigManager(tmpDir)

		// Create config to save
		defaultManager := "npm"
		timeout := 90
		cfg := &Config{
			DefaultManager:   &defaultManager,
			OperationTimeout: &timeout,
			IgnorePatterns:   []string{".git/", "*.bak"},
		}

		// Save config
		if err := manager.Save(cfg); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify it was saved
		loader := NewConfigLoader(tmpDir)
		if !loader.Exists() {
			t.Error("Config file should exist after save")
		}

		// Load and verify content
		loaded, err := manager.Load()
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loaded.DefaultManager == nil || *loaded.DefaultManager != "npm" {
			t.Error("Saved config should have npm as default manager")
		}
		if loaded.OperationTimeout == nil || *loaded.OperationTimeout != 90 {
			t.Error("Saved config should have operation timeout of 90")
		}
		if len(loaded.IgnorePatterns) != 2 {
			t.Error("Saved config should have 2 ignore patterns")
		}
	})

	t.Run("creates directory if needed", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "nested", "config", "dir")
		manager := NewConfigManager(nestedDir)

		cfg := &Config{}
		if err := manager.Save(cfg); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify directory was created
		if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
			t.Error("Config directory should have been created")
		}
	})
}

func TestLoadConfigWithDefaults(t *testing.T) {
	t.Run("convenience function works", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Should return empty config for missing file
		cfg := LoadConfigWithDefaults(tmpDir)
		if cfg == nil {
			t.Error("LoadConfigWithDefaults should never return nil")
		}

		// Write a config
		configPath := filepath.Join(tmpDir, "plonk.yaml")
		validConfig := `default_manager: cargo`
		if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
			t.Fatal(err)
		}

		// Should load the config
		cfg = LoadConfigWithDefaults(tmpDir)
		if cfg.DefaultManager == nil || *cfg.DefaultManager != "cargo" {
			t.Error("Expected config to be loaded")
		}
	})
}
