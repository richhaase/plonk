// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*TableBuilder)
		expected string
	}{
		{
			name: "empty builder returns empty string",
			build: func(tb *TableBuilder) {
				// Don't add anything
			},
			expected: "",
		},
		{
			name: "single title",
			build: func(tb *TableBuilder) {
				tb.AddTitle("Test Title")
			},
			expected: "Test Title\n==========\n",
		},
		{
			name: "title with different lengths",
			build: func(tb *TableBuilder) {
				tb.AddTitle("Short")
			},
			expected: "Short\n=====\n",
		},
		{
			name: "single line",
			build: func(tb *TableBuilder) {
				tb.AddLine("Hello %s", "World")
			},
			expected: "Hello World\n",
		},
		{
			name: "multiple lines",
			build: func(tb *TableBuilder) {
				tb.AddLine("Line 1")
				tb.AddLine("Line 2")
				tb.AddLine("Line 3")
			},
			expected: "Line 1\nLine 2\nLine 3\n",
		},
		{
			name: "newline adds empty line",
			build: func(tb *TableBuilder) {
				tb.AddLine("Before")
				tb.AddNewline()
				tb.AddLine("After")
			},
			expected: "Before\n\nAfter\n",
		},
		{
			name: "complex output",
			build: func(tb *TableBuilder) {
				tb.AddTitle("Package Status")
				tb.AddLine("Total: %d", 5)
				tb.AddLine("Installed: %d", 3)
				tb.AddLine("Missing: %d", 2)
				tb.AddNewline()
				tb.AddTitle("Next Steps")
				tb.AddLine("Run 'plonk apply' to install missing packages")
			},
			expected: "Package Status\n==============\nTotal: 5\nInstalled: 3\nMissing: 2\n\nNext Steps\n==========\nRun 'plonk apply' to install missing packages\n",
		},
		{
			name: "chaining methods",
			build: func(tb *TableBuilder) {
				tb.AddTitle("Chained").
					AddLine("Line 1").
					AddNewline().
					AddLine("Line 2")
			},
			expected: "Chained\n=======\nLine 1\n\nLine 2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := NewTableBuilder()
			tt.build(tb)
			result := tb.Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTableBuilder_FormattingEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "no format args",
			format:   "Plain text",
			args:     []interface{}{},
			expected: "Plain text\n",
		},
		{
			name:     "multiple format specifiers",
			format:   "%s: %d/%d",
			args:     []interface{}{"Progress", 5, 10},
			expected: "Progress: 5/10\n",
		},
		{
			name:     "escaped percent",
			format:   "100%% complete",
			args:     []interface{}{},
			expected: "100% complete\n",
		},
		{
			name:     "unicode in title",
			format:   "✓ Success",
			args:     []interface{}{},
			expected: "✓ Success\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := NewTableBuilder()
			tb.AddLine(tt.format, tt.args...)
			result := tb.Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewTableBuilder(t *testing.T) {
	tb := NewTableBuilder()
	assert.NotNil(t, tb)
	assert.Equal(t, "", tb.Build(), "new builder should produce empty string")
}

func TestFormatValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		expected string
		want     string
	}{
		{
			name:     "simple validation error",
			field:    "manager",
			value:    "unknown",
			expected: "must be one of: brew, npm, cargo",
			want:     `invalid manager "unknown": must be one of: brew, npm, cargo`,
		},
		{
			name:     "timeout validation",
			field:    "timeout",
			value:    "-5s",
			expected: "must be positive",
			want:     `invalid timeout "-5s": must be positive`,
		},
		{
			name:     "empty value",
			field:    "name",
			value:    "",
			expected: "cannot be empty",
			want:     `invalid name "": cannot be empty`,
		},
		{
			name:     "special characters in value",
			field:    "path",
			value:    "path/with spaces",
			expected: "must not contain spaces",
			want:     `invalid path "path/with spaces": must not contain spaces`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValidationError(tt.field, tt.value, tt.expected)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFormatNotFoundError(t *testing.T) {
	tests := []struct {
		name        string
		itemType    string
		itemName    string
		suggestions []string
		want        string
	}{
		{
			name:        "no suggestions",
			itemType:    "package",
			itemName:    "unknown",
			suggestions: []string{},
			want:        `package "unknown" not found`,
		},
		{
			name:        "single suggestion",
			itemType:    "command",
			itemName:    "installl",
			suggestions: []string{"install"},
			want:        "command \"installl\" not found\nDid you mean: install",
		},
		{
			name:        "multiple suggestions",
			itemType:    "manager",
			itemName:    "pip3",
			suggestions: []string{"pip", "npm", "gem"},
			want:        "manager \"pip3\" not found\nValid options: pip, npm, gem",
		},
		{
			name:        "empty suggestions slice not nil",
			itemType:    "file",
			itemName:    "missing.txt",
			suggestions: []string{},
			want:        `file "missing.txt" not found`,
		},
		{
			name:        "special characters in name",
			itemType:    "path",
			itemName:    "~/config/file",
			suggestions: []string{"~/.config/file"},
			want:        "path \"~/config/file\" not found\nDid you mean: ~/.config/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNotFoundError(tt.itemType, tt.itemName, tt.suggestions)
			assert.Equal(t, tt.want, result)
		})
	}
}
