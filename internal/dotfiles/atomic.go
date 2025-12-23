// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileWriter defines atomic write/copy operations for files
type FileWriter interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
	CopyFile(ctx context.Context, src, dst string, perm os.FileMode) error
}

// AtomicFileWriter handles atomic file operations using temp file + rename pattern
type AtomicFileWriter struct{}

// NewAtomicFileWriter creates a new atomic file writer
func NewAtomicFileWriter() *AtomicFileWriter {
	return &AtomicFileWriter{}
}

// WriteFile atomically writes data to a file
func (a *AtomicFileWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return a.writeFileInternal(filename, func(tmpFile *os.File) error {
		_, err := tmpFile.Write(data)
		return err
	}, perm)
}

// CopyFile atomically copies a file from source to destination
func (a *AtomicFileWriter) CopyFile(ctx context.Context, src, dst string, perm os.FileMode) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	// Get source file info for permissions if perm is 0
	var finalPerm os.FileMode
	if perm == 0 {
		srcInfo, err := srcFile.Stat()
		if err != nil {
			return fmt.Errorf("failed to get source file info %s: %w", src, err)
		}
		finalPerm = srcInfo.Mode()
	} else {
		finalPerm = perm
	}

	// Use atomic write with reader
	return a.writeFileInternal(dst, func(tmpFile *os.File) error {
		// Check for context cancellation before copying
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, err := io.Copy(tmpFile, srcFile)
		return err
	}, finalPerm)
}

// writeFileInternal handles the common atomic write pattern
func (a *AtomicFileWriter) writeFileInternal(filename string, writeFunc func(*os.File) error, perm os.FileMode) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(filename)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	// Create temporary file in same directory as destination
	tmpFile, err := os.CreateTemp(destDir, ".tmp-"+filepath.Base(filename))
	if err != nil {
		return fmt.Errorf("failed to create temporary file in %s for %s: %w", destDir, filename, err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on failure
	defer func() {
		if tmpFile != nil {
			_ = tmpFile.Close()
		}
		_ = os.Remove(tmpPath)
	}()

	// Write data using provided function
	if err := writeFunc(tmpFile); err != nil {
		return fmt.Errorf("failed to write file contents for %s (tmp: %s): %w", filename, tmpPath, err)
	}

	// Sync to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file %s for %s: %w", tmpPath, filename, err)
	}

	// Close temporary file
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file %s for %s: %w", tmpPath, filename, err)
	}
	tmpFile = nil // Mark as closed for defer cleanup

	// Set permissions
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set file permissions %v on %s for %s: %w", perm, tmpPath, filename, err)
	}

	// Atomic rename - this is the critical atomic operation
	if err := os.Rename(tmpPath, filename); err != nil {
		return fmt.Errorf("failed to rename temporary file %s to %s: %w", tmpPath, filename, err)
	}

	return nil
}
