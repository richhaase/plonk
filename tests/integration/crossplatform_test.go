//go:build integration
// +build integration

package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCrossPlatformIntegration(t *testing.T) {
	// Skip if not on CI
	if os.Getenv("CI") != "true" {
		t.Skip("Integration tests only run in CI")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	plonkDir := filepath.Join(tempDir, ".plonk")

	// Set PLONK_DIR for isolation
	t.Setenv("PLONK_DIR", plonkDir)
	t.Setenv("HOME", tempDir)

	// Build plonk binary for testing
	plonkBinary := buildPlonk(t)

	// Create a mixed-platform config
	require.NoError(t, os.MkdirAll(plonkDir, 0755))

	config := map[string]interface{}{
		"default_manager": "brew",
	}

	configData, err := yaml.Marshal(config)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(plonkDir, "plonk.yaml"), configData, 0644))

	// Create a lock file with Homebrew packages
	lockContent := `version: 2
packages:
  brew:
    - name: jq
      version: "1.6"
`
	require.NoError(t, os.WriteFile(filepath.Join(plonkDir, "plonk.lock"), []byte(lockContent), 0644))

	t.Run("DoctorShowsPlatformSpecific", func(t *testing.T) {
		cmd := exec.Command(plonkBinary, "doctor")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Doctor should succeed: %s", output)

		outputStr := string(output)

		// Homebrew availability depends on whether it's installed
		require.Contains(t, outputStr, "brew:")
	})

	t.Run("StatusShowsOnlyRelevantPackages", func(t *testing.T) {
		cmd := exec.Command(plonkBinary, "status", "--packages")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Status should succeed: %s", output)

		outputStr := string(output)

		// Homebrew packages shown if Homebrew is available
		if _, err := exec.LookPath("brew"); err == nil {
			require.Contains(t, outputStr, "jq")
		}
	})
}

func TestPackageManagerDetection(t *testing.T) {
	// Skip if not on CI
	if os.Getenv("CI") != "true" {
		t.Skip("Integration tests only run in CI")
	}

	// Build plonk binary for testing
	plonkBinary := buildPlonk(t)

	t.Run("SearchWithUnavailableManager", func(t *testing.T) {
		// Try to use a fake manager
		cmd := exec.Command(plonkBinary, "search", "fake:nginx")
		output, err := cmd.CombinedOutput()
		require.Error(t, err, "Fake manager search should fail")
		require.Contains(t, string(output), "unsupported")
	})
}
