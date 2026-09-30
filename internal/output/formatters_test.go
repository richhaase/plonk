// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"strings"
	"testing"
)

func TestDotfileAddOutput_TableOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   DotfileAddOutput
		wantStrs []string // strings that should appear in output
		noWant   []string // strings that should NOT appear
	}{
		{
			name: "successful add",
			output: DotfileAddOutput{
				Source:      "~/.config/plonk/.vimrc",
				Destination: "~/.vimrc",
				Action:      "added",
				Path:        "/home/user/.vimrc",
			},
			wantStrs: []string{
				"added  /home/user/.vimrc",
				"copied to ~/.config/plonk/.vimrc",
				"/home/user/.vimrc",
				"/home/user/.vimrc",
				"copied to ~/.config/plonk/.vimrc",
			},
			noWant: []string{"Dry run", "Would add"},
		},
		{
			name: "successful update",
			output: DotfileAddOutput{
				Source:      "~/.config/plonk/.bashrc",
				Destination: "~/.bashrc",
				Action:      "updated",
				Path:        "/home/user/.bashrc",
			},
			wantStrs: []string{
				"updated  /home/user/.bashrc",
				"~/.config/plonk/.bashrc",
				"copied to ~/.config/plonk/.bashrc",
			},
		},
		{
			name: "dry run would add",
			output: DotfileAddOutput{
				Source:      "~/.config/plonk/.bashrc",
				Destination: "~/.bashrc",
				Action:      "would-add",
				Path:        "/home/user/.bashrc",
			},
			wantStrs: []string{
				"would add",
				"Dry run",
				"~/.config/plonk/.bashrc",
			},
			noWant: []string{"has been copied"},
		},
		{
			name: "dry run would update",
			output: DotfileAddOutput{
				Source:      "~/.config/plonk/.gitconfig",
				Destination: "~/.gitconfig",
				Action:      "would-update",
				Path:        "/home/user/.gitconfig",
			},
			wantStrs: []string{
				"would update",
				"Dry run",
			},
		},
		{
			name: "failed action",
			output: DotfileAddOutput{
				Action: "failed",
				Path:   "/home/user/.config/test",
				Error:  "permission denied",
			},
			wantStrs: []string{
				"failed",
				"/home/user/.config/test",
				"permission denied",
			},
			noWant: []string{"Source:", "Destination:"},
		},
		{
			name: "empty path",
			output: DotfileAddOutput{
				Source:      "~/.config/plonk/.zshrc",
				Destination: "~/.zshrc",
				Action:      "added",
				Path:        "",
			},
			wantStrs: []string{
				"added",
				"copied to ~/.config/plonk/.zshrc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.TableOutput()

			// Check for expected strings
			for _, want := range tt.wantStrs {
				if !strings.Contains(result, want) {
					t.Errorf("TableOutput() missing %q in:\n%s",
						want, result)
				}
			}

			// Check for strings that should NOT appear
			for _, noWant := range tt.noWant {
				if strings.Contains(result, noWant) {
					t.Errorf("TableOutput() should not contain %q in:\n%s",
						noWant, result)
				}
			}
		})
	}
}

func TestDotfileBatchAddOutput_TableOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   DotfileBatchAddOutput
		wantStrs []string
	}{
		{
			name: "mixed results",
			output: DotfileBatchAddOutput{
				TotalFiles: 3,
				AddedFiles: []DotfileAddOutput{
					{Action: "added", Path: "file1", Destination: "dest1", Source: "src1"},
					{Action: "updated", Path: "file2", Destination: "dest2", Source: "src2"},
					{Action: "would-add", Path: "file3", Destination: "dest3", Source: "src3"},
				},
			},
			wantStrs: []string{
				"Dry run",
				"1 added, 1 updated, 1 planned, 0 failed",
				"would add  file3",
			},
		},
		{
			name: "with errors",
			output: DotfileBatchAddOutput{
				TotalFiles: 2,
				AddedFiles: []DotfileAddOutput{
					{Action: "added", Path: "file1", Destination: "dest1", Source: "src1"},
				},
				Errors: []string{
					"Error processing file2: permission denied",
					"Error processing file3: file not found",
				},
			},
			wantStrs: []string{
				"1 added, 0 updated, 0 planned, 2 failed",
				"error  add",
				"Error processing file2: permission denied",
				"Error processing file3: file not found",
			},
		},
		{
			name: "dry run",
			output: DotfileBatchAddOutput{
				TotalFiles: 2,
				AddedFiles: []DotfileAddOutput{
					{Action: "would-add", Path: "file1", Destination: "dest1", Source: "src1"},
					{Action: "would-update", Path: "file2", Destination: "dest2", Source: "src2"},
				},
			},
			wantStrs: []string{
				"0 added, 0 updated, 2 planned, 0 failed",
				"would add  file1",
				"would update  file2",
			},
		},
		{
			name: "no files",
			output: DotfileBatchAddOutput{
				TotalFiles: 0,
				AddedFiles: []DotfileAddOutput{},
			},
			wantStrs: []string{
				"No dotfiles to add.",
				"No dotfiles to add.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.output.TableOutput()

			for _, want := range tt.wantStrs {
				if !strings.Contains(result, want) {
					t.Errorf("TableOutput() missing %q in:\n%s",
						want, result)
				}
			}
		})
	}
}

func TestMapStatusToAction(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"added", "added"},
		{"updated", "updated"},
		{"would-add", "would-add"},
		{"would-update", "would-update"},
		{"unknown", "failed"},
		{"error", "failed"},
		{"", "failed"},
		{"random-status", "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := MapStatusToAction(tt.status); got != tt.want {
				t.Errorf("MapStatusToAction(%q) = %q, want %q",
					tt.status, got, tt.want)
			}
		})
	}
}
