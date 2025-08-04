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
		Command string `json:"command"`
		Summary struct {
			Packages struct {
				Installed int `json:"installed"`
				Failed    int `json:"failed"`
			} `json:"packages"`
		} `json:"summary"`
		DryRun bool `json:"dry_run"`
	}

	err = env.RunJSON(&applyResult, "apply", "--packages-only")
	require.NoError(t, err, "Apply should succeed")

	// Verify results
	assert.Equal(t, "apply", applyResult.Command)
	assert.Equal(t, 2, applyResult.Summary.Packages.Installed)
	assert.Equal(t, 0, applyResult.Summary.Packages.Failed)
	assert.False(t, applyResult.DryRun)

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
		Command string `json:"command"`
		Summary struct {
			Dotfiles struct {
				Deployed int `json:"deployed"`
				Failed   int `json:"failed"`
			} `json:"dotfiles"`
		} `json:"summary"`
	}

	err = env.RunJSON(&applyResult, "apply", "--dotfiles-only")
	require.NoError(t, err, "Apply should succeed")

	// Verify results
	assert.Equal(t, 1, applyResult.Summary.Dotfiles.Deployed)
	assert.Equal(t, 0, applyResult.Summary.Dotfiles.Failed)

	// Verify symlink exists
	output, err := env.Exec("ls", "-la", "/home/testuser/.testrc")
	require.NoError(t, err, "Symlink should exist")
	assert.Contains(t, output, "->", "Should be a symlink")
	assert.Contains(t, output, ".config/plonk/testrc", "Should point to config directory")
}
