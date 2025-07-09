// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"testing"
)

func TestItemState_String(t *testing.T) {
	tests := []struct {
		state    ItemState
		expected string
	}{
		{StateManaged, "managed"},
		{StateMissing, "missing"},
		{StateUntracked, "untracked"},
		{ItemState(999), "unknown"}, // Invalid state
	}

	for _, test := range tests {
		result := test.state.String()
		if result != test.expected {
			t.Errorf("ItemState(%d).String() = %s, expected %s", test.state, result, test.expected)
		}
	}
}

func TestResult_Count(t *testing.T) {
	result := Result{
		Domain: "test",
		Managed: []Item{
			{Name: "managed1", State: StateManaged},
			{Name: "managed2", State: StateManaged},
		},
		Missing: []Item{
			{Name: "missing1", State: StateMissing},
		},
		Untracked: []Item{
			{Name: "untracked1", State: StateUntracked},
			{Name: "untracked2", State: StateUntracked},
			{Name: "untracked3", State: StateUntracked},
		},
	}

	expected := 6 // 2 managed + 1 missing + 3 untracked
	count := result.Count()
	if count != expected {
		t.Errorf("Result.Count() = %d, expected %d", count, expected)
	}
}

func TestResult_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{
			name:     "empty result",
			result:   Result{Domain: "test"},
			expected: true,
		},
		{
			name: "result with managed items",
			result: Result{
				Domain:  "test",
				Managed: []Item{{Name: "test", State: StateManaged}},
			},
			expected: false,
		},
		{
			name: "result with missing items",
			result: Result{
				Domain:  "test",
				Missing: []Item{{Name: "test", State: StateMissing}},
			},
			expected: false,
		},
		{
			name: "result with untracked items",
			result: Result{
				Domain:    "test",
				Untracked: []Item{{Name: "test", State: StateUntracked}},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isEmpty := test.result.IsEmpty()
			if isEmpty != test.expected {
				t.Errorf("Result.IsEmpty() = %t, expected %t", isEmpty, test.expected)
			}
		})
	}
}

func TestResult_AddToSummary(t *testing.T) {
	result := Result{
		Domain: "test",
		Managed: []Item{
			{Name: "managed1", State: StateManaged},
			{Name: "managed2", State: StateManaged},
		},
		Missing: []Item{
			{Name: "missing1", State: StateMissing},
		},
		Untracked: []Item{
			{Name: "untracked1", State: StateUntracked},
			{Name: "untracked2", State: StateUntracked},
			{Name: "untracked3", State: StateUntracked},
		},
	}

	summary := Summary{}
	result.AddToSummary(&summary)

	if summary.TotalManaged != 2 {
		t.Errorf("Summary.TotalManaged = %d, expected 2", summary.TotalManaged)
	}
	if summary.TotalMissing != 1 {
		t.Errorf("Summary.TotalMissing = %d, expected 1", summary.TotalMissing)
	}
	if summary.TotalUntracked != 3 {
		t.Errorf("Summary.TotalUntracked = %d, expected 3", summary.TotalUntracked)
	}
	if len(summary.Results) != 1 {
		t.Errorf("len(Summary.Results) = %d, expected 1", len(summary.Results))
	}
	if summary.Results[0].Domain != "test" {
		t.Errorf("Summary.Results[0].Domain = %s, expected test", summary.Results[0].Domain)
	}
}

func TestSummary_MultipleResults(t *testing.T) {
	summary := Summary{}

	// Add first result
	result1 := Result{
		Domain:  "package",
		Managed: []Item{{Name: "git", State: StateManaged}},
		Missing: []Item{{Name: "curl", State: StateMissing}},
	}
	result1.AddToSummary(&summary)

	// Add second result
	result2 := Result{
		Domain:    "dotfile",
		Managed:   []Item{{Name: ".zshrc", State: StateManaged}},
		Untracked: []Item{{Name: ".vimrc", State: StateUntracked}},
	}
	result2.AddToSummary(&summary)

	// Verify aggregated counts
	if summary.TotalManaged != 2 {
		t.Errorf("Summary.TotalManaged = %d, expected 2", summary.TotalManaged)
	}
	if summary.TotalMissing != 1 {
		t.Errorf("Summary.TotalMissing = %d, expected 1", summary.TotalMissing)
	}
	if summary.TotalUntracked != 1 {
		t.Errorf("Summary.TotalUntracked = %d, expected 1", summary.TotalUntracked)
	}
	if len(summary.Results) != 2 {
		t.Errorf("len(Summary.Results) = %d, expected 2", len(summary.Results))
	}
}
