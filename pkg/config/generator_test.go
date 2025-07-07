// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		name     string
		results  DiscoveryResults
		expected Config
	}{
		{
			name: "complete discovery results",
			results: DiscoveryResults{
				HomebrewPackages: []string{"git", "jq", "node"},
				AsdfTools:        []string{"nodejs 20.0.0", "python 3.11.3", "ruby 3.0.0"},
				NpmPackages:      []string{"claude-code", "@angular/cli", "typescript"},
				Dotfiles:         []string{".zshrc", ".gitconfig", ".zshenv"},
			},
			expected: Config{
				Settings: Settings{DefaultManager: "homebrew"},
				Dotfiles: []string{".zshrc", ".gitconfig", ".zshenv"},
				Homebrew: HomebrewConfig{
					Brews: []HomebrewPackage{
						{Name: "git"},
						{Name: "jq"},
						{Name: "node"},
					},
				},
				ASDF: []ASDFTool{
					{Name: "nodejs", Version: "20.0.0"},
					{Name: "python", Version: "3.11.3"},
					{Name: "ruby", Version: "3.0.0"},
				},
				NPM: []NPMPackage{
					{Name: "claude-code"},
					{Name: "@angular/cli"},
					{Name: "typescript"},
				},
			},
		},
		{
			name: "partial discovery results",
			results: DiscoveryResults{
				HomebrewPackages: []string{"git"},
				AsdfTools:        []string{},
				NpmPackages:      []string{"typescript"},
				Dotfiles:         []string{".zshrc"},
			},
			expected: Config{
				Settings: Settings{DefaultManager: "homebrew"},
				Dotfiles: []string{".zshrc"},
				Homebrew: HomebrewConfig{
					Brews: []HomebrewPackage{{Name: "git"}},
				},
				ASDF: []ASDFTool{},
				NPM:  []NPMPackage{{Name: "typescript"}},
			},
		},
		{
			name: "empty discovery results",
			results: DiscoveryResults{
				HomebrewPackages: []string{},
				AsdfTools:        []string{},
				NpmPackages:      []string{},
				Dotfiles:         []string{},
			},
			expected: Config{
				Settings: Settings{DefaultManager: "homebrew"},
				Dotfiles: []string{},
				Homebrew: HomebrewConfig{Brews: []HomebrewPackage{}},
				ASDF:     []ASDFTool{},
				NPM:      []NPMPackage{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GenerateConfig(tt.results)

			// Check settings
			if config.Settings.DefaultManager != tt.expected.Settings.DefaultManager {
				t.Errorf("Expected default manager %s, got %s",
					tt.expected.Settings.DefaultManager, config.Settings.DefaultManager)
			}

			// Check dotfiles
			if len(config.Dotfiles) != len(tt.expected.Dotfiles) {
				t.Errorf("Expected %d dotfiles, got %d", len(tt.expected.Dotfiles), len(config.Dotfiles))
			}
			for i, dotfile := range config.Dotfiles {
				if i < len(tt.expected.Dotfiles) && dotfile != tt.expected.Dotfiles[i] {
					t.Errorf("Expected dotfile %s, got %s", tt.expected.Dotfiles[i], dotfile)
				}
			}

			// Check homebrew packages
			if len(config.Homebrew.Brews) != len(tt.expected.Homebrew.Brews) {
				t.Errorf("Expected %d homebrew packages, got %d",
					len(tt.expected.Homebrew.Brews), len(config.Homebrew.Brews))
			}
			for i, pkg := range config.Homebrew.Brews {
				if i < len(tt.expected.Homebrew.Brews) && pkg.Name != tt.expected.Homebrew.Brews[i].Name {
					t.Errorf("Expected homebrew package %s, got %s",
						tt.expected.Homebrew.Brews[i].Name, pkg.Name)
				}
			}

			// Check ASDF tools
			if len(config.ASDF) != len(tt.expected.ASDF) {
				t.Errorf("Expected %d ASDF tools, got %d", len(tt.expected.ASDF), len(config.ASDF))
			}
			for i, tool := range config.ASDF {
				if i < len(tt.expected.ASDF) {
					expected := tt.expected.ASDF[i]
					if tool.Name != expected.Name || tool.Version != expected.Version {
						t.Errorf("Expected ASDF tool %s %s, got %s %s",
							expected.Name, expected.Version, tool.Name, tool.Version)
					}
				}
			}

			// Check NPM packages
			if len(config.NPM) != len(tt.expected.NPM) {
				t.Errorf("Expected %d NPM packages, got %d", len(tt.expected.NPM), len(config.NPM))
			}
			for i, pkg := range config.NPM {
				if i < len(tt.expected.NPM) && pkg.Name != tt.expected.NPM[i].Name {
					t.Errorf("Expected NPM package %s, got %s", tt.expected.NPM[i].Name, pkg.Name)
				}
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		expectedYAML string
		expectError  bool
	}{
		{
			name: "save complete config",
			config: Config{
				Settings: Settings{DefaultManager: "homebrew"},
				Dotfiles: []string{".zshrc", ".gitconfig", ".zshenv"},
				Homebrew: HomebrewConfig{
					Brews: []HomebrewPackage{
						{Name: "git"},
						{Name: "jq"},
					},
				},
				ASDF: []ASDFTool{
					{Name: "nodejs", Version: "20.0.0"},
					{Name: "python", Version: "3.11.3"},
				},
				NPM: []NPMPackage{
					{Name: "claude-code"},
					{Name: "@angular/cli"},
				},
			},
			expectedYAML: `settings:
  default_manager: homebrew
dotfiles:
  - .zshrc
  - .gitconfig
  - .zshenv
homebrew:
  brews:
    - git
    - jq
asdf:
  - name: nodejs
    version: 20.0.0
  - name: python
    version: 3.11.3
npm:
  - claude-code
  - '@angular/cli'
`,
			expectError: false,
		},
		{
			name: "save minimal config",
			config: Config{
				Settings: Settings{DefaultManager: "homebrew"},
				Dotfiles: []string{".zshrc"},
			},
			expectedYAML: `settings:
  default_manager: homebrew
dotfiles:
  - .zshrc
`,
			expectError: false,
		},
		{
			name: "save empty config",
			config: Config{
				Settings: Settings{DefaultManager: "homebrew"},
			},
			expectedYAML: `settings:
  default_manager: homebrew
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "plonk.yaml")

			// Save the config
			err := SaveConfig(tt.config, configPath)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Read the saved file
				savedContent, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read saved config: %v", err)
				}

				// Compare with expected YAML
				if string(savedContent) != tt.expectedYAML {
					t.Errorf("YAML mismatch.\nExpected:\n%s\nGot:\n%s",
						tt.expectedYAML, string(savedContent))
				}
			}
		})
	}
}
