// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		expected    OutputFormat
		expectError bool
		errorMsg    string
	}{
		{
			name:        "table format",
			format:      "table",
			expected:    OutputTable,
			expectError: false,
		},
		{
			name:        "json format",
			format:      "json",
			expected:    OutputJSON,
			expectError: false,
		},
		{
			name:        "yaml format",
			format:      "yaml",
			expected:    OutputYAML,
			expectError: false,
		},
		{
			name:        "invalid format",
			format:      "xml",
			expected:    OutputTable, // default on error
			expectError: true,
			errorMsg:    "unsupported format 'xml' (use: table, json, or yaml)",
		},
		{
			name:        "empty format",
			format:      "",
			expected:    OutputTable, // default on error
			expectError: true,
			errorMsg:    "unsupported format '' (use: table, json, or yaml)",
		},
		{
			name:        "uppercase format",
			format:      "JSON",
			expected:    OutputTable, // default on error
			expectError: true,
			errorMsg:    "unsupported format 'JSON' (use: table, json, or yaml)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseOutputFormat(tt.format)

			if tt.expectError {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}
