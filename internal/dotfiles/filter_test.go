// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"testing"
	"time"
)

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name  string
	isDir bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }

func TestFilter_ShouldSkip(t *testing.T) {
	tests := []struct {
		name           string
		ignorePatterns []string
		configDir      string
		skipConfigDir  bool
		relPath        string
		fileInfo       os.FileInfo
		expected       bool
	}{
		{
			name:           "always skip plonk.yaml",
			ignorePatterns: []string{},
			relPath:        "plonk.yaml",
			fileInfo:       mockFileInfo{name: "plonk.yaml", isDir: false},
			expected:       true,
		},
		{
			name:           "skip config directory",
			ignorePatterns: []string{},
			configDir:      "/home/user/.config/plonk",
			skipConfigDir:  true,
			relPath:        ".config/plonk",
			fileInfo:       mockFileInfo{name: "plonk", isDir: true},
			expected:       true,
		},
		{
			name:           "skip file in config directory",
			ignorePatterns: []string{},
			configDir:      "/home/user/.config/plonk",
			skipConfigDir:  true,
			relPath:        ".config/plonk/dotfiles/vimrc",
			fileInfo:       mockFileInfo{name: "vimrc", isDir: false},
			expected:       true,
		},
		{
			name:           "don't skip config dir when flag is false",
			ignorePatterns: []string{},
			configDir:      "/home/user/.config/plonk",
			skipConfigDir:  false,
			relPath:        ".config/plonk",
			fileInfo:       mockFileInfo{name: "plonk", isDir: true},
			expected:       false,
		},
		{
			name:           "skip by exact match",
			ignorePatterns: []string{".gitignore"},
			relPath:        ".gitignore",
			fileInfo:       mockFileInfo{name: ".gitignore", isDir: false},
			expected:       true,
		},
		{
			name:           "skip by glob pattern",
			ignorePatterns: []string{"*.bak"},
			relPath:        "file.bak",
			fileInfo:       mockFileInfo{name: "file.bak", isDir: false},
			expected:       true,
		},
		{
			name:           "skip by path pattern",
			ignorePatterns: []string{".cache/*"},
			relPath:        ".cache/data",
			fileInfo:       mockFileInfo{name: "data", isDir: false},
			expected:       true,
		},
		{
			name:           "don't skip non-matching",
			ignorePatterns: []string{".gitignore", "*.bak"},
			relPath:        ".vimrc",
			fileInfo:       mockFileInfo{name: ".vimrc", isDir: false},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.ignorePatterns, tt.configDir, tt.skipConfigDir)
			result := filter.ShouldSkip(tt.relPath, tt.fileInfo)
			if result != tt.expected {
				t.Errorf("ShouldSkip() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilter_ShouldSkipFilesOnly(t *testing.T) {
	filter := NewFilter([]string{}, "", false)

	// Test directory - should be skipped
	dirInfo := mockFileInfo{name: "dir", isDir: true}
	if !filter.ShouldSkipFilesOnly("dir", dirInfo) {
		t.Error("Expected directory to be skipped in files-only mode")
	}

	// Test file - should not be skipped
	fileInfo := mockFileInfo{name: "file", isDir: false}
	if filter.ShouldSkipFilesOnly("file", fileInfo) {
		t.Error("Expected file to not be skipped in files-only mode")
	}
}
