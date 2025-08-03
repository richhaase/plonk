// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"testing"

	"github.com/richhaase/plonk/internal/testutil"
)

func TestProgressUpdate(t *testing.T) {
	// Save and restore original progress writer
	originalProgressWriter := progressWriter
	defer func() { progressWriter = originalProgressWriter }()

	tests := []struct {
		name      string
		current   int
		total     int
		operation string
		item      string
		want      string
	}{
		{
			name:      "single item shows simple format",
			current:   1,
			total:     1,
			operation: "Installing",
			item:      "vim",
			want:      "Installing: vim\n",
		},
		{
			name:      "multiple items shows progress",
			current:   2,
			total:     5,
			operation: "Installing",
			item:      "git",
			want:      "[2/5] Installing: git\n",
		},
		{
			name:      "zero total shows nothing",
			current:   1,
			total:     0,
			operation: "Installing",
			item:      "test",
			want:      "",
		},
		{
			name:      "first of many",
			current:   1,
			total:     10,
			operation: "Updating",
			item:      "ripgrep",
			want:      "[1/10] Updating: ripgrep\n",
		},
		{
			name:      "last of many",
			current:   10,
			total:     10,
			operation: "Updating",
			item:      "fd",
			want:      "[10/10] Updating: fd\n",
		},
		{
			name:      "negative total shows nothing",
			current:   1,
			total:     -1,
			operation: "Installing",
			item:      "test",
			want:      "",
		},
		{
			name:      "long operation and item names",
			current:   3,
			total:     7,
			operation: "Processing configuration",
			item:      "very-long-package-name-with-many-parts",
			want:      "[3/7] Processing configuration: very-long-package-name-with-many-parts\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := testutil.NewBufferWriter(true)
			progressWriter = buf

			ProgressUpdate(tt.current, tt.total, tt.operation, tt.item)

			if got := buf.String(); got != tt.want {
				t.Errorf("ProgressUpdate() output = %q, want %q",
					got, tt.want)
			}
		})
	}
}

func TestStageUpdate(t *testing.T) {
	originalProgressWriter := progressWriter
	defer func() { progressWriter = originalProgressWriter }()

	tests := []struct {
		name  string
		stage string
		want  string
	}{
		{
			name:  "simple stage",
			stage: "Cloning repository...",
			want:  "Cloning repository...\n",
		},
		{
			name:  "empty stage",
			stage: "",
			want:  "\n",
		},
		{
			name:  "stage with special characters",
			stage: "Running: git clone https://github.com/user/repo.git",
			want:  "Running: git clone https://github.com/user/repo.git\n",
		},
		{
			name:  "multiline stage text",
			stage: "Line1\nLine2",
			want:  "Line1\nLine2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := testutil.NewBufferWriter(true)
			progressWriter = buf

			StageUpdate(tt.stage)

			if got := buf.String(); got != tt.want {
				t.Errorf("StageUpdate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProgressUpdateNonTerminal(t *testing.T) {
	// Save and restore original progress writer
	originalProgressWriter := progressWriter
	defer func() { progressWriter = originalProgressWriter }()

	// Test that output is the same for terminal and non-terminal
	buf := testutil.NewBufferWriter(false) // non-terminal
	progressWriter = buf

	ProgressUpdate(2, 5, "Installing", "package")
	want := "[2/5] Installing: package\n"

	if got := buf.String(); got != want {
		t.Errorf("ProgressUpdate() in non-terminal = %q, want %q",
			got, want)
	}
}
