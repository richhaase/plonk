// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHomeDir(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	t.Run("returns home directory", func(t *testing.T) {
		// GetHomeDir uses os.UserHomeDir() which doesn't depend on HOME env var
		// It returns the actual user's home directory
		result := GetHomeDir()
		assert.NotEmpty(t, result)
		assert.True(t, filepath.IsAbs(result))
	})

	t.Run("returns current user home", func(t *testing.T) {
		// GetHomeDir uses os.UserHomeDir() which doesn't depend on HOME env var
		result := GetHomeDir()
		// Should return a valid home directory path
		assert.NotEmpty(t, result)
		assert.True(t, filepath.IsAbs(result))
	})
}

func TestGetConfigDir(t *testing.T) {
	// Save original PLONK_DIR
	originalPlonkDir := os.Getenv("PLONK_DIR")
	defer os.Setenv("PLONK_DIR", originalPlonkDir)

	t.Run("uses default config directory", func(t *testing.T) {
		os.Unsetenv("PLONK_DIR")

		result := GetConfigDir()
		expected := GetDefaultConfigDirectory()
		assert.Equal(t, expected, result)
	})

	t.Run("respects PLONK_DIR environment variable", func(t *testing.T) {
		testDir := "/custom/plonk/dir"
		os.Setenv("PLONK_DIR", testDir)

		result := GetConfigDir()
		assert.Equal(t, testDir, result)
	})
}

func TestGetDefaults(t *testing.T) {
	defaults := GetDefaults()

	assert.NotNil(t, defaults)
	assert.Equal(t, "brew", defaults.DefaultManager)
	assert.Equal(t, 300, defaults.OperationTimeout)
	assert.Equal(t, 180, defaults.PackageTimeout)
	assert.Equal(t, 60, defaults.DotfileTimeout)
	assert.Contains(t, defaults.ExpandDirectories, ".config")
	assert.Greater(t, len(defaults.IgnorePatterns), 0)
	assert.Contains(t, defaults.IgnorePatterns, ".DS_Store")
	assert.Greater(t, len(defaults.Dotfiles.UnmanagedFilters), 0)
}

func TestNewSimpleValidator(t *testing.T) {
	validator := NewSimpleValidator()
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.validator)
}

func TestValidateConfigFromYAML(t *testing.T) {
	validator := NewSimpleValidator()

	t.Run("valid config", func(t *testing.T) {
		validYAML := []byte(`
default_manager: brew
operation_timeout: 300
package_timeout: 180
dotfile_timeout: 60
`)
		result := validator.ValidateConfigFromYAML(validYAML)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("invalid YAML", func(t *testing.T) {
		invalidYAML := []byte(`
default_manager: brew
invalid yaml content {{
`)
		result := validator.ValidateConfigFromYAML(invalidYAML)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
		assert.Contains(t, result.Errors[0], "invalid YAML")
	})

	t.Run("invalid config values", func(t *testing.T) {
		invalidConfig := []byte(`
default_manager: invalid_manager
operation_timeout: -1
`)
		result := validator.ValidateConfigFromYAML(invalidConfig)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("empty config uses defaults", func(t *testing.T) {
		emptyYAML := []byte(``)
		result := validator.ValidateConfigFromYAML(emptyYAML)
		// Empty config is valid because defaults are applied
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})
}

func TestGetDefaultConfigDirectory_WithTilde(t *testing.T) {
	// Save original PLONK_DIR and HOME
	originalPlonkDir := os.Getenv("PLONK_DIR")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("PLONK_DIR", originalPlonkDir)
		os.Setenv("HOME", originalHome)
	}()

	t.Run("expands tilde in PLONK_DIR", func(t *testing.T) {
		testHome := "/test/home"
		os.Setenv("HOME", testHome)
		os.Setenv("PLONK_DIR", "~/custom/plonk")

		result := GetDefaultConfigDirectory()
		expected := filepath.Join(testHome, "custom/plonk")
		assert.Equal(t, expected, result)
	})

	t.Run("handles absolute path in PLONK_DIR", func(t *testing.T) {
		absolutePath := "/absolute/path/to/plonk"
		os.Setenv("PLONK_DIR", absolutePath)

		result := GetDefaultConfigDirectory()
		assert.Equal(t, absolutePath, result)
	})
}
