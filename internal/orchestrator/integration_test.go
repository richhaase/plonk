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

func TestOrchestratorSyncWithResources(t *testing.T) {
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
		DefaultManager: "homebrew",
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
	lockFile := &lock.LockFile{
		Version:  2,
		Packages: make(map[string][]lock.PackageEntry),
	}
	require.NoError(t, lockService.Save(lockFile))

	// Create context
	ctx := context.Background()

	// Test 1: Dotfile reconciliation only (doesn't require external tools)
	t.Run("DotfileReconciliation", func(t *testing.T) {
		result, err := orchestrator.ReconcileDotfiles(ctx, homeDir, configDir)
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
		reconciled, err := orchestrator.ReconcileResource(ctx, dotfileResource)
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

	// Test 3: Test dotfile sync operation
	t.Run("DotfileSync", func(t *testing.T) {
		// Run sync in dry-run mode
		syncResult, err := orchestrator.SyncDotfiles(ctx, configDir, homeDir, cfg, true, false)
		require.NoError(t, err)

		// Verify sync result
		assert.True(t, syncResult.DryRun)
		assert.Equal(t, 1, syncResult.TotalFiles)
		assert.Equal(t, 1, syncResult.Summary.Added)
		assert.Equal(t, 0, syncResult.Summary.Failed)
		assert.Len(t, syncResult.Actions, 1)

		if len(syncResult.Actions) > 0 {
			action := syncResult.Actions[0]
			assert.Equal(t, "would-copy", action.Action)
			assert.Equal(t, "would-add", action.Status)
		}
	})

	// Test 4: Verify lock file v2 structure
	t.Run("LockFileV2Structure", func(t *testing.T) {
		// Load the lock file we created
		loadedLock, err := lockService.Load()
		require.NoError(t, err)

		// Verify v2 structure with empty packages
		assert.Equal(t, 2, loadedLock.Version)
		assert.NotNil(t, loadedLock.Packages)
		assert.Empty(t, loadedLock.Packages, "should have no packages configured")
	})
}
