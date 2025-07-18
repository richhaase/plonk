// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/errors"
)

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

// WriteFromReader atomically writes from a reader to a file
func (a *AtomicFileWriter) WriteFromReader(filename string, reader io.Reader, perm os.FileMode) error {
	return a.writeFileInternal(filename, func(tmpFile *os.File) error {
		_, err := io.Copy(tmpFile, reader)
		return err
	}, perm)
}

// CopyFile atomically copies a file from source to destination
func (a *AtomicFileWriter) CopyFile(ctx context.Context, src, dst string, perm os.FileMode) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy",
			"failed to open source file").WithMetadata("source", src)
	}
	defer srcFile.Close()

	// Get source file info for permissions if perm is 0
	var finalPerm os.FileMode
	if perm == 0 {
		srcInfo, err := srcFile.Stat()
		if err != nil {
			return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy",
				"failed to get source file info").WithMetadata("source", src)
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
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "write",
			"failed to create destination directory").WithMetadata("directory", destDir).WithSuggestionMessage("Check directory permissions and disk space")
	}

	// Create temporary file in same directory as destination
	tmpFile, err := os.CreateTemp(destDir, ".tmp-"+filepath.Base(filename))
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "write",
			"failed to create temporary file").WithMetadata("directory", destDir).WithMetadata("filename", filename).WithSuggestionMessage("Check directory permissions and disk space")
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
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "write",
			"failed to write file contents").WithMetadata("filename", filename).WithMetadata("tmpPath", tmpPath)
	}

	// Sync to disk
	if err := tmpFile.Sync(); err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "write",
			"failed to sync temporary file").WithMetadata("filename", filename).WithMetadata("tmpPath", tmpPath)
	}

	// Close temporary file
	if err := tmpFile.Close(); err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "write",
			"failed to close temporary file").WithMetadata("filename", filename).WithMetadata("tmpPath", tmpPath)
	}
	tmpFile = nil // Mark as closed for defer cleanup

	// Set permissions
	if err := os.Chmod(tmpPath, perm); err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainDotfiles, "write",
			"failed to set file permissions").WithMetadata("filename", filename).WithMetadata("tmpPath", tmpPath).WithMetadata("permissions", perm).WithSuggestionMessage("Check file ownership and directory permissions")
	}

	// Atomic rename - this is the critical atomic operation
	if err := os.Rename(tmpPath, filename); err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "write",
			"failed to rename temporary file").WithMetadata("filename", filename).WithMetadata("tmpPath", tmpPath).WithSuggestionMessage("Check destination directory permissions and disk space")
	}

	return nil
}
