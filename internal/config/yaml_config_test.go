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

	configContent := `settings:
  default_manager: homebrew

homebrew:
  - aichat
  - aider
  - name: neovim
  - font-hack-nerd-font

npm:
  - "@anthropic-ai/claude-code"
  - name: some-tool
    package: "@scope/different-name"
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
	if config.Settings.DefaultManager != "homebrew" {
		t.Errorf("Expected default_manager 'homebrew', got '%s'", config.Settings.DefaultManager)
	}

	// Verify homebrew packages
	if len(config.Homebrew) != 4 {
		t.Errorf("Expected 4 homebrew packages, got %d", len(config.Homebrew))
	}

	// Check simple package
	if config.Homebrew[0].Name != "aichat" {
		t.Errorf("Expected first package 'aichat', got '%s'", config.Homebrew[0].Name)
	}

	// Check package with config
	neovim := config.Homebrew[2]
	if neovim.Name != "neovim" {
		t.Errorf("Expected neovim name 'neovim', got '%s'", neovim.Name)
	}

	// ASDF functionality has been removed

	// Verify npm packages
	if len(config.NPM) != 2 {
		t.Errorf("Expected 2 npm packages, got %d", len(config.NPM))
	}

	claudeCode := config.NPM[0]
	if claudeCode.Name != "@anthropic-ai/claude-code" {
		t.Errorf("Expected claude-code name '@anthropic-ai/claude-code', got '%s'", claudeCode.Name)
	}

	someTool := config.NPM[1]
	if someTool.Name != "some-tool" || someTool.Package != "@scope/different-name" {
		t.Errorf("some-tool not parsed correctly: %+v", someTool)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()

	_, err := LoadConfig(tempDir)
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}

func TestConfigValidation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plonk.yaml")

	// Test invalid default manager should fail
	configContent := `settings:
  default_manager: invalid_manager

dotfiles:
  - source: test
    destination: ~/.test
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

