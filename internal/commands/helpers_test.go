// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCompleteDotfilePaths(t *testing.T) {
	tests := []struct {
		name       string
		toComplete string
		wantLen    int
		checkStrs  []string
	}{
		{
			name:       "empty input returns all suggestions",
			toComplete: "",
			wantLen:    24, // Count of all commonDotfiles
			checkStrs:  []string{"~/.zshrc", "~/.bashrc", "~/.vimrc"},
		},
		{
			name:       "tilde path filters suggestions",
			toComplete: "~/.z",
			wantLen:    3, // .zshrc, .zprofile, .zshenv
			checkStrs:  []string{"~/.zshrc", "~/.zprofile", "~/.zshenv"},
		},
		{
			name:       "config directory path",
			toComplete: "~/.config/",
			wantLen:    4,
			checkStrs:  []string{"~/.config/", "~/.config/nvim/", "~/.config/fish/"},
		},
		{
			name:       "relative dotfile",
			toComplete: ".v",
			wantLen:    1,
			checkStrs:  []string{".vimrc"},
		},
		{
			name:       "no matches for tilde path",
			toComplete: "~/.nonexistent",
			wantLen:    0,
			checkStrs:  []string{},
		},
		{
			name:       "absolute path returns default",
			toComplete: "/etc/",
			wantLen:    0,
			checkStrs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions, directive := CompleteDotfilePaths(nil, nil, tt.toComplete)

			if tt.toComplete == "/etc/" {
				// For absolute paths, we expect default directive
				if directive != cobra.ShellCompDirectiveDefault {
					t.Errorf("Expected ShellCompDirectiveDefault for absolute path, got %v", directive)
				}
			} else if len(suggestions) > 0 {
				// For matches, we expect NoSpace directive
				if directive != cobra.ShellCompDirectiveNoSpace {
					t.Errorf("Expected ShellCompDirectiveNoSpace when suggestions exist, got %v", directive)
				}
			}

			if len(suggestions) != tt.wantLen {
				t.Errorf("Got %d suggestions, want %d", len(suggestions), tt.wantLen)
			}

			for _, check := range tt.checkStrs {
				found := false
				for _, s := range suggestions {
					if s == check {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find %q in suggestions", check)
				}
			}
		})
	}
}
func TestNormalizeDisplayFlags(t *testing.T) {
	tests := []struct {
		name         string
		showPackages bool
		showDotfiles bool
		wantPackages bool
		wantDotfiles bool
	}{
		{
			name:         "both false returns both true",
			showPackages: false,
			showDotfiles: false,
			wantPackages: true,
			wantDotfiles: true,
		},
		{
			name:         "packages only",
			showPackages: true,
			showDotfiles: false,
			wantPackages: true,
			wantDotfiles: false,
		},
		{
			name:         "dotfiles only",
			showPackages: false,
			showDotfiles: true,
			wantPackages: false,
			wantDotfiles: true,
		},
		{
			name:         "both true stays both true",
			showPackages: true,
			showDotfiles: true,
			wantPackages: true,
			wantDotfiles: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packages, dotfiles := normalizeDisplayFlags(tt.showPackages, tt.showDotfiles)
			assert.Equal(t, tt.wantPackages, packages)
			assert.Equal(t, tt.wantDotfiles, dotfiles)
		})
	}
}
