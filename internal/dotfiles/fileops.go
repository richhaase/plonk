// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileOperations handles file system operations for dotfiles
type FileOperations struct {
	manager *Manager
}

// NewFileOperations creates a new file operations handler
func NewFileOperations(manager *Manager) *FileOperations {
	return &FileOperations{
		manager: manager,
	}
}

// CopyOptions configures file copy operations
type CopyOptions struct {
	CreateBackup bool
	BackupSuffix string
	OverwriteExisting bool
}

// DefaultCopyOptions returns default copy options
func DefaultCopyOptions() CopyOptions {
	return CopyOptions{
		CreateBackup: true,
		BackupSuffix: ".backup",
		OverwriteExisting: true,
	}
}

// CopyFile copies a file from source to destination with options
func (f *FileOperations) CopyFile(source, destination string, options CopyOptions) error {
	sourcePath := f.manager.GetSourcePath(source)
	destPath := f.manager.GetDestinationPath(destination)
	
	// Check if source exists
	if !f.manager.FileExists(sourcePath) {
		return fmt.Errorf("source file does not exist: %s", sourcePath)
	}
	
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Handle backup if destination exists
	if f.manager.FileExists(destPath) {
		if options.CreateBackup {
			backupPath := destPath + options.BackupSuffix
			if err := f.createBackup(destPath, backupPath); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
		}
		
		if !options.OverwriteExisting {
			return fmt.Errorf("destination file exists and overwrite is disabled: %s", destPath)
		}
	}
	
	// Copy the file
	return f.copyFileContents(sourcePath, destPath)
}

// CopyDirectory copies a directory recursively from source to destination
func (f *FileOperations) CopyDirectory(source, destination string, options CopyOptions) error {
	sourcePath := f.manager.GetSourcePath(source)
	destPath := f.manager.GetDestinationPath(destination)
	
	// Check if source exists and is a directory
	if !f.manager.IsDirectory(sourcePath) {
		return fmt.Errorf("source is not a directory: %s", sourcePath)
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
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destDir, err)
		}
		
		// Handle backup if destination exists
		if f.manager.FileExists(destFilePath) {
			if options.CreateBackup {
				backupPath := destFilePath + options.BackupSuffix
				if err := f.createBackup(destFilePath, backupPath); err != nil {
					return fmt.Errorf("failed to create backup for %s: %w", destFilePath, err)
				}
			}
			
			if !options.OverwriteExisting {
				return fmt.Errorf("destination file exists and overwrite is disabled: %s", destFilePath)
			}
		}
		
		// Copy the file
		return f.copyFileContents(path, destFilePath)
	})
}

// createBackup creates a backup of the file with timestamp
func (f *FileOperations) createBackup(source, backupPath string) error {
	// Add timestamp to backup path
	timestamp := time.Now().Format("20060102-150405")
	backupPath = backupPath + "." + timestamp
	
	return f.copyFileContents(source, backupPath)
}

// copyFileContents copies the contents of one file to another
func (f *FileOperations) copyFileContents(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()
	
	// Copy file contents
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}
	
	// Copy file permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}
	
	if err := destFile.Chmod(sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	
	return nil
}

// RemoveFile removes a file from the destination
func (f *FileOperations) RemoveFile(destination string) error {
	destPath := f.manager.GetDestinationPath(destination)
	
	if !f.manager.FileExists(destPath) {
		return nil // File doesn't exist, nothing to remove
	}
	
	return os.Remove(destPath)
}

// FileNeedsUpdate checks if a file needs to be updated based on modification time
func (f *FileOperations) FileNeedsUpdate(source, destination string) (bool, error) {
	sourcePath := f.manager.GetSourcePath(source)
	destPath := f.manager.GetDestinationPath(destination)
	
	// Check if source exists first
	if !f.manager.FileExists(sourcePath) {
		return false, fmt.Errorf("source file does not exist: %s", sourcePath)
	}
	
	// If destination doesn't exist, it needs to be created
	if !f.manager.FileExists(destPath) {
		return true, nil
	}
	
	// Compare modification times
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return false, fmt.Errorf("failed to stat source file: %w", err)
	}
	
	destInfo, err := os.Stat(destPath)
	if err != nil {
		return false, fmt.Errorf("failed to stat destination file: %w", err)
	}
	
	// Source is newer than destination
	return sourceInfo.ModTime().After(destInfo.ModTime()), nil
}

// GetFileInfo returns information about a file
func (f *FileOperations) GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}