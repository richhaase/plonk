// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/richhaase/plonk/internal/resources"
)

// StandardTableBuilder provides consistent table formatting across commands
type StandardTableBuilder struct {
	title   string
	headers []string
	rows    [][]string
	summary string
	errors  []string
}

// NewStandardTableBuilder creates a new standardized table builder
func NewStandardTableBuilder(title string) *StandardTableBuilder {
	return &StandardTableBuilder{
		title: title,
		rows:  make([][]string, 0),
	}
}

// SetHeaders sets the table column headers
func (t *StandardTableBuilder) SetHeaders(headers ...string) *StandardTableBuilder {
	t.headers = headers
	return t
}

// AddRow adds a data row to the table
func (t *StandardTableBuilder) AddRow(values ...string) *StandardTableBuilder {
	t.rows = append(t.rows, values)
	return t
}

// SetSummary sets the summary line displayed after the table
func (t *StandardTableBuilder) SetSummary(summary string) *StandardTableBuilder {
	t.summary = summary
	return t
}

// Build constructs the final table output
func (t *StandardTableBuilder) Build() string {
	var output strings.Builder

	// Title
	if t.title != "" {
		output.WriteString(t.title + "\n")
		output.WriteString(strings.Repeat("=", len(t.title)) + "\n\n")
	}

	// Table with proper alignment
	if len(t.headers) > 0 || len(t.rows) > 0 {
		var tableOutput strings.Builder
		writer := tabwriter.NewWriter(&tableOutput, 0, 2, 2, ' ', 0)

		// Headers
		if len(t.headers) > 0 {
			fmt.Fprintln(writer, strings.Join(t.headers, "\t"))
		}

		// Rows
		for _, row := range t.rows {
			fmt.Fprintln(writer, strings.Join(row, "\t"))
		}

		writer.Flush()
		output.WriteString(tableOutput.String())
		output.WriteString("\n")
	}

	// Summary
	if t.summary != "" {
		output.WriteString(t.summary + "\n")
	}

	// Errors
	if len(t.errors) > 0 {
		output.WriteString("\nErrors:\n")
		for _, err := range t.errors {
			output.WriteString(fmt.Sprintf("  %s %s\n", IconError, err))
		}
	}

	return output.String()
}

// Package-specific output types

// SerializableOperationResult wraps OperationResult for proper JSON serialization
type SerializableOperationResult struct {
	Name           string                 `json:"name"`
	Manager        string                 `json:"manager,omitempty"`
	Version        string                 `json:"version,omitempty"`
	Status         string                 `json:"status"`
	Error          string                 `json:"error,omitempty"`
	AlreadyManaged bool                   `json:"already_managed,omitempty"`
	FilesProcessed int                    `json:"files_processed,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConvertOperationResults converts OperationResult to SerializableOperationResult
func ConvertOperationResults(results []resources.OperationResult) []SerializableOperationResult {
	converted := make([]SerializableOperationResult, len(results))
	for i, result := range results {
		errorMsg := ""
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		converted[i] = SerializableOperationResult{
			Name:           result.Name,
			Manager:        result.Manager,
			Version:        result.Version,
			Status:         result.Status,
			Error:          errorMsg,
			AlreadyManaged: result.AlreadyManaged,
			FilesProcessed: result.FilesProcessed,
			Metadata:       result.Metadata,
		}
	}
	return converted
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

// TableOutput generates human-friendly output for package operations
func (p PackageOperationOutput) TableOutput() string {
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
func (p PackageOperationOutput) StructuredData() any {
	return p
}

// Calculation helpers

// CalculatePackageOperationSummary calculates summary from operation results
func CalculatePackageOperationSummary(results []resources.OperationResult) PackageOperationSummary {
	summary := PackageOperationSummary{}
	for _, result := range results {
		switch result.Status {
		case "added", "removed", "installed", "uninstalled", "success":
			summary.Succeeded++
		case "skipped", "already-installed", "already-configured":
			summary.Skipped++
		case "failed", "error":
			summary.Failed++
		}
	}
	return summary
}
