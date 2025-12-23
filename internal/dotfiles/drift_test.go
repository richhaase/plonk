// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestDriftDetectionIntegration(t *testing.T) {
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
	cfg := config.LoadWithDefaults(configDir)

	t.Run("full drift detection flow", func(t *testing.T) {
		// Step 1: Create source file in config
		sourceFile := filepath.Join(configDir, "vimrc")
		originalContent := []byte("set number\nset autoindent\n")
		if err := os.WriteFile(sourceFile, originalContent, 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Step 2: Deploy the file (simulate apply)
		destFile := filepath.Join(homeDir, ".vimrc")
		if err := os.WriteFile(destFile, originalContent, 0644); err != nil {
			t.Fatalf("Failed to create dest file: %v", err)
		}

		// Step 3: Reconcile - should show as managed
		result, err := ReconcileWithConfig(ctx, homeDir, configDir, cfg)
		if err != nil {
			t.Fatalf("First reconcile failed: %v", err)
		}

		// Find the vimrc item
		var vimrcItem *DotfileItem
		for i := range result.Managed {
			if result.Managed[i].Name == ".vimrc" {
				vimrcItem = &result.Managed[i]
				break
			}
		}

		if vimrcItem == nil {
			t.Fatal("vimrc not found in managed items")
		}

		if vimrcItem.State != StateManaged {
			t.Errorf("Expected vimrc to be StateManaged, got %v", vimrcItem.State)
		}

		// Step 4: Modify the deployed file (drift occurs)
		modifiedContent := []byte("set nonumber\nset autoindent\nset ruler\n")
		if err := os.WriteFile(destFile, modifiedContent, 0644); err != nil {
			t.Fatalf("Failed to modify dest file: %v", err)
		}

		// Step 5: Reconcile again - should show as drifted
		result2, err := ReconcileWithConfig(ctx, homeDir, configDir, cfg)
		if err != nil {
			t.Fatalf("Second reconcile failed: %v", err)
		}

		// Find the vimrc item again
		var driftedItem *DotfileItem
		for i := range result2.Managed {
			if result2.Managed[i].Name == ".vimrc" {
				driftedItem = &result2.Managed[i]
				break
			}
		}

		if driftedItem == nil {
			t.Fatal("vimrc not found in managed items after drift")
		}

		if driftedItem.State != StateDegraded {
			t.Errorf("Expected vimrc to be StateDegraded (drifted), got %v", driftedItem.State)
		}

		// Check drift metadata
		if driftedItem.Metadata == nil || driftedItem.Metadata["drift_status"] != "modified" {
			t.Errorf("Expected drift_status=modified in metadata, got %v", driftedItem.Metadata)
		}
	})

	t.Run("missing file not detected as drift", func(t *testing.T) {
		// Create source file
		sourceFile := filepath.Join(configDir, "gitconfig")
		content := []byte("[user]\n  name = Test\n")
		if err := os.WriteFile(sourceFile, content, 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Don't create destination file (missing)

		// Reconcile
		result, err := ReconcileWithConfig(ctx, homeDir, configDir, cfg)
		if err != nil {
			t.Fatalf("Reconcile failed: %v", err)
		}

		// Should be in missing, not managed
		if len(result.Missing) == 0 {
			t.Error("Expected gitconfig to be in missing items")
		}

		// Should not be in managed
		for _, item := range result.Managed {
			if item.Name == ".gitconfig" {
				t.Error("gitconfig should not be in managed items when missing")
			}
		}
	})
}
