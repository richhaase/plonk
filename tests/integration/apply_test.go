//go:build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyPackages(t *testing.T) {
	env := NewTestEnv(t)

	// Create a lock file with packages
	lockContent := `version: 2
resources:
  - type: package
    id: brew:curl
    metadata:
      manager: brew
      name: curl
    installed_at: "2025-01-01T00:00:00Z"
  - type: package
    id: brew:htop
    metadata:
      manager: brew
      name: htop
    installed_at: "2025-01-01T00:00:00Z"
`

	// Write lock file
	err := env.WriteFile("/home/testuser/.config/plonk/plonk.lock", []byte(lockContent))
	require.NoError(t, err, "Should write lock file")

	// Run apply
	var applyResult struct {
		DryRun   bool   `json:"dry_run"`
		Scope    string `json:"scope"`
		Packages struct {
			TotalInstalled int `json:"total_installed"`
			TotalFailed    int `json:"total_failed"`
		} `json:"packages"`
		Success bool `json:"success"`
	}

	err = env.RunJSON(&applyResult, "apply", "--packages")
	require.NoError(t, err, "Apply should succeed")

	// Verify results
	assert.Equal(t, "packages", applyResult.Scope)
	assert.Equal(t, 2, applyResult.Packages.TotalInstalled)
	assert.Equal(t, 0, applyResult.Packages.TotalFailed)
	assert.False(t, applyResult.DryRun)
	assert.True(t, applyResult.Success)

	// Verify packages are actually installed
	brewList, err := env.Exec("brew", "list")
	require.NoError(t, err)
	assert.Contains(t, brewList, "curl")
	assert.Contains(t, brewList, "htop")
}

func TestApplyDotfiles(t *testing.T) {
	env := NewTestEnv(t)

	// Create test dotfiles
	testConfig := "# Test config file\ntest=true\n"
	err := env.WriteFile("/home/testuser/.config/plonk/testrc", []byte(testConfig))
	require.NoError(t, err)

	// Create lock file with dotfile
	lockContent := `version: 2
resources:
  - type: dotfile
    id: .testrc
    metadata:
      source: testrc
      destination: ~/.testrc
    installed_at: "2025-01-01T00:00:00Z"
`
	err = env.WriteFile("/home/testuser/.config/plonk/plonk.lock", []byte(lockContent))
	require.NoError(t, err)

	// Run apply
	var applyResult struct {
		DryRun   bool   `json:"dry_run"`
		Scope    string `json:"scope"`
		Dotfiles struct {
			Summary struct {
				Added  int `json:"added"`
				Failed int `json:"failed"`
			} `json:"summary"`
		} `json:"dotfiles"`
		Success bool `json:"success"`
	}

	err = env.RunJSON(&applyResult, "apply", "--dotfiles")
	require.NoError(t, err, "Apply should succeed")

	// Verify results
	assert.Equal(t, "dotfiles", applyResult.Scope)
	assert.Equal(t, 1, applyResult.Dotfiles.Summary.Added)
	assert.Equal(t, 0, applyResult.Dotfiles.Summary.Failed)
	assert.True(t, applyResult.Success)

	// Verify the dotfile was deployed
	_, err = env.Exec("ls", "/home/testuser/.testrc")
	require.NoError(t, err, "Dotfile should exist")

	// Verify the content is correct
	content, err := env.Exec("cat", "/home/testuser/.testrc")
	require.NoError(t, err)
	assert.Equal(t, testConfig, content, "Dotfile content should match")
}
