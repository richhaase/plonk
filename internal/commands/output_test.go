// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/resources"
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

func TestDotfileRemovalOutput_StructuredData(t *testing.T) {
	output := DotfileRemovalOutput{
		TotalFiles: 3,
		Summary: DotfileRemovalSummary{
			Removed: 2,
			Skipped: 0,
			Failed:  1,
		},
	}

	result := output.StructuredData()
	assert.Equal(t, output, result)
}

func TestSearchOutput_StructuredData(t *testing.T) {
	output := SearchOutput{
		Package: "ripgrep",
		Status:  "found",
		Message: "Found package 'ripgrep' in brew",
		Results: []SearchResultEntry{
			{
				Manager:  "brew",
				Packages: []string{"ripgrep"},
			},
		},
	}

	result := output.StructuredData()
	assert.Equal(t, output, result)
}

func TestStatusOutput_StructuredData(t *testing.T) {
	output := StatusOutput{
		ConfigPath:   "/home/user/.config/plonk/plonk.yaml",
		LockPath:     "/home/user/.config/plonk/plonk.lock",
		ConfigExists: true,
		ConfigValid:  true,
		LockExists:   true,
		StateSummary: resources.Summary{
			TotalManaged:   10,
			TotalMissing:   2,
			TotalUntracked: 3,
			Results:        []resources.Result{},
		},
	}

	result := output.StructuredData()

	// The StructuredData method returns StatusOutputSummary, not a pointer
	summary, ok := result.(StatusOutputSummary)
	assert.True(t, ok)
	assert.Equal(t, output.ConfigPath, summary.ConfigPath)
	assert.Equal(t, output.LockPath, summary.LockPath)
	assert.Equal(t, output.ConfigExists, summary.ConfigExists)
	assert.Equal(t, output.ConfigValid, summary.ConfigValid)
	assert.Equal(t, output.LockExists, summary.LockExists)
	assert.Equal(t, 10, summary.Summary.TotalManaged)
	assert.Equal(t, 2, summary.Summary.TotalMissing)
	assert.Equal(t, 3, summary.Summary.TotalUntracked)
}
