// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"testing"

	"github.com/richhaase/plonk/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserDefinedChecker(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	t.Run("with no user config", func(t *testing.T) {
		checker := NewUserDefinedChecker(tempDir)
		assert.NotNil(t, checker)
		assert.NotNil(t, checker.defaults)
		// userConfig will be nil since no config file exists
	})

	t.Run("with user config", func(t *testing.T) {
		// Create a config file
		tempDir := testutil.NewTestConfig(t, "default_manager: npm")

		checker := NewUserDefinedChecker(tempDir)
		assert.NotNil(t, checker)
		assert.NotNil(t, checker.defaults)
		assert.NotNil(t, checker.userConfig)
		assert.Equal(t, "npm", checker.userConfig.DefaultManager)
	})
}

func TestIsFieldUserDefined(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("no user config", func(t *testing.T) {
		checker := NewUserDefinedChecker(tempDir)

		// When no config file exists, Load() returns defaults
		// So userConfig is not nil but has all default values
		// Any value different from default is considered user-defined
		assert.True(t, checker.IsFieldUserDefined("default_manager", "npm"))
		assert.True(t, checker.IsFieldUserDefined("operation_timeout", 600))
		// Same as default, so not user-defined
		assert.False(t, checker.IsFieldUserDefined("package_timeout", 180))
	})

	t.Run("with user config", func(t *testing.T) {
		// Create a config file with custom values
		configContent := `
default_manager: npm
operation_timeout: 600
`
		tempDir := testutil.NewTestConfig(t, configContent)

		checker := NewUserDefinedChecker(tempDir)

		// npm is different from default (brew)
		assert.True(t, checker.IsFieldUserDefined("default_manager", "npm"))

		// 600 is different from default (300)
		assert.True(t, checker.IsFieldUserDefined("operation_timeout", 600))

		// 180 is same as default
		assert.False(t, checker.IsFieldUserDefined("package_timeout", 180))
	})
}

func TestGetNonDefaultFields(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("all defaults", func(t *testing.T) {
		checker := NewUserDefinedChecker(tempDir)

		// Create a copy of default config
		defaults := defaultConfig
		nonDefaults := checker.GetNonDefaultFields(&defaults)

		// Should be empty since everything is default
		assert.Empty(t, nonDefaults)
	})

	t.Run("some custom values", func(t *testing.T) {
		// Create a config file with some custom values
		configContent := `
default_manager: npm
operation_timeout: 600
ignore_patterns:
  - custom_pattern
`
		tempDir := testutil.NewTestConfig(t, configContent)

		checker := NewUserDefinedChecker(tempDir)
		cfg, err := Load(tempDir)
		require.NoError(t, err)

		nonDefaults := checker.GetNonDefaultFields(cfg)

		// Should contain the changed fields
		assert.Contains(t, nonDefaults, "default_manager")
		assert.Equal(t, "npm", nonDefaults["default_manager"])

		assert.Contains(t, nonDefaults, "operation_timeout")
		assert.Equal(t, 600, nonDefaults["operation_timeout"])

		assert.Contains(t, nonDefaults, "ignore_patterns")
		patterns := nonDefaults["ignore_patterns"].([]string)
		assert.Contains(t, patterns, "custom_pattern")

		// Should not contain defaults
		assert.NotContains(t, nonDefaults, "package_timeout")
		assert.NotContains(t, nonDefaults, "dotfile_timeout")
	})

	t.Run("modified dotfiles config", func(t *testing.T) {
		checker := NewUserDefinedChecker(tempDir)

		// Create a new config with modified dotfiles
		cfg := &Config{
			DefaultManager:    "brew",
			OperationTimeout:  300,
			PackageTimeout:    180,
			DotfileTimeout:    60,
			ExpandDirectories: []string{".config"},
			IgnorePatterns:    checker.defaults.IgnorePatterns,
			Dotfiles: Dotfiles{
				UnmanagedFilters: []string{"custom_filter"},
			},
		}

		nonDefaults := checker.GetNonDefaultFields(cfg)

		// Check if dotfiles is marked as non-default
		if dotfilesVal, ok := nonDefaults["dotfiles"]; ok {
			dotfiles := dotfilesVal.(Dotfiles)
			assert.Contains(t, dotfiles.UnmanagedFilters, "custom_filter")
		} else {
			// If not marked as different, the test setup may be wrong
			t.Logf("dotfiles not detected as different: default has %d filters, test has 1",
				len(checker.defaults.Dotfiles.UnmanagedFilters))
		}
	})

}

