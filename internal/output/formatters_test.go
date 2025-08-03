// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"reflect"
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
				"Added dotfile to plonk configuration",
				"Source: ~/.config/plonk/.vimrc",
				"Destination: ~/.vimrc",
				"Original: /home/user/.vimrc",
				"has been copied to your plonk config",
			},
			noWant: []string{"(dry-run)", "Would add"},
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
				"Updated existing dotfile in plonk configuration",
				"Source: ~/.config/plonk/.bashrc",
				"has been copied to your plonk config directory, overwriting the previous version",
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
				"Would add dotfile",
				"(dry-run)",
				"Source: ~/.config/plonk/.bashrc",
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
				"Would update existing dotfile in plonk configuration",
				"(dry-run)",
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
				"✗",
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
				"Added dotfile",
				"Source: ~/.config/plonk/.zshrc",
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
				"Dotfile Directory Add",
				"Would add 3 files to plonk configuration - dry-run",
				"+ dest3 → src3",
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
				"Added 2 files to plonk configuration",
				"Warnings:",
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
				"Would process 2 files (1 add, 1 update) - dry-run",
				"+ dest1 → src1",
				"↻ dest2 → src2",
			},
		},
		{
			name: "no files",
			output: DotfileBatchAddOutput{
				TotalFiles: 0,
				AddedFiles: []DotfileAddOutput{},
			},
			wantStrs: []string{
				"Added 0 files to plonk configuration",
				"All files have been copied to your plonk config directory",
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

func TestDotfileAddOutput_StructuredData(t *testing.T) {
	output := DotfileAddOutput{
		Path:        "~/.bashrc",
		Source:      "~/.config/plonk/.bashrc",
		Destination: "~/.bashrc",
		Action:      "added",
	}

	result := output.StructuredData()
	// Compare the values since result is interface{}
	if actualOutput, ok := result.(DotfileAddOutput); !ok || actualOutput != output {
		t.Errorf("StructuredData() = %v, want %v", result, output)
	}
}

func TestDotfileBatchAddOutput_StructuredData(t *testing.T) {
	output := DotfileBatchAddOutput{
		TotalFiles: 3,
		AddedFiles: []DotfileAddOutput{
			{
				Path:   "~/.bashrc",
				Action: "added",
			},
		},
	}

	result := output.StructuredData()
	// Use reflect.DeepEqual for slice comparison
	if !reflect.DeepEqual(result, output) {
		t.Errorf("StructuredData() = %v, want %v", result, output)
	}
}
