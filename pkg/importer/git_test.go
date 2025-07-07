// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"os"
	"path/filepath"
	"testing"

	"plonk/pkg/config"
)

func TestGitDiscoverer(t *testing.T) {
	tests := []struct {
		name             string
		gitconfigContent string
		expectedConfig   config.GitConfig
	}{
		{
			name: "parse basic git configuration",
			gitconfigContent: `[user]
	name = Rich Haase
	email = haaserdh@gmail.com
[core]
	pager = delta
	excludesfile = ~/.gitignore_global
[alias]
	st = status
	co = checkout
[delta]
	navigate = true
	side-by-side = true`,
			expectedConfig: config.GitConfig{
				User: map[string]string{
					"name":  "Rich Haase",
					"email": "haaserdh@gmail.com",
				},
				Core: map[string]string{
					"pager":        "delta",
					"excludesfile": "~/.gitignore_global",
				},
				Aliases: map[string]string{
					"st": "status",
					"co": "checkout",
				},
				Delta: map[string]string{
					"navigate":     "true",
					"side-by-side": "true",
				},
			},
		},
		{
			name:           "empty configuration",
			expectedConfig: config.GitConfig{},
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
			if tt.gitconfigContent != "" {
				gitconfigPath := filepath.Join(tempHome, ".gitconfig")
				err := os.WriteFile(gitconfigPath, []byte(tt.gitconfigContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create .gitconfig: %v", err)
				}
			}

			discoverer := NewGitDiscoverer()
			gitConfig, err := discoverer.DiscoverGitConfig()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify user settings
			if len(gitConfig.User) != len(tt.expectedConfig.User) {
				t.Errorf("Expected %d user settings, got %d", len(tt.expectedConfig.User), len(gitConfig.User))
			}
			for key, expected := range tt.expectedConfig.User {
				if actual, ok := gitConfig.User[key]; !ok || actual != expected {
					t.Errorf("Expected user %s=%s, got %s", key, expected, actual)
				}
			}

			// Verify core settings
			if len(gitConfig.Core) != len(tt.expectedConfig.Core) {
				t.Errorf("Expected %d core settings, got %d", len(tt.expectedConfig.Core), len(gitConfig.Core))
			}
			for key, expected := range tt.expectedConfig.Core {
				if actual, ok := gitConfig.Core[key]; !ok || actual != expected {
					t.Errorf("Expected core %s=%s, got %s", key, expected, actual)
				}
			}

			// Verify aliases
			if len(gitConfig.Aliases) != len(tt.expectedConfig.Aliases) {
				t.Errorf("Expected %d aliases, got %d", len(tt.expectedConfig.Aliases), len(gitConfig.Aliases))
			}
			for key, expected := range tt.expectedConfig.Aliases {
				if actual, ok := gitConfig.Aliases[key]; !ok || actual != expected {
					t.Errorf("Expected alias %s=%s, got %s", key, expected, actual)
				}
			}
		})
	}
}
