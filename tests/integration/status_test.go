//go:build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusWithMultiplePackages(t *testing.T) {
	env := NewTestEnv(t)

	// Install multiple packages
	packages := []string{"brew:jq", "brew:tree", "brew:wget"}

	for _, pkg := range packages {
		var installResult struct {
			Results []struct {
				Status string `json:"status"`
			} `json:"results"`
		}

		err := env.RunJSON(&installResult, "install", pkg)
		require.NoError(t, err, "Install %s should succeed", pkg)
		assert.Equal(t, "added", installResult.Results[0].Status)
	}

	// Run status command
	var statusResult struct {
		Summary struct {
			TotalManaged int `json:"total_managed"`
			Domains      []struct {
				Domain       string `json:"domain"`
				ManagedCount int    `json:"managed_count"`
			} `json:"domains"`
		} `json:"summary"`
		ManagedItems []struct {
			Name    string `json:"name"`
			Domain  string `json:"domain"`
			Manager string `json:"manager"`
		} `json:"managed_items"`
	}

	err := env.RunJSON(&statusResult, "status")
	require.NoError(t, err, "Status should succeed")

	// Verify summary counts
	assert.Equal(t, 3, statusResult.Summary.TotalManaged, "Should have 3 managed items")

	// Find package domain in summary
	packageCount := 0
	for _, d := range statusResult.Summary.Domains {
		if d.Domain == "package" {
			packageCount = d.ManagedCount
			break
		}
	}
	assert.Equal(t, 3, packageCount, "Package domain should have 3 items")

	// Verify all packages are in managed items
	expectedPackages := map[string]bool{
		"jq":   false,
		"tree": false,
		"wget": false,
	}

	for _, item := range statusResult.ManagedItems {
		if item.Domain == "package" && item.Manager == "brew" {
			if _, ok := expectedPackages[item.Name]; ok {
				expectedPackages[item.Name] = true
			}
		}
	}

	// Check all packages were found
	for pkg, found := range expectedPackages {
		assert.True(t, found, "Package %s should be in status output", pkg)
	}
}
