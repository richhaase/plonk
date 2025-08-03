// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEditor(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "VISUAL set",
			envVars:  map[string]string{"VISUAL": "nvim", "EDITOR": "vim"},
			expected: "nvim",
		},
		{
			name:     "only EDITOR set",
			envVars:  map[string]string{"EDITOR": "emacs"},
			expected: "emacs",
		},
		{
			name:     "neither set, default to vim",
			envVars:  map[string]string{},
			expected: "vim",
		},
		{
			name:     "VISUAL empty, use EDITOR",
			envVars:  map[string]string{"VISUAL": "", "EDITOR": "nano"},
			expected: "nano",
		},
		{
			name:     "both empty, use default",
			envVars:  map[string]string{"VISUAL": "", "EDITOR": ""},
			expected: "vim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origVisual := os.Getenv("VISUAL")
			origEditor := os.Getenv("EDITOR")
			defer func() {
				os.Setenv("VISUAL", origVisual)
				os.Setenv("EDITOR", origEditor)
			}()

			// Clear env vars
			os.Unsetenv("VISUAL")
			os.Unsetenv("EDITOR")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			result := getEditor()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name      string
		configDir string
		expected  string
	}{
		{
			name:      "simple path",
			configDir: "/home/user/.config/plonk",
			expected:  "/home/user/.config/plonk/plonk.yaml",
		},
		{
			name:      "path with trailing slash",
			configDir: "/home/user/.config/plonk/",
			expected:  "/home/user/.config/plonk/plonk.yaml",
		},
		{
			name:      "relative path",
			configDir: "config",
			expected:  "config/plonk.yaml",
		},
		{
			name:      "empty path",
			configDir: "",
			expected:  "plonk.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getConfigPath(tt.configDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}
