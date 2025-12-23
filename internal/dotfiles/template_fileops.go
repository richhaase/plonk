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

// TemplateFileOperations wraps FileOperations to handle template rendering during copy
type TemplateFileOperations struct {
	baseOps           *FileOperations
	templateProcessor TemplateProcessor
	pathResolver      PathResolver
}

// NewTemplateFileOperations creates a new template-aware file operations handler
func NewTemplateFileOperations(pathResolver PathResolver, templateProc TemplateProcessor) *TemplateFileOperations {
	return &TemplateFileOperations{
		baseOps:           NewFileOperations(pathResolver),
		templateProcessor: templateProc,
		pathResolver:      pathResolver,
	}
}

// CopyFile copies a file from source to destination, rendering templates as needed
// For template files (.tmpl), it renders the content and writes the result
// For regular files, it delegates to the base FileOperations
func (tfo *TemplateFileOperations) CopyFile(ctx context.Context, source, destination string, options CopyOptions) error {
	sourcePath := tfo.pathResolver.GetSourcePath(source)

	// Check if source is a template
	if tfo.templateProcessor.IsTemplate(sourcePath) {
		return tfo.copyTemplateFile(ctx, source, destination, options)
	}

	// Not a template, use standard copy
	return tfo.baseOps.CopyFile(ctx, source, destination, options)
}

// copyTemplateFile renders a template and writes the result to the destination
func (tfo *TemplateFileOperations) copyTemplateFile(ctx context.Context, source, destination string, options CopyOptions) error {
	sourcePath := tfo.pathResolver.GetSourcePath(source)
	destPath, err := tfo.pathResolver.GetDestinationPath(destination)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path %s: %w", destination, err)
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check if source template exists
	srcInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template file %s does not exist at %s", source, sourcePath)
		}
		return fmt.Errorf("failed to stat template file %s: %w", sourcePath, err)
	}

	// Render the template
	rendered, err := tfo.templateProcessor.RenderToBytes(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", source, err)
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory %s for %s: %w", destDir, destination, err)
	}

	// Handle backup if destination exists
	if tfo.fileExists(destPath) {
		if options.CreateBackup {
			if err := tfo.createBackup(ctx, destPath, options.BackupSuffix); err != nil {
				return fmt.Errorf("failed to create backup for %s: %w", destination, err)
			}
		}

		if !options.OverwriteExisting {
			return fmt.Errorf("destination file %s exists at %s and overwrite is disabled", destination, destPath)
		}
	}

	// Write rendered content atomically
	writer := NewAtomicFileWriter()
	if err := writer.WriteFile(destPath, rendered, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to write rendered template to %s: %w", destPath, err)
	}

	return nil
}

// fileExists checks if a file exists at the given path
func (tfo *TemplateFileOperations) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// createBackup creates a timestamped backup of an existing file
func (tfo *TemplateFileOperations) createBackup(ctx context.Context, filePath, suffix string) error {
	// Delegate to base operations which handles backup creation
	// We need to copy the file to the backup location
	writer := NewAtomicFileWriter()

	// Generate backup path with timestamp
	backupPath := filePath + suffix + "." + tfo.getTimestamp()

	return writer.CopyFile(ctx, filePath, backupPath, 0)
}

// getTimestamp returns a timestamp string for backup files
func (tfo *TemplateFileOperations) getTimestamp() string {
	return time.Now().Format("20060102-150405")
}

// GetTemplateProcessor returns the underlying template processor
func (tfo *TemplateFileOperations) GetTemplateProcessor() TemplateProcessor {
	return tfo.templateProcessor
}

// GetBaseOperations returns the underlying FileOperations
func (tfo *TemplateFileOperations) GetBaseOperations() *FileOperations {
	return tfo.baseOps
}
