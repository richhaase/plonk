// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverDotfiles(t *testing.T) {
	tests := []struct {
		name          string
		existingFiles []string
		expectedFiles []string
	}{
		{
			name:          "all target dotfiles exist",
			existingFiles: []string{".zshrc", ".gitconfig", ".zshenv"},
			expectedFiles: []string{".zshrc", ".gitconfig", ".zshenv"},
		},
		{
			name:          "some target dotfiles exist",
			existingFiles: []string{".zshrc", ".gitconfig"},
			expectedFiles: []string{".zshrc", ".gitconfig"},
		},
		{
			name:          "single target dotfile exists",
			existingFiles: []string{".zshrc"},
			expectedFiles: []string{".zshrc"},
		},
		{
			name:          "no target dotfiles exist - returns empty list",
			existingFiles: []string{},
			expectedFiles: []string{},
		},
		{
			name:          "other dotfiles exist but not our managed types",
			existingFiles: []string{".bashrc", ".vimrc", "README.md"},
			expectedFiles: []string{},
		},
		{
			name:          "mix of managed and unmanaged dotfiles",
			existingFiles: []string{".zshrc", ".bashrc", ".gitconfig", ".vimrc"},
			expectedFiles: []string{".zshrc", ".gitconfig"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary home directory
			tempHome := t.TempDir()
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempHome)
			defer os.Setenv("HOME", originalHome)

			// Create test files
			for _, file := range tt.existingFiles {
				filePath := filepath.Join(tempHome, file)
				err := os.WriteFile(filePath, []byte("# test content"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file %s: %v", file, err)
				}
			}

			discoverer := NewDotfileDiscoverer()
			dotfiles, err := discoverer.DiscoverDotfiles()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(dotfiles) != len(tt.expectedFiles) {
				t.Errorf("Expected %d dotfiles, got %d", len(tt.expectedFiles), len(dotfiles))
				t.Errorf("Expected: %v", tt.expectedFiles)
				t.Errorf("Got: %v", dotfiles)
			}

			// Check that all expected files are found and no unexpected ones
			for _, expected := range tt.expectedFiles {
				found := false
				for _, actual := range dotfiles {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dotfile %s not found in results", expected)
				}
			}
		})
	}
}
