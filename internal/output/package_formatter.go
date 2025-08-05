// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
)

// SerializableOperationResult represents an operation result for serialization
type SerializableOperationResult struct {
	Name     string                 `json:"name" yaml:"name"`
	Status   string                 `json:"status" yaml:"status"`
	Manager  string                 `json:"manager,omitempty" yaml:"manager,omitempty"`
	Path     string                 `json:"path,omitempty" yaml:"path,omitempty"`
	Error    error                  `json:"error,omitempty" yaml:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// PackageOperationOutput standardized output for package operations (install/uninstall)
type PackageOperationOutput struct {
	Command    string                        `json:"command" yaml:"command"`
	TotalItems int                           `json:"total_items" yaml:"total_items"`
	Results    []SerializableOperationResult `json:"results" yaml:"results"`
	Summary    PackageOperationSummary       `json:"summary" yaml:"summary"`
	DryRun     bool                          `json:"dry_run,omitempty" yaml:"dry_run,omitempty"`
}

// PackageOperationSummary provides summary for package operations
type PackageOperationSummary struct {
	Succeeded int `json:"succeeded" yaml:"succeeded"`
	Skipped   int `json:"skipped" yaml:"skipped"`
	Failed    int `json:"failed" yaml:"failed"`
}

// PackageOperationFormatter formats package operation output
type PackageOperationFormatter struct {
	Data PackageOperationOutput
}

// NewPackageOperationFormatter creates a new formatter
func NewPackageOperationFormatter(data PackageOperationOutput) PackageOperationFormatter {
	return PackageOperationFormatter{Data: data}
}

// TableOutput generates human-friendly output for package operations
func (f PackageOperationFormatter) TableOutput() string {
	p := f.Data
	// strings.Title is deprecated, use simple capitalization
	commandTitle := p.Command
	if len(commandTitle) > 0 {
		commandTitle = strings.ToUpper(commandTitle[:1]) + commandTitle[1:]
	}
	title := fmt.Sprintf("Package %s", commandTitle)
	if p.DryRun {
		title += " (Dry Run)"
	}

	builder := NewStandardTableBuilder(title)

	if len(p.Results) > 0 {
		builder.SetHeaders("PACKAGE", "MANAGER", "STATUS")

		for _, result := range p.Results {
			status := result.Status
			if p.DryRun && status == "added" {
				status = "would-install"
			} else if p.DryRun && status == "removed" {
				status = "would-remove"
			}

			icon := GetStatusIcon(status)
			statusText := fmt.Sprintf("%s %s", icon, status)

			builder.AddRow(result.Name, result.Manager, statusText)
		}
	}

	// Build summary
	summaryText := fmt.Sprintf("Total: %d processed", p.TotalItems)
	if p.Summary.Succeeded > 0 {
		summaryText += fmt.Sprintf(", %d succeeded", p.Summary.Succeeded)
	}
	if p.Summary.Skipped > 0 {
		summaryText += fmt.Sprintf(", %d skipped", p.Summary.Skipped)
	}
	if p.Summary.Failed > 0 {
		summaryText += fmt.Sprintf(", %d failed", p.Summary.Failed)
	}

	builder.SetSummary(summaryText)

	return builder.Build()
}

// StructuredData returns the structured data for serialization
func (f PackageOperationFormatter) StructuredData() any {
	return f.Data
}
