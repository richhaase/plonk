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
