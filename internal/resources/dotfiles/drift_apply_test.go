// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/resources"
)

func TestApplyDriftedFiles(t *testing.T) {
	// Create temp directories
	homeDir, err := os.MkdirTemp("", "plonk-home-*")
	if err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	defer os.RemoveAll(homeDir)

	configDir, err := os.MkdirTemp("", "plonk-config-*")
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	defer os.RemoveAll(configDir)

	ctx := context.Background()

	t.Run("apply restores drifted file", func(t *testing.T) {
		// Create source file
		sourceFile := filepath.Join(configDir, "bashrc")
		originalContent := []byte("export PATH=/usr/local/bin:$PATH\n")
		if err := os.WriteFile(sourceFile, originalContent, 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Create drifted destination file
		destFile := filepath.Join(homeDir, ".bashrc")
		driftedContent := []byte("export PATH=/opt/bin:$PATH\n# Modified by user\n")
		if err := os.WriteFile(destFile, driftedContent, 0644); err != nil {
			t.Fatalf("Failed to create dest file: %v", err)
		}

		// Create manager
		manager := NewManager(homeDir, configDir)

		// Get desired and actual states
		desired, err := manager.GetConfiguredDotfiles()
		if err != nil {
			t.Fatalf("Failed to get configured dotfiles: %v", err)
		}

		actual, err := manager.GetActualDotfiles(ctx)
		if err != nil {
			t.Fatalf("Failed to get actual dotfiles: %v", err)
		}

		// Reconcile to detect drift
		reconciled := resources.ReconcileItems(desired, actual)

		// Find drifted item
		var driftedItem *resources.Item
		for i := range reconciled {
			if reconciled[i].Name == ".bashrc" && reconciled[i].State == resources.StateDegraded {
				driftedItem = &reconciled[i]
				break
			}
		}

		if driftedItem == nil {
			t.Fatal("Drifted bashrc not found")
		}

		// Apply the drifted item using the helper function
		opts := ApplyOptions{
			DryRun: false,
			Backup: true,
		}
		err = applyDotfileItem(ctx, manager, *driftedItem, opts)
		if err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		// Verify file was restored to original content
		restoredContent, err := os.ReadFile(destFile)
		if err != nil {
			t.Fatalf("Failed to read restored file: %v", err)
		}

		if string(restoredContent) != string(originalContent) {
			t.Errorf("File not restored correctly.\nExpected: %s\nGot: %s",
				string(originalContent), string(restoredContent))
		}

		// Verify backup was created
		backupFiles, err := filepath.Glob(destFile + ".backup.*")
		if err != nil {
			t.Fatalf("Failed to check for backup files: %v", err)
		}
		if len(backupFiles) == 0 {
			t.Error("No backup file created")
		}
	})
}
