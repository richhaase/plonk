// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompareFiles(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "plonk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir, tmpDir)

	t.Run("identical files return true", func(t *testing.T) {
		content := []byte("test content\n")
		file1 := filepath.Join(tmpDir, "compare1")
		file2 := filepath.Join(tmpDir, "compare2")

		if err := os.WriteFile(file1, content, 0644); err != nil {
			t.Fatalf("Failed to write file1: %v", err)
		}
		if err := os.WriteFile(file2, content, 0644); err != nil {
			t.Fatalf("Failed to write file2: %v", err)
		}

		identical, err := manager.CompareFiles(file1, file2)
		if err != nil {
			t.Fatalf("CompareFiles failed: %v", err)
		}
		if !identical {
			t.Error("CompareFiles returned false for identical files")
		}
	})

	t.Run("different files return false", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "compare3")
		file2 := filepath.Join(tmpDir, "compare4")

		if err := os.WriteFile(file1, []byte("content A\n"), 0644); err != nil {
			t.Fatalf("Failed to write file1: %v", err)
		}
		if err := os.WriteFile(file2, []byte("content B\n"), 0644); err != nil {
			t.Fatalf("Failed to write file2: %v", err)
		}

		identical, err := manager.CompareFiles(file1, file2)
		if err != nil {
			t.Fatalf("CompareFiles failed: %v", err)
		}
		if identical {
			t.Error("CompareFiles returned true for different files")
		}
	})

	t.Run("error when file doesn't exist", func(t *testing.T) {
		existingFile := filepath.Join(tmpDir, "exists")
		nonExistentFile := filepath.Join(tmpDir, "does-not-exist")

		if err := os.WriteFile(existingFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		_, err := manager.CompareFiles(existingFile, nonExistentFile)
		if err == nil {
			t.Error("Expected error when comparing with non-existent file")
		}
	})

	t.Run("same file returns true", func(t *testing.T) {
		file := filepath.Join(tmpDir, "samefile")
		if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		identical, err := manager.CompareFiles(file, file)
		if err != nil {
			t.Fatalf("CompareFiles failed: %v", err)
		}
		if !identical {
			t.Error("CompareFiles returned false for same file")
		}
	})
}
