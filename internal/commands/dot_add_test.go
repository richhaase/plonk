// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/operations"
)

func TestDotAddCommand_Creation(t *testing.T) {
	// Test that the dot add command is created correctly
	if dotAddCmd == nil {
		t.Fatal("dotAddCmd is nil")
	}

	if dotAddCmd.Use != "add <dotfile1> [dotfile2] ..." {
		t.Errorf("Expected Use to be 'add <dotfile1> [dotfile2] ...', got '%s'", dotAddCmd.Use)
	}

	if dotAddCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if dotAddCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if dotAddCmd.RunE == nil {
		t.Error("RunE should not be nil")
	}
}

func TestDotAddCommand_Flags(t *testing.T) {
	// Test that flags are set up correctly
	flag := dotAddCmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Error("dry-run flag not found")
	}
}

func TestAddSingleFileNew(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	configDir := filepath.Join(tempDir, "config")

	// Create directories
	err := os.MkdirAll(homeDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create test file
	testFile := filepath.Join(homeDir, ".vimrc")
	testContent := "set number\nset syntax=on\n"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create minimal config
	cfg := &config.Config{}

	tests := []struct {
		name           string
		filePath       string
		dryRun         bool
		setupExisting  bool
		expectedStatus string
	}{
		{
			name:           "new file add",
			filePath:       testFile,
			dryRun:         false,
			setupExisting:  false,
			expectedStatus: "added",
		},
		{
			name:           "existing file update",
			filePath:       testFile,
			dryRun:         false,
			setupExisting:  true,
			expectedStatus: "updated",
		},
		{
			name:           "dry run new file",
			filePath:       testFile,
			dryRun:         true,
			setupExisting:  false,
			expectedStatus: "would-add",
		},
		{
			name:           "dry run existing file",
			filePath:       testFile,
			dryRun:         true,
			setupExisting:  true,
			expectedStatus: "would-update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable to use test config directory
			originalPlonkDir := os.Getenv("PLONK_DIR")
			os.Setenv("PLONK_DIR", configDir)
			defer func() {
				if originalPlonkDir == "" {
					os.Unsetenv("PLONK_DIR")
				} else {
					os.Setenv("PLONK_DIR", originalPlonkDir)
				}
			}()

			// Setup existing file if needed
			if tt.setupExisting {
				source, _ := generatePaths(tt.filePath, homeDir)
				sourcePath := filepath.Join(configDir, source)
				err := os.MkdirAll(filepath.Dir(sourcePath), 0755)
				if err != nil {
					t.Fatal(err)
				}
				err = os.WriteFile(sourcePath, []byte("existing content"), 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Execute function
			ctx := context.Background()
			result := addSingleFileNew(ctx, cfg, tt.filePath, homeDir, configDir, tt.dryRun)

			// Verify results
			if result.Name != tt.filePath {
				t.Errorf("Name: got %s, want %s", result.Name, tt.filePath)
			}
			if result.Status != tt.expectedStatus {
				t.Errorf("Status: got %s, want %s", result.Status, tt.expectedStatus)
			}
			if result.FilesProcessed != 1 {
				t.Errorf("FilesProcessed: got %d, want 1", result.FilesProcessed)
			}

			// Check metadata
			if result.Metadata["source"] == nil {
				t.Error("Source metadata should not be nil")
			}
			if result.Metadata["destination"] == nil {
				t.Error("Destination metadata should not be nil")
			}

			// For non-dry-run, verify file was actually copied
			if !tt.dryRun && result.Status != "failed" {
				source, _ := generatePaths(tt.filePath, homeDir)
				sourcePath := filepath.Join(configDir, source)
				if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
					t.Error("File should have been copied to config directory")
				}

				// Verify content is correct
				copiedContent, err := os.ReadFile(sourcePath)
				if err != nil {
					t.Fatal(err)
				}
				if string(copiedContent) != testContent {
					t.Error("Copied file content does not match original")
				}
			}

			// Clean up for next test
			if tt.setupExisting {
				source, _ := generatePaths(tt.filePath, homeDir)
				sourcePath := filepath.Join(configDir, source)
				os.RemoveAll(filepath.Dir(sourcePath))
			}
		})
	}
}

func TestAddDirectoryFilesNew(t *testing.T) {
	// Create temp directories
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	configDir := filepath.Join(tempDir, "config")
	testDir := filepath.Join(homeDir, ".config", "nvim")

	// Create directories
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create test files
	testFiles := map[string]string{
		"init.lua":        "-- nvim config\nvim.opt.number = true\n",
		"lua/config.lua":  "-- lua config\nlocal M = {}\nreturn M\n",
		"plugin/test.vim": "\" vim plugin\necho 'hello'\n",
		".gitignore":      "*.tmp\n",
	}

	for relPath, content := range testFiles {
		fullPath := filepath.Join(testDir, relPath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Set environment variable to use test config directory
	originalPlonkDir := os.Getenv("PLONK_DIR")
	os.Setenv("PLONK_DIR", configDir)
	defer func() {
		if originalPlonkDir == "" {
			os.Unsetenv("PLONK_DIR")
		} else {
			os.Setenv("PLONK_DIR", originalPlonkDir)
		}
	}()

	// Create minimal config
	cfg := &config.Config{}

	tests := []struct {
		name           string
		dryRun         bool
		expectedFiles  int
		expectedStatus string
	}{
		{
			name:           "add directory",
			dryRun:         false,
			expectedFiles:  4,
			expectedStatus: "added",
		},
		{
			name:           "dry run directory",
			dryRun:         true,
			expectedFiles:  4,
			expectedStatus: "would-add",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh config directory for each test to avoid interference
			testConfigDir := filepath.Join(tempDir, "config_"+tt.name)
			err := os.MkdirAll(testConfigDir, 0755)
			if err != nil {
				t.Fatal(err)
			}

			// Update PLONK_DIR for this specific test
			os.Setenv("PLONK_DIR", testConfigDir)

			// Execute function
			ctx := context.Background()
			results := addDirectoryFilesNew(ctx, cfg, testDir, homeDir, testConfigDir, tt.dryRun)

			// Verify results
			if len(results) != tt.expectedFiles {
				t.Errorf("Expected %d files, got %d", tt.expectedFiles, len(results))
			}

			// Check that all results have the expected status
			for _, result := range results {
				if result.Status != tt.expectedStatus && result.Status != "failed" {
					t.Errorf("Unexpected status: got %s, want %s", result.Status, tt.expectedStatus)
				}
				if result.FilesProcessed != 1 {
					t.Errorf("FilesProcessed: got %d, want 1", result.FilesProcessed)
				}
			}

			// For non-dry-run, verify files were actually copied
			if !tt.dryRun {
				for _, result := range results {
					if result.Status == "added" {
						source := result.Metadata["source"].(string)
						sourcePath := filepath.Join(testConfigDir, source)
						if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
							t.Errorf("File should have been copied: %s", sourcePath)
						}
					}
				}
			}
		})
	}
}

func TestCopyFileWithAttributes(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	content := "test content\n"
	err := os.WriteFile(srcFile, []byte(content), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Get original file info
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	// Copy file
	dstFile := filepath.Join(tempDir, "dest.txt")
	err = copyFileWithAttributes(srcFile, dstFile)
	if err != nil {
		t.Fatal(err)
	}

	// Verify destination file exists and has correct content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(dstContent) != content {
		t.Error("Content does not match")
	}

	// Verify permissions are preserved
	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatal(err)
	}
	if dstInfo.Mode() != srcInfo.Mode() {
		t.Errorf("Permissions not preserved: got %v, want %v", dstInfo.Mode(), srcInfo.Mode())
	}

	// Verify modification time is preserved
	if !dstInfo.ModTime().Equal(srcInfo.ModTime()) {
		t.Error("Modification time not preserved")
	}
}

