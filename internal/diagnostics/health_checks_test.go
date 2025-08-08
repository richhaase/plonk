// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package diagnostics

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/richhaase/plonk/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckSystemRequirements(t *testing.T) {
	check := checkSystemRequirements()

	assert.Equal(t, "System Requirements", check.Name)
	assert.Equal(t, "system", check.Category)

	// Check should pass on supported platforms (darwin, linux)
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		assert.Equal(t, "pass", check.Status)
		assert.Equal(t, "System requirements met", check.Message)
	} else {
		assert.Equal(t, "fail", check.Status)
		assert.Contains(t, check.Issues[0], "Unsupported operating system")
	}

	// Verify details include OS and architecture
	assert.Contains(t, check.Details[len(check.Details)-2], "OS:")
	assert.Contains(t, check.Details[len(check.Details)-1], "Architecture:")
}

func TestCheckEnvironmentVariables(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	t.Run("with PATH set", func(t *testing.T) {
		os.Setenv("PATH", "/usr/bin:/usr/local/bin")
		check := checkEnvironmentVariables()

		assert.Equal(t, "Environment Variables", check.Name)
		assert.Equal(t, "environment", check.Category)
		assert.Equal(t, "pass", check.Status)
		assert.Equal(t, "Environment variables configured", check.Message)

		// Should have HOME, PLONK_DIR, and PATH details
		assert.GreaterOrEqual(t, len(check.Details), 3)
	})

	t.Run("without PATH", func(t *testing.T) {
		os.Unsetenv("PATH")
		check := checkEnvironmentVariables()

		assert.Equal(t, "fail", check.Status)
		assert.Equal(t, "Critical environment variables missing", check.Message)
		assert.Contains(t, check.Issues[0], "PATH environment variable is not set")
	})
}

func TestCheckPermissions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override config directory for test
	oldPlonkDir := os.Getenv("PLONK_DIR")
	os.Setenv("PLONK_DIR", tempDir)
	defer os.Setenv("PLONK_DIR", oldPlonkDir)

	check := checkPermissions()

	assert.Equal(t, "Permissions", check.Name)
	assert.Equal(t, "permissions", check.Category)
	assert.Equal(t, "pass", check.Status)
	assert.Equal(t, "File permissions are correct", check.Message)
	assert.Contains(t, check.Details[0], "Config directory is writable")
}

func TestCheckConfigurationFile(t *testing.T) {
	t.Run("config file exists", func(t *testing.T) {
		tempDir := testutil.NewTestConfig(t, "default_manager: brew")
		testutil.SetEnv(t, "PLONK_DIR", tempDir)

		check := checkConfigurationFile()

		assert.Equal(t, "Configuration File", check.Name)
		assert.Equal(t, "configuration", check.Category)
		assert.Equal(t, "pass", check.Status)
		assert.Contains(t, check.Message, "Configuration file exists")
		assert.Contains(t, check.Details[0], "Config file size: 21 bytes")
	})

	t.Run("config file does not exist", func(t *testing.T) {
		tempDir := testutil.NewTestConfig(t, "")
		testutil.SetEnv(t, "PLONK_DIR", tempDir)

		check := checkConfigurationFile()

		assert.Equal(t, "info", check.Status)
		assert.Contains(t, check.Message, "Configuration file does not exist")
		assert.Contains(t, check.Details[0], "Will use default configuration")
	})
}

func TestCheckConfigurationValidity(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		validConfig := `default_manager: brew
ignore_patterns:
  - "*.tmp"
  - ".DS_Store"
`
		tempDir := testutil.NewTestConfig(t, validConfig)
		testutil.SetEnv(t, "PLONK_DIR", tempDir)

		check := checkConfigurationValidity()

		assert.Equal(t, "Configuration Validity", check.Name)
		assert.Equal(t, "pass", check.Status)
		assert.Contains(t, check.Details[0], "Default manager: brew")
		assert.Contains(t, check.Details[1], "Ignore patterns: 2")
	})

	t.Run("invalid config", func(t *testing.T) {
		invalidConfig := `invalid yaml content {{`
		tempDir := testutil.NewTestConfig(t, invalidConfig)
		testutil.SetEnv(t, "PLONK_DIR", tempDir)

		check := checkConfigurationValidity()

		assert.Equal(t, "fail", check.Status)
		assert.Contains(t, check.Message, "Configuration has format errors")
		assert.Contains(t, check.Issues[0], "Configuration is invalid")
	})

	t.Run("no config file", func(t *testing.T) {
		tempDir := t.TempDir()

		oldPlonkDir := os.Getenv("PLONK_DIR")
		os.Setenv("PLONK_DIR", tempDir)
		defer os.Setenv("PLONK_DIR", oldPlonkDir)

		check := checkConfigurationValidity()

		// When no config file exists, it loads defaults successfully
		assert.Equal(t, "pass", check.Status)
		assert.Contains(t, check.Message, "Configuration is valid")
	})
}

