// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/resources"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestParsePackageSpec(t *testing.T) {
	tests := []struct {
		name            string
		spec            string
		expectedManager string
		expectedPackage string
	}{
		{
			name:            "package without prefix",
			spec:            "htop",
			expectedManager: "",
			expectedPackage: "htop",
		},
		{
			name:            "package with brew prefix",
			spec:            "brew:git",
			expectedManager: "brew",
			expectedPackage: "git",
		},
		{
			name:            "package with npm prefix",
			spec:            "npm:lodash",
			expectedManager: "npm",
			expectedPackage: "lodash",
		},
		{
			name:            "go package with module path",
			spec:            "go:golang.org/x/tools/cmd/gopls",
			expectedManager: "go",
			expectedPackage: "golang.org/x/tools/cmd/gopls",
		},
		{
			name:            "empty prefix with colon",
			spec:            ":package",
			expectedManager: "",
			expectedPackage: "package",
		},
		{
			name:            "multiple colons in spec",
			spec:            "npm:@types/node:lts",
			expectedManager: "npm",
			expectedPackage: "@types/node:lts",
		},
		{
			name:            "empty package name",
			spec:            "brew:",
			expectedManager: "brew",
			expectedPackage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, packageName := ParsePackageSpec(tt.spec)
			if manager != tt.expectedManager {
				t.Errorf("ParsePackageSpec(%q) manager = %q, want %q", tt.spec, manager, tt.expectedManager)
			}
			if packageName != tt.expectedPackage {
				t.Errorf("ParsePackageSpec(%q) packageName = %q, want %q", tt.spec, packageName, tt.expectedPackage)
			}
		})
	}
}

func TestIsValidManager(t *testing.T) {
	tests := []struct {
		name     string
		manager  string
		expected bool
	}{
		{
			name:     "valid brew manager",
			manager:  "brew",
			expected: true,
		},
		{
			name:     "valid npm manager",
			manager:  "npm",
			expected: true,
		},
		{
			name:     "valid cargo manager",
			manager:  "cargo",
			expected: true,
		},
		{
			name:     "valid uv manager",
			manager:  "uv",
			expected: true,
		},
		{
			name:     "valid gem manager",
			manager:  "gem",
			expected: true,
		},
		{
			name:     "valid go manager",
			manager:  "go",
			expected: true,
		},
		{
			name:     "invalid manager",
			manager:  "invalid",
			expected: false,
		},
		{
			name:     "empty manager",
			manager:  "",
			expected: false,
		},
		{
			name:     "case sensitive check",
			manager:  "Brew",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidManager(tt.manager)
			if result != tt.expected {
				t.Errorf("IsValidManager(%q) = %v, want %v", tt.manager, result, tt.expected)
			}
		})
	}
}

func TestGetValidManagers(t *testing.T) {
	managers := GetValidManagers()

	// Should return all 12 supported managers
	expectedCount := 12
	if len(managers) != expectedCount {
		t.Errorf("GetValidManagers() returned %d managers, want %d", len(managers), expectedCount)
	}

	// Check that all expected managers are present
	expectedManagers := map[string]bool{
		"brew":     false,
		"npm":      false,
		"pnpm":     false,
		"cargo":    false,
		"uv":       false,
		"gem":      false,
		"go":       false,
		"pixi":     false,
		"composer": false,
		"dotnet":   false,
		"pipx":     false,
		"conda":    false,
	}

	for _, manager := range managers {
		if _, ok := expectedManagers[manager]; ok {
			expectedManagers[manager] = true
		} else {
			t.Errorf("GetValidManagers() returned unexpected manager: %s", manager)
		}
	}

	// Verify all expected managers were found
	for manager, found := range expectedManagers {
		if !found {
			t.Errorf("GetValidManagers() missing expected manager: %s", manager)
		}
	}
}

func TestGetMetadataString(t *testing.T) {
	tests := []struct {
		name     string
		result   resources.OperationResult
		key      string
		expected string
	}{
		{
			name: "valid string metadata",
			result: resources.OperationResult{
				Metadata: map[string]interface{}{
					"version": "1.2.3",
					"author":  "John Doe",
				},
			},
			key:      "version",
			expected: "1.2.3",
		},
		{
			name: "key not found",
			result: resources.OperationResult{
				Metadata: map[string]interface{}{
					"version": "1.2.3",
				},
			},
			key:      "missing",
			expected: "",
		},
		{
			name: "nil metadata",
			result: resources.OperationResult{
				Metadata: nil,
			},
			key:      "any",
			expected: "",
		},
		{
			name: "non-string value",
			result: resources.OperationResult{
				Metadata: map[string]interface{}{
					"count":   42,
					"enabled": true,
					"data":    []string{"a", "b"},
				},
			},
			key:      "count",
			expected: "",
		},
		{
			name: "empty string value",
			result: resources.OperationResult{
				Metadata: map[string]interface{}{
					"empty": "",
				},
			},
			key:      "empty",
			expected: "",
		},
		{
			name: "mixed case key",
			result: resources.OperationResult{
				Metadata: map[string]interface{}{
					"MixedCase": "value",
				},
			},
			key:      "MixedCase",
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMetadataString(tt.result, tt.key)
			if result != tt.expected {
				t.Errorf("GetMetadataString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

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
func TestValidateStatusFlags(t *testing.T) {
	tests := []struct {
		name          string
		showUnmanaged bool
		showMissing   bool
		wantErr       bool
		errContains   string
	}{
		{
			name:          "both false is valid",
			showUnmanaged: false,
			showMissing:   false,
			wantErr:       false,
		},
		{
			name:          "unmanaged only is valid",
			showUnmanaged: true,
			showMissing:   false,
			wantErr:       false,
		},
		{
			name:          "missing only is valid",
			showUnmanaged: false,
			showMissing:   true,
			wantErr:       false,
		},
		{
			name:          "both true is invalid",
			showUnmanaged: true,
			showMissing:   true,
			wantErr:       true,
			errContains:   "mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStatusFlags(tt.showUnmanaged, tt.showMissing)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
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
