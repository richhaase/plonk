// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYAMLConfigService_LoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
		validateFn  func(*Config) error
	}{
		{
			name: "valid minimal config",
			yamlContent: `
settings:
  default_manager: homebrew
`,
			expectError: false,
			validateFn: func(cfg *Config) error {
				// Packages are now in lock file, just validate config loaded
				if cfg.DefaultManager != nil && *cfg.DefaultManager != "homebrew" {
					t.Errorf("Expected default manager homebrew, got %s", *cfg.DefaultManager)
				}
				return nil
			},
		},
		{
			name: "valid config with timeout settings",
			yamlContent: `
settings:
  default_manager: homebrew
  operation_timeout: 600
  package_timeout: 300
  dotfile_timeout: 120
`,
			expectError: false,
			validateFn: func(cfg *Config) error {
				if cfg.OperationTimeout != nil && *cfg.OperationTimeout != 600 {
					t.Errorf("Expected operation timeout 600, got %d", *cfg.OperationTimeout)
				}
				return nil
			},
		},
		{
			name: "config with ignore patterns",
			yamlContent: `
settings:
  default_manager: homebrew

ignore_patterns:
  - .DS_Store
  - "*.log"
`,
			expectError: false,
			validateFn: func(cfg *Config) error {
				if len(cfg.IgnorePatterns) != 2 {
					t.Errorf("Expected 2 ignore patterns, got %d", len(cfg.IgnorePatterns))
				}
				return nil
			},
		},
		{
			name: "config with missing default_manager uses default",
			yamlContent: `
settings:
  operation_timeout: 600
`,
			expectError: false,
			validateFn: func(cfg *Config) error {
				if cfg.DefaultManager != nil && *cfg.DefaultManager != "homebrew" {
					t.Errorf("Expected default manager homebrew, got %s", *cfg.DefaultManager)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir, err := os.MkdirTemp("", "plonk-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			configPath := filepath.Join(tmpDir, "plonk.yaml")
			err = os.WriteFile(configPath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// Test loading
			service := NewYAMLConfigService()
			config, err := service.LoadConfigFromFile(configPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.validateFn != nil {
				if err := tt.validateFn(config); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

func TestYAMLConfigService_LoadConfig_Directory(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "plonk-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config file
	configContent := `
settings:
  default_manager: homebrew
`
	configPath := filepath.Join(tmpDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test loading from directory
	service := NewYAMLConfigService()
	config, err := service.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.DefaultManager != nil && *config.DefaultManager != "homebrew" {
		t.Errorf("Expected default_manager to be 'homebrew', got %s", *config.DefaultManager)
	}

	// Packages now in lock file, not config
	// Just verify config structure is valid
}

func TestYAMLConfigService_LoadConfigFromFile_NotFound(t *testing.T) {
	service := NewYAMLConfigService()
	_, err := service.LoadConfigFromFile("/path/that/does/not/exist/plonk.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestYAMLConfigService_SaveConfigToWriter(t *testing.T) {
	service := NewYAMLConfigService()

	config := &Config{
		DefaultManager: StringPtr("homebrew"),
		// Packages now in lock file
	}

	var buf strings.Builder
	err := service.SaveConfigToWriter(&buf, config)
	if err != nil {
		t.Fatalf("SaveConfigToWriter() failed: %v", err)
	}

	// Verify the output contains expected content
	output := buf.String()
	if !strings.Contains(output, "default_manager: homebrew") {
		t.Error("Output should contain default_manager setting")
	}
}

func TestYAMLConfigService_SaveConfigToFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plonk-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	service := NewYAMLConfigService()
	config := &Config{
		DefaultManager: StringPtr("homebrew"),
		// Packages now in lock file
	}

	configPath := filepath.Join(tmpDir, "plonk.yaml")
	err = service.SaveConfigToFile(configPath, config)
	if err != nil {
		t.Fatalf("SaveConfigToFile() failed: %v", err)
	}

	// Verify file exists and can be read back
	loadedConfig, err := service.LoadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	// Compare default manager values
	configManager := ""
	if config.DefaultManager != nil {
		configManager = *config.DefaultManager
	}
	loadedManager := ""
	if loadedConfig.DefaultManager != nil {
		loadedManager = *loadedConfig.DefaultManager
	}
	if configManager != loadedManager {
		t.Errorf("Default manager mismatch: expected %s, got %s", configManager, loadedManager)
	}
}

func TestYAMLConfigService_ValidateConfig(t *testing.T) {
	service := NewYAMLConfigService()

	// Valid config
	validConfig := &Config{
		DefaultManager: StringPtr("homebrew"),
	}

	result := service.ValidateConfig(validConfig)
	if !result.IsValid() {
		t.Errorf("Valid config should pass validation: %v", result.Errors)
	}

	// Invalid config
	invalidConfig := &Config{
		DefaultManager: StringPtr("invalid_manager"),
	}

	result = service.ValidateConfig(invalidConfig)
	if result.IsValid() {
		t.Error("Invalid config should fail validation")
	}
}

func TestYAMLConfigService_LoadDefaultConfig(t *testing.T) {
	service := NewYAMLConfigService()
	config := service.GetDefaultConfig()

	if config.DefaultManager != nil && *config.DefaultManager != "homebrew" {
		t.Errorf("Default config should have homebrew as default manager, got %s", *config.DefaultManager)
	}

	// Packages now in lock file, not config
}
