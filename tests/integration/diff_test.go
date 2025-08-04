//go:build integration

package integration_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffDotfile(t *testing.T) {
	env := NewTestEnv(t)

	// Create and add a dotfile
	originalContent := "# Original content\noriginal=true\n"
	err := env.WriteFile("/home/testuser/.testdiff", []byte(originalContent))
	require.NoError(t, err)

	var addResult struct {
		Results []struct {
			Status string `json:"status"`
		} `json:"results"`
	}

	err = env.RunJSON(&addResult, "add", "~/.testdiff")
	require.NoError(t, err, "Add should succeed")

	// Modify the deployed version to create drift
	modifiedContent := "# Modified content\noriginal=false\nmodified=true\n"
	err = env.WriteFile("/home/testuser/.testdiff", []byte(modifiedContent))
	require.NoError(t, err)

	// Run diff command
	diffOutput, err := env.Run("diff", ".testdiff")
	require.NoError(t, err, "Diff should succeed")

	// Verify diff output shows the changes
	assert.Contains(t, diffOutput, "original=true", "Should show original content")
	assert.Contains(t, diffOutput, "original=false", "Should show modified content")
	assert.Contains(t, diffOutput, "modified=true", "Should show added line")

	// Test diff with no drift
	// First restore original content
	err = env.WriteFile("/home/testuser/.testdiff", []byte(originalContent))
	require.NoError(t, err)

	// Apply to ensure symlink points to correct content
	_, err = env.Run("apply", "--dotfiles-only", "-o", "json")
	require.NoError(t, err)

	// Now diff should show no differences
	diffOutput2, err := env.Run("diff")
	if err == nil {
		// Some diff tools return success even with no differences
		assert.True(t,
			strings.Contains(diffOutput2, "No drifted dotfiles found") ||
				len(strings.TrimSpace(diffOutput2)) == 0,
			"Should indicate no differences")
	}
}
