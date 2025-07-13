// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"strings"
	"testing"
)

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		// Success statuses
		{"managed status", "managed", IconSuccess},
		{"added status", "added", IconSuccess},
		{"installed status", "installed", IconSuccess},
		{"removed status", "removed", IconSuccess},
		{"success status", "success", IconSuccess},
		{"completed status", "completed", IconSuccess},

		// Warning statuses
		{"missing status", "missing", IconWarning},
		{"warn status", "warn", IconWarning},
		{"warning status", "warning", IconWarning},

		// Error statuses
		{"failed status", "failed", IconError},
		{"error status", "error", IconError},
		{"fail status", "fail", IconError},

		// Unknown statuses
		{"untracked status", "untracked", IconUnknown},
		{"unknown status", "unknown", IconUnknown},

		// Info statuses
		{"skipped status", "skipped", IconInfo},
		{"already-configured status", "already-configured", IconInfo},
		{"already-installed status", "already-installed", IconInfo},

		// Default
		{"unrecognized status", "something-else", IconSkipped},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatusIcon(tt.status)
			if result != tt.expected {
				t.Errorf("GetStatusIcon(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

func TestGetActionIcon(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected string
	}{
		{"error action", "Package installation error", IconError},
		{"failed action", "Command failed with exit code 1", IconError},
		{"already action", "Package already installed", IconInfo},
		{"not found action", "Package not found in registry", IconInfo},
		{"success action", "Package installed successfully", IconSuccess},
		{"generic action", "Processing package", IconSuccess},
		// Case insensitive
		{"ERROR uppercase", "ERROR: something went wrong", IconError},
		{"Failed mixed case", "Failed to process", IconError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetActionIcon(tt.action)
			if result != tt.expected {
				t.Errorf("GetActionIcon(%q) = %q, want %q", tt.action, result, tt.expected)
			}
		})
	}
}

func TestTableBuilder(t *testing.T) {
	t.Run("basic table construction", func(t *testing.T) {
		tb := NewTableBuilder()
		output := tb.
			AddTitle("Test Title").
			AddNewline().
			AddLine("Test line 1").
			AddLine("Test line %d", 2).
			Build()

		expected := `Test Title
==========

Test line 1
Test line 2
`
		if output != expected {
			t.Errorf("TableBuilder output mismatch:\ngot:\n%s\nwant:\n%s", output, expected)
		}
	})

	t.Run("action list", func(t *testing.T) {
		tb := NewTableBuilder()
		actions := []string{
			"Package installed successfully",
			"Package already exists",
			"Installation failed",
		}
		output := tb.AddActionList(actions).Build()

		expected := `✓ Package installed successfully
• Package already exists
✗ Installation failed
`
		if output != expected {
			t.Errorf("TableBuilder action list mismatch:\ngot:\n%s\nwant:\n%s", output, expected)
		}
	})

	t.Run("summary line", func(t *testing.T) {
		tb := NewTableBuilder()
		counts := map[string]int{
			"added":   3,
			"skipped": 1,
			"failed":  0,
		}
		output := tb.AddSummaryLine("Summary:", counts).Build()

		// The order might vary due to map iteration, so check parts
		if !strings.Contains(output, "Summary:") {
			t.Error("Summary line should contain 'Summary:'")
		}
		if !strings.Contains(output, "3 added") {
			t.Error("Summary line should contain '3 added'")
		}
		if !strings.Contains(output, "1 skipped") {
			t.Error("Summary line should contain '1 skipped'")
		}
		if strings.Contains(output, "0 failed") {
			t.Error("Summary line should not contain '0 failed'")
		}
	})
}

func TestCalculateOperationSummary(t *testing.T) {
	tests := []struct {
		name         string
		statusCounts map[string]int
		expected     OperationSummary
	}{
		{
			name: "mixed statuses",
			statusCounts: map[string]int{
				"added":              3,
				"already-configured": 2,
				"failed":             1,
			},
			expected: OperationSummary{
				Total:     6,
				Succeeded: 3,
				Skipped:   2,
				Failed:    1,
			},
		},
		{
			name: "all success",
			statusCounts: map[string]int{
				"installed": 5,
				"updated":   2,
			},
			expected: OperationSummary{
				Total:     7,
				Succeeded: 7,
				Skipped:   0,
				Failed:    0,
			},
		},
		{
			name: "would-actions counted as success",
			statusCounts: map[string]int{
				"would-add":    2,
				"would-remove": 1,
			},
			expected: OperationSummary{
				Total:     3,
				Succeeded: 3,
				Skipped:   0,
				Failed:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateOperationSummary(tt.statusCounts)
			if result != tt.expected {
				t.Errorf("CalculateOperationSummary() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestFormatTableHeader(t *testing.T) {
	header := FormatTableHeader("Status", "Package", "Manager")
	expected := `Status Package Manager
------ ------- -------
`
	if header != expected {
		t.Errorf("FormatTableHeader() = %q, want %q", header, expected)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"longer than max", "hello world", 8, "hello..."},
		{"very short max", "hello", 2, "he"},
		{"zero length", "hello", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}
