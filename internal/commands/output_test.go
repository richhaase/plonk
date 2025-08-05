// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/output"
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
	outputData := DotfileRemovalOutput{
		TotalFiles: 3,
		Summary: DotfileRemovalSummary{
			Removed: 2,
			Skipped: 0,
			Failed:  1,
		},
	}

	// Convert to output package type and create formatter
	formatterData := output.DotfileRemovalOutput{
		TotalFiles: outputData.TotalFiles,
		Results:    outputData.Results,
		Summary: output.DotfileRemovalSummary{
			Removed: outputData.Summary.Removed,
			Skipped: outputData.Summary.Skipped,
			Failed:  outputData.Summary.Failed,
		},
	}
	formatter := output.NewDotfileRemovalFormatter(formatterData)
	result := formatter.StructuredData()

	// The structured data should match the formatter data
	assert.Equal(t, formatterData, result)
}

func TestSearchOutput_StructuredData(t *testing.T) {
	outputData := SearchOutput{
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

	// Convert to output package type and create formatter
	formatterData := output.SearchOutput{
		Package: outputData.Package,
		Status:  outputData.Status,
		Message: outputData.Message,
		Results: convertSearchResults(outputData.Results),
	}
	formatter := output.NewSearchFormatter(formatterData)
	result := formatter.StructuredData()

	// The structured data should match the formatter data
	assert.Equal(t, formatterData, result)
}

func TestStatusOutput_StructuredData(t *testing.T) {
	outputData := StatusOutput{
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
		ShowPackages:  true,
		ShowDotfiles:  true,
		ShowUnmanaged: false,
		ShowMissing:   false,
		ConfigDir:     "/home/user/.config/plonk",
	}

	// Convert to output package type and create formatter
	formatterData := output.StatusOutput{
		ConfigPath:    outputData.ConfigPath,
		LockPath:      outputData.LockPath,
		ConfigExists:  outputData.ConfigExists,
		ConfigValid:   outputData.ConfigValid,
		LockExists:    outputData.LockExists,
		StateSummary:  convertSummary(outputData.StateSummary),
		ShowPackages:  outputData.ShowPackages,
		ShowDotfiles:  outputData.ShowDotfiles,
		ShowUnmanaged: outputData.ShowUnmanaged,
		ShowMissing:   outputData.ShowMissing,
		ConfigDir:     outputData.ConfigDir,
	}
	formatter := output.NewStatusFormatter(formatterData)
	result := formatter.StructuredData()

	// Check that the result structure is correct
	assert.NotNil(t, result)
}
