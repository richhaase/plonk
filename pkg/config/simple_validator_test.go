// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"strings"
	"testing"
)

func TestSimpleValidator_ValidateConfig_ValidConfigs(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "minimal valid config",
			config: &Config{
				Settings: Settings{
					DefaultManager: "homebrew",
				},
			},
		},
		{
			name: "config with valid packages",
			config: &Config{
				Settings: Settings{
					DefaultManager: "homebrew",
				},
				Dotfiles: []string{"zshrc", "vimrc"},
				Homebrew: HomebrewConfig{
					Brews: []HomebrewPackage{
						{Name: "git"},
						{Name: "neovim", Config: "config/nvim/"},
					},
				},
				ASDF: []ASDFTool{
					{Name: "nodejs", Version: "20.0.0"},
				},
				NPM: []NPMPackage{
					{Name: "@vue/cli"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewSimpleValidator()
			result := validator.ValidateConfig(tt.config)

			if !result.IsValid() {
				t.Errorf("Expected valid config, got errors: %v", result.Errors)
			}
		})
	}
}

func TestSimpleValidator_ValidateConfig_InvalidConfigs(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError string
	}{
		{
			name: "missing default manager",
			config: &Config{
				Settings: Settings{}, // DefaultManager is empty
			},
			expectError: "DefaultManager is required",
		},
		{
			name: "invalid default manager",
			config: &Config{
				Settings: Settings{
					DefaultManager: "invalid",
				},
			},
			expectError: "must be one of: homebrew asdf npm",
		},
		{
			name: "invalid package name",
			config: &Config{
				Settings: Settings{DefaultManager: "homebrew"},
				Homebrew: HomebrewConfig{
					Brews: []HomebrewPackage{
						{Name: "invalid package name"},
					},
				},
			},
			expectError: "invalid package name",
		},
		{
			name: "invalid file path",
			config: &Config{
				Settings: Settings{DefaultManager: "homebrew"},
				Dotfiles: []string{"/absolute/path"},
			},
			expectError: "invalid file path",
		},
		{
			name: "missing ASDF version",
			config: &Config{
				Settings: Settings{DefaultManager: "asdf"},
				ASDF: []ASDFTool{
					{Name: "nodejs"}, // Missing version
				},
			},
			expectError: "Version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewSimpleValidator()
			result := validator.ValidateConfig(tt.config)

			if result.IsValid() {
				t.Errorf("Expected validation error, but config was valid")
			}

			found := false
			for _, err := range result.Errors {
				if strings.Contains(strings.ToLower(err), strings.ToLower(tt.expectError)) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error containing %q, got: %v", tt.expectError, result.Errors)
			}
		})
	}
}

func TestSimpleValidator_ValidateConfigFromYAML_ValidYAML(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "minimal valid YAML",
			yaml: `settings:
  default_manager: homebrew`,
		},
		{
			name: "complete valid YAML",
			yaml: `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
  - vimrc

homebrew:
  brews:
    - git
    - name: neovim
      config: config/nvim/

asdf:
  - name: nodejs
    version: "20.0.0"

npm:
  - "@vue/cli"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewSimpleValidator()
			result := validator.ValidateConfigFromYAML([]byte(tt.yaml))

			if !result.IsValid() {
				t.Errorf("Expected valid YAML, got errors: %v", result.Errors)
			}
		})
	}
}

func TestSimpleValidator_ValidateConfigFromYAML_InvalidYAML(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		expectError string
	}{
		{
			name: "invalid YAML syntax",
			yaml: `settings:
  default_manager: homebrew
    invalid_indent: value`,
			expectError: "YAML syntax error",
		},
		{
			name: "invalid default manager",
			yaml: `settings:
  default_manager: invalid`,
			expectError: "must be one of",
		},
		{
			name: "invalid package name",
			yaml: `settings:
  default_manager: homebrew
homebrew:
  brews:
    - "invalid package"`,
			expectError: "invalid package name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewSimpleValidator()
			result := validator.ValidateConfigFromYAML([]byte(tt.yaml))

			if result.IsValid() {
				t.Errorf("Expected validation error, but YAML was valid")
			}

			found := false
			for _, err := range result.Errors {
				if strings.Contains(strings.ToLower(err), strings.ToLower(tt.expectError)) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error containing %q, got: %v", tt.expectError, result.Errors)
			}
		})
	}
}

func TestSimpleValidator_Warnings(t *testing.T) {
	config := &Config{
		Settings: Settings{
			DefaultManager: "npm", // Should trigger warning
		},
	}

	validator := NewSimpleValidator()
	result := validator.ValidateConfig(config)

	if len(result.Warnings) == 0 {
		t.Error("Expected warning about npm default manager")
	}

	if !strings.Contains(result.Warnings[0], "npm as default manager may be slower") {
		t.Errorf("Expected npm warning, got: %s", result.Warnings[0])
	}
}

func TestValidationResult_Summary(t *testing.T) {
	tests := []struct {
		name   string
		result *ValidationResult
		expect string
	}{
		{
			name: "valid result",
			result: &ValidationResult{
				Valid:    true,
				Errors:   []string{},
				Warnings: []string{},
			},
			expect: "Configuration is valid",
		},
		{
			name: "result with errors",
			result: &ValidationResult{
				Valid:  false,
				Errors: []string{"error1", "error2"},
			},
			expect: "2 errors",
		},
		{
			name: "result with errors and warnings",
			result: &ValidationResult{
				Valid:    false,
				Errors:   []string{"error1"},
				Warnings: []string{"warning1"},
			},
			expect: "1 errors, 1 warnings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.result.GetSummary()
			if !strings.Contains(summary, tt.expect) {
				t.Errorf("GetSummary() = %q, expected to contain %q", summary, tt.expect)
			}
		})
	}
}
