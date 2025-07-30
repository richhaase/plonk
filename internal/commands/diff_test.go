// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/resources"
)

func TestNormalizePath(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set a known HOME for testing
	testHome := "/test/home"
	os.Setenv("HOME", testHome)

	// Save original working directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "tilde expansion",
			input:    "~/.zshrc",
			expected: filepath.Join(testHome, ".zshrc"),
			wantErr:  false,
		},
		{
			name:     "HOME variable expansion",
			input:    "$HOME/.bashrc",
			expected: filepath.Join(testHome, ".bashrc"),
			wantErr:  false,
		},
		{
			name:     "absolute path unchanged",
			input:    "/usr/local/bin/tool",
			expected: "/usr/local/bin/tool",
			wantErr:  false,
		},
		{
			name:     "nested path with tilde",
			input:    "~/.config/nvim/init.lua",
			expected: filepath.Join(testHome, ".config/nvim/init.lua"),
			wantErr:  false,
		},
		{
			name:     "multiple slashes cleaned",
			input:    "/usr//local///bin/tool",
			expected: "/usr/local/bin/tool",
			wantErr:  false,
		},
		{
			name:     "env var in middle of path",
			input:    "/usr/$USER/files",
			expected: "/usr/" + os.Getenv("USER") + "/files",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("normalizePath() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetSourceNameFromItem(t *testing.T) {
	tests := []struct {
		name     string
		item     resources.Item
		expected string
	}{
		{
			name: "source in metadata",
			item: resources.Item{
				Name: ".zshrc",
				Metadata: map[string]interface{}{
					"source": "zshrc",
				},
			},
			expected: "zshrc",
		},
		{
			name: "no metadata, remove dot",
			item: resources.Item{
				Name: ".bashrc",
			},
			expected: "bashrc",
		},
		{
			name: "no metadata, no dot",
			item: resources.Item{
				Name: "gitconfig",
			},
			expected: "gitconfig",
		},
		{
			name: "nested path in metadata",
			item: resources.Item{
				Name: ".config/nvim/init.lua",
				Metadata: map[string]interface{}{
					"source": "config/nvim/init.lua",
				},
			},
			expected: "config/nvim/init.lua",
		},
		{
			name: "nested path no metadata",
			item: resources.Item{
				Name: ".config/helix/config.toml",
			},
			expected: "config/helix/config.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSourceNameFromItem(tt.item)
			if got != tt.expected {
				t.Errorf("getSourceNameFromItem() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFilterDriftedFile(t *testing.T) {
	// Create test items
	driftedFiles := []resources.Item{
		{
			Name: ".zshrc",
			Metadata: map[string]interface{}{
				"destination": "~/.zshrc",
			},
		},
		{
			Name: ".bashrc",
			Metadata: map[string]interface{}{
				"destination": "~/.bashrc",
			},
		},
		{
			Name: ".config/nvim/init.lua",
			Metadata: map[string]interface{}{
				"destination": "~/.config/nvim/init.lua",
			},
		},
	}

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set a known HOME for testing
	testHome := "/test/home"
	os.Setenv("HOME", testHome)

	tests := []struct {
		name      string
		arg       string
		wantFound bool
		wantName  string
	}{
		{
			name:      "find by tilde path",
			arg:       "~/.zshrc",
			wantFound: true,
			wantName:  ".zshrc",
		},
		{
			name:      "find by HOME var",
			arg:       "$HOME/.bashrc",
			wantFound: true,
			wantName:  ".bashrc",
		},
		{
			name:      "find by absolute path",
			arg:       filepath.Join(testHome, ".zshrc"),
			wantFound: true,
			wantName:  ".zshrc",
		},
		{
			name:      "nested path",
			arg:       "~/.config/nvim/init.lua",
			wantFound: true,
			wantName:  ".config/nvim/init.lua",
		},
		{
			name:      "not found",
			arg:       "~/.vimrc",
			wantFound: false,
			wantName:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterDriftedFile(tt.arg, driftedFiles)
			if tt.wantFound {
				if got == nil {
					t.Errorf("filterDriftedFile() = nil, want item with name %v", tt.wantName)
				} else if got.Name != tt.wantName {
					t.Errorf("filterDriftedFile() returned item with name %v, want %v", got.Name, tt.wantName)
				}
			} else {
				if got != nil {
					t.Errorf("filterDriftedFile() = %v, want nil", got.Name)
				}
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set a known HOME for testing
	testHome := "/test/home"
	os.Setenv("HOME", testHome)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde slash",
			input:    "~/file",
			expected: filepath.Join(testHome, "file"),
		},
		{
			name:     "no tilde",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "tilde without slash",
			input:    "~file",
			expected: "~file",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "nested tilde path",
			input:    "~/.config/app/config",
			expected: filepath.Join(testHome, ".config/app/config"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandHome(tt.input)
			if got != tt.expected {
				t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
