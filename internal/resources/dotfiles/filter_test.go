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

func TestFilter_UnmanagedPatterns(t *testing.T) {
	// Test patterns specifically for unmanaged filtering
	tests := []struct {
		name     string
		pattern  string
		relPath  string
		fileInfo os.FileInfo
		expected bool
	}{
		{
			name:     "skip log files",
			pattern:  "*.log",
			relPath:  "gcloud/logs/2025.04.16/17.57.27.876860.log",
			fileInfo: mockFileInfo{name: "17.57.27.876860.log", isDir: false},
			expected: true,
		},
		{
			name:     "skip map files",
			pattern:  "*.map",
			relPath:  "raycast/extensions/uuid/file.js.map",
			fileInfo: mockFileInfo{name: "file.js.map", isDir: false},
			expected: true,
		},
		{
			name:     "skip node_modules directories",
			pattern:  "**/node_modules/**",
			relPath:  "project/node_modules/package/file.js",
			fileInfo: mockFileInfo{name: "file.js", isDir: false},
			expected: true,
		},
		{
			name:     "skip UUID directories",
			pattern:  "**/*-*-*-*-*/**",
			relPath:  "raycast/extensions/4d342edf-4371-498e-8ead-a424d65f933f/file.js",
			fileInfo: mockFileInfo{name: "file.js", isDir: false},
			expected: true,
		},
		{
			name:     "skip cache directories",
			pattern:  "**/*cache*/**",
			relPath:  "app/cache/data/file.tmp",
			fileInfo: mockFileInfo{name: "file.tmp", isDir: false},
			expected: true,
		},
		{
			name:     "skip git internals",
			pattern:  "**/.git/**",
			relPath:  "repo/.git/objects/pack/pack-abc.pack",
			fileInfo: mockFileInfo{name: "pack-abc.pack", isDir: false},
			expected: true,
		},
		{
			name:     "don't skip regular files",
			pattern:  "*.log",
			relPath:  "config/app.conf",
			fileInfo: mockFileInfo{name: "app.conf", isDir: false},
			expected: false,
		},
		{
			name:     "don't skip UUID-like but not matching pattern",
			pattern:  "**/*-*-*-*-*/**",
			relPath:  "file-with-dashes.txt",
			fileInfo: mockFileInfo{name: "file-with-dashes.txt", isDir: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter([]string{tt.pattern}, "", true)
			result := filter.ShouldSkip(tt.relPath, tt.fileInfo)
			if result != tt.expected {
				t.Errorf("ShouldSkip(%q) with pattern %q = %v, want %v", tt.relPath, tt.pattern, result, tt.expected)
			}
		})
	}
}

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
			name:           "always skip .plonk directory",
			ignorePatterns: []string{},
			relPath:        ".plonk",
			fileInfo:       mockFileInfo{name: ".plonk", isDir: true},
			expected:       true,
		},
		{
			name:           "always skip files in .plonk directory",
			ignorePatterns: []string{},
			relPath:        ".plonk/hooks/pre-apply.sh",
			fileInfo:       mockFileInfo{name: "pre-apply.sh", isDir: false},
			expected:       true,
		},
		{
			name:           "always skip subdirectories in .plonk",
			ignorePatterns: []string{},
			relPath:        ".plonk/templates/config",
			fileInfo:       mockFileInfo{name: "config", isDir: false},
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
