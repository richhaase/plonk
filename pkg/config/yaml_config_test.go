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

dotfiles:
  - zshrc
  - zshenv
  - plugins.zsh

homebrew:
  brews:
    - aichat
    - aider
    - name: neovim
      config: config/nvim/
  casks:
    - font-hack-nerd-font

asdf:
  - name: nodejs
    version: "24.2.0"
    config: config/npm/
  - name: python
    version: "3.13.2"

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

	// Verify dotfiles
	expectedDotfiles := []string{"zshrc", "zshenv", "plugins.zsh"}
	if len(config.Dotfiles) != len(expectedDotfiles) {
		t.Errorf("Expected %d dotfiles, got %d", len(expectedDotfiles), len(config.Dotfiles))
	}
	for i, expected := range expectedDotfiles {
		if i >= len(config.Dotfiles) || config.Dotfiles[i] != expected {
			t.Errorf("Expected dotfile '%s', got '%s'", expected, config.Dotfiles[i])
		}
	}

	// Verify homebrew packages
	if len(config.Homebrew.Brews) != 3 {
		t.Errorf("Expected 3 homebrew brews, got %d", len(config.Homebrew.Brews))
	}

	// Check simple brew
	if config.Homebrew.Brews[0].Name != "aichat" {
		t.Errorf("Expected first brew 'aichat', got '%s'", config.Homebrew.Brews[0].Name)
	}

	// Check brew with config
	neovim := config.Homebrew.Brews[2]
	if neovim.Name != "neovim" {
		t.Errorf("Expected neovim name 'neovim', got '%s'", neovim.Name)
	}
	if neovim.Config != "config/nvim/" {
		t.Errorf("Expected neovim config 'config/nvim/', got '%s'", neovim.Config)
	}

	// Verify asdf tools
	if len(config.ASDF) != 2 {
		t.Errorf("Expected 2 asdf tools, got %d", len(config.ASDF))
	}

	nodejs := config.ASDF[0]
	if nodejs.Name != "nodejs" || nodejs.Version != "24.2.0" || nodejs.Config != "config/npm/" {
		t.Errorf("nodejs tool not parsed correctly: %+v", nodejs)
	}

	python := config.ASDF[1]
	if python.Name != "python" || python.Version != "3.13.2" || python.Config != "" {
		t.Errorf("python tool not parsed correctly: %+v", python)
	}

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

	// Test ASDF tool without version should fail
	configContent := `settings:
  default_manager: homebrew

asdf:
  - name: nodejs
    # Missing version for ASDF tool
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err = LoadConfig(tempDir)
	if err == nil {
		t.Error("Expected error for ASDF tool without version")
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
		{"dot_gitconfig", "~/.gitconfig"},
		{"editorconfig", "~/.editorconfig"},
	}

	for _, test := range tests {
		result := sourceToTarget(test.source)
		if result != test.expected {
			t.Errorf("sourceToTarget(%s) = %s, expected %s", test.source, result, test.expected)
		}
	}
}

func TestGetDotfileTargets(t *testing.T) {
	config := &Config{
		Dotfiles: []string{"zshrc", "config/nvim/", "dot_gitconfig"},
	}

	targets := config.GetDotfileTargets()

	expected := map[string]string{
		"zshrc":         "~/.zshrc",
		"config/nvim/":  "~/.config/nvim/",
		"dot_gitconfig": "~/.gitconfig",
	}

	for source, expectedTarget := range expected {
		if target, exists := targets[source]; !exists {
			t.Errorf("Missing target for source %s", source)
		} else if target != expectedTarget {
			t.Errorf("Target for %s = %s, expected %s", source, target, expectedTarget)
		}
	}
}
