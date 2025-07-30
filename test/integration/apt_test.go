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

// Test package to use - tree is small, harmless, and useful
const testPackage = "tree"

func TestAPTIntegration(t *testing.T) {
	// Skip if not on CI
	if os.Getenv("CI") != "true" {
		t.Skip("Integration tests only run in CI")
	}

	// Skip if not on Linux
	if _, err := exec.LookPath("apt-get"); err != nil {
		t.Skip("APT not available on this system")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	plonkDir := filepath.Join(tempDir, ".plonk")

	// Set PLONK_DIR for isolation
	t.Setenv("PLONK_DIR", plonkDir)
	t.Setenv("HOME", tempDir)

	// Build plonk binary for testing
	plonkBinary := buildPlonk(t)

	t.Run("InstallPackage", func(t *testing.T) {
		// First ensure package is not installed
		_ = exec.Command("sudo", "apt-get", "remove", "-y", testPackage).Run()

		// Install package using plonk
		cmd := exec.Command("sudo", plonkBinary, "install", "apt:"+testPackage)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to install package: %s", output)

		// Verify package is installed
		cmd = exec.Command("dpkg", "-l", testPackage)
		err = cmd.Run()
		require.NoError(t, err, "Package should be installed")

		// Verify lock file contains the package
		lockPath := filepath.Join(plonkDir, "plonk.lock")
		require.FileExists(t, lockPath)

		lockContent, err := os.ReadFile(lockPath)
		require.NoError(t, err)
		require.Contains(t, string(lockContent), testPackage)
	})

	t.Run("StatusShowsPackage", func(t *testing.T) {
		cmd := exec.Command(plonkBinary, "status", "--packages")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to run status: %s", output)
		require.Contains(t, string(output), testPackage)
		require.Contains(t, string(output), "apt")
	})

	t.Run("UninstallPackage", func(t *testing.T) {
		// Uninstall package using plonk
		cmd := exec.Command("sudo", plonkBinary, "uninstall", "apt:"+testPackage)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to uninstall package: %s", output)

		// Verify package is removed
		cmd = exec.Command("dpkg", "-l", testPackage)
		err = cmd.Run()
		require.Error(t, err, "Package should be uninstalled")

		// Verify lock file no longer contains the package
		lockContent, err := os.ReadFile(filepath.Join(plonkDir, "plonk.lock"))
		require.NoError(t, err)
		require.NotContains(t, string(lockContent), testPackage)
	})
}

func TestAPTSearchIntegration(t *testing.T) {
	// Skip if not on CI
	if os.Getenv("CI") != "true" {
		t.Skip("Integration tests only run in CI")
	}

	// Skip if not on Linux
	if _, err := exec.LookPath("apt-cache"); err != nil {
		t.Skip("APT not available on this system")
	}

	plonkBinary := buildPlonk(t)

	// Search for a common package
	cmd := exec.Command(plonkBinary, "search", "apt:curl")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Search should succeed: %s", output)
	require.Contains(t, string(output), "curl", "Should find curl package")
}

// buildPlonk builds the plonk binary for testing
func buildPlonk(t *testing.T) string {
	t.Helper()

	// Build in a temporary directory
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "plonk")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/plonk")
	cmd.Dir = getRootDir(t)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build plonk: %s", output)

	return binaryPath
}

// getRootDir finds the project root directory
func getRootDir(t *testing.T) string {
	t.Helper()

	// Start from test directory and walk up
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root")
		}
		dir = parent
	}
}
