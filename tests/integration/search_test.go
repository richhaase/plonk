//go:build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchPackages(t *testing.T) {
	env := NewTestEnv(t)

	// Search for a common package
	var searchResult struct {
		Query   string `json:"query"`
		Manager string `json:"manager"`
		Results []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"results"`
	}

	err := env.RunJSON(&searchResult, "search", "brew:git")
	require.NoError(t, err, "Search should succeed")

	// Verify search metadata
	assert.Equal(t, "git", searchResult.Query)
	assert.Equal(t, "brew", searchResult.Manager)

	// Should find git
	assert.Greater(t, len(searchResult.Results), 0, "Should find at least one result")

	// Look for exact match
	foundGit := false
	for _, result := range searchResult.Results {
		if result.Name == "git" {
			foundGit = true
			assert.NotEmpty(t, result.Description, "Git should have a description")
			break
		}
	}
	assert.True(t, foundGit, "Should find git in search results")
}

func TestSearchWithoutManager(t *testing.T) {
	env := NewTestEnv(t)

	// Search without specifying manager (uses default)
	var searchResult struct {
		Query   string `json:"query"`
		Manager string `json:"manager"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}

	err := env.RunJSON(&searchResult, "search", "wget")
	require.NoError(t, err, "Search should succeed")

	// Should use brew as default
	assert.Equal(t, "wget", searchResult.Query)
	assert.Equal(t, "brew", searchResult.Manager)
	assert.Greater(t, len(searchResult.Results), 0, "Should find results")
}
