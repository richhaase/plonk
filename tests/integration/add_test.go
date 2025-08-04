//go:build integration

package integration_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddDotfile(t *testing.T) {
	env := NewTestEnv(t)

	// Create test dotfiles
	testContent1 := "# Test RC file\nexport TEST=true\n"
	testContent2 := "[user]\n  name = Test User\n  email = test@example.com\n"

	err := env.WriteFile("/home/testuser/.testrc", []byte(testContent1))
	require.NoError(t, err)

	err = env.WriteFile("/home/testuser/.gitconfig.test", []byte(testContent2))
	require.NoError(t, err)

	// Add first dotfile
	var addResult1 struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Action      string `json:"action"`
		Path        string `json:"path"`
	}

	err = env.RunJSON(&addResult1, "add", "~/.testrc")
	require.NoError(t, err, "Add first dotfile should succeed")

	// Verify first add
	assert.Equal(t, "testrc", addResult1.Source)
	assert.Equal(t, "~/.testrc", addResult1.Destination)
	assert.Equal(t, "added", addResult1.Action)
	assert.Contains(t, addResult1.Path, ".testrc")

	// Add second dotfile
	var addResult2 struct {
		Action string `json:"action"`
	}

	err = env.RunJSON(&addResult2, "add", "~/.gitconfig.test")
	require.NoError(t, err, "Add second dotfile should succeed")
	assert.Equal(t, "added", addResult2.Action)

	// Verify files exist in config directory
	configFiles, err := env.Exec("ls", "/home/testuser/.config/plonk/")
	require.NoError(t, err)
	assert.Contains(t, configFiles, "testrc")
	assert.Contains(t, configFiles, "gitconfig.test")

	// Verify lock file exists and contains both dotfiles
	lockContent, err := env.Exec("cat", "/home/testuser/.config/plonk/plonk.lock")
	require.NoError(t, err, "Lock file should exist after adding dotfiles")
	assert.Contains(t, lockContent, "dotfile")
	assert.Contains(t, lockContent, ".testrc")
	assert.Contains(t, lockContent, ".gitconfig.test")

	// Verify symlinks were created
	link1, err := env.Exec("readlink", "/home/testuser/.testrc")
	require.NoError(t, err)
	assert.Contains(t, strings.TrimSpace(link1), ".config/plonk/testrc")

	link2, err := env.Exec("readlink", "/home/testuser/.gitconfig.test")
	require.NoError(t, err)
	assert.Contains(t, strings.TrimSpace(link2), ".config/plonk/gitconfig.test")
}
