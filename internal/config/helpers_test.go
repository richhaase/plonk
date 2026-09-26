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
		result, err := GetHomeDir()
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.True(t, filepath.IsAbs(result))
	})

	t.Run("returns current user home", func(t *testing.T) {
		// GetHomeDir uses os.UserHomeDir() which doesn't depend on HOME env var
		result, err := GetHomeDir()
		assert.NoError(t, err)
		// Should return a valid home directory path
		assert.NotEmpty(t, result)
		assert.True(t, filepath.IsAbs(result))
	})
}

func TestGetDefaults(t *testing.T) {
	defaults := GetDefaults()

	assert.NotNil(t, defaults)
	assert.Equal(t, "brew", defaults.DefaultManager)
	assert.Equal(t, 300, defaults.OperationTimeout)
	assert.Equal(t, 60, defaults.DotfileTimeout)
	assert.Contains(t, defaults.ExpandDirectories, ".config")
	assert.Greater(t, len(defaults.IgnorePatterns), 0)
	assert.Contains(t, defaults.IgnorePatterns, ".DS_Store")
	assert.Greater(t, len(defaults.Dotfiles.UnmanagedFilters), 0)
}

func TestParse(t *testing.T) {
	tests := []struct {
		name, yaml, wantError string
	}{
		{"valid", "default_manager: brew\noperation_timeout: 300\ndotfile_timeout: 60\n", ""},
		{"invalid YAML", "default_manager: [", "invalid YAML"},
		{"invalid values", "default_manager: invalid_manager\noperation_timeout: -1\n", "failed on"},
		{"custom manager", "default_manager: custom-manager\n", "validmanager"},
		{"empty uses defaults", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse([]byte(tt.yaml))
			if tt.wantError != "" {
				assert.ErrorContains(t, err, tt.wantError)
				return
			}
			if assert.NoError(t, err) && assert.NotNil(t, cfg) {
				assert.Equal(t, "brew", cfg.DefaultManager)
				assert.Equal(t, 300, cfg.OperationTimeout)
				assert.Equal(t, 60, cfg.DotfileTimeout)
			}
		})
	}
}

func TestParsePreservesEditorDefaults(t *testing.T) {
	cfg, err := Parse(nil)
	assert.NoError(t, err)
	if assert.NotNil(t, cfg) {
		// The editor has always left omitted nested settings empty, whereas
		// loading a config file merges them with the runtime defaults.
		assert.Empty(t, cfg.Dotfiles.UnmanagedFilters)
	}
	loaded, err := Load(t.TempDir())
	assert.NoError(t, err)
	if assert.NotNil(t, loaded) {
		assert.NotEmpty(t, loaded.Dotfiles.UnmanagedFilters)
	}
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
