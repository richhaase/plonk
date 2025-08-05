// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"github.com/richhaase/plonk/internal/resources"
)

// DotfileRemovalOutput represents the output for dotfile removal
type DotfileRemovalOutput struct {
	TotalFiles int                         `json:"total_files" yaml:"total_files"`
	Results    []resources.OperationResult `json:"results" yaml:"results"`
	Summary    DotfileRemovalSummary       `json:"summary" yaml:"summary"`
}

// DotfileRemovalSummary provides summary for dotfile removal
type DotfileRemovalSummary struct {
	Removed int `json:"removed" yaml:"removed"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// DotfileRemovalFormatter formats dotfile removal output
type DotfileRemovalFormatter struct {
	Data DotfileRemovalOutput
}

// NewDotfileRemovalFormatter creates a new formatter
func NewDotfileRemovalFormatter(data DotfileRemovalOutput) DotfileRemovalFormatter {
	return DotfileRemovalFormatter{Data: data}
}

// TableOutput generates human-friendly output
func (f DotfileRemovalFormatter) TableOutput() string {
	d := f.Data
	tb := NewTableBuilder()

	// For single file operations, show inline result
	if d.TotalFiles == 1 && len(d.Results) == 1 {
		result := d.Results[0]
		switch result.Status {
		case "removed":
			tb.AddLine("Removed dotfile from plonk management")
			tb.AddLine("   File: %s", result.Name)
			if source, ok := result.Metadata["source"].(string); ok {
				tb.AddLine("   Source: %s (removed from config)", source)
			}
		case "would-remove":
			tb.AddLine("Would remove dotfile from plonk management (dry-run)")
			tb.AddLine("   File: %s", result.Name)
			if source, ok := result.Metadata["source"].(string); ok {
				tb.AddLine("   Source: %s", source)
			}
		case "skipped":
			tb.AddLine("Skipped: %s", result.Name)
			if result.Error != nil {
				tb.AddLine("   Reason: %s", result.Error.Error())
			}
		case "failed":
			tb.AddLine("Failed: %s", result.Name)
			if result.Error != nil {
				tb.AddLine("   Error: %s", result.Error.Error())
			}
		}
		return tb.Build()
	}

	// For batch operations, show summary
	tb.AddTitle("Dotfile Removal")
	tb.AddNewline()

	// Check if this is a dry run
	isDryRun := false
	wouldRemoveCount := 0
	for _, result := range d.Results {
		if result.Status == "would-remove" {
			isDryRun = true
			wouldRemoveCount++
		}
	}

	if isDryRun {
		if wouldRemoveCount > 0 {
			tb.AddLine("Would remove %d dotfiles (dry-run)", wouldRemoveCount)
		}
	} else {
		if d.Summary.Removed > 0 {
			tb.AddLine("📄 Removed %d dotfiles", d.Summary.Removed)
		}
	}

	if d.Summary.Skipped > 0 {
		tb.AddLine("%d skipped", d.Summary.Skipped)
	}
	if d.Summary.Failed > 0 {
		tb.AddLine("%d failed", d.Summary.Failed)
	}

	tb.AddNewline()

	// Show individual files
	for _, result := range d.Results {
		switch result.Status {
		case "removed":
			tb.AddLine("   ✓ %s", result.Name)
		case "would-remove":
			tb.AddLine("   - %s", result.Name)
		case "skipped":
			tb.AddLine("   %s (not managed)", result.Name)
		case "failed":
			tb.AddLine("   ✗ %s", result.Name)
		}
	}

	tb.AddNewline()
	tb.AddLine("Total: %d dotfiles processed", d.TotalFiles)

	return tb.Build()
}

// StructuredData returns the structured data for serialization
func (f DotfileRemovalFormatter) StructuredData() any {
	return f.Data
}
