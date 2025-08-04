//go:build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPackages(t *testing.T) {
	env := NewTestEnv(t)

	// Install some packages first
	packages := []string{"brew:jq", "brew:tree"}

	for _, pkg := range packages {
		var installResult struct {
			Results []struct {
				Status string `json:"status"`
			} `json:"results"`
		}

		err := env.RunJSON(&installResult, "install", pkg)
		require.NoError(t, err, "Install %s should succeed", pkg)
	}

	// Test list command
	var listResult struct {
		ManagedCount int `json:"managed_count"`
		Managers     []struct {
			Name         string `json:"name"`
			ManagedCount int    `json:"managed_count"`
			Packages     []struct {
				Name  string `json:"name"`
				State string `json:"state"`
			} `json:"packages"`
		} `json:"managers"`
	}

	err := env.RunJSON(&listResult, "list")
	require.NoError(t, err, "List should succeed")

	// Verify counts
	assert.Equal(t, 2, listResult.ManagedCount, "Should have 2 managed packages")

	// Find brew manager
	var brewManager struct {
		Name         string
		ManagedCount int
		Packages     []struct {
			Name  string `json:"name"`
			State string `json:"state"`
		}
	}

	for _, mgr := range listResult.Managers {
		if mgr.Name == "brew" {
			brewManager.Name = mgr.Name
			brewManager.ManagedCount = mgr.ManagedCount
			brewManager.Packages = mgr.Packages
			break
		}
	}

	assert.Equal(t, "brew", brewManager.Name)
	assert.Equal(t, 2, brewManager.ManagedCount)
	assert.Len(t, brewManager.Packages, 2)

	// Verify packages
	foundPackages := make(map[string]string)
	for _, pkg := range brewManager.Packages {
		foundPackages[pkg.Name] = pkg.State
	}

	assert.Equal(t, "managed", foundPackages["jq"])
	assert.Equal(t, "managed", foundPackages["tree"])
}