func TestCheckLockFile(t *testing.T) {
	t.Run("lock file exists", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")
		lockContent := `version: 2
resources: []`
		require.NoError(t, os.WriteFile(lockPath, []byte(lockContent), 0644))

		oldPlonkDir := os.Getenv("PLONK_DIR")
		os.Setenv("PLONK_DIR", tempDir)
		defer os.Setenv("PLONK_DIR", oldPlonkDir)

		check := checkLockFile()

		assert.Equal(t, "Lock File", check.Name)
		assert.Equal(t, "pass", check.Status)
		assert.Contains(t, check.Message, "Lock file exists")
	})

	t.Run("lock file does not exist", func(t *testing.T) {
		tempDir := t.TempDir()

		oldPlonkDir := os.Getenv("PLONK_DIR")
		os.Setenv("PLONK_DIR", tempDir)
		defer os.Setenv("PLONK_DIR", oldPlonkDir)

		check := checkLockFile()

		assert.Equal(t, "info", check.Status)
		assert.Contains(t, check.Message, "Lock file does not exist")
	})

	t.Run("empty lock file", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")
		require.NoError(t, os.WriteFile(lockPath, []byte(""), 0644))

		oldPlonkDir := os.Getenv("PLONK_DIR")
		os.Setenv("PLONK_DIR", tempDir)
		defer os.Setenv("PLONK_DIR", oldPlonkDir)

		check := checkLockFile()

		assert.Equal(t, "warn", check.Status)
		assert.Contains(t, check.Message, "Lock file is empty")
	})
}

func TestCheckLockFileValidity(t *testing.T) {
	t.Run("valid lock file with packages", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")
		lockContent := `version: 2
resources:
  - type: package
    name: ripgrep
    metadata:
      manager: brew
      version: "13.0.0"
  - type: package
    name: prettier
    metadata:
      manager: npm
      version: "2.8.0"`
		require.NoError(t, os.WriteFile(lockPath, []byte(lockContent), 0644))

		oldPlonkDir := os.Getenv("PLONK_DIR")
		os.Setenv("PLONK_DIR", tempDir)
		defer os.Setenv("PLONK_DIR", oldPlonkDir)

		check := checkLockFileValidity()

		assert.Equal(t, "Lock File Validity", check.Name)
		assert.Equal(t, "pass", check.Status)
		assert.Contains(t, check.Message, "Lock file is valid")

		// Check details for package counts
		hasBrewDetail := false
		hasNpmDetail := false
		hasTotalDetail := false
		for _, detail := range check.Details {
			if detail == "brew packages: 1" {
				hasBrewDetail = true
			}
			if detail == "npm packages: 1" {
				hasNpmDetail = true
			}
			if detail == "Total managed packages: 2" {
				hasTotalDetail = true
			}
		}
		assert.True(t, hasBrewDetail, "Should have brew package count")
		assert.True(t, hasNpmDetail, "Should have npm package count")
		assert.True(t, hasTotalDetail, "Should have total package count")
	})

	t.Run("valid but empty lock file", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")
		lockContent := `version: 2
resources: []`
		require.NoError(t, os.WriteFile(lockPath, []byte(lockContent), 0644))

		oldPlonkDir := os.Getenv("PLONK_DIR")
		os.Setenv("PLONK_DIR", tempDir)
		defer os.Setenv("PLONK_DIR", oldPlonkDir)

		check := checkLockFileValidity()

		assert.Equal(t, "info", check.Status)
		assert.Contains(t, check.Message, "Lock file is valid but contains no packages")
	})

	t.Run("invalid lock file", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")
		require.NoError(t, os.WriteFile(lockPath, []byte("invalid yaml {{"), 0644))

		oldPlonkDir := os.Getenv("PLONK_DIR")
		os.Setenv("PLONK_DIR", tempDir)
		defer os.Setenv("PLONK_DIR", oldPlonkDir)

		check := checkLockFileValidity()

		assert.Equal(t, "fail", check.Status)
		assert.Contains(t, check.Message, "Lock file has format errors")
		assert.Contains(t, check.Issues[0], "Lock file is invalid")
	})
}

func TestCheckExecutablePath(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	t.Run("plonk in PATH", func(t *testing.T) {
		// This test assumes plonk might be in PATH during development
		// If not, the test will still pass but with warning status
		check := checkExecutablePath()

		assert.Equal(t, "Executable Path", check.Name)
		assert.Equal(t, "installation", check.Category)

		// Either pass (found) or warn (not found) is acceptable
		assert.Contains(t, []string{"pass", "warn"}, check.Status)
	})

	t.Run("plonk not in PATH", func(t *testing.T) {
		// Set PATH to a directory that definitely doesn't have plonk
		os.Setenv("PATH", "/nonexistent")
		check := checkExecutablePath()

		assert.Equal(t, "warn", check.Status)
		assert.Contains(t, check.Message, "Executable not in PATH")
		assert.Contains(t, check.Issues[0], "plonk executable not found in PATH")
	})
}

func TestHealthReportStructures(t *testing.T) {
	t.Run("HealthStatus", func(t *testing.T) {
		status := HealthStatus{
			Status:  "healthy",
			Message: "All systems operational",
		}
		assert.Equal(t, "healthy", status.Status)
		assert.Equal(t, "All systems operational", status.Message)
	})

	t.Run("HealthCheck", func(t *testing.T) {
		check := HealthCheck{
			Name:        "Test Check",
			Category:    "test",
			Status:      "pass",
			Message:     "Test passed",
			Details:     []string{"Detail 1", "Detail 2"},
			Issues:      []string{"Issue 1"},
			Suggestions: []string{"Suggestion 1"},
		}
		assert.Equal(t, "Test Check", check.Name)
		assert.Equal(t, "test", check.Category)
		assert.Equal(t, "pass", check.Status)
		assert.Len(t, check.Details, 2)
		assert.Len(t, check.Issues, 1)
		assert.Len(t, check.Suggestions, 1)
	})

	t.Run("HealthReport", func(t *testing.T) {
		report := HealthReport{
			Overall: HealthStatus{
				Status:  "healthy",
				Message: "All good",
			},
			Checks: []HealthCheck{
				{Name: "Check 1", Status: "pass"},
				{Name: "Check 2", Status: "pass"},
			},
		}
		assert.Equal(t, "healthy", report.Overall.Status)
		assert.Len(t, report.Checks, 2)
	})
}
