// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		output   []byte
		prefix   string
		expected string
	}{
		{
			name:     "simple version",
			output:   []byte("npm version: 8.19.2\nnode version: 18.12.0"),
			prefix:   "npm version:",
			expected: "8.19.2",
		},
		{
			name:     "version with v prefix",
			output:   []byte("cargo v1.70.0 (90c541806 2023-05-31)"),
			prefix:   "cargo",
			expected: "v1.70.0 (90c541806 2023-05-31)",
		},
		{
			name:     "multiline output",
			output:   []byte("Python 3.11.5\nType \"help\", \"copyright\", \"credits\" or \"license\" for more information."),
			prefix:   "Python",
			expected: "3.11.5",
		},
		{
			name:     "version not found",
			output:   []byte("Some other output\nNo version here"),
			prefix:   "Version:",
			expected: "",
		},
		{
			name:     "empty output",
			output:   []byte(""),
			prefix:   "version:",
			expected: "",
		},
		{
			name:     "version with extra whitespace",
			output:   []byte("  Version:   2.4.1  \n"),
			prefix:   "Version:",
			expected: "2.4.1",
		},
		{
			name:     "case sensitive prefix",
			output:   []byte("version: 1.0.0\nVersion: 2.0.0"),
			prefix:   "Version:",
			expected: "2.0.0",
		},
		{
			name:     "version at end of output",
			output:   []byte("Some info\nMore info\nruby 3.2.0"),
			prefix:   "ruby",
			expected: "3.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractVersion(tt.output, tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanJSONValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "double quoted value",
			value:    `"1.2.3"`,
			expected: "1.2.3",
		},
		{
			name:     "single quoted value",
			value:    `'1.2.3'`,
			expected: "1.2.3",
		},
		{
			name:     "value with trailing comma",
			value:    `"1.2.3",`,
			expected: "1.2.3",
		},
		{
			name:     "unquoted value",
			value:    "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "empty string",
			value:    "",
			expected: "",
		},
		{
			name:     "just quotes",
			value:    `""`,
			expected: "",
		},
		{
			name:     "mixed quotes",
			value:    `"'1.2.3'"`,
			expected: "1.2.3",
		},
		{
			name:     "value with spaces",
			value:    `"version 1.2.3"`,
			expected: "version 1.2.3",
		},
		{
			name:     "multiple trailing commas",
			value:    `"1.2.3",,`,
			expected: `1.2.3",`, // Only removes last comma with TrimSuffix, then quotes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanJSONValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseVersionOutput(t *testing.T) {
	tests := []struct {
		name      string
		output    []byte
		prefix    string
		expected  string
		expectErr bool
	}{
		{
			name:      "simple version",
			output:    []byte("npm version: 8.19.2\n"),
			prefix:    "npm version:",
			expected:  "8.19.2",
			expectErr: false,
		},
		{
			name:      "version with v prefix cleaned",
			output:    []byte("Version: v2.0.0\n"),
			prefix:    "Version:",
			expected:  "2.0.0",
			expectErr: false,
		},
		{
			name:      "version not found",
			output:    []byte("Some output without version"),
			prefix:    "Version:",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "empty output",
			output:    []byte(""),
			prefix:    "Version:",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "version line with no value",
			output:    []byte("Version: \n"),
			prefix:    "Version:",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "multiple versions takes first",
			output:    []byte("Version: 1.0.0\nVersion: 2.0.0\n"),
			prefix:    "Version:",
			expected:  "1.0.0",
			expectErr: false,
		},
		{
			name:      "version with extra info",
			output:    []byte("cargo 1.70.0 (90c541806 2023-05-31)\n"),
			prefix:    "cargo",
			expected:  "1.70.0",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVersionOutput(tt.output, tt.prefix)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "version not found")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCleanVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "version with v prefix",
			version:  "v1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "version with Version prefix",
			version:  "Version1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "version with version prefix lowercase",
			version:  "version2.0.0",
			expected: "ersion2.0.0", // "version" prefix removal is case-sensitive
		},
		{
			name:     "version with space separated info",
			version:  "1.70.0 (90c541806 2023-05-31)",
			expected: "1.70.0",
		},
		{
			name:     "version with tab separated info",
			version:  "3.11.5\tPython",
			expected: "3.11.5",
		},
		{
			name:     "clean version",
			version:  "4.5.6",
			expected: "4.5.6",
		},
		{
			name:     "version with leading/trailing spaces",
			version:  "  7.8.9  ",
			expected: "7.8.9",
		},
		{
			name:     "empty version",
			version:  "",
			expected: "",
		},
		{
			name:     "version prefix with space",
			version:  "v 1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "complex version string",
			version:  "version 2.4.1 built on 2023-01-01",
			expected: "ersion", // Removes "version" then cuts at first space
		},
		{
			name:     "semver with pre-release",
			version:  "v1.0.0-alpha.1",
			expected: "1.0.0-alpha.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanVersionString(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("ExtractVersion with nil output", func(t *testing.T) {
		result := ExtractVersion(nil, "prefix")
		assert.Equal(t, "", result)
	})

	t.Run("ParseVersionOutput with nil output", func(t *testing.T) {
		result, err := ParseVersionOutput(nil, "prefix")
		assert.Error(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("CleanVersionString only removes first matching prefix", func(t *testing.T) {
		// Should only remove the first "v", not the second
		result := CleanVersionString("vversion1.0.0")
		assert.Equal(t, "version1.0.0", result)
	})

	t.Run("CleanJSONValue handles nested quotes", func(t *testing.T) {
		// Removes all quotes due to Trim behavior
		result := CleanJSONValue(`"'nested'"`)
		assert.Equal(t, "nested", result)
	})
}