func TestMapStatusToAction(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"added", "added"},
		{"would-add", "added"},
		{"updated", "updated"},
		{"would-update", "updated"},
		{"failed", "failed"},
		{"unknown", "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := mapStatusToAction(tt.status)
			if result != tt.expected {
				t.Errorf("mapStatusToAction(%s): got %s, want %s", tt.status, result, tt.expected)
			}
		})
	}
}

func TestConvertResultsToDotfileAdd(t *testing.T) {
	results := []operations.OperationResult{
		{
			Name:   "/home/user/.vimrc",
			Status: "added",
			Metadata: map[string]interface{}{
				"source":      "dot_vimrc",
				"destination": "~/.vimrc",
			},
		},
		{
			Name:   "/home/user/.zshrc",
			Status: "failed",
			Error:  fmt.Errorf("permission denied"),
		},
		{
			Name:   "/home/user/.tmux.conf",
			Status: "updated",
			Metadata: map[string]interface{}{
				"source":      "dot_tmux.conf",
				"destination": "~/.tmux.conf",
			},
		},
	}

	outputs := convertResultsToDotfileAdd(results)

	// Should only include non-failed results
	if len(outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(outputs))
		return
	}

	// Test first result (added)
	if outputs[0].Source != "dot_vimrc" {
		t.Errorf("Source: got %s, want dot_vimrc", outputs[0].Source)
	}
	if outputs[0].Action != "added" {
		t.Errorf("Action: got %s, want added", outputs[0].Action)
	}

	// Test second result (updated)
	if outputs[1].Source != "dot_tmux.conf" {
		t.Errorf("Source: got %s, want dot_tmux.conf", outputs[1].Source)
	}
	if outputs[1].Action != "updated" {
		t.Errorf("Action: got %s, want updated", outputs[1].Action)
	}
}

func TestExtractErrorMessages(t *testing.T) {
	results := []operations.OperationResult{
		{
			Name:   "/home/user/.vimrc",
			Status: "added",
		},
		{
			Name:   "/home/user/.zshrc",
			Status: "failed",
			Error:  fmt.Errorf("permission denied"),
		},
		{
			Name:   "/home/user/.bashrc",
			Status: "failed",
			Error:  fmt.Errorf("file not found"),
		},
	}

	errors := extractErrorMessages(results)

	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
		return
	}

	expectedErrors := []string{
		"failed to add /home/user/.zshrc: permission denied",
		"failed to add /home/user/.bashrc: file not found",
	}

	for i, expectedError := range expectedErrors {
		if errors[i] != expectedError {
			t.Errorf("Error %d: got %q, want %q", i, errors[i], expectedError)
		}
	}
}
