// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/richhaase/plonk/internal/testutil"
)

func TestInitColors(t *testing.T) {
	// Save original state
	originalWriter := writer
	originalNoColor := color.NoColor
	defer func() {
		writer = originalWriter
		color.NoColor = originalNoColor
	}()

	tests := []struct {
		name        string
		isTerminal  bool
		wantNoColor bool
	}{
		{
			name:        "terminal output enables colors",
			isTerminal:  true,
			wantNoColor: false,
		},
		{
			name:        "non-terminal output disables colors",
			isTerminal:  false,
			wantNoColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test writer
			writer = testutil.NewBufferWriter(tt.isTerminal)
			color.NoColor = false // Reset state

			InitColors()

			if color.NoColor != tt.wantNoColor {
				t.Errorf("color.NoColor = %v, want %v",
					color.NoColor, tt.wantNoColor)
			}
		})
	}
}

func TestStatusWords(t *testing.T) {
	// Save original state
	originalNoColor := color.NoColor
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Test with colors enabled
	color.NoColor = false

	tests := []struct {
		name      string
		fn        func() string
		wantText  string
		wantColor bool
	}{
		{
			name:      "Available shows green",
			fn:        Available,
			wantText:  "available",
			wantColor: true,
		},
		{
			name:      "Missing shows red",
			fn:        Missing,
			wantText:  "missing",
			wantColor: true,
		},
		{
			name:      "Drifted shows yellow",
			fn:        Drifted,
			wantText:  "drifted",
			wantColor: true,
		},
		{
			name:      "Deployed shows green",
			fn:        Deployed,
			wantText:  "deployed",
			wantColor: true,
		},
		{
			name:      "Managed shows green",
			fn:        Managed,
			wantText:  "managed",
			wantColor: true,
		},
		{
			name:      "Success shows green",
			fn:        Success,
			wantText:  "success",
			wantColor: true,
		},
		{
			name:      "Valid shows green",
			fn:        Valid,
			wantText:  "Valid",
			wantColor: true,
		},
		{
			name:      "Invalid shows red",
			fn:        Invalid,
			wantText:  "Invalid",
			wantColor: true,
		},
		{
			name:      "NotAvailable shows red",
			fn:        NotAvailable,
			wantText:  "not available",
			wantColor: true,
		},
		{
			name:      "Unmanaged shows yellow",
			fn:        Unmanaged,
			wantText:  "unmanaged",
			wantColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()

			// Check text content
			if !strings.Contains(result, tt.wantText) {
				t.Errorf("result %q does not contain %q",
					result, tt.wantText)
			}

			// With colors enabled, output should have ANSI codes
			if tt.wantColor && !strings.Contains(result, "\033[") {
				t.Error("expected colored output but got plain text")
			}
		})
	}
}

func TestStatusWordsNoColor(t *testing.T) {
	// Save original state
	originalNoColor := color.NoColor
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Test with colors disabled
	color.NoColor = true

	tests := []struct {
		name     string
		fn       func() string
		wantText string
	}{
		{
			name:     "Available returns plain text",
			fn:       Available,
			wantText: "available",
		},
		{
			name:     "Missing returns plain text",
			fn:       Missing,
			wantText: "missing",
		},
		{
			name:     "Invalid returns plain text",
			fn:       Invalid,
			wantText: "Invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()

			// Should be exactly the text, no ANSI codes
			if result != tt.wantText {
				t.Errorf("got %q, want %q", result, tt.wantText)
			}

			// Ensure no ANSI codes
			if strings.Contains(result, "\033[") {
				t.Error("unexpected ANSI codes in output when colors are disabled")
			}
		})
	}
}

func TestColorHelpers(t *testing.T) {
	// Save original state
	originalNoColor := color.NoColor
	defer func() {
		color.NoColor = originalNoColor
	}()

	tests := []struct {
		name      string
		fn        func(string) string
		input     string
		wantText  string
		wantColor bool
	}{
		{
			name:      "ColorError applies red",
			fn:        ColorError,
			input:     "test error",
			wantText:  "test error",
			wantColor: true,
		},
		{
			name:      "ColorInfo applies blue",
			fn:        ColorInfo,
			input:     "test info",
			wantText:  "test info",
			wantColor: true,
		},
	}

	// Test with colors enabled
	color.NoColor = false

	for _, tt := range tests {
		t.Run(tt.name+" with color", func(t *testing.T) {
			result := tt.fn(tt.input)

			if !strings.Contains(result, tt.wantText) {
				t.Errorf("result %q does not contain %q",
					result, tt.wantText)
			}

			if tt.wantColor && !strings.Contains(result, "\033[") {
				t.Error("expected colored output but got plain text")
			}
		})
	}

	// Test with colors disabled
	color.NoColor = true

	for _, tt := range tests {
		t.Run(tt.name+" without color", func(t *testing.T) {
			result := tt.fn(tt.input)

			if result != tt.wantText {
				t.Errorf("got %q, want %q", result, tt.wantText)
			}
		})
	}
}
