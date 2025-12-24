// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
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

// ScanDotfiles scans for dotfiles in the home directory
func (s *Scanner) ScanDotfiles(ctx context.Context) ([]ScanResult, error) {
	var results []ScanResult

	// List all entries in home directory
	entries, err := os.ReadDir(s.homeDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Skip non-dotfiles
		if len(entry.Name()) == 0 || entry.Name()[0] != '.' {
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
