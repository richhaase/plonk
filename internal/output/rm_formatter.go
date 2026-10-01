// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
)

// SerializableRemovalResult represents a removal result for serialization
type SerializableRemovalResult struct {
	Name     string                 `json:"name" yaml:"name"`
	Status   string                 `json:"status" yaml:"status"`
	Error    string                 `json:"error,omitempty" yaml:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// DotfileRemovalOutput represents the output for dotfile removal
type DotfileRemovalOutput struct {
	TotalFiles int                         `json:"total_files" yaml:"total_files"`
	Results    []SerializableRemovalResult `json:"results" yaml:"results"`
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

// TableOutput generates compact removal results, retaining every failure reason.
func (f DotfileRemovalFormatter) TableOutput() string {
	d := f.Data
	var w strings.Builder
	planned := 0
	for _, r := range d.Results {
		if r.Status == "would-remove" {
			planned++
		}
	}
	if planned > 0 {
		w.WriteString("Dry run · no changes will be made\n\n")
	}
	for _, r := range d.Results {
		detail := r.Error
		if r.Status == "removed" {
			detail = "source removed from configuration; deployed file kept"
		}
		if r.Status == "would-remove" {
			detail = "would remove source from configuration; deployed file would be kept"
		}
		if force, _ := r.Metadata["force"].(bool); force {
			if r.Status == "removed" {
				detail = "source and deployed file removed"
			}
			if r.Status == "would-remove" {
				detail = "would remove source and deployed file"
			}
		}
		if source, ok := r.Metadata["source"].(string); ok && (r.Status == "removed" || r.Status == "would-remove") {
			detail += "\n  source: " + source
		}
		if target, ok := r.Metadata["destination"].(string); ok && target != "" && (r.Status == "removed" || r.Status == "would-remove") {
			detail += "\n  target: " + target
		}
		WriteAction(&w, r.Status, r.Name, detail, false)
	}
	if len(d.Results) > 1 {
		fmt.Fprintf(&w, "\n%d removed, %d planned, %d skipped, %d failed\n", d.Summary.Removed, planned, d.Summary.Skipped, d.Summary.Failed)
	}
	if len(d.Results) == 0 {
		w.WriteString("No dotfiles to remove.\n")
	}
	return w.String()
}
