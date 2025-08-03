// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package clone

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/diagnostics"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetManagerDescription(t *testing.T) {
	tests := []struct {
		name     string
		manager  string
		expected string
	}{
		{
			name:     "homebrew",
			manager:  "homebrew",
			expected: "Homebrew (macOS/Linux package manager)",
		},
		{
			name:     "brew alias",
			manager:  "brew",
			expected: "Homebrew (macOS/Linux package manager)",
		},
		{
			name:     "cargo",
			manager:  "cargo",
			expected: "Cargo (Rust package manager)",
		},
		{
			name:     "npm",
			manager:  "npm",
			expected: "npm (Node.js package manager)",
		},
		{
			name:     "pip",
			manager:  "pip",
			expected: "pip (Python package manager)",
		},
		{
			name:     "gem",
			manager:  "gem",
			expected: "gem (Ruby package manager)",
		},
		{
			name:     "go",
			manager:  "go",
			expected: "go (Go package manager)",
		},
		{
			name:     "unknown",
			manager:  "unknown-manager",
			expected: "unknown-manager package manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getManagerDescription(tt.manager)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetManualInstallInstructions(t *testing.T) {
	tests := []struct {
		name     string
		manager  string
		expected string
	}{
		{
			name:     "homebrew",
			manager:  "homebrew",
			expected: "Visit https://brew.sh for installation instructions (prerequisite)",
		},
		{
			name:     "brew alias",
			manager:  "brew",
			expected: "Visit https://brew.sh for installation instructions (prerequisite)",
		},
		{
			name:     "cargo",
			manager:  "cargo",
			expected: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
		},
		{
			name:     "npm",
			manager:  "npm",
			expected: "Install Node.js from https://nodejs.org/ or use brew install node",
		},
		{
			name:     "pip",
			manager:  "pip",
			expected: "Install Python from https://python.org/ or use brew install python",
		},
		{
			name:     "gem",
			manager:  "gem",
			expected: "Install Ruby from https://ruby-lang.org/ or use brew install ruby",
		},
		{
			name:     "go",
			manager:  "go",
			expected: "Install Go from https://golang.org/dl/ or use brew install go",
		},
		{
			name:     "unknown",
			manager:  "unknown-manager",
			expected: "See official documentation for installation instructions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getManualInstallInstructions(tt.manager)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectRequiredManagers(t *testing.T) {
	t.Run("v2 lock file with metadata", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")

		// Create a v2 lock file with manager in metadata
		lockFile := &lock.Lock{
			Version: 2,
			Resources: []lock.ResourceEntry{
				{
					Type: "package",
					ID:   "brew:ripgrep",
					Metadata: map[string]interface{}{
						"manager": "brew",
						"version": "13.0.0",
					},
				},
				{
					Type: "package",
					ID:   "npm:prettier",
					Metadata: map[string]interface{}{
						"manager": "npm",
						"version": "2.8.0",
					},
				},
				{
					Type: "package",
					ID:   "brew:fd",
					Metadata: map[string]interface{}{
						"manager": "brew",
						"version": "8.7.0",
					},
				},
				{
					Type: "dotfile",
					ID:   "dotfile:.vimrc",
					Metadata: map[string]interface{}{
						"name": ".vimrc",
						"path": "vimrc",
					},
				},
			},
		}

		// Write lock file
		lockService := lock.NewYAMLLockService(tempDir)
		require.NoError(t, lockService.Write(lockFile))

		// Detect managers
		managers, err := DetectRequiredManagers(lockPath)
		require.NoError(t, err)

		// Should have brew and npm (unique)
		assert.Len(t, managers, 2)
		assert.Contains(t, managers, "brew")
		assert.Contains(t, managers, "npm")
	})

	t.Run("v2 lock file with ID prefix fallback", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")

		// Create a v2 lock file where manager is extracted from ID
		lockFile := &lock.Lock{
			Version: 2,
			Resources: []lock.ResourceEntry{
				{
					Type: "package",
					ID:   "cargo:tokio",
					Metadata: map[string]interface{}{
						"version": "1.35.0",
					},
				},
				{
					Type: "package",
					ID:   "pip:requests",
					Metadata: map[string]interface{}{
						"version": "2.31.0",
					},
				},
				{
					Type: "package",
					ID:   "cargo:serde",
					Metadata: map[string]interface{}{
						"version": "1.0.0",
					},
				},
			},
		}

		// Write lock file
		lockService := lock.NewYAMLLockService(tempDir)
		require.NoError(t, lockService.Write(lockFile))

		// Detect managers
		managers, err := DetectRequiredManagers(lockPath)
		require.NoError(t, err)

		// Should have cargo and pip (unique)
		assert.Len(t, managers, 2)
		assert.Contains(t, managers, "cargo")
		assert.Contains(t, managers, "pip")
	})

	t.Run("empty lock file", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")

		lockFile := &lock.Lock{
			Version:   2,
			Resources: []lock.ResourceEntry{},
		}

		// Write lock file
		lockService := lock.NewYAMLLockService(tempDir)
		require.NoError(t, lockService.Write(lockFile))

		// Detect managers
		managers, err := DetectRequiredManagers(lockPath)
		require.NoError(t, err)

		assert.Empty(t, managers)
	})

	t.Run("non-existent lock file", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")

		// Don't create the file - the lock service returns empty lock
		managers, err := DetectRequiredManagers(lockPath)
		assert.NoError(t, err)
		assert.Empty(t, managers)
	})

	t.Run("only dotfiles in lock file", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")

		lockFile := &lock.Lock{
			Version: 2,
			Resources: []lock.ResourceEntry{
				{
					Type: "dotfile",
					ID:   "dotfile:.bashrc",
					Metadata: map[string]interface{}{
						"name": ".bashrc",
						"path": "bashrc",
					},
				},
				{
					Type: "dotfile",
					ID:   "dotfile:.vimrc",
					Metadata: map[string]interface{}{
						"name": ".vimrc",
						"path": "vimrc",
					},
				},
			},
		}

		// Write lock file
		lockService := lock.NewYAMLLockService(tempDir)
		require.NoError(t, lockService.Write(lockFile))

		// Detect managers
		managers, err := DetectRequiredManagers(lockPath)
		require.NoError(t, err)

		assert.Empty(t, managers)
	})
}

