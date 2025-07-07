// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"os"
	"path/filepath"
	"testing"

	"plonk/pkg/config"
)

func TestZSHDiscoverer(t *testing.T) {
	tests := []struct {
		name           string
		zshrcContent   string
		zshenvContent  string
		expectedConfig config.ZSHConfig
	}{
		{
			name: "parse basic zsh configuration",
			zshrcContent: `export EDITOR=nvim
export PATH="/usr/local/bin:$PATH"
alias ll="ls -la"
eval "$(starship init zsh)"`,
			zshenvContent: `eval "$(/opt/homebrew/bin/brew shellenv)"`,
			expectedConfig: config.ZSHConfig{
				EnvVars: map[string]string{
					"EDITOR": "nvim",
					"PATH":   "/usr/local/bin:$PATH",
				},
				Aliases: map[string]string{
					"ll": "ls -la",
				},
				Inits: []string{
					"starship init zsh",
				},
				SourceBefore: []string{
					`eval "$(/opt/homebrew/bin/brew shellenv)"`,
				},
			},
		},
		{
			name: "empty configuration",
			expectedConfig: config.ZSHConfig{
				EnvVars: map[string]string{},
				Aliases: map[string]string{},
				Inits:   []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary home directory
			tempHome := t.TempDir()
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempHome)
			defer os.Setenv("HOME", originalHome)

			// Create test files
			if tt.zshrcContent != "" {
				zshrcPath := filepath.Join(tempHome, ".zshrc")
				err := os.WriteFile(zshrcPath, []byte(tt.zshrcContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create .zshrc: %v", err)
				}
			}

			if tt.zshenvContent != "" {
				zshenvPath := filepath.Join(tempHome, ".zshenv")
				err := os.WriteFile(zshenvPath, []byte(tt.zshenvContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create .zshenv: %v", err)
				}
			}

			discoverer := NewZSHDiscoverer()
			zshConfig, err := discoverer.DiscoverZSHConfig()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify env vars
			if len(zshConfig.EnvVars) != len(tt.expectedConfig.EnvVars) {
				t.Errorf("Expected %d env vars, got %d", len(tt.expectedConfig.EnvVars), len(zshConfig.EnvVars))
			}
			for key, expected := range tt.expectedConfig.EnvVars {
				if actual, ok := zshConfig.EnvVars[key]; !ok || actual != expected {
					t.Errorf("Expected env var %s=%s, got %s", key, expected, actual)
				}
			}

			// Verify aliases
			if len(zshConfig.Aliases) != len(tt.expectedConfig.Aliases) {
				t.Errorf("Expected %d aliases, got %d", len(tt.expectedConfig.Aliases), len(zshConfig.Aliases))
			}
			for key, expected := range tt.expectedConfig.Aliases {
				if actual, ok := zshConfig.Aliases[key]; !ok || actual != expected {
					t.Errorf("Expected alias %s=%s, got %s", key, expected, actual)
				}
			}

			// Verify inits
			if len(zshConfig.Inits) != len(tt.expectedConfig.Inits) {
				t.Errorf("Expected %d inits, got %d", len(tt.expectedConfig.Inits), len(zshConfig.Inits))
			}
			for i, expected := range tt.expectedConfig.Inits {
				if i >= len(zshConfig.Inits) || zshConfig.Inits[i] != expected {
					t.Errorf("Expected init[%d]=%s, got %s", i, expected, zshConfig.Inits[i])
				}
			}
		})
	}
}
