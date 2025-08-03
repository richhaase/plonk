// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemState_String(t *testing.T) {
	tests := []struct {
		state    ItemState
		expected string
	}{
		{StateManaged, "managed"},
		{StateMissing, "missing"},
		{StateUntracked, "untracked"},
		{StateDegraded, "drifted"},
		{ItemState(999), "unknown"}, // invalid state
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestResult_Count(t *testing.T) {
	result := Result{
		Managed:   []Item{{}, {}, {}}, // 3 items
		Missing:   []Item{{}, {}},     // 2 items
		Untracked: []Item{{}},         // 1 item
	}

	assert.Equal(t, 6, result.Count())
}

func TestResult_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{
			name:     "empty result",
			result:   Result{},
			expected: true,
		},
		{
			name: "result with managed items",
			result: Result{
				Managed: []Item{{}},
			},
			expected: false,
		},
		{
			name: "result with missing items",
			result: Result{
				Missing: []Item{{}},
			},
			expected: false,
		},
		{
			name: "result with untracked items",
			result: Result{
				Untracked: []Item{{}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.IsEmpty())
		})
	}
}

func TestResult_AddToSummary(t *testing.T) {
	// Create initial summary
	summary := Summary{
		TotalManaged:   5,
		TotalMissing:   2,
		TotalUntracked: 3,
		Results: []Result{
			{
				Domain:    "packages",
				Manager:   "brew",
				Managed:   []Item{{}, {}, {}},
				Missing:   []Item{{}},
				Untracked: []Item{{}},
			},
		},
	}

	// Create a new result to add
	newResult := Result{
		Domain:    "packages",
		Manager:   "npm",
		Managed:   []Item{{}, {}}, // 2 items
		Missing:   []Item{{}},     // 1 item
		Untracked: []Item{{}, {}}, // 2 items
	}

	// Add the result
	newResult.AddToSummary(&summary)

	// Verify the summary was updated
	assert.Equal(t, 7, summary.TotalManaged)   // 5 + 2
	assert.Equal(t, 3, summary.TotalMissing)   // 2 + 1
	assert.Equal(t, 5, summary.TotalUntracked) // 3 + 2
	assert.Len(t, summary.Results, 2)          // Should have 2 results now
}

func TestCalculateSummary(t *testing.T) {
	results := []OperationResult{
		{Status: "added"},
		{Status: "added"},
		{Status: "would-add"},
		{Status: "updated"},
		{Status: "would-update"},
		{Status: "removed"},
		{Status: "would-remove"},
		{Status: "unlinked"},
		{Status: "would-unlink"},
		{Status: "skipped"},
		{Status: "failed"},
		{Status: "unknown"}, // should not match any case
		{Status: "added", FilesProcessed: 5},
		{Status: "updated", FilesProcessed: 3},
	}

	summary := CalculateSummary(results)

	assert.Equal(t, 14, summary.Total)
	assert.Equal(t, 4, summary.Added)    // 2 added + 1 would-add + 1 added with files
	assert.Equal(t, 3, summary.Updated)  // 1 updated + 1 would-update + 1 updated with files
	assert.Equal(t, 2, summary.Removed)  // 1 removed + 1 would-remove
	assert.Equal(t, 2, summary.Unlinked) // 1 unlinked + 1 would-unlink
	assert.Equal(t, 1, summary.Skipped)
	assert.Equal(t, 1, summary.Failed)
	assert.Equal(t, 8, summary.FilesProcessed) // 5 + 3
}