func TestFindMissingPackageManagers(t *testing.T) {
	t.Run("extract missing managers", func(t *testing.T) {
		report := diagnostics.HealthReport{
			Checks: []diagnostics.HealthCheck{
				{
					Name:     "Package Manager Availability",
					Category: "package-managers",
					Details: []string{
						"brew: available",
						"npm: not available",
						"pip: available",
						"cargo: not available",
						"gem: not available",
						"go: available",
					},
				},
			},
		}

		missing := findMissingPackageManagers(report)

		assert.Len(t, missing, 3)
		assert.Contains(t, missing, "npm")
		assert.Contains(t, missing, "cargo")
		assert.Contains(t, missing, "gem")
	})

	t.Run("no missing managers", func(t *testing.T) {
		report := diagnostics.HealthReport{
			Checks: []diagnostics.HealthCheck{
				{
					Name:     "Package Manager Availability",
					Category: "package-managers",
					Details: []string{
						"brew: available",
						"npm: available",
					},
				},
			},
		}

		missing := findMissingPackageManagers(report)
		assert.Empty(t, missing)
	})

	t.Run("empty report", func(t *testing.T) {
		report := diagnostics.HealthReport{
			Checks: []diagnostics.HealthCheck{},
		}

		missing := findMissingPackageManagers(report)
		assert.Empty(t, missing)
	})

	t.Run("different check category", func(t *testing.T) {
		report := diagnostics.HealthReport{
			Checks: []diagnostics.HealthCheck{
				{
					Name:     "System Check",
					Category: "system",
					Details: []string{
						"brew: not available",
						"npm: not available",
					},
				},
			},
		}

		missing := findMissingPackageManagers(report)
		assert.Empty(t, missing)
	})

	t.Run("malformed details", func(t *testing.T) {
		report := diagnostics.HealthReport{
			Checks: []diagnostics.HealthCheck{
				{
					Name:     "Package Manager Availability",
					Category: "package-managers",
					Details: []string{
						"brew available", // Missing colon
						"npm: not available",
						"no status here",    // No colon at all
						": not available",   // Empty manager name
						"cargo : available", // Extra spaces
					},
				},
			},
		}

		missing := findMissingPackageManagers(report)

		// Should only extract npm
		assert.Len(t, missing, 1)
		assert.Contains(t, missing, "npm")
	})
}

func TestCreateDefaultConfig(t *testing.T) {
	tempDir := t.TempDir()

	err := createDefaultConfig(tempDir)
	require.NoError(t, err)

	// Verify file was created
	configPath := filepath.Join(tempDir, "plonk.yaml")
	assert.FileExists(t, configPath)

	// Read and verify content
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	contentStr := string(content)

	// Check for required sections
	assert.Contains(t, contentStr, "# Plonk Configuration File")
	assert.Contains(t, contentStr, "default_manager:")
	assert.Contains(t, contentStr, "operation_timeout:")
	assert.Contains(t, contentStr, "package_timeout:")
	assert.Contains(t, contentStr, "dotfile_timeout:")
	assert.Contains(t, contentStr, "expand_directories:")
	assert.Contains(t, contentStr, "ignore_patterns:")

	// Check for specific values from defaults
	assert.Contains(t, contentStr, "default_manager: brew")
	assert.Contains(t, contentStr, "- .config") // in expand_directories
	assert.Contains(t, contentStr, `- ".ssh"`)  // in ignore_patterns
	assert.Contains(t, contentStr, `- ".DS_Store"`)
	assert.Contains(t, contentStr, `- "*.swp"`)
}

func TestConfigStruct(t *testing.T) {
	// Test the Config struct fields
	cfg := Config{
		Interactive: true,
		Verbose:     true,
		NoApply:     false,
	}

	assert.True(t, cfg.Interactive)
	assert.True(t, cfg.Verbose)
	assert.False(t, cfg.NoApply)
}
