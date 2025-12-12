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

func TestSuccess(t *testing.T) {
	// Save original state
	originalNoColor := color.NoColor
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Test with colors enabled
	color.NoColor = false
	result := Success()
	if !strings.Contains(result, "success") {
		t.Errorf("result %q does not contain 'success'", result)
	}
	if !strings.Contains(result, "\033[") {
		t.Error("expected colored output but got plain text")
	}

	// Test with colors disabled
	color.NoColor = true
	result = Success()
	if result != "success" {
		t.Errorf("got %q, want 'success'", result)
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
