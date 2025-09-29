// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDotfileResource_ID(t *testing.T) {
	resource := &DotfileResource{}
	assert.Equal(t, "dotfiles", resource.ID())
}

func TestDotfileResource_DryRun(t *testing.T) {
	tests := []struct {
		name          string
		dryRun        bool
		expectFileOps bool
		expectBackup  bool
	}{
		{
			name:          "dry-run mode does not modify filesystem",
			dryRun:        true,
			expectFileOps: false,
			expectBackup:  false,
		},
		{
			name:          "normal mode modifies filesystem",
			dryRun:        false,
			expectFileOps: true,
			expectBackup:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directories
			tmpDir := t.TempDir()
			homeDir := filepath.Join(tmpDir, "home")
			configDir := filepath.Join(tmpDir, "config")

			require.NoError(t, os.MkdirAll(homeDir, 0755))
			require.NoError(t, os.MkdirAll(configDir, 0755))

			// Create a source dotfile in config directory
			sourceFile := filepath.Join(configDir, "bashrc")
			sourceContent := []byte("export PATH=/usr/local/bin:$PATH\n")
			require.NoError(t, os.WriteFile(sourceFile, sourceContent, 0644))

			// Create manager and resource with specified dry-run mode
			manager := NewManager(homeDir, configDir)
			resource := NewDotfileResource(manager, tt.dryRun)

			// Create item to apply (use tilde path for destination like the real code does)
			item := resources.Item{
				Name:   ".bashrc",
				Path:   "bashrc",
				Domain: "dotfile",
				State:  resources.StateMissing,
				Metadata: map[string]interface{}{
					"source":      "bashrc",    // Relative to configDir
					"destination": "~/.bashrc", // Tilde notation for home
				},
			}

			// Apply the item
			err := resource.Apply(context.Background(), item)

			if tt.expectFileOps {
				// Normal mode: file should be created
				assert.NoError(t, err)
				destFile := filepath.Join(homeDir, ".bashrc")
				assert.FileExists(t, destFile)

				content, err := os.ReadFile(destFile)
				require.NoError(t, err)
				assert.Equal(t, sourceContent, content)
			} else {
				// Dry-run mode: file should NOT be created
				destFile := filepath.Join(homeDir, ".bashrc")
				assert.NoFileExists(t, destFile, "dry-run should not create destination file")
			}
		})
	}
}

func TestDotfileResource_DryRunPropagation(t *testing.T) {
	// Verify that the dry-run flag is properly stored and accessible
	manager := &Manager{} // Minimal manager for testing

	t.Run("dry-run enabled", func(t *testing.T) {
		resource := NewDotfileResource(manager, true)
		assert.True(t, resource.dryRun, "dry-run flag should be true")
	})

	t.Run("dry-run disabled", func(t *testing.T) {
		resource := NewDotfileResource(manager, false)
		assert.False(t, resource.dryRun, "dry-run flag should be false")
	})
}
