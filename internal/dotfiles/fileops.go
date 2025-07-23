// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/richhaase/plonk/internal/errors"
)

// FileOperations handles file system operations for dotfiles
type FileOperations struct {
	manager      *Manager
	atomicWriter *AtomicFileWriter
}

// NewFileOperations creates a new file operations handler
func NewFileOperations(manager *Manager) *FileOperations {
	return &FileOperations{
		manager:      manager,
		atomicWriter: NewAtomicFileWriter(),
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
	sourcePath := f.manager.GetSourcePath(source)
	destPath, err := f.manager.GetDestinationPath(destination)
	if err != nil {
		return errors.Wrap(err, errors.ErrPathValidation, errors.DomainDotfiles, "copy",
			"failed to resolve destination path").
			WithItem(destination)
	}

	// Check if source exists
	if !f.manager.FileExists(sourcePath) {
		return errors.DotfileError(errors.ErrFileNotFound, "copy",
			"source file does not exist").
			WithItem(source).
			WithMetadata("source_path", sourcePath)
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return errors.Wrap(err, errors.ErrDirectoryCreate, errors.DomainDotfiles, "copy",
			"failed to create destination directory").
			WithItem(destination).
			WithMetadata("dest_dir", destDir)
	}

	// Handle backup if destination exists
	if f.manager.FileExists(destPath) {
		if options.CreateBackup {
			backupPath := destPath + options.BackupSuffix
			if err := f.createBackup(ctx, destPath, backupPath); err != nil {
				return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy",
					"failed to create backup").
					WithItem(destination).
					WithMetadata("backup_path", backupPath)
			}
		}

		if !options.OverwriteExisting {
			return errors.DotfileError(errors.ErrFilePermission, "copy",
				"destination file exists and overwrite is disabled").
				WithItem(destination).
				WithMetadata("dest_path", destPath)
		}
	}

	// Copy the file
	return f.copyFileContents(ctx, sourcePath, destPath)
}

// CopyDirectory copies a directory recursively from source to destination
func (f *FileOperations) CopyDirectory(ctx context.Context, source, destination string, options CopyOptions) error {
	sourcePath := f.manager.GetSourcePath(source)
	destPath, err := f.manager.GetDestinationPath(destination)
	if err != nil {
		return errors.Wrap(err, errors.ErrPathValidation, errors.DomainDotfiles, "copy-directory",
			"failed to resolve destination path").
			WithItem(destination)
	}

	// Check if source exists and is a directory
	if !f.manager.IsDirectory(sourcePath) {
		return errors.DotfileError(errors.ErrInvalidInput, "copy-directory",
			"source is not a directory").
			WithItem(source).
			WithMetadata("source_path", sourcePath).
			WithSuggestionMessage("Ensure the source path points to a directory, not a file")
	}

	// Walk the source directory and copy each file
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path and destination
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		destFilePath := filepath.Join(destPath, relPath)

		// Create destination directory
		destDir := filepath.Dir(destFilePath)
		if err := os.MkdirAll(destDir, 0750); err != nil {
			return errors.Wrap(err, errors.ErrDirectoryCreate, errors.DomainDotfiles, "copy-directory",
				"failed to create destination directory").
				WithMetadata("dest_dir", destDir).
				WithSuggestionMessage("Check directory permissions and available disk space")
		}

		// Handle backup if destination exists
		if f.manager.FileExists(destFilePath) {
			if options.CreateBackup {
				backupPath := destFilePath + options.BackupSuffix
				if err := f.createBackup(ctx, destFilePath, backupPath); err != nil {
					return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy-directory",
						"failed to create backup").
						WithMetadata("dest_path", destFilePath).
						WithSuggestionMessage("Check file permissions and available disk space")
				}
			}

			if !options.OverwriteExisting {
				return errors.DotfileError(errors.ErrFilePermission, "copy-directory",
					"destination file exists and overwrite is disabled").
					WithMetadata("dest_path", destFilePath).
					WithSuggestionMessage("Use --force to overwrite existing files or enable overwrite in copy options")
			}
		}

		// Copy the file
		return f.copyFileContents(ctx, path, destFilePath)
	})
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
	return f.atomicWriter.CopyFile(ctx, source, destination, 0)
}

// RemoveFile removes a file from the destination
func (f *FileOperations) RemoveFile(destination string) error {
	destPath, err := f.manager.GetDestinationPath(destination)
	if err != nil {
		return errors.Wrap(err, errors.ErrPathValidation, errors.DomainDotfiles, "remove",
			"failed to resolve destination path").
			WithItem(destination)
	}

	if !f.manager.FileExists(destPath) {
		return nil // File doesn't exist, nothing to remove
	}

	return os.Remove(destPath)
}

// FileNeedsUpdate checks if a file needs to be updated based on modification time
func (f *FileOperations) FileNeedsUpdate(ctx context.Context, source, destination string) (bool, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	sourcePath := f.manager.GetSourcePath(source)
	destPath, err := f.manager.GetDestinationPath(destination)
	if err != nil {
		return false, errors.Wrap(err, errors.ErrPathValidation, errors.DomainDotfiles, "check",
			"failed to resolve destination path").
			WithItem(destination)
	}

	// Check if source exists first
	if !f.manager.FileExists(sourcePath) {
		return false, errors.DotfileError(errors.ErrFileNotFound, "check",
			"source file does not exist").
			WithItem(source).
			WithMetadata("source_path", sourcePath)
	}

	// If destination doesn't exist, it needs to be created
	if !f.manager.FileExists(destPath) {
		return true, nil
	}

	// Check for context cancellation before file operations
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	// Compare modification times
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return false, errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "check",
			"failed to stat source file").
			WithMetadata("source_path", sourcePath).
			WithSuggestionMessage("Check if the source file exists and is accessible")
	}

	destInfo, err := os.Stat(destPath)
	if err != nil {
		return false, errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "check",
			"failed to stat destination file").
			WithMetadata("dest_path", destPath).
			WithSuggestionMessage("Check if the destination file exists and is accessible")
	}

	// Source is newer than destination
	return sourceInfo.ModTime().After(destInfo.ModTime()), nil
}

// GetFileInfo returns information about a file
func (f *FileOperations) GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// CopyFileWithAttributes is a simple utility function that copies a file while preserving attributes
// This function creates the destination directory if needed and preserves file permissions and timestamps
func CopyFileWithAttributes(src, dst string) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(dst)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.Wrap(err, errors.ErrDirectoryCreate, errors.DomainDotfiles, "copy", "failed to create destination directory")
	}

	// Get source file info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", "failed to stat source file")
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", "failed to read source file")
	}

	// Write to destination with same permissions
	if err := os.WriteFile(dst, data, srcInfo.Mode()); err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", "failed to write destination file")
	}

	// Preserve timestamps
	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}
