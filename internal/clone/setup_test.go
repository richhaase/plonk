// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package clone

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetManagerDescription(t *testing.T) {
	customCfg := &config.Config{
		Managers: map[string]config.ManagerConfig{
			"custom": {
				Description: "Custom Manager",
			},
		},
	}

	tests := []struct {
		name     string
		cfg      *config.Config
		manager  string
		expected string
	}{
		{
			name:     "homebrew",
			cfg:      nil,
			manager:  "homebrew",
			expected: "homebrew package manager",
		},
		{
			name:     "brew alias",
			cfg:      nil,
			manager:  "brew",
			expected: "Homebrew (macOS/Linux package manager)",
		},
		{
			name:     "cargo",
			cfg:      nil,
			manager:  "cargo",
			expected: "Cargo (Rust package manager)",
		},
		{
			name:     "npm",
			cfg:      nil,
			manager:  "npm",
			expected: "npm (Node.js package manager)",
		},
		{
			name:     "uv",
			cfg:      nil,
			manager:  "uv",
			expected: "uv (Python package manager)",
		},
		{
			name:     "gem",
			cfg:      nil,
			manager:  "gem",
			expected: "gem (Ruby package manager)",
		},
		{
			name:     "unknown",
			cfg:      nil,
			manager:  "unknown-manager",
			expected: "unknown-manager package manager",
		},
		{
			name:     "custom description from config",
			cfg:      customCfg,
			manager:  "custom",
			expected: "Custom Manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getManagerDescription(tt.cfg, tt.manager)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetManualInstallInstructions(t *testing.T) {
	customCfg := &config.Config{
		Managers: map[string]config.ManagerConfig{
			"custom": {
				InstallHint: "Install custom via custom-installer",
			},
		},
	}

	tests := []struct {
		name     string
		cfg      *config.Config
		manager  string
		expected string
	}{
		{
			name:     "homebrew",
			cfg:      nil,
			manager:  "homebrew",
			expected: "See official documentation for installation instructions",
		},
		{
			name:     "brew alias",
			cfg:      nil,
			manager:  "brew",
			expected: "Visit https://brew.sh for installation instructions (prerequisite)",
		},
		{
			name:     "cargo",
			cfg:      nil,
			manager:  "cargo",
			expected: "Install Rust from https://rustup.rs/",
		},
		{
			name:     "npm",
			cfg:      nil,
			manager:  "npm",
			expected: "Install Node.js from https://nodejs.org/ or use brew install node",
		},
		{
			name:     "uv",
			cfg:      nil,
			manager:  "uv",
			expected: "Install UV from https://docs.astral.sh/uv/ or use brew install uv",
		},
		{
			name:     "gem",
			cfg:      nil,
			manager:  "gem",
			expected: "Install Ruby from https://ruby-lang.org/ or use brew install ruby",
		},
		{
			name:     "unknown",
			cfg:      nil,
			manager:  "unknown-manager",
			expected: "See official documentation for installation instructions",
		},
		{
			name:     "custom hint from config",
			cfg:      customCfg,
			manager:  "custom",
			expected: "Install custom via custom-installer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getManualInstallInstructions(tt.cfg, tt.manager)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectRequiredManagers(t *testing.T) {
	t.Run("lock file with metadata", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")

		// Create a lock file with manager in metadata
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

	t.Run("lock file with ID prefix fallback", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "plonk.lock")

		// Create a lock file where manager is extracted from ID
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
					ID:   "uv:requests",
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

		// Should have cargo and uv (unique)
		assert.Len(t, managers, 2)
		assert.Contains(t, managers, "cargo")
		assert.Contains(t, managers, "uv")
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

// TestFindMissingPackageManagers test removed - functionality replaced by SelfInstall interface

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
	}

	assert.True(t, cfg.Interactive)
	assert.True(t, cfg.Verbose)
}
