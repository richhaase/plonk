// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"testing"

	"github.com/richhaase/plonk/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNonDefaultFields(t *testing.T) {

	t.Run("all defaults", func(t *testing.T) {

		// Create a copy of default config
		defaults := defaultConfig
		nonDefaults := GetNonDefaultFields(&defaults)

		// Should be empty since everything is default
		assert.Empty(t, nonDefaults)
	})

	t.Run("some custom values", func(t *testing.T) {
		// Create a config file with some custom values
		configContent := `
default_manager: cargo
operation_timeout: 600
diff_tool: delta
ignore_patterns:
  - custom_pattern
`
		tempDir := testutil.NewTestConfig(t, configContent)

		cfg, err := Load(tempDir)
		require.NoError(t, err)

		nonDefaults := GetNonDefaultFields(cfg)

		// Should contain the changed fields
		assert.Contains(t, nonDefaults, "default_manager")
		assert.Equal(t, "cargo", nonDefaults["default_manager"])

		assert.Contains(t, nonDefaults, "operation_timeout")
		assert.Equal(t, 600, nonDefaults["operation_timeout"])

		assert.Contains(t, nonDefaults, "diff_tool")
		assert.Equal(t, "delta", nonDefaults["diff_tool"])

		assert.Contains(t, nonDefaults, "ignore_patterns")
		patterns := nonDefaults["ignore_patterns"].([]string)
		assert.Contains(t, patterns, "custom_pattern")

		// Should not contain defaults
		assert.NotContains(t, nonDefaults, "dotfile_timeout")
	})

	t.Run("modified dotfiles config", func(t *testing.T) {

		// Create a new config with modified dotfiles
		cfg := &Config{
			DefaultManager:    "brew",
			OperationTimeout:  300,
			DotfileTimeout:    60,
			ExpandDirectories: []string{".config"},
			IgnorePatterns:    defaultConfig.IgnorePatterns,
			Dotfiles: Dotfiles{
				UnmanagedFilters: []string{"custom_filter"},
			},
		}

		nonDefaults := GetNonDefaultFields(cfg)

		// Check if dotfiles is marked as non-default
		require.Contains(t, nonDefaults, "dotfiles")
		dotfiles := nonDefaults["dotfiles"].(Dotfiles)
		assert.Contains(t, dotfiles.UnmanagedFilters, "custom_filter")
	})

}
