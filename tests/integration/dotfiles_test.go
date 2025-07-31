//go:build integration
// +build integration

package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDotfilesIntegration(t *testing.T) {
	// Skip if not on CI
	if os.Getenv("CI") != "true" {
		t.Skip("Integration tests only run in CI")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	plonkDir := filepath.Join(tempDir, ".plonk")
	homeDir := filepath.Join(tempDir, "home")

	// Create home directory
	require.NoError(t, os.MkdirAll(homeDir, 0755))

	// Set environment for isolation
	t.Setenv("PLONK_DIR", plonkDir)
	t.Setenv("HOME", homeDir)

	// Build plonk binary for testing
	plonkBinary := buildPlonk(t)

	// Create a test dotfile
	testFile := filepath.Join(homeDir, ".testrc")
	testContent := "# Test configuration file\ntest_setting=true\n"
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0644))

	t.Run("AddDotfile", func(t *testing.T) {
		// Add the dotfile
		cmd := exec.Command(plonkBinary, "add", testFile)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to add dotfile: %s", output)

		// Verify source file exists in PLONK_DIR (without leading dot)
		sourceFile := filepath.Join(plonkDir, "testrc")
		require.FileExists(t, sourceFile)

		// Verify content matches
		content, err := os.ReadFile(sourceFile)
		require.NoError(t, err)
		require.Equal(t, testContent, string(content))
	})

	t.Run("StatusShowsDotfile", func(t *testing.T) {
		cmd := exec.Command(plonkBinary, "status", "--dotfiles")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to run status: %s", output)
		require.Contains(t, string(output), ".testrc")
		require.Contains(t, string(output), "managed")
	})

	t.Run("DriftDetection", func(t *testing.T) {
		// Modify the deployed file
		modifiedContent := testContent + "# Modified\n"
		require.NoError(t, os.WriteFile(testFile, []byte(modifiedContent), 0644))

		// Check status shows drift
		cmd := exec.Command(plonkBinary, "status", "--dotfiles")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to run status: %s", output)
		require.Contains(t, string(output), "drifted")

		// Restore with apply
		cmd = exec.Command(plonkBinary, "apply", "--dotfiles")
		output, err = cmd.CombinedOutput()
		require.NoError(t, err, "Failed to apply: %s", output)

		// Verify content is restored
		content, err := os.ReadFile(testFile)
		require.NoError(t, err)
		require.Equal(t, testContent, string(content))
	})

	t.Run("RemoveDotfile", func(t *testing.T) {
		// Remove the dotfile - need to use the full path relative to HOME
		cmd := exec.Command(plonkBinary, "rm", ".testrc")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to remove dotfile: %s", output)

		// Verify source file is removed from PLONK_DIR
		sourceFile := filepath.Join(plonkDir, "testrc")
		require.NoFileExists(t, sourceFile)

		// Verify deployed file still exists (plonk rm only removes from management)
		require.FileExists(t, testFile)
	})
}

func TestNestedDotfilesIntegration(t *testing.T) {
	// Skip if not on CI
	if os.Getenv("CI") != "true" {
		t.Skip("Integration tests only run in CI")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	plonkDir := filepath.Join(tempDir, ".plonk")
	homeDir := filepath.Join(tempDir, "home")

	// Create home directory
	require.NoError(t, os.MkdirAll(homeDir, 0755))

	// Set environment for isolation
	t.Setenv("PLONK_DIR", plonkDir)
	t.Setenv("HOME", homeDir)

	// Build plonk binary for testing
	plonkBinary := buildPlonk(t)

	// Create a nested test dotfile
	configDir := filepath.Join(homeDir, ".config", "myapp")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	testFile := filepath.Join(configDir, "config.yaml")
	testContent := "setting: value\n"
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0644))

	t.Run("AddNestedDotfile", func(t *testing.T) {
		// Add the nested dotfile
		cmd := exec.Command(plonkBinary, "add", testFile)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to add nested dotfile: %s", output)

		// Verify source structure in PLONK_DIR
		sourceFile := filepath.Join(plonkDir, "config", "myapp", "config.yaml")
		require.FileExists(t, sourceFile)

		// Verify content matches
		content, err := os.ReadFile(sourceFile)
		require.NoError(t, err)
		require.Equal(t, testContent, string(content))
	})
}
