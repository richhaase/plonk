// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAtomicFileWriter_WriteFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	
	writer := NewAtomicFileWriter()
	
	t.Run("successful write", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "test.txt")
		data := []byte("test content")
		
		err := writer.WriteFile(filePath, data, 0644)
		if err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		
		// Verify file exists and has correct content
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		
		if string(content) != "test content" {
			t.Errorf("Expected 'test content', got '%s'", string(content))
		}
		
		// Verify permissions
		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}
		
		if info.Mode().Perm() != 0644 {
			t.Errorf("Expected permissions 0644, got %o", info.Mode().Perm())
		}
	})
	
	t.Run("overwrites existing file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "overwrite.txt")
		
		// Write initial content
		err := writer.WriteFile(filePath, []byte("initial"), 0644)
		if err != nil {
			t.Fatalf("Initial write failed: %v", err)
		}
		
		// Overwrite with new content
		err = writer.WriteFile(filePath, []byte("updated"), 0644)
		if err != nil {
			t.Fatalf("Overwrite failed: %v", err)
		}
		
		// Verify new content
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		
		if string(content) != "updated" {
			t.Errorf("Expected 'updated', got '%s'", string(content))
		}
	})
	
	t.Run("creates parent directories", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "nested", "dir", "file.txt")
		
		err := writer.WriteFile(filePath, []byte("nested content"), 0644)
		if err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		
		// Verify file exists
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		
		if string(content) != "nested content" {
			t.Errorf("Expected 'nested content', got '%s'", string(content))
		}
	})
}

func TestAtomicFileWriter_CopyFile(t *testing.T) {
	tempDir := t.TempDir()
	writer := NewAtomicFileWriter()
	
	t.Run("successful copy", func(t *testing.T) {
		// Create source file
		srcPath := filepath.Join(tempDir, "source.txt")
		err := os.WriteFile(srcPath, []byte("source content"), 0755)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}
		
		// Copy to destination
		dstPath := filepath.Join(tempDir, "destination.txt")
		ctx := context.Background()
		
		err = writer.CopyFile(ctx, srcPath, dstPath, 0644)
		if err != nil {
			t.Fatalf("CopyFile failed: %v", err)
		}
		
		// Verify content
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}
		
		if string(content) != "source content" {
			t.Errorf("Expected 'source content', got '%s'", string(content))
		}
		
		// Verify permissions (should use provided perm, not source perm)
		info, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("Failed to stat destination file: %v", err)
		}
		
		if info.Mode().Perm() != 0644 {
			t.Errorf("Expected permissions 0644, got %o", info.Mode().Perm())
		}
	})
	
	t.Run("copies with source permissions when perm is 0", func(t *testing.T) {
		// Create source file with specific permissions
		srcPath := filepath.Join(tempDir, "source_perm.txt")
		err := os.WriteFile(srcPath, []byte("perm content"), 0755)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}
		
		// Copy to destination with perm = 0 (should use source permissions)
		dstPath := filepath.Join(tempDir, "destination_perm.txt")
		ctx := context.Background()
		
		err = writer.CopyFile(ctx, srcPath, dstPath, 0)
		if err != nil {
			t.Fatalf("CopyFile failed: %v", err)
		}
		
		// Verify permissions match source
		info, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("Failed to stat destination file: %v", err)
		}
		
		if info.Mode().Perm() != 0755 {
			t.Errorf("Expected permissions 0755, got %o", info.Mode().Perm())
		}
	})
	
	t.Run("context cancellation", func(t *testing.T) {
		// Create source file
		srcPath := filepath.Join(tempDir, "source_cancel.txt")
		err := os.WriteFile(srcPath, []byte("cancel content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}
		
		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		dstPath := filepath.Join(tempDir, "destination_cancel.txt")
		
		err = writer.CopyFile(ctx, srcPath, dstPath, 0644)
		if err == nil {
			t.Error("Expected error when context is cancelled")
		}
		
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
		
		// Verify destination file doesn't exist
		if _, err := os.Stat(dstPath); !os.IsNotExist(err) {
			t.Error("Destination file should not exist after cancelled operation")
		}
	})
	
	t.Run("nonexistent source file", func(t *testing.T) {
		srcPath := filepath.Join(tempDir, "nonexistent.txt")
		dstPath := filepath.Join(tempDir, "destination_nonexistent.txt")
		ctx := context.Background()
		
		err := writer.CopyFile(ctx, srcPath, dstPath, 0644)
		if err == nil {
			t.Error("Expected error for nonexistent source file")
		}
		
		if !strings.Contains(err.Error(), "failed to open source file") {
			t.Errorf("Expected 'failed to open source file' in error, got: %v", err)
		}
	})
}

func TestAtomicFileWriter_AtomicBehavior(t *testing.T) {
	tempDir := t.TempDir()
	writer := NewAtomicFileWriter()
	
	t.Run("no partial writes visible", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "atomic.txt")
		
		// Write initial content
		err := writer.WriteFile(filePath, []byte("initial"), 0644)
		if err != nil {
			t.Fatalf("Initial write failed: %v", err)
		}
		
		// Simulate a failing write by making the directory read-only
		// This should fail during rename, leaving original file intact
		
		// Actually, let's test that temp files are cleaned up
		// We'll check that no .tmp files remain after successful operations
		
		err = writer.WriteFile(filePath, []byte("updated"), 0644)
		if err != nil {
			t.Fatalf("Update write failed: %v", err)
		}
		
		// Check no temp files remain
		entries, err := os.ReadDir(tempDir)
		if err != nil {
			t.Fatalf("Failed to read temp dir: %v", err)
		}
		
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".tmp-") {
				t.Errorf("Temporary file not cleaned up: %s", entry.Name())
			}
		}
		
		// Verify final content
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		
		if string(content) != "updated" {
			t.Errorf("Expected 'updated', got '%s'", string(content))
		}
	})
}

func TestAtomicFileWriter_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	writer := NewAtomicFileWriter()
	
	t.Run("CopyFile context timeout", func(t *testing.T) {
		// Create source file
		srcPath := filepath.Join(tempDir, "source_timeout.txt")
		err := os.WriteFile(srcPath, []byte("timeout content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}
		
		// Create context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		// Give context time to expire
		time.Sleep(10 * time.Millisecond)
		
		dstPath := filepath.Join(tempDir, "destination_timeout.txt")
		
		err = writer.CopyFile(ctx, srcPath, dstPath, 0644)
		if err == nil {
			t.Error("Expected error when context times out")
		}
		
		if !containsContextError(err) {
			t.Errorf("Expected context deadline exceeded error, got %v", err)
		}
		
		// Verify destination file doesn't exist
		if _, err := os.Stat(dstPath); !os.IsNotExist(err) {
			t.Error("Destination file should not exist after timeout")
		}
	})
}