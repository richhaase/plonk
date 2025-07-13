// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathResolver_ResolveDotfilePath(t *testing.T) {
	// Setup test directories
	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	resolver := NewPathResolver(homeDir, configDir)

	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:     "tilde path",
			input:    "~/.zshrc",
			expected: "/tmp/test-home/.zshrc",
		},
		{
			name:     "tilde with subdirectory",
			input:    "~/.config/nvim/init.lua",
			expected: "/tmp/test-home/.config/nvim/init.lua",
		},
		{
			name:        "absolute path outside home",
			input:       "/etc/passwd",
			shouldError: true,
		},
		{
			name:     "absolute path inside home",
			input:    "/tmp/test-home/.bashrc",
			expected: "/tmp/test-home/.bashrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolveDotfilePath(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for input %s, but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for input %s: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPathResolver_GenerateDestinationPath(t *testing.T) {
	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	resolver := NewPathResolver(homeDir, configDir)

	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:     "home directory file",
			input:    "/tmp/test-home/.zshrc",
			expected: "~/.zshrc",
		},
		{
			name:     "subdirectory file",
			input:    "/tmp/test-home/.config/nvim/init.lua",
			expected: "~/.config/nvim/init.lua",
		},
		{
			name:     "path outside home creates relative path",
			input:    "/etc/passwd",
			expected: "~/../../etc/passwd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.GenerateDestinationPath(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for input %s, but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for input %s: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPathResolver_GenerateSourcePath(t *testing.T) {
	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	resolver := NewPathResolver(homeDir, configDir)

	tests := []struct {
		name        string
		destination string
		expected    string
	}{
		{
			name:        "zshrc file",
			destination: "~/.zshrc",
			expected:    "zshrc",
		},
		{
			name:        "config directory file",
			destination: "~/.config/nvim/init.lua",
			expected:    "config/nvim/init.lua",
		},
		{
			name:        "editorconfig file",
			destination: "~/.editorconfig",
			expected:    "editorconfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.GenerateSourcePath(tt.destination)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPathResolver_GeneratePaths(t *testing.T) {
	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	resolver := NewPathResolver(homeDir, configDir)

	resolvedPath := "/tmp/test-home/.zshrc"
	expectedSource := "zshrc"
	expectedDestination := "~/.zshrc"

	source, destination, err := resolver.GeneratePaths(resolvedPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if source != expectedSource {
		t.Errorf("expected source %s, got %s", expectedSource, source)
	}

	if destination != expectedDestination {
		t.Errorf("expected destination %s, got %s", expectedDestination, destination)
	}
}

func TestPathResolver_ValidatePath(t *testing.T) {
	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	resolver := NewPathResolver(homeDir, configDir)

	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:  "valid tilde path",
			input: "~/.zshrc",
		},
		{
			name:        "directory traversal",
			input:       "~/../etc/passwd",
			shouldError: true,
		},
		{
			name:        "parent directory access",
			input:       "~/.ssh/../../../etc/passwd",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resolver.ValidatePath(tt.input)

			if tt.shouldError && err == nil {
				t.Errorf("expected error for input %s, but got none", tt.input)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error for input %s: %v", tt.input, err)
			}
		})
	}
}

func TestPathResolver_ExpandDirectory(t *testing.T) {
	// Create temporary test directory structure
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	configDir := filepath.Join(tempDir, "config")

	// Create test directory and files
	testDir := filepath.Join(homeDir, ".config", "nvim")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create test files
	testFiles := []string{"init.lua", "plugins.lua", "keymaps.lua"}
	for _, file := range testFiles {
		filePath := filepath.Join(testDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file %s: %v", file, err)
		}
	}

	resolver := NewPathResolver(homeDir, configDir)

	// Test directory expansion
	entries, err := resolver.ExpandDirectory(filepath.Join(homeDir, ".config", "nvim"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != len(testFiles) {
		t.Errorf("expected %d entries, got %d", len(testFiles), len(entries))
	}

	// Verify entries - use map for order-independent comparison since filepath.Walk doesn't guarantee order
	foundFiles := make(map[string]bool)
	for _, entry := range entries {
		foundFiles[entry.RelativePath] = true

		// Verify full path is correct
		expectedFullPath := filepath.Join(testDir, entry.RelativePath)
		if entry.FullPath != expectedFullPath {
			t.Errorf("expected full path %s, got %s", expectedFullPath, entry.FullPath)
		}
	}

	// Verify all expected files were found
	for _, expectedFile := range testFiles {
		if !foundFiles[expectedFile] {
			t.Errorf("expected file %s not found in entries", expectedFile)
		}
	}
}

func TestPathResolver_GetSourcePath(t *testing.T) {
	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	resolver := NewPathResolver(homeDir, configDir)

	source := "zshrc"
	expected := "/tmp/test-config/zshrc"

	result := resolver.GetSourcePath(source)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestPathResolver_GetDestinationPath(t *testing.T) {
	homeDir := "/tmp/test-home"
	configDir := "/tmp/test-config"
	resolver := NewPathResolver(homeDir, configDir)

	tests := []struct {
		name        string
		destination string
		expected    string
		shouldError bool
	}{
		{
			name:        "tilde path",
			destination: "~/.zshrc",
			expected:    "/tmp/test-home/.zshrc",
		},
		{
			name:        "absolute path",
			destination: "/tmp/test-home/.bashrc",
			expected:    "/tmp/test-home/.bashrc",
		},
		{
			name:        "relative path",
			destination: ".vimrc",
			expected:    "/tmp/test-home/.vimrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.GetDestinationPath(tt.destination)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for destination %s, but got none", tt.destination)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for destination %s: %v", tt.destination, err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
