// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"testing"

	"github.com/richhaase/plonk/internal/testutil"
)

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
