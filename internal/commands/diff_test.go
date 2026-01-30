// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/dotfiles"
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

func TestFilterDriftedStatus(t *testing.T) {
	// Create test items using DotfileStatus
	driftedFiles := []dotfiles.DotfileStatus{
		{
			Dotfile: dotfiles.Dotfile{
				Name:   "zshrc",
				Target: "~/.zshrc",
			},
			State: dotfiles.SyncStateDrifted,
		},
		{
			Dotfile: dotfiles.Dotfile{
				Name:   "bashrc",
				Target: "~/.bashrc",
			},
			State: dotfiles.SyncStateDrifted,
		},
		{
			Dotfile: dotfiles.Dotfile{
				Name:   "config/nvim/init.lua",
				Target: "~/.config/nvim/init.lua",
			},
			State: dotfiles.SyncStateDrifted,
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
			wantName:  "zshrc",
		},
		{
			name:      "find by HOME var",
			arg:       "$HOME/.bashrc",
			wantFound: true,
			wantName:  "bashrc",
		},
		{
			name:      "find by absolute path",
			arg:       filepath.Join(testHome, ".zshrc"),
			wantFound: true,
			wantName:  "zshrc",
		},
		{
			name:      "nested path",
			arg:       "~/.config/nvim/init.lua",
			wantFound: true,
			wantName:  "config/nvim/init.lua",
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
			got := filterDriftedStatus(tt.arg, driftedFiles, testHome)
			if tt.wantFound {
				if got == nil {
					t.Errorf("filterDriftedStatus() = nil, want item with name %v", tt.wantName)
				} else if got.Name != tt.wantName {
					t.Errorf("filterDriftedStatus() returned item with name %v, want %v", got.Name, tt.wantName)
				}
			} else {
				if got != nil {
					t.Errorf("filterDriftedStatus() = %v, want nil", got.Name)
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
