// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package testutil

import (
	"testing"
)

func TestBufferWriter(t *testing.T) {
	t.Run("captures output", func(t *testing.T) {
		buf := NewBufferWriter(true)

		buf.Printf("Hello %s", "World")

		if got := buf.String(); got != "Hello World" {
			t.Errorf("String() = %q, want %q", got, "Hello World")
		}
	})

	t.Run("terminal detection", func(t *testing.T) {
		tests := []struct {
			name       string
			isTerminal bool
		}{
			{"terminal", true},
			{"non-terminal", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				buf := NewBufferWriter(tt.isTerminal)
				if got := buf.IsTerminal(); got != tt.isTerminal {
					t.Errorf("IsTerminal() = %v, want %v", got, tt.isTerminal)
				}
			})
		}
	})

	t.Run("reset clears buffer", func(t *testing.T) {
		buf := NewBufferWriter(true)

		buf.Printf("Some content")
		buf.Reset()

		if got := buf.String(); got != "" {
			t.Errorf("After Reset(), String() = %q, want empty", got)
		}
	})
}
