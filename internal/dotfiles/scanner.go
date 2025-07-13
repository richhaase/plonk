// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/errors"
)

// ScanResult represents a scanned file or directory
type ScanResult struct {
	Name     string                 // Relative path from scan root
	Path     string                 // Absolute path
	Info     os.FileInfo            // File info
	Metadata map[string]interface{} // Additional metadata
}

// Scanner handles file system scanning for dotfiles
type Scanner struct {
	homeDir string
	filter  *Filter
}

// NewScanner creates a new scanner
func NewScanner(homeDir string, filter *Filter) *Scanner {
	return &Scanner{
		homeDir: homeDir,
		filter:  filter,
	}
}

// ScanDirectory scans a directory for dotfiles
func (s *Scanner) ScanDirectory(ctx context.Context, dir string, maxDepth int) ([]ScanResult, error) {
	var results []ScanResult

	err := s.walkDirectory(ctx, dir, dir, 0, maxDepth, &results)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "scan",
			"failed to scan directory")
	}

	return results, nil
}

// ScanDotfiles scans for dotfiles in the home directory
func (s *Scanner) ScanDotfiles(ctx context.Context) ([]ScanResult, error) {
	var results []ScanResult

	// List all entries in home directory
	entries, err := os.ReadDir(s.homeDir)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "scan",
			"failed to read home directory")
	}

	for _, entry := range entries {
		// Skip non-dotfiles
		if !isDotfile(entry.Name()) {
			continue
		}

		fullPath := filepath.Join(s.homeDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		// Apply filter
		if s.filter != nil && s.filter.ShouldSkip(entry.Name(), info) {
			continue
		}

		results = append(results, ScanResult{
			Name: entry.Name(),
			Path: fullPath,
			Info: info,
			Metadata: map[string]interface{}{
				"path": fullPath,
			},
		})
	}

	return results, nil
}

// walkDirectory recursively walks a directory up to maxDepth
func (s *Scanner) walkDirectory(ctx context.Context, root, dir string, currentDepth, maxDepth int, results *[]ScanResult) error {
	if currentDepth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		// Skip directories we can't read
		return nil
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fullPath := filepath.Join(dir, entry.Name())
		relPath, err := filepath.Rel(root, fullPath)
		if err != nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Apply filter
		if s.filter != nil && s.filter.ShouldSkip(relPath, info) {
			continue
		}

		*results = append(*results, ScanResult{
			Name: relPath,
			Path: fullPath,
			Info: info,
			Metadata: map[string]interface{}{
				"path": fullPath,
			},
		})

		// Recurse into directories
		if entry.IsDir() && currentDepth < maxDepth {
			if err := s.walkDirectory(ctx, root, fullPath, currentDepth+1, maxDepth, results); err != nil {
				return err
			}
		}
	}

	return nil
}

// isDotfile checks if a filename represents a dotfile
func isDotfile(name string) bool {
	return len(name) > 0 && name[0] == '.'
}
