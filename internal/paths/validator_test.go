// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package paths

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPathValidator_ValidateSecure(t *testing.T) {
	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	validator := NewPathValidator(homeDir, configDir, []string{})

	tests := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{
			name: "valid path in home",
			path: "/tmp/test-home/.zshrc",
		},
		{
			name:        "directory traversal with dots",
			path:        "/tmp/test-home/../etc/passwd",
			shouldError: true,
		},
		{
			name:        "path outside home",
			path:        "/etc/passwd",
			shouldError: true,
		},
		{
			name:        "null bytes in path",
			path:        "/tmp/test-home/.zshrc\x00evil",
			shouldError: true,
		},
		{
			name:        "config directory access",
			path:        "/tmp/test-config/plonk.yaml",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSecure(tt.path)

			if tt.shouldError && err == nil {
				t.Errorf("expected error for path %s, but got none", tt.path)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error for path %s: %v", tt.path, err)
			}
		})
	}
}

func TestPathValidator_ShouldSkipPath(t *testing.T) {
	ignorePatterns := []string{
		".DS_Store",
		".git",
		"*.backup",
		"*.tmp",
		"node_modules/",
	}

	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	validator := NewPathValidator(homeDir, configDir, ignorePatterns)

	tests := []struct {
		name     string
		relPath  string
		isDir    bool
		expected bool
	}{
		{
			name:     "plonk config file",
			relPath:  "plonk.yaml",
			expected: true,
		},
		{
			name:     "plonk lock file",
			relPath:  "plonk.lock",
			expected: true,
		},
		{
			name:     "DS_Store file",
			relPath:  ".DS_Store",
			expected: true,
		},
		{
			name:     "git directory",
			relPath:  ".git",
			isDir:    true,
			expected: true,
		},
		{
			name:     "backup file",
			relPath:  "config.backup",
			expected: true,
		},
		{
			name:     "tmp file",
			relPath:  "temp.tmp",
			expected: true,
		},
		{
			name:     "node_modules directory",
			relPath:  "node_modules",
			isDir:    true,
			expected: true,
		},
		{
			name:     "file in node_modules",
			relPath:  "node_modules/package.json",
			expected: true,
		},
		{
			name:     "normal dotfile",
			relPath:  ".zshrc",
			expected: false,
		},
		{
			name:     "normal config file",
			relPath:  ".config/nvim/init.lua",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &mockFileInfo{isDir: tt.isDir}
			result := validator.ShouldSkipPath(tt.relPath, info)

			if result != tt.expected {
				t.Errorf("expected %t for path %s, got %t", tt.expected, tt.relPath, result)
			}
		})
	}
}

func TestPathValidator_ValidateDirectory(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	configDir := filepath.Join(tempDir, "config")

	// Create test directories
	testDir := filepath.Join(homeDir, ".config")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create a test file (not directory)
	testFile := filepath.Join(homeDir, "testfile")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	validator := NewPathValidator(homeDir, configDir, []string{})

	tests := []struct {
		name        string
		dirPath     string
		shouldError bool
	}{
		{
			name:    "valid directory",
			dirPath: testDir,
		},
		{
			name:        "non-existent directory",
			dirPath:     filepath.Join(homeDir, "nonexistent"),
			shouldError: true,
		},
		{
			name:        "path is file not directory",
			dirPath:     testFile,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDirectory(tt.dirPath)

			if tt.shouldError && err == nil {
				t.Errorf("expected error for directory %s, but got none", tt.dirPath)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error for directory %s: %v", tt.dirPath, err)
			}
		})
	}
}

func TestPathValidator_ValidateFile(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	configDir := filepath.Join(tempDir, "config")

	// Create test directory
	testDir := filepath.Join(homeDir, ".config")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(homeDir, "testfile")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	validator := NewPathValidator(homeDir, configDir, []string{})

	tests := []struct {
		name        string
		filePath    string
		shouldError bool
	}{
		{
			name:     "valid file",
			filePath: testFile,
		},
		{
			name:        "non-existent file",
			filePath:    filepath.Join(homeDir, "nonexistent"),
			shouldError: true,
		},
		{
			name:        "path is directory not file",
			filePath:    testDir,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFile(tt.filePath)

			if tt.shouldError && err == nil {
				t.Errorf("expected error for file %s, but got none", tt.filePath)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error for file %s: %v", tt.filePath, err)
			}
		})
	}
}

func TestPathValidator_IgnorePatterns(t *testing.T) {
	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	initialPatterns := []string{".DS_Store", "*.tmp"}

	validator := NewPathValidator(homeDir, configDir, initialPatterns)

	// Test getting patterns
	patterns := validator.GetIgnorePatterns()
	if len(patterns) != len(initialPatterns) {
		t.Errorf("expected %d patterns, got %d", len(initialPatterns), len(patterns))
	}

	// Test adding pattern
	validator.AddIgnorePattern("*.backup")
	patterns = validator.GetIgnorePatterns()
	if len(patterns) != len(initialPatterns)+1 {
		t.Errorf("expected %d patterns after adding, got %d", len(initialPatterns)+1, len(patterns))
	}

	// Verify the new pattern was added
	found := false
	for _, pattern := range patterns {
		if pattern == "*.backup" {
			found = true
			break
		}
	}
	if !found {
		t.Error("newly added pattern '*.backup' not found in patterns")
	}
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	isDir bool
}

func (m *mockFileInfo) Name() string       { return "test" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }
