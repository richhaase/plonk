// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"testing"

	"github.com/richhaase/plonk/internal/testutil"
)

// TestConfig_CompatibilityWithExisting ensures the new implementation
// can read configs that work with the current system
func TestConfig_CompatibilityWithExisting(t *testing.T) {
	tests := []struct {
		name    string
		content string
		check   func(t *testing.T, cfg *Config)
	}{
		{
			name: "minimal config with version",
			content: `version: 1
default_manager: brew
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.DefaultManager != "brew" {
					t.Errorf("Expected default_manager brew, got %s", cfg.DefaultManager)
				}
			},
		},
		{
			name: "config with all fields",
			content: `version: 1
default_manager: cargo
operation_timeout: 600
package_timeout: 300
dotfile_timeout: 120
expand_directories:
  - .config
  - .vim
ignore_patterns:
  - "*.tmp"
  - "*.log"
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.DefaultManager != "cargo" {
					t.Errorf("Expected default_manager cargo, got %s", cfg.DefaultManager)
				}
				if cfg.OperationTimeout != 600 {
					t.Errorf("Expected operation_timeout 600, got %d", cfg.OperationTimeout)
				}
				if cfg.PackageTimeout != 300 {
					t.Errorf("Expected package_timeout 300, got %d", cfg.PackageTimeout)
				}
				if cfg.DotfileTimeout != 120 {
					t.Errorf("Expected dotfile_timeout 120, got %d", cfg.DotfileTimeout)
				}
				if len(cfg.ExpandDirectories) != 2 {
					t.Errorf("Expected 2 expand_directories, got %d", len(cfg.ExpandDirectories))
				}
				if len(cfg.IgnorePatterns) != 2 {
					t.Errorf("Expected 2 ignore_patterns, got %d", len(cfg.IgnorePatterns))
				}
			},
		},
		{
			name: "config with unknown fields (should ignore)",
			content: `version: 1
default_manager: uv
unknown_field: should be ignored
packages:
  - name: git
    manager: brew
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.DefaultManager != "uv" {
					t.Errorf("Expected default_manager uv, got %s", cfg.DefaultManager)
				}
				// Should successfully parse despite unknown fields
			},
		},
		{
			name:    "empty config",
			content: ``,
			check: func(t *testing.T, cfg *Config) {
				// Should get all defaults
				if cfg.DefaultManager != "brew" {
					t.Errorf("Expected default brew, got %s", cfg.DefaultManager)
				}
			},
		},
		{
			name: "config with empty arrays gets defaults",
			content: `default_manager: pnpm
expand_directories: []
ignore_patterns: []
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.DefaultManager != "pnpm" {
					t.Errorf("Expected default_manager pnpm, got %s", cfg.DefaultManager)
				}
				// Empty arrays in YAML get default values applied
				if len(cfg.ExpandDirectories) == 0 {
					t.Errorf("Expected expand_directories to have defaults, got empty")
				}
				if len(cfg.IgnorePatterns) == 0 {
					t.Errorf("Expected ignore_patterns to have defaults, got empty")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := testutil.NewTestConfig(t, tc.content)

			cfg, err := Load(tempDir)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			tc.check(t, cfg)
		})
	}
}