func TestConvertResultsToSummary(t *testing.T) {
	results := map[string]Result{
		"packages": {
			Domain:  "packages",
			Manager: "brew",
			Managed: []Item{
				{Name: "package1"},
				{Name: "package2"},
				{Name: "package3"},
			},
			Missing: []Item{
				{Name: "package4"},
			},
			Untracked: []Item{
				{Name: "package5"},
				{Name: "package6"},
			},
		},
		"dotfiles": {
			Domain: "dotfiles",
			Managed: []Item{
				{Name: "file1"},
				{Name: "file2"},
			},
			Missing: []Item{},
			Untracked: []Item{
				{Name: "file3"},
			},
		},
	}

	summary := ConvertResultsToSummary(results)

	assert.Equal(t, 5, summary.TotalManaged)   // 3 + 2
	assert.Equal(t, 1, summary.TotalMissing)   // 1 + 0
	assert.Equal(t, 3, summary.TotalUntracked) // 2 + 1
	assert.Len(t, summary.Results, 2)

	// Find and verify each result
	for _, result := range summary.Results {
		if result.Domain == "packages" {
			assert.Equal(t, "brew", result.Manager)
			assert.Len(t, result.Managed, 3)
			assert.Len(t, result.Missing, 1)
			assert.Len(t, result.Untracked, 2)
		} else if result.Domain == "dotfiles" {
			assert.Empty(t, result.Manager)
			assert.Len(t, result.Managed, 2)
			assert.Len(t, result.Missing, 0)
			assert.Len(t, result.Untracked, 1)
		}
	}
}

func TestCreateDomainSummary(t *testing.T) {
	results := []Result{
		{
			Domain:    "packages",
			Manager:   "brew",
			Managed:   []Item{{}, {}}, // 2 items
			Missing:   []Item{{}},     // 1 item
			Untracked: []Item{{}},     // 1 item
		},
		{
			Domain:    "packages",
			Manager:   "npm",
			Managed:   []Item{{}}, // 1 item
			Missing:   []Item{},   // 0 items
			Untracked: []Item{{}}, // 1 item
		},
		{
			Domain:    "dotfiles",
			Manager:   "",
			Managed:   []Item{{}, {}}, // 2 items
			Missing:   []Item{{}},     // 1 item
			Untracked: []Item{},       // 0 items
		},
		{
			Domain:    "empty",
			Manager:   "test",
			Managed:   []Item{},
			Missing:   []Item{},
			Untracked: []Item{},
		}, // This should be skipped because it's empty
	}

	summaries := CreateDomainSummary(results)

	assert.Len(t, summaries, 3) // brew, npm, dotfiles (empty skipped)

	// Find and verify each summary
	for _, summary := range summaries {
		switch summary.Domain {
		case "packages":
			if summary.Manager == "brew" {
				assert.Equal(t, 2, summary.ManagedCount)
				assert.Equal(t, 1, summary.MissingCount)
				assert.Equal(t, 1, summary.UntrackedCount)
			} else if summary.Manager == "npm" {
				assert.Equal(t, 1, summary.ManagedCount)
				assert.Equal(t, 0, summary.MissingCount)
				assert.Equal(t, 1, summary.UntrackedCount)
			}
		case "dotfiles":
			assert.Empty(t, summary.Manager)
			assert.Equal(t, 2, summary.ManagedCount)
			assert.Equal(t, 1, summary.MissingCount)
			assert.Equal(t, 0, summary.UntrackedCount)
		}
	}
}

func TestValidateOperationResults(t *testing.T) {
	t.Run("valid results", func(t *testing.T) {
		results := []OperationResult{
			{Status: "added"},
			{Status: "updated"},
		}
		assert.NoError(t, ValidateOperationResults(results, "apply"))
	})

	t.Run("empty results", func(t *testing.T) {
		results := []OperationResult{}
		assert.NoError(t, ValidateOperationResults(results, "apply"))
	})

	t.Run("all failed", func(t *testing.T) {
		results := []OperationResult{
			{Status: "failed"},
			{Status: "failed"},
		}
		err := ValidateOperationResults(results, "apply")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "apply operation failed")
	})

	t.Run("some succeeded", func(t *testing.T) {
		results := []OperationResult{
			{Status: "failed"},
			{Status: "added"},
			{Status: "failed"},
		}
		assert.NoError(t, ValidateOperationResults(results, "update"))
	})
}
