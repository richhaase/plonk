// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYAMLConfigService_LoadConfigFromReader(t *testing.T) {
	service := NewYAMLConfigService()
	
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
				if cfg.Settings.DefaultManager != "homebrew" {
					t.Errorf("Expected default_manager to be 'homebrew', got %s", cfg.Settings.DefaultManager)
				}
				return nil
			},
		},
		{
			name: "valid config with packages",
			yamlContent: `
settings:
  default_manager: homebrew
homebrew:
  - name: git
  - name: curl
  - name: firefox
npm:
  - name: typescript
  - name: prettier
`,
			expectError: false,
			validateFn: func(cfg *Config) error {
				if len(cfg.Homebrew) != 3 {
					t.Errorf("Expected 3 homebrew packages, got %d", len(cfg.Homebrew))
				}
				if len(cfg.NPM) != 2 {
					t.Errorf("Expected 2 npm packages, got %d", len(cfg.NPM))
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
				if cfg.Settings.GetOperationTimeout() != 600 {
					t.Errorf("Expected operation timeout 600, got %d", cfg.Settings.GetOperationTimeout())
				}
				if cfg.Settings.GetPackageTimeout() != 300 {
					t.Errorf("Expected package timeout 300, got %d", cfg.Settings.GetPackageTimeout())
				}
				if cfg.Settings.GetDotfileTimeout() != 120 {
					t.Errorf("Expected dotfile timeout 120, got %d", cfg.Settings.GetDotfileTimeout())
				}
				return nil
			},
		},
		{
			name: "invalid config with timeout too high",
			yamlContent: `
settings:
  default_manager: homebrew
  operation_timeout: 5000
`,
			expectError: true,
		},
		{
			name: "invalid config with timeout too low",
			yamlContent: `
settings:
  default_manager: homebrew
  package_timeout: -1
`,
			expectError: true,
		},
		{
			name: "valid config with timeout zero (use default)",
			yamlContent: `
settings:
  default_manager: homebrew
  package_timeout: 0
`,
			expectError: false,
			validateFn: func(cfg *Config) error {
				if cfg.Settings.GetPackageTimeout() != 180 {
					t.Errorf("Expected package timeout default 180, got %d", cfg.Settings.GetPackageTimeout())
				}
				return nil
			},
		},
		{
			name: "invalid yaml syntax",
			yamlContent: `
settings:
  default_manager: homebrew
invalid: [
`,
			expectError: true,
		},
		{
			name: "invalid default manager",
			yamlContent: `
settings:
  default_manager: invalid_manager
`,
			expectError: true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.yamlContent)
			config, err := service.LoadConfigFromReader(reader)
			
			if test.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}
			
			if test.validateFn != nil {
				if err := test.validateFn(config); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

func TestYAMLConfigService_LoadConfigFromFile(t *testing.T) {
	service := NewYAMLConfigService()
	
	// Create temporary file
	tempFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	// Write test config
	testConfig := `
settings:
  default_manager: homebrew
homebrew:
  - name: git
`
	
	if _, err := tempFile.WriteString(testConfig); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	tempFile.Close()
	
	// Test loading from file
	config, err := service.LoadConfigFromFile(tempFile.Name())
	if err != nil {
		t.Fatalf("LoadConfigFromFile() failed: %v", err)
	}
	
	if config.Settings.DefaultManager != "homebrew" {
		t.Errorf("Expected default_manager to be 'homebrew', got %s", config.Settings.DefaultManager)
	}
	
	if len(config.Homebrew) != 1 {
		t.Errorf("Expected 1 homebrew package, got %d", len(config.Homebrew))
	}
}

func TestYAMLConfigService_LoadConfigFromFile_NotFound(t *testing.T) {
	service := NewYAMLConfigService()
	
	_, err := service.LoadConfigFromFile("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file but got none")
	}
}

func TestYAMLConfigService_SaveConfigToWriter(t *testing.T) {
	service := NewYAMLConfigService()
	
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew",
		},
		Homebrew: []HomebrewPackage{
			{Name: "git"},
			{Name: "curl"},
		},
		NPM: []NPMPackage{
			{Name: "typescript"},
		},
	}
	
	var buf strings.Builder
	err := service.SaveConfigToWriter(&buf, config)
	if err != nil {
		t.Fatalf("SaveConfigToWriter() failed: %v", err)
	}
	
	// Verify the output contains expected content
	output := buf.String()
	if !strings.Contains(output, "default_manager: homebrew") {
		t.Error("Output should contain default_manager: homebrew")
	}
	// Check for either "name: git" or "- git" (simple string format)
	if !strings.Contains(output, "name: git") && !strings.Contains(output, "- git") {
		t.Error("Output should contain git package")
	}
	// Check for either "name: typescript" or "- typescript" (simple string format)
	if !strings.Contains(output, "name: typescript") && !strings.Contains(output, "- typescript") {
		t.Error("Output should contain typescript package")
	}
}

func TestYAMLConfigService_SaveConfigToFile(t *testing.T) {
	service := NewYAMLConfigService()
	
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew",
		},
		Homebrew: []HomebrewPackage{
			{Name: "git"},
		},
	}
	
	// Create temporary file
	tempFile, err := os.CreateTemp("", "test_save_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())
	
	// Save config to file
	err = service.SaveConfigToFile(tempFile.Name(), config)
	if err != nil {
		t.Fatalf("SaveConfigToFile() failed: %v", err)
	}
	
	// Verify file was created and contains expected content
	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	
	contentStr := string(content)
	if !strings.Contains(contentStr, "default_manager: homebrew") {
		t.Error("Saved file should contain default_manager: homebrew")
	}
	// Check for either "name: git" or "- git" (simple string format)
	if !strings.Contains(contentStr, "name: git") && !strings.Contains(contentStr, "- git") {
		t.Error("Saved file should contain git package")
	}
}

