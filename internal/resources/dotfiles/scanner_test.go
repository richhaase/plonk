// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanner_ScanDotfiles(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		".bashrc",
		".vimrc",
		".config/nvim/init.lua",
		"regular-file.txt", // Should be skipped
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create scanner
	scanner := NewScanner(tmpDir, nil)

	// Scan dotfiles
	results, err := scanner.ScanDotfiles(context.Background())
	if err != nil {
		t.Fatalf("ScanDotfiles failed: %v", err)
	}

	// Verify results
	expectedCount := 3 // .bashrc, .vimrc, .config (dir)
	if len(results) != expectedCount {
		t.Errorf("Expected %d results, got %d", expectedCount, len(results))
	}

	// Check that regular-file.txt was skipped
	for _, result := range results {
		if result.Name == "regular-file.txt" {
			t.Error("Non-dotfile should have been skipped")
		}
	}
}

func TestScanner_WithFilter(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		".bashrc",
		".gitignore",
		".config/plonk/config.yaml",
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create filter that skips .gitignore
	filter := NewFilter([]string{".gitignore"}, "", false)

	// Create scanner with filter
	scanner := NewScanner(tmpDir, filter)

	// Scan dotfiles
	results, err := scanner.ScanDotfiles(context.Background())
	if err != nil {
		t.Fatalf("ScanDotfiles failed: %v", err)
	}

	// Verify .gitignore was filtered
	for _, result := range results {
		if result.Name == ".gitignore" {
			t.Error(".gitignore should have been filtered")
		}
	}

	// Should still have .bashrc and .config
	foundBashrc := false
	foundConfig := false
	for _, result := range results {
		if result.Name == ".bashrc" {
			foundBashrc = true
		}
		if result.Name == ".config" {
			foundConfig = true
		}
	}

	if !foundBashrc {
		t.Error("Expected to find .bashrc")
	}
	if !foundConfig {
		t.Error("Expected to find .config")
	}
}
