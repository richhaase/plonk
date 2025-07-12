// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package operations

import (
	"testing"
)

func TestCalculateSummary(t *testing.T) {
	tests := []struct {
		name     string
		results  []OperationResult
		expected ResultSummary
	}{
		{
			name: "mixed results",
			results: []OperationResult{
				{Status: "added", FilesProcessed: 1},
				{Status: "updated", FilesProcessed: 1},
				{Status: "skipped"},
				{Status: "failed"},
				{Status: "added", FilesProcessed: 3},
			},
			expected: ResultSummary{
				Total:          5,
				Added:          2,
				Updated:        1,
				Skipped:        1,
				Failed:         1,
				FilesProcessed: 5,
			},
		},
		{
			name:    "empty results",
			results: []OperationResult{},
			expected: ResultSummary{
				Total: 0,
			},
		},
		{
			name: "all failed",
			results: []OperationResult{
				{Status: "failed"},
				{Status: "failed"},
			},
			expected: ResultSummary{
				Total:  2,
				Failed: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSummary(tt.results)

			if result.Total != tt.expected.Total {
				t.Errorf("Total: got %d, want %d", result.Total, tt.expected.Total)
			}
			if result.Added != tt.expected.Added {
				t.Errorf("Added: got %d, want %d", result.Added, tt.expected.Added)
			}
			if result.Updated != tt.expected.Updated {
				t.Errorf("Updated: got %d, want %d", result.Updated, tt.expected.Updated)
			}
			if result.Skipped != tt.expected.Skipped {
				t.Errorf("Skipped: got %d, want %d", result.Skipped, tt.expected.Skipped)
			}
			if result.Failed != tt.expected.Failed {
				t.Errorf("Failed: got %d, want %d", result.Failed, tt.expected.Failed)
			}
			if result.FilesProcessed != tt.expected.FilesProcessed {
				t.Errorf("FilesProcessed: got %d, want %d", result.FilesProcessed, tt.expected.FilesProcessed)
			}
		})
	}
}

func TestCountByStatus(t *testing.T) {
	results := []OperationResult{
		{Status: "added"},
		{Status: "failed"},
		{Status: "added"},
		{Status: "skipped"},
		{Status: "failed"},
	}

	tests := []struct {
		status   string
		expected int
	}{
		{"added", 2},
		{"failed", 2},
		{"skipped", 1},
		{"updated", 0},
		{"nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := CountByStatus(results, tt.status)
			if result != tt.expected {
				t.Errorf("CountByStatus(%q): got %d, want %d", tt.status, result, tt.expected)
			}
		})
	}
}
