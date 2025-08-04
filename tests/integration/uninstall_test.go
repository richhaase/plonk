//go:build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUninstallPackage(t *testing.T) {
	env := NewTestEnv(t)

	// First install a package
	var installResult struct {
		Command    string `json:"command"`
		TotalItems int    `json:"total_items"`
		Results    []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"results"`
	}

	err := env.RunJSON(&installResult, "install", "brew:jq")
	require.NoError(t, err, "Install should succeed")
	assert.Equal(t, "added", installResult.Results[0].Status)

	// Verify it's installed via brew
	brewListBefore, err := env.Exec("brew", "list")
	require.NoError(t, err)
	assert.Contains(t, brewListBefore, "jq", "jq should be in brew list")

	// Now uninstall it
	var uninstallResult struct {
		Command    string `json:"command"`
		TotalItems int    `json:"total_items"`
		Results    []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"results"`
	}

	err = env.RunJSON(&uninstallResult, "uninstall", "brew:jq")
	require.NoError(t, err, "Uninstall should succeed")

	// Verify uninstall result
	assert.Equal(t, "uninstall", uninstallResult.Command)
	assert.Equal(t, 1, uninstallResult.TotalItems)
	assert.Equal(t, "removed", uninstallResult.Results[0].Status)

	// Verify it's not in brew anymore
	brewListAfter, err := env.Exec("brew", "list")
	require.NoError(t, err)
	assert.NotContains(t, brewListAfter, "jq", "jq should not be in brew list")

	// Verify it's not in plonk status
	var statusResult struct {
		ManagedItems []struct {
			Name    string `json:"name"`
			Manager string `json:"manager"`
		} `json:"managed_items"`
	}

	err = env.RunJSON(&statusResult, "status")
	require.NoError(t, err)

	// Should not find jq in managed items
	found := false
	for _, item := range statusResult.ManagedItems {
		if item.Name == "jq" && item.Manager == "brew" {
			found = true
			break
		}
	}
	assert.False(t, found, "jq should not be in managed items after uninstall")
}
