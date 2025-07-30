// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeFileHash(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "plonk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir, tmpDir)

	t.Run("identical files have same hash", func(t *testing.T) {
		// Create two identical files
		content := []byte("test content\n")
		file1 := filepath.Join(tmpDir, "file1")
		file2 := filepath.Join(tmpDir, "file2")

		if err := os.WriteFile(file1, content, 0644); err != nil {
			t.Fatalf("Failed to write file1: %v", err)
		}
		if err := os.WriteFile(file2, content, 0644); err != nil {
			t.Fatalf("Failed to write file2: %v", err)
		}

		hash1, err := manager.computeFileHash(file1)
		if err != nil {
			t.Fatalf("Failed to compute hash for file1: %v", err)
		}

		hash2, err := manager.computeFileHash(file2)
		if err != nil {
			t.Fatalf("Failed to compute hash for file2: %v", err)
		}

		if hash1 != hash2 {
			t.Errorf("Identical files have different hashes: %s != %s", hash1, hash2)
		}
	})

	t.Run("different files have different hashes", func(t *testing.T) {
		// Create two different files
		file1 := filepath.Join(tmpDir, "file3")
		file2 := filepath.Join(tmpDir, "file4")

		if err := os.WriteFile(file1, []byte("content 1\n"), 0644); err != nil {
			t.Fatalf("Failed to write file1: %v", err)
		}
		if err := os.WriteFile(file2, []byte("content 2\n"), 0644); err != nil {
			t.Fatalf("Failed to write file2: %v", err)
		}

		hash1, err := manager.computeFileHash(file1)
		if err != nil {
			t.Fatalf("Failed to compute hash for file1: %v", err)
		}

		hash2, err := manager.computeFileHash(file2)
		if err != nil {
			t.Fatalf("Failed to compute hash for file2: %v", err)
		}

		if hash1 == hash2 {
			t.Errorf("Different files have same hash: %s", hash1)
		}
	})

	t.Run("error on non-existent file", func(t *testing.T) {
		nonExistent := filepath.Join(tmpDir, "does-not-exist")
		_, err := manager.computeFileHash(nonExistent)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("empty file has consistent hash", func(t *testing.T) {
		emptyFile := filepath.Join(tmpDir, "empty")
		if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		hash, err := manager.computeFileHash(emptyFile)
		if err != nil {
			t.Fatalf("Failed to compute hash for empty file: %v", err)
		}

		// SHA256 hash of empty string
		expectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		if hash != expectedHash {
			t.Errorf("Empty file hash mismatch: got %s, want %s", hash, expectedHash)
		}
	})
}

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

func TestCreateCompareFunc(t *testing.T) {
	// Create temporary directories
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

	manager := NewManager(homeDir, configDir)

	t.Run("compare func detects identical files", func(t *testing.T) {
		// Create source file in config
		source := "vimrc"
		sourcePath := filepath.Join(configDir, source)
		content := []byte("vim settings\n")
		if err := os.WriteFile(sourcePath, content, 0644); err != nil {
			t.Fatalf("Failed to write source: %v", err)
		}

		// Create destination file in home
		destination := "~/.vimrc"
		destPath := filepath.Join(homeDir, ".vimrc")
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			t.Fatalf("Failed to write destination: %v", err)
		}

		// Create and test compare function
		compareFn := manager.createCompareFunc(source, destination)
		identical, err := compareFn()
		if err != nil {
			t.Fatalf("Compare function failed: %v", err)
		}
		if !identical {
			t.Error("Compare function returned false for identical files")
		}
	})

	t.Run("compare func detects different files", func(t *testing.T) {
		// Create source file in config
		source := "zshrc"
		sourcePath := filepath.Join(configDir, source)
		if err := os.WriteFile(sourcePath, []byte("source content\n"), 0644); err != nil {
			t.Fatalf("Failed to write source: %v", err)
		}

		// Create different destination file in home
		destination := "~/.zshrc"
		destPath := filepath.Join(homeDir, ".zshrc")
		if err := os.WriteFile(destPath, []byte("different content\n"), 0644); err != nil {
			t.Fatalf("Failed to write destination: %v", err)
		}

		// Create and test compare function
		compareFn := manager.createCompareFunc(source, destination)
		identical, err := compareFn()
		if err != nil {
			t.Fatalf("Compare function failed: %v", err)
		}
		if identical {
			t.Error("Compare function returned true for different files")
		}
	})

	t.Run("compare func returns false when destination missing", func(t *testing.T) {
		// Create source file in config
		source := "gitconfig"
		sourcePath := filepath.Join(configDir, source)
		if err := os.WriteFile(sourcePath, []byte("git config\n"), 0644); err != nil {
			t.Fatalf("Failed to write source: %v", err)
		}

		// Destination doesn't exist
		destination := "~/.gitconfig"

		// Create and test compare function
		compareFn := manager.createCompareFunc(source, destination)
		identical, err := compareFn()
		if err != nil {
			t.Fatalf("Compare function failed: %v", err)
		}
		if identical {
			t.Error("Compare function returned true when destination missing")
		}
	})
}
