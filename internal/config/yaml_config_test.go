// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_BasicStructure(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	configContent := `default_manager: homebrew
operation_timeout: 600

ignore_patterns:
  - .DS_Store
  - "*.tmp"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load configuration
	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify settings
	if config.DefaultManager != nil && *config.DefaultManager != "homebrew" {
		t.Errorf("Expected default_manager 'homebrew', got '%s'", *config.DefaultManager)
	}

	// Verify ignore patterns
	if len(config.IgnorePatterns) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(config.IgnorePatterns))
	}

	// Check ignore patterns
	if config.IgnorePatterns[0] != ".DS_Store" {
		t.Errorf("Expected first ignore pattern '.DS_Store', got '%s'", config.IgnorePatterns[0])
	}

	if config.IgnorePatterns[1] != "*.tmp" {
		t.Errorf("Expected second ignore pattern '*.tmp', got '%s'", config.IgnorePatterns[1])
	}

	// Packages are now in lock file, not config
	// Verify no package fields exist
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()

	config, err := LoadConfig(tempDir)
	if err != nil {
		t.Errorf("Expected no error for non-existent config file (zero-config), got: %v", err)
	}

	if config == nil {
		t.Error("Expected default config when file doesn't exist")
	}

	// Verify it resolves to defaults
	resolved := config.Resolve()
	if resolved.GetDefaultManager() != "homebrew" {
		t.Errorf("Expected default manager 'homebrew', got %s", resolved.GetDefaultManager())
	}
}

func TestConfigValidation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	// Test invalid default manager should fail
	configContent := `default_manager: invalid_manager
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err = LoadConfig(tempDir)
	if err == nil {
		t.Error("Expected error for invalid default manager")
	}
}

func TestSourceToTarget(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"zshrc", "~/.zshrc"},
		{"zshenv", "~/.zshenv"},
		{"config/nvim/", "~/.config/nvim/"},
		{"config/mcfly/config.yaml", "~/.config/mcfly/config.yaml"},
		{"gitconfig", "~/.gitconfig"},
		{"editorconfig", "~/.editorconfig"},
	}

	for _, test := range tests {
		result := sourceToTarget(test.source)
		if result != test.expected {
			t.Errorf("sourceToTarget(%s) = %s, expected %s", test.source, result, test.expected)
		}
	}
}
