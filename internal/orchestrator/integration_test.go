// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestOrchestratorApplyWithResources(t *testing.T) {
	// This test focuses on the Resource abstraction and doesn't require actual package managers

	// Test setup with temp directory
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	configDir := filepath.Join(tempDir, "config")

	// Create directories
	require.NoError(t, os.MkdirAll(homeDir, 0755))
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, ".config", "plonk"), 0755))

	// Create a test config file
	cfg := &config.Config{
		DefaultManager: "brew",
		IgnorePatterns: []string{"*.tmp", "*.swp"},
	}

	// Write config file manually since there's no SaveConfig
	configPath := filepath.Join(configDir, "plonk.yaml")
	configBytes, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, configBytes, 0644))

	// Configure 1 dotfile (e.g., .testrc)
	dotfileSrc := filepath.Join(configDir, ".config", "plonk", ".testrc")
	dotfileContent := "# Test RC file\nexport TEST=1\n"
	require.NoError(t, os.WriteFile(dotfileSrc, []byte(dotfileContent), 0644))

	// Create an empty lock file v2 (no packages to avoid external dependencies)
	lockService := lock.NewYAMLLockService(configDir)
	lockFile := &lock.Lock{
		Version:   2,
		Resources: []lock.ResourceEntry{},
	}
	require.NoError(t, lockService.Write(lockFile))

	// Create context
	ctx := context.Background()

	// Test 1: Dotfile reconciliation only (doesn't require external tools)
	t.Run("DotfileReconciliation", func(t *testing.T) {
		result, err := dotfiles.Reconcile(ctx, homeDir, configDir)
		require.NoError(t, err)

		// Check dotfile reconciliation
		assert.Equal(t, "dotfile", result.Domain)
		assert.Len(t, result.Missing, 1, "should have 1 missing dotfile")
		// The name includes the relative path from config dir
		assert.Contains(t, result.Missing[0].Name, ".testrc")
		assert.Empty(t, result.Managed, "no dotfiles should be managed yet")
		assert.Empty(t, result.Untracked, "no untracked dotfiles")
	})

	// Test 2: Test Resource-based reconciliation directly
	t.Run("ResourceBasedReconciliation", func(t *testing.T) {
		// Create dotfile resource
		manager := dotfiles.NewManager(homeDir, configDir)
		dotfileResource := dotfiles.NewDotfileResource(manager)

		// Set desired state
		configured, err := manager.GetConfiguredDotfiles()
		require.NoError(t, err)
		dotfileResource.SetDesired(configured)

		// Reconcile
		reconciled, err := resources.ReconcileResource(ctx, dotfileResource)
		require.NoError(t, err)

		// Check results
		assert.NotEmpty(t, reconciled)
		foundMissing := false
		for _, item := range reconciled {
			if item.State == resources.StateMissing {
				foundMissing = true
				assert.Contains(t, item.Name, ".testrc")
			}
		}
		assert.True(t, foundMissing, "should find missing .testrc")
	})

	// Test 3: Test dotfile apply operation
	t.Run("DotfileApply", func(t *testing.T) {
		// Run apply in dry-run mode
		applyResult, err := orchestrator.ApplyDotfiles(ctx, configDir, homeDir, cfg, true)
		require.NoError(t, err)

		// Verify apply result
		assert.True(t, applyResult.DryRun)
		assert.Equal(t, 1, applyResult.TotalFiles)
		assert.Equal(t, 1, applyResult.Summary.Added)
		assert.Equal(t, 0, applyResult.Summary.Failed)
		assert.Len(t, applyResult.Actions, 1)

		if len(applyResult.Actions) > 0 {
			action := applyResult.Actions[0]
			assert.Equal(t, "would-copy", action.Action)
			assert.Equal(t, "would-add", action.Status)
		}
	})

	// Test 4: Verify lock file v2 structure
	t.Run("LockFileV2Structure", func(t *testing.T) {
		// Load the lock file we created
		loadedLock, err := lockService.Read()
		require.NoError(t, err)

		// Verify v2 structure with empty resources
		assert.Equal(t, 2, loadedLock.Version)
		assert.NotNil(t, loadedLock.Resources)
		assert.Empty(t, loadedLock.Resources, "should have no resources configured")
	})
}
