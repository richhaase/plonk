// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGetEditor(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "VISUAL set",
			envVars:  map[string]string{"VISUAL": "nvim", "EDITOR": "vim"},
			expected: "nvim",
		},
		{
			name:     "only EDITOR set",
			envVars:  map[string]string{"EDITOR": "emacs"},
			expected: "emacs",
		},
		{
			name:     "neither set, default to vim",
			envVars:  map[string]string{},
			expected: "vim",
		},
		{
			name:     "VISUAL empty, use EDITOR",
			envVars:  map[string]string{"VISUAL": "", "EDITOR": "nano"},
			expected: "nano",
		},
		{
			name:     "both empty, use default",
			envVars:  map[string]string{"VISUAL": "", "EDITOR": ""},
			expected: "vim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origVisual := os.Getenv("VISUAL")
			origEditor := os.Getenv("EDITOR")
			defer func() {
				os.Setenv("VISUAL", origVisual)
				os.Setenv("EDITOR", origEditor)
			}()

			// Clear env vars
			os.Unsetenv("VISUAL")
			os.Unsetenv("EDITOR")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			result := getEditor()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name      string
		configDir string
		expected  string
	}{
		{
			name:      "simple path",
			configDir: "/home/user/.config/plonk",
			expected:  "/home/user/.config/plonk/plonk.yaml",
		},
		{
			name:      "path with trailing slash",
			configDir: "/home/user/.config/plonk/",
			expected:  "/home/user/.config/plonk/plonk.yaml",
		},
		{
			name:      "relative path",
			configDir: "config",
			expected:  "config/plonk.yaml",
		},
		{
			name:      "empty path",
			configDir: "",
			expected:  "plonk.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getConfigPath(tt.configDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigEditRoundTripPreservesTopLevelFields(t *testing.T) {
	// Create a test config with some non-default values.
	configContent := `
default_manager: npm
operation_timeout: 600
package_timeout: 200
dotfile_timeout: 90
expand_directories:
  - ".config"
  - ".local/share"
ignore_patterns:
  - "custom_pattern"
`
	configDir := testutil.NewTestConfig(t, configContent)

	// Capture the effective config before edit.
	originalCfg := config.LoadWithDefaults(configDir)

	// Simulate the config edit pipeline without launching an editor.
	tempFile, err := createTempConfigFile(configDir)
	require.NoError(t, err)
	defer os.Remove(tempFile)

	editedCfg, err := parseAndValidateConfig(tempFile)
	require.NoError(t, err)

	// Save non-default values back to plonk.yaml.
	err = saveNonDefaultValues(configDir, editedCfg)
	require.NoError(t, err)

	// Reload config and ensure effective top-level values are preserved.
	reloadedCfg := config.LoadWithDefaults(configDir)

	assert.Equal(t, originalCfg.DefaultManager, reloadedCfg.DefaultManager)
	assert.Equal(t, originalCfg.OperationTimeout, reloadedCfg.OperationTimeout)
	assert.Equal(t, originalCfg.PackageTimeout, reloadedCfg.PackageTimeout)
	assert.Equal(t, originalCfg.DotfileTimeout, reloadedCfg.DotfileTimeout)
	assert.Equal(t, originalCfg.ExpandDirectories, reloadedCfg.ExpandDirectories)
	assert.Equal(t, originalCfg.IgnorePatterns, reloadedCfg.IgnorePatterns)
	assert.Equal(t, originalCfg.Dotfiles, reloadedCfg.Dotfiles)

	// The on-disk config should not contain a managers section yet.
	data, err := os.ReadFile(filepath.Join(configDir, "plonk.yaml"))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "managers:")
}

func TestSaveNonDefaultValuesIncludesManagerDiffs(t *testing.T) {
	configDir := testutil.NewTestConfig(t, "")

	// Start from the effective defaults (includes default managers).
	cfg := config.LoadWithDefaults(configDir)

	if cfg.Managers == nil {
		cfg.Managers = make(map[string]config.ManagerConfig)
	}

	// Add a custom manager and override a built-in manager.
	cfg.Managers["custom-manager"] = config.ManagerConfig{
		Binary: "custom-binary",
	}

	// Create an overridden npm config based on the shipped default.
	defaults := config.GetDefaultManagers()
	if npmDefault, ok := defaults["npm"]; ok {
		overridden := npmDefault
		overridden.Binary = "npm-custom"
		cfg.Managers["npm"] = overridden
	}

	// Save only non-default values.
	err := saveNonDefaultValues(configDir, cfg)
	require.NoError(t, err)

	// plonk.yaml should contain a managers section with custom/overridden managers only.
	data, err := os.ReadFile(filepath.Join(configDir, "plonk.yaml"))
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "managers:")
	assert.Contains(t, content, "custom-manager")
	assert.Contains(t, content, "custom-binary")
	// We should see an npm entry, but not a full dump of all defaults.
	assert.Contains(t, content, "npm:")

	// Reload config and confirm effective managers include defaults plus the overrides.
	reloadedCfg := config.LoadWithDefaults(configDir)

	// Custom manager should be present.
	custom, ok := reloadedCfg.Managers["custom-manager"]
	if assert.True(t, ok) {
		assert.Equal(t, "custom-binary", custom.Binary)
	}

	// Overridden npm manager should reflect the custom binary.
	reloadedNPM, ok := reloadedCfg.Managers["npm"]
	if assert.True(t, ok) {
		assert.Equal(t, "npm-custom", reloadedNPM.Binary)
	}

	// A default manager we didn't touch (e.g., brew) should still be present and unchanged.
	if brewDefault, ok := defaults["brew"]; ok {
		reloadedBrew, ok := reloadedCfg.Managers["brew"]
		if assert.True(t, ok) {
			assert.Equal(t, brewDefault, reloadedBrew)
		}
	}
}

func TestCreateTempConfigFileWritesFullConfig(t *testing.T) {
	// Create a test config with some values, including a managers block.
	configContent := `
default_manager: npm
operation_timeout: 450
managers:
  npm:
    binary: "npm"
`
	configDir := testutil.NewTestConfig(t, configContent)

	// Load the effective config (what config show would use).
	cfg := config.LoadWithDefaults(configDir)
	expectedYAML, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	// Create the temp config file used by config edit.
	tempFile, err := createTempConfigFile(configDir)
	require.NoError(t, err)
	defer os.Remove(tempFile)

	data, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	// Strip header comments from the temp file content to get the raw config YAML.
	lines := strings.Split(string(data), "\n")
	var bodyLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		bodyLines = append(bodyLines, line)
	}
	body := strings.TrimSpace(strings.Join(bodyLines, "\n"))

	expected := strings.TrimSpace(string(expectedYAML))

	assert.Equal(t, expected, body)
}
