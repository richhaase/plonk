//go:build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneRepository(t *testing.T) {
	env := NewTestEnv(t)

	// Create a test git repository with dotfiles
	_, err := env.Exec("mkdir", "-p", "/tmp/test-dotfiles")
	require.NoError(t, err)

	// Initialize git repo
	_, err = env.Exec("git", "-C", "/tmp/test-dotfiles", "init")
	require.NoError(t, err)

	// Configure git
	_, err = env.Exec("git", "-C", "/tmp/test-dotfiles", "config", "user.email", "test@example.com")
	require.NoError(t, err)
	_, err = env.Exec("git", "-C", "/tmp/test-dotfiles", "config", "user.name", "Test User")
	require.NoError(t, err)

	// Create test files
	testRC := "# Test RC from clone\nexport CLONED=true\n"
	err = env.WriteFile("/tmp/test-dotfiles/bashrc", []byte(testRC))
	require.NoError(t, err)

	testConfig := "# Test config from clone\n[test]\n  value = cloned\n"
	err = env.WriteFile("/tmp/test-dotfiles/testconfig", []byte(testConfig))
	require.NoError(t, err)

	// Create plonk.lock
	lockContent := `version: 2
resources:
  - type: dotfile
    id: .bashrc
    metadata:
      source: bashrc
      destination: ~/.bashrc
    installed_at: "2025-01-01T00:00:00Z"
  - type: dotfile
    id: .testconfig
    metadata:
      source: testconfig
      destination: ~/.testconfig
    installed_at: "2025-01-01T00:00:00Z"
`
	err = env.WriteFile("/tmp/test-dotfiles/plonk.lock", []byte(lockContent))
	require.NoError(t, err)

	// Commit files
	_, err = env.Exec("git", "-C", "/tmp/test-dotfiles", "add", ".")
	require.NoError(t, err)
	_, err = env.Exec("git", "-C", "/tmp/test-dotfiles", "commit", "-m", "Initial commit")
	require.NoError(t, err)

	// Remove any existing plonk directory
	env.Exec("rm", "-rf", "/home/testuser/.config/plonk")

	// Clone the repository (using --no-apply to skip auto-apply)
	cloneOutput, err := env.Run("clone", "--no-apply", "/tmp/test-dotfiles")
	require.NoError(t, err, "Clone should succeed")

	// Verify clone messages
	assert.Contains(t, cloneOutput, "Repository cloned successfully")
	assert.Contains(t, cloneOutput, "Detected required package managers from lock file")

	// Verify files were cloned
	files, err := env.Exec("ls", "/home/testuser/.config/plonk/")
	require.NoError(t, err)
	assert.Contains(t, files, "bashrc")
	assert.Contains(t, files, "testconfig")
	assert.Contains(t, files, "plonk.lock")

	// Verify lock file was read correctly by checking status
	var statusResult struct {
		ManagedItems []struct {
			Name   string `json:"name"`
			Domain string `json:"domain"`
		} `json:"managed_items"`
	}

	err = env.RunJSON(&statusResult, "status")
	require.NoError(t, err)

	// Should have the dotfiles from the cloned repo
	dotfileCount := 0
	for _, item := range statusResult.ManagedItems {
		if item.Domain == "dotfile" {
			dotfileCount++
		}
	}
	assert.Equal(t, 2, dotfileCount, "Should have 2 dotfiles from cloned repo")
}
