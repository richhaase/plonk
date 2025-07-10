// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestZeroConfigScenarios tests comprehensive zero-config behavior
func TestZeroConfigScenarios(t *testing.T) {
	t.Run("LoadConfig with no file returns empty config", func(t *testing.T) {
		tmpDir := t.TempDir()

		config, err := LoadConfig(tmpDir)
		if err != nil {
			t.Errorf("Expected no error for missing config file, got: %v", err)
		}

		if config == nil {
			t.Fatal("Expected config to be returned, got nil")
		}

		// Config should be empty (all nil/zero values)
		if config.DefaultManager != nil {
			t.Error("Expected DefaultManager to be nil in empty config")
		}
		if config.OperationTimeout != nil {
			t.Error("Expected OperationTimeout to be nil in empty config")
		}

		if len(config.IgnorePatterns) != 0 {
			t.Errorf("Expected empty IgnorePatterns, got %d items", len(config.IgnorePatterns))
		}
	})

	t.Run("Empty config resolves to all defaults", func(t *testing.T) {
		config := &Config{} // Completely empty config

		resolved := config.Resolve()

		// Verify all defaults are applied
		if resolved.DefaultManager != "homebrew" {
			t.Errorf("Expected default manager 'homebrew', got '%s'", resolved.DefaultManager)
		}

		if resolved.OperationTimeout != 300 {
			t.Errorf("Expected operation timeout 300, got %d", resolved.OperationTimeout)
		}

		if resolved.PackageTimeout != 180 {
			t.Errorf("Expected package timeout 180, got %d", resolved.PackageTimeout)
		}

		if resolved.DotfileTimeout != 60 {
			t.Errorf("Expected dotfile timeout 60, got %d", resolved.DotfileTimeout)
		}

		expectedDirs := []string{".config", ".ssh", ".aws", ".kube", ".docker", ".gnupg", ".local"}
		if len(resolved.ExpandDirectories) != len(expectedDirs) {
			t.Errorf("Expected %d expand directories, got %d", len(expectedDirs), len(resolved.ExpandDirectories))
		}

		expectedPatterns := []string{".DS_Store", ".git", "*.backup", "*.tmp", "*.swp"}
		if len(resolved.IgnorePatterns) != len(expectedPatterns) {
			t.Errorf("Expected %d ignore patterns, got %d", len(expectedPatterns), len(resolved.IgnorePatterns))
		}
	})

	t.Run("YAMLConfigService.LoadConfig handles missing files", func(t *testing.T) {
		tmpDir := t.TempDir()
		service := NewYAMLConfigService()

		config, err := service.LoadConfig(tmpDir)
		if err != nil {
			t.Errorf("Expected no error for missing config file, got: %v", err)
		}

		if config == nil {
			t.Fatal("Expected config to be returned, got nil")
		}

		// Should resolve to defaults
		resolved := config.Resolve()
		if resolved.DefaultManager != "homebrew" {
			t.Errorf("Expected default manager 'homebrew', got '%s'", resolved.DefaultManager)
		}
	})

	t.Run("GetOrCreateConfig works with missing files", func(t *testing.T) {
		tmpDir := t.TempDir()

		config, err := GetOrCreateConfig(tmpDir)
		if err != nil {
			t.Errorf("Expected no error for missing config file, got: %v", err)
		}

		if config == nil {
			t.Fatal("Expected config to be returned, got nil")
		}

		// Directory should be created
		if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
			t.Error("Expected config directory to be created")
		}

		// Config should resolve to defaults
		resolved := config.Resolve()
		if resolved.DefaultManager != "homebrew" {
			t.Errorf("Expected default manager 'homebrew', got '%s'", resolved.DefaultManager)
		}
	})
}