func TestYAMLConfigService_SaveConfig(t *testing.T) {
	service := NewYAMLConfigService()
	
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew",
		},
	}
	
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Save config to directory
	err := service.SaveConfig(tempDir, config)
	if err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}
	
	// Verify plonk.yaml was created
	configPath := filepath.Join(tempDir, "plonk.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read plonk.yaml: %v", err)
	}
	
	if !strings.Contains(string(content), "default_manager: homebrew") {
		t.Error("plonk.yaml should contain default_manager: homebrew")
	}
}

func TestYAMLConfigService_ValidateConfig(t *testing.T) {
	service := NewYAMLConfigService()
	
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Settings: Settings{
					DefaultManager: "homebrew",
				},
			},
			expectError: false,
		},
		{
			name: "invalid default manager",
			config: &Config{
				Settings: Settings{
					DefaultManager: "invalid",
				},
			},
			expectError: true,
		},
		{
			name: "missing default manager",
			config: &Config{
				Settings: Settings{},
			},
			expectError: true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.ValidateConfig(test.config)
			
			if test.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestYAMLConfigService_ValidateConfigFromReader(t *testing.T) {
	service := NewYAMLConfigService()
	
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
	}{
		{
			name: "valid config",
			yamlContent: `
settings:
  default_manager: homebrew
`,
			expectError: false,
		},
		{
			name: "invalid config",
			yamlContent: `
settings:
  default_manager: invalid
`,
			expectError: true,
		},
		{
			name: "invalid yaml",
			yamlContent: `invalid: [`,
			expectError: true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.yamlContent)
			err := service.ValidateConfigFromReader(reader)
			
			if test.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestYAMLConfigService_LoadConfig_Integration(t *testing.T) {
	service := NewYAMLConfigService()
	
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create plonk.yaml
	configPath := filepath.Join(tempDir, "plonk.yaml")
	configContent := `
settings:
  default_manager: homebrew
homebrew:
  - name: git
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test LoadConfig (uses existing LoadConfig function)
	config, err := service.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}
	
	if config.Settings.DefaultManager != "homebrew" {
		t.Errorf("Expected default_manager to be 'homebrew', got %s", config.Settings.DefaultManager)
	}
	
	if len(config.Homebrew) != 1 {
		t.Errorf("Expected 1 homebrew package, got %d", len(config.Homebrew))
	}
}

// Test that YAMLConfigService implements all required interfaces
func TestYAMLConfigService_ImplementsInterfaces(t *testing.T) {
	service := NewYAMLConfigService()
	
	// Test that it implements ConfigReader
	var _ ConfigReader = service
	
	// Test that it implements ConfigWriter
	var _ ConfigWriter = service
	
	// Test that it implements ConfigValidator
	var _ ConfigValidator = service
	
	// Test that it implements ConfigReadWriter
	var _ ConfigReadWriter = service
	
	// Note: YAMLConfigService doesn't implement DotfileConfigReader or PackageConfigReader
	// because those require a Config instance - that's handled by ConfigAdapter
}

func TestYAMLConfigService_ErrorHandling(t *testing.T) {
	service := NewYAMLConfigService()
	
	// Test SaveConfigToWriter with failing writer
	config := &Config{
		Settings: Settings{DefaultManager: "homebrew"},
	}
	
	failingWriter := &failingWriter{}
	err := service.SaveConfigToWriter(failingWriter, config)
	if err == nil {
		t.Error("Expected error when writing to failing writer")
	}
}

// failingWriter always returns an error on Write
type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrShortWrite
}

func TestSettings_TimeoutMethods(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
		expected struct {
			operation int
			pkg       int
			dotfile   int
		}
	}{
		{
			name:     "default timeouts when not set",
			settings: Settings{DefaultManager: "homebrew"},
			expected: struct {
				operation int
				pkg       int
				dotfile   int
			}{
				operation: 300,
				pkg:       180,
				dotfile:   60,
			},
		},
		{
			name: "custom timeouts when set",
			settings: Settings{
				DefaultManager:   "homebrew",
				OperationTimeout: 600,
				PackageTimeout:   300,
				DotfileTimeout:   120,
			},
			expected: struct {
				operation int
				pkg       int
				dotfile   int
			}{
				operation: 600,
				pkg:       300,
				dotfile:   120,
			},
		},
		{
			name: "fallback to defaults when set to zero",
			settings: Settings{
				DefaultManager:   "homebrew",
				OperationTimeout: 0,
				PackageTimeout:   0,
				DotfileTimeout:   0,
			},
			expected: struct {
				operation int
				pkg       int
				dotfile   int
			}{
				operation: 300,
				pkg:       180,
				dotfile:   60,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.settings.GetOperationTimeout(); got != test.expected.operation {
				t.Errorf("GetOperationTimeout() = %d, want %d", got, test.expected.operation)
			}
			if got := test.settings.GetPackageTimeout(); got != test.expected.pkg {
				t.Errorf("GetPackageTimeout() = %d, want %d", got, test.expected.pkg)
			}
			if got := test.settings.GetDotfileTimeout(); got != test.expected.dotfile {
				t.Errorf("GetDotfileTimeout() = %d, want %d", got, test.expected.dotfile)
			}
		})
	}
}