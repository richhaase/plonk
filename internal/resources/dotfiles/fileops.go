// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileOperations handles file system operations for dotfiles
type FileOperations struct {
	pathResolver PathResolver
	writer       FileWriter
}

// NewFileOperations creates a new file operations handler
func NewFileOperations(pathResolver PathResolver) *FileOperations {
	return &FileOperations{
		pathResolver: pathResolver,
		writer:       NewAtomicFileWriter(),
	}
}

// NewFileOperationsWithWriter allows injecting a custom FileWriter (for testing)
func NewFileOperationsWithWriter(pathResolver PathResolver, writer FileWriter) *FileOperations {
	if writer == nil {
		writer = NewAtomicFileWriter()
	}
	return &FileOperations{
		pathResolver: pathResolver,
		writer:       writer,
	}
}

// CopyOptions configures file copy operations
type CopyOptions struct {
	CreateBackup      bool
	BackupSuffix      string
	OverwriteExisting bool
}

// DefaultCopyOptions returns default copy options
func DefaultCopyOptions() CopyOptions {
	return CopyOptions{
		CreateBackup:      true,
		BackupSuffix:      ".backup",
		OverwriteExisting: true,
	}
}

// CopyFile copies a file from source to destination with options
func (f *FileOperations) CopyFile(ctx context.Context, source, destination string, options CopyOptions) error {
	sourcePath := f.pathResolver.GetSourcePath(source)
	destPath, err := f.pathResolver.GetDestinationPath(destination)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path %s: %w", destination, err)
	}

	// Check if source exists
	if !f.fileExists(sourcePath) {
		return fmt.Errorf("source file %s does not exist at %s", source, sourcePath)
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory %s for %s: %w", destDir, destination, err)
	}

	// Handle backup if destination exists
	if f.fileExists(destPath) {
		if options.CreateBackup {
			backupPath := destPath + options.BackupSuffix
			if err := f.createBackup(ctx, destPath, backupPath); err != nil {
				return fmt.Errorf("failed to create backup %s for %s: %w", backupPath, destination, err)
			}
		}

		if !options.OverwriteExisting {
			return fmt.Errorf("destination file %s exists at %s and overwrite is disabled", destination, destPath)
		}
	}

	// Copy the file
	return f.copyFileContents(ctx, sourcePath, destPath)
}

// createBackup creates a backup of the file with timestamp
func (f *FileOperations) createBackup(ctx context.Context, source, backupPath string) error {
	// Add timestamp to backup path
	timestamp := time.Now().Format("20060102-150405")
	backupPath = backupPath + "." + timestamp

	return f.copyFileContents(ctx, source, backupPath)
}

// copyFileContents copies the contents of one file to another atomically
func (f *FileOperations) copyFileContents(ctx context.Context, source, destination string) error {
	// Use atomic file writer to copy file with proper permissions
	return f.writer.CopyFile(ctx, source, destination, 0)
}

// CopyFileWithAttributes is a simple utility function that copies a file while preserving attributes
// This function creates the destination directory if needed and preserves file permissions and timestamps
func CopyFileWithAttributes(src, dst string) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(dst)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	// Get source file info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", src, err)
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", src, err)
	}

	// Write to destination with same permissions
	if err := os.WriteFile(dst, data, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dst, err)
	}

	// Preserve timestamps
	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}

// fileExists checks if a file exists at the given path
func (f *FileOperations) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
