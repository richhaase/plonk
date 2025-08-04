//go:build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigShow(t *testing.T) {
	env := NewTestEnv(t)

	// Test config show returns valid JSON
	var configResult struct {
		DefaultManager    string   `json:"default_manager"`
		OperationTimeout  int      `json:"operation_timeout"`
		PackageTimeout    int      `json:"package_timeout"`
		DotfileTimeout    int      `json:"dotfile_timeout"`
		ExpandDirectories []string `json:"expand_directories"`
		IgnorePatterns    []string `json:"ignore_patterns"`
	}

	err := env.RunJSON(&configResult, "config", "show")
	require.NoError(t, err, "Config show should succeed")

	// Verify default values
	assert.Equal(t, "brew", configResult.DefaultManager)
	assert.Equal(t, 300, configResult.OperationTimeout)
	assert.Equal(t, 120, configResult.PackageTimeout)
	assert.Equal(t, 30, configResult.DotfileTimeout)
	assert.Contains(t, configResult.ExpandDirectories, ".config")
	assert.Contains(t, configResult.IgnorePatterns, ".git")
}

func TestConfigWithCustomValues(t *testing.T) {
	env := NewTestEnv(t)

	// Create custom config
	configContent := `default_manager: npm
package_timeout: 60
ignore_patterns:
  - .git
  - .DS_Store
  - "*.swp"
  - custom_ignore
`
	err := env.WriteFile("/home/testuser/.config/plonk/plonk.yaml", []byte(configContent))
	require.NoError(t, err)

	// Test config show reflects custom values
	var configResult struct {
		DefaultManager   string   `json:"default_manager"`
		PackageTimeout   int      `json:"package_timeout"`
		IgnorePatterns   []string `json:"ignore_patterns"`
		OperationTimeout int      `json:"operation_timeout"`
	}

	err = env.RunJSON(&configResult, "config", "show")
	require.NoError(t, err)

	// Verify custom values
	assert.Equal(t, "npm", configResult.DefaultManager)
	assert.Equal(t, 60, configResult.PackageTimeout)
	assert.Contains(t, configResult.IgnorePatterns, "custom_ignore")
	assert.Contains(t, configResult.IgnorePatterns, "*.swp")

	// Verify defaults still apply for non-specified values
	assert.Equal(t, 300, configResult.OperationTimeout)

	// Test that custom default manager is used
	var searchResult struct {
		Manager string `json:"manager"`
	}
	err = env.RunJSON(&searchResult, "search", "typescript")
	require.NoError(t, err)
	assert.Equal(t, "npm", searchResult.Manager, "Should use npm as default manager")
}