// TestPartialConfigOverrides tests config resolution with partial user settings
func TestPartialConfigOverrides(t *testing.T) {
	t.Run("partial settings override defaults", func(t *testing.T) {
		config := &Config{
			DefaultManager:   StringPtr("npm"),
			OperationTimeout: IntPtr(600),
			// Other settings left as nil - should use defaults
		}

		resolved := config.Resolve()

		// Overridden values
		if resolved.DefaultManager != "npm" {
			t.Errorf("Expected overridden default manager 'npm', got '%s'", resolved.DefaultManager)
		}

		if resolved.OperationTimeout != 600 {
			t.Errorf("Expected overridden operation timeout 600, got %d", resolved.OperationTimeout)
		}

		// Default values for unspecified settings
		if resolved.PackageTimeout != 180 {
			t.Errorf("Expected default package timeout 180, got %d", resolved.PackageTimeout)
		}

		if resolved.DotfileTimeout != 60 {
			t.Errorf("Expected default dotfile timeout 60, got %d", resolved.DotfileTimeout)
		}
	})

	t.Run("ignore patterns override completely", func(t *testing.T) {
		customPatterns := []string{"custom1", "custom2"}
		config := &Config{
			IgnorePatterns: customPatterns,
		}

		resolved := config.Resolve()

		if len(resolved.IgnorePatterns) != 2 {
			t.Errorf("Expected 2 custom ignore patterns, got %d", len(resolved.IgnorePatterns))
		}

		if resolved.IgnorePatterns[0] != "custom1" || resolved.IgnorePatterns[1] != "custom2" {
			t.Errorf("Expected custom patterns, got %v", resolved.IgnorePatterns)
		}
	})

	t.Run("expand directories override completely", func(t *testing.T) {
		customDirs := []string{".custom1", ".custom2"}
		config := &Config{
			ExpandDirectories: &customDirs,
		}

		resolved := config.Resolve()

		if len(resolved.ExpandDirectories) != 2 {
			t.Errorf("Expected 2 custom expand directories, got %d", len(resolved.ExpandDirectories))
		}

		if resolved.ExpandDirectories[0] != ".custom1" || resolved.ExpandDirectories[1] != ".custom2" {
			t.Errorf("Expected custom directories, got %v", resolved.ExpandDirectories)
		}
	})
}

// TestZeroConfigIntegration tests integration scenarios
func TestZeroConfigIntegration(t *testing.T) {
	t.Run("config adapter works with zero config", func(t *testing.T) {
		config := &Config{} // Empty config
		adapter := NewConfigAdapter(config)

		// Should be able to get dotfile targets (though may be empty)
		targets := adapter.GetDotfileTargets()
		if targets == nil {
			t.Error("Expected targets map to be non-nil")
		}

		// Should be able to get packages (should be empty - now in lock file)
		for _, manager := range []string{"homebrew", "npm", "cargo"} {
			packages, err := adapter.GetPackagesForManager(manager)
			if err != nil {
				t.Errorf("Expected no error for manager %s, got: %v", manager, err)
			}
			if len(packages) != 0 {
				t.Errorf("Expected empty packages for %s (now in lock file), got %d", manager, len(packages))
			}
		}
	})

	t.Run("validation works with zero config", func(t *testing.T) {
		config := &Config{} // Empty config
		validator := NewSimpleValidator()

		result := validator.ValidateConfig(config)
		if !result.IsValid() {
			t.Errorf("Expected empty config to be valid, got errors: %v", result.Errors)
		}
	})

	t.Run("YAML service default config", func(t *testing.T) {
		service := NewYAMLConfigService()
		config := service.GetDefaultConfig()

		if config == nil {
			t.Fatal("Expected default config to be returned, got nil")
		}

		// Should have defaults populated - fields should be set

		if config.DefaultManager == nil || *config.DefaultManager != "homebrew" {
			t.Error("Expected default manager to be 'homebrew' in default config")
		}

		if len(config.IgnorePatterns) == 0 {
			t.Error("Expected ignore patterns to be populated in default config")
		}
	})
}

// TestZeroConfigFileOperations tests file operations with zero config
func TestZeroConfigFileOperations(t *testing.T) {
	t.Run("load from missing directory", func(t *testing.T) {
		// Use a directory that doesn't exist
		nonExistentDir := filepath.Join(t.TempDir(), "does-not-exist")

		config, err := LoadConfig(nonExistentDir)
		if err != nil {
			t.Errorf("Expected no error for missing directory, got: %v", err)
		}

		if config == nil {
			t.Fatal("Expected config to be returned, got nil")
		}

		// Should resolve to defaults
		resolved := config.Resolve()
		if resolved.DefaultManager != "homebrew" {
			t.Errorf("Expected default manager 'homebrew', got '%s'", resolved.DefaultManager)
		}
	})

	t.Run("corrupted config file still fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "plonk.yaml")

		// Write invalid YAML
		err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadConfig(tmpDir)
		if err == nil {
			t.Error("Expected error for corrupted config file")
		}
	})
}
