//go:build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveDotfile(t *testing.T) {
	env := NewTestEnv(t)

	// First add a dotfile
	testContent := "# Test file to remove\n"
	err := env.WriteFile("/home/testuser/.testremove", []byte(testContent))
	require.NoError(t, err)

	var addResult struct {
		Results []struct {
			Status string `json:"status"`
		} `json:"results"`
	}

	err = env.RunJSON(&addResult, "add", "~/.testremove")
	require.NoError(t, err, "Add should succeed")
	assert.Equal(t, "added", addResult.Results[0].Status)

	// Verify file exists in config directory
	_, err = env.Exec("ls", "/home/testuser/.config/plonk/testremove")
	require.NoError(t, err, "Source file should exist")

	// Verify symlink exists
	_, err = env.Exec("readlink", "/home/testuser/.testremove")
	require.NoError(t, err, "Symlink should exist")

	// Now remove it
	var rmResult struct {
		TotalItems int `json:"total_items"`
		Results    []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Type   string `json:"type"`
		} `json:"results"`
		Summary struct {
			Removed int `json:"removed"`
			Failed  int `json:"failed"`
		} `json:"summary"`
	}

	err = env.RunJSON(&rmResult, "rm", ".testremove")
	require.NoError(t, err, "Remove should succeed")

	// Verify removal result
	assert.Equal(t, 1, rmResult.TotalItems)
	assert.Equal(t, ".testremove", rmResult.Results[0].Name)
	assert.Equal(t, "removed", rmResult.Results[0].Status)
	assert.Equal(t, "dotfile", rmResult.Results[0].Type)
	assert.Equal(t, 1, rmResult.Summary.Removed)
	assert.Equal(t, 0, rmResult.Summary.Failed)

	// Verify source file is deleted
	_, err = env.Exec("ls", "/home/testuser/.config/plonk/testremove")
	assert.Error(t, err, "Source file should be deleted")

	// Verify symlink is removed
	_, err = env.Exec("readlink", "/home/testuser/.testremove")
	assert.Error(t, err, "Symlink should be removed")

	// Verify it's not in the lock file
	lockContent, err := env.Exec("cat", "/home/testuser/.config/plonk/plonk.lock")
	require.NoError(t, err)
	assert.NotContains(t, lockContent, ".testremove")
}