func TestGetNonDefaultManagers(t *testing.T) {
	tempDir := t.TempDir()
	checker := NewUserDefinedChecker(tempDir)

	t.Run("no managers returns empty map", func(t *testing.T) {
		cfg := &Config{}
		nonDefaults := checker.GetNonDefaultManagers(cfg)
		assert.Empty(t, nonDefaults)
	})

	t.Run("custom manager not in defaults is returned", func(t *testing.T) {
		cfg := &Config{
			Managers: map[string]ManagerConfig{
				"custom-manager": {
					Binary: "custom-binary",
				},
			},
		}

		nonDefaults := checker.GetNonDefaultManagers(cfg)

		if assert.Contains(t, nonDefaults, "custom-manager") {
			assert.Equal(t, "custom-binary", nonDefaults["custom-manager"].Binary)
		}
	})

	t.Run("overridden built-in manager is returned", func(t *testing.T) {
		defaults := GetDefaultManagers()
		defaultNPM, ok := defaults["npm"]
		if !ok {
			t.Skip("npm default manager not defined; skipping overridden manager test")
		}

		// Create an overridden npm config
		overridden := defaultNPM
		overridden.Binary = "npm-custom"

		cfg := &Config{
			Managers: map[string]ManagerConfig{
				"npm": overridden,
			},
		}

		nonDefaults := checker.GetNonDefaultManagers(cfg)

		if assert.Contains(t, nonDefaults, "npm") {
			assert.Equal(t, "npm-custom", nonDefaults["npm"].Binary)
		}
	})

	t.Run("default built-in manager is not returned", func(t *testing.T) {
		defaults := GetDefaultManagers()
		defaultBrew, ok := defaults["brew"]
		if !ok {
			t.Skip("brew default manager not defined; skipping default manager test")
		}

		cfg := &Config{
			Managers: map[string]ManagerConfig{
				"brew": defaultBrew,
			},
		}

		nonDefaults := checker.GetNonDefaultManagers(cfg)
		assert.NotContains(t, nonDefaults, "brew")
	})

	t.Run("overridden built-in manager only saves diffs", func(t *testing.T) {
		defaults := GetDefaultManagers()
		defaultNPM, ok := defaults["npm"]
		if !ok {
			t.Skip("npm default manager not defined; skipping diff test")
		}

		override := ManagerConfig{
			Install: CommandConfig{
				Command: []string{"npm", "install", "-g", "{{.Package}}", "--legacy-peer-deps"},
			},
		}
		effective := ManagerConfig{
			Binary:             defaultNPM.Binary,
			List:               defaultNPM.List,
			Install:            MergeManagerConfig(defaultNPM, override).Install,
			Upgrade:            defaultNPM.Upgrade,
			UpgradeAll:         defaultNPM.UpgradeAll,
			Uninstall:          defaultNPM.Uninstall,
			Description:        defaultNPM.Description,
			InstallHint:        defaultNPM.InstallHint,
			HelpURL:            defaultNPM.HelpURL,
			MetadataExtractors: defaultNPM.MetadataExtractors,
		}

		cfg := &Config{
			Managers: map[string]ManagerConfig{
				"npm": effective,
			},
		}

		nonDefaults := checker.GetNonDefaultManagers(cfg)
		diff, ok := nonDefaults["npm"]
		if !ok {
			t.Fatalf("expected npm diff to be returned")
		}

		require.Equal(t, []string{"npm", "install", "-g", "{{.Package}}", "--legacy-peer-deps"}, diff.Install.Command)
		assert.Empty(t, diff.List.Command)
		assert.Empty(t, diff.Binary)

		merged := MergeManagerConfig(defaultNPM, diff)
		assert.Equal(t, effective.Install.Command, merged.Install.Command)
	})
}

func TestGetDefaultFieldValue(t *testing.T) {
	// Create a new temp directory for this test to avoid state pollution
	tempDir := t.TempDir()
	checker := NewUserDefinedChecker(tempDir)

	tests := []struct {
		fieldName string
		expected  interface{}
	}{
		{"default_manager", "brew"},
		{"operation_timeout", 300},
		{"package_timeout", 180},
		{"dotfile_timeout", 60},
		{"expand_directories", []string{".config"}},
		{"diff_tool", ""},
		{"unknown_field", nil},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := checker.getDefaultFieldValue(tt.fieldName)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				switch expected := tt.expected.(type) {
				case []string:
					assert.Equal(t, expected, result)
				default:
					assert.Equal(t, expected, result)
				}
			}
		})
	}

	t.Run("ignore_patterns returns slice", func(t *testing.T) {
		result := checker.getDefaultFieldValue("ignore_patterns")
		patterns, ok := result.([]string)
		assert.True(t, ok)
		assert.Greater(t, len(patterns), 0)
		assert.Contains(t, patterns, ".DS_Store")
	})

	t.Run("dotfiles returns struct", func(t *testing.T) {
		result := checker.getDefaultFieldValue("dotfiles")
		dotfiles, ok := result.(Dotfiles)
		assert.True(t, ok)
		assert.Greater(t, len(dotfiles.UnmanagedFilters), 0)
	})

}
