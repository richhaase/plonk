// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/resources"
	"github.com/stretchr/testify/assert"
)

func TestCalculatePackageOperationSummary(t *testing.T) {
	tests := []struct {
		name     string
		results  []resources.OperationResult
		expected PackageOperationSummary
	}{
		{
			name:    "empty results",
			results: []resources.OperationResult{},
			expected: PackageOperationSummary{
				Succeeded: 0,
				Skipped:   0,
				Failed:    0,
			},
		},
		{
			name: "all succeeded",
			results: []resources.OperationResult{
				{Status: "added"},
				{Status: "removed"},
				{Status: "installed"},
				{Status: "uninstalled"},
				{Status: "success"},
			},
			expected: PackageOperationSummary{
				Succeeded: 5,
				Skipped:   0,
				Failed:    0,
			},
		},
		{
			name: "all skipped",
			results: []resources.OperationResult{
				{Status: "skipped"},
				{Status: "already-installed"},
				{Status: "already-configured"},
			},
			expected: PackageOperationSummary{
				Succeeded: 0,
				Skipped:   3,
				Failed:    0,
			},
		},
		{
			name: "all failed",
			results: []resources.OperationResult{
				{Status: "failed"},
				{Status: "error"},
			},
			expected: PackageOperationSummary{
				Succeeded: 0,
				Skipped:   0,
				Failed:    2,
			},
		},
		{
			name: "mixed results",
			results: []resources.OperationResult{
				{Status: "added"},
				{Status: "installed"},
				{Status: "skipped"},
				{Status: "already-installed"},
				{Status: "failed"},
				{Status: "error"},
			},
			expected: PackageOperationSummary{
				Succeeded: 2,
				Skipped:   2,
				Failed:    2,
			},
		},
		{
			name: "unknown status ignored",
			results: []resources.OperationResult{
				{Status: "added"},
				{Status: "unknown"},
				{Status: "pending"},
				{Status: "would-add"},
			},
			expected: PackageOperationSummary{
				Succeeded: 1,
				Skipped:   0,
				Failed:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePackageOperationSummary(tt.results)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewStandardTableBuilder(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "simple title",
			title:    "Test Title",
			expected: "Test Title\n==========\n\n",
		},
		{
			name:     "empty title",
			title:    "",
			expected: "",
		},
		{
			name:     "title with spaces",
			title:    "Multiple Word Title",
			expected: "Multiple Word Title\n===================\n\n",
		},
		{
			name:     "title with special characters",
			title:    "Title: With Special Characters!",
			expected: "Title: With Special Characters!\n===============================\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewStandardTableBuilder(tt.title)
			result := builder.Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStandardTableBuilder_Methods(t *testing.T) {
	t.Run("SetHeaders and AddRow", func(t *testing.T) {
		builder := NewStandardTableBuilder("Test")
		builder.SetHeaders("Name", "Status").
			AddRow("package1", "installed").
			AddRow("package2", "missing")
		result := builder.Build()
		assert.Contains(t, result, "Test")
		assert.Contains(t, result, "Name")
		assert.Contains(t, result, "Status")
		assert.Contains(t, result, "package1")
		assert.Contains(t, result, "installed")
	})

	t.Run("SetSummary", func(t *testing.T) {
		builder := NewStandardTableBuilder("Test")
		builder.SetHeaders("Item").
			AddRow("item1").
			SetSummary("Total: 1 item")
		result := builder.Build()
		assert.Contains(t, result, "Total: 1 item")
	})

	t.Run("Empty table", func(t *testing.T) {
		builder := NewStandardTableBuilder("Empty Table")
		result := builder.Build()
		expected := "Empty Table\n===========\n\n"
		assert.Equal(t, expected, result)
	})

	t.Run("Table without headers", func(t *testing.T) {
		builder := NewStandardTableBuilder("No Headers")
		builder.AddRow("row1", "data1").
			AddRow("row2", "data2")
		result := builder.Build()
		assert.Contains(t, result, "No Headers")
		assert.Contains(t, result, "row1")
		assert.Contains(t, result, "data1")
	})

	t.Run("Complex table", func(t *testing.T) {
		builder := NewStandardTableBuilder("Package Status")
		builder.SetHeaders("Package", "Manager", "Status").
			AddRow("htop", "brew", "installed").
			AddRow("vim", "brew", "installed").
			AddRow("jq", "brew", "missing").
			SetSummary("Total: 3 packages (2 installed, 1 missing)")

		result := builder.Build()
		assert.Contains(t, result, "Package Status")
		assert.Contains(t, result, "Package")
		assert.Contains(t, result, "Manager")
		assert.Contains(t, result, "Status")
		assert.Contains(t, result, "htop")
		assert.Contains(t, result, "Total: 3 packages")
	})
}

func TestConvertOperationResults(t *testing.T) {
	tests := []struct {
		name     string
		results  []resources.OperationResult
		expected []SerializableOperationResult
	}{
		{
			name:     "empty results",
			results:  []resources.OperationResult{},
			expected: []SerializableOperationResult{},
		},
		{
			name: "single result with all fields",
			results: []resources.OperationResult{
				{
					Name:           "htop",
					Manager:        "brew",
					Status:         "installed",
					AlreadyManaged: false,
					Error:          nil,
				},
			},
			expected: []SerializableOperationResult{
				{
					Name:           "htop",
					Manager:        "brew",
					Status:         "installed",
					AlreadyManaged: false,
					Error:          "",
				},
			},
		},
		{
			name: "result with error",
			results: []resources.OperationResult{
				{
					Name:           "invalid-package",
					Manager:        "npm",
					Status:         "failed",
					AlreadyManaged: false,
					Error:          assert.AnError,
				},
			},
			expected: []SerializableOperationResult{
				{
					Name:           "invalid-package",
					Manager:        "npm",
					Status:         "failed",
					AlreadyManaged: false,
					Error:          "assert.AnError general error for testing",
				},
			},
		},
		{
			name: "multiple results",
			results: []resources.OperationResult{
				{
					Name:           "package1",
					Manager:        "pip",
					Status:         "added",
					AlreadyManaged: false,
					Error:          nil,
				},
				{
					Name:           "package2",
					Manager:        "gem",
					Status:         "skipped",
					AlreadyManaged: true,
					Error:          nil,
				},
			},
			expected: []SerializableOperationResult{
				{
					Name:           "package1",
					Manager:        "pip",
					Status:         "added",
					AlreadyManaged: false,
					Error:          "",
				},
				{
					Name:           "package2",
					Manager:        "gem",
					Status:         "skipped",
					AlreadyManaged: true,
					Error:          "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertOperationResults(tt.results)
			assert.Equal(t, tt.expected, result)
		})
	}
}
