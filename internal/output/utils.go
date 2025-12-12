// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// Common status icons used across all commands
const (
	IconSuccess = "✓"
	IconWarning = "⚠"
	IconError   = "✗"
	IconInfo    = "•"
	IconUnknown = "?"
	IconSkipped = "-"
)

// GetStatusIcon returns the appropriate icon for a given status
func GetStatusIcon(status string) string {
	switch status {
	case "managed", "added", "installed", "removed", "success", "completed", "deployed":
		return IconSuccess
	case "missing", "warn", "warning", "would-install", "would-remove", "would-add", "would-update":
		return IconWarning
	case "failed", "error", "fail":
		return IconError
	case "untracked", "unknown", "available":
		return IconUnknown
	case "skipped", "already-configured", "already-installed", "already-managed":
		return IconInfo
	default:
		return IconSkipped
	}
}

// TableBuilder helps construct consistent table outputs
type TableBuilder struct {
	output strings.Builder
}

// NewTableBuilder creates a new TableBuilder
func NewTableBuilder() *TableBuilder {
	return &TableBuilder{}
}

// AddTitle adds a title with underline
func (t *TableBuilder) AddTitle(title string) *TableBuilder {
	t.output.WriteString(title + "\n")
	t.output.WriteString(strings.Repeat("=", len(title)) + "\n")
	return t
}

// AddLine adds a single line
func (t *TableBuilder) AddLine(format string, args ...interface{}) *TableBuilder {
	t.output.WriteString(fmt.Sprintf(format, args...) + "\n")
	return t
}

// AddNewline adds an empty line
func (t *TableBuilder) AddNewline() *TableBuilder {
	t.output.WriteString("\n")
	return t
}

// Build returns the constructed output
func (t *TableBuilder) Build() string {
	return t.output.String()
}

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

// OperationSummary is a generic summary for operations
type OperationSummary struct {
	Total     int `json:"total" yaml:"total"`
	Succeeded int `json:"succeeded" yaml:"succeeded"`
	Skipped   int `json:"skipped" yaml:"skipped"`
	Failed    int `json:"failed" yaml:"failed"`
}

// CommonSummary provides fields common to many operations
type CommonSummary struct {
	Added   int `json:"added,omitempty" yaml:"added,omitempty"`
	Updated int `json:"updated,omitempty" yaml:"updated,omitempty"`
	Removed int `json:"removed,omitempty" yaml:"removed,omitempty"`
	Skipped int `json:"skipped,omitempty" yaml:"skipped,omitempty"`
	Failed  int `json:"failed,omitempty" yaml:"failed,omitempty"`
}

// StateSummary provides fields for state-based operations
type StateSummary struct {
	Total     int `json:"total" yaml:"total"`
	Managed   int `json:"managed" yaml:"managed"`
	Missing   int `json:"missing" yaml:"missing"`
	Untracked int `json:"untracked,omitempty" yaml:"untracked,omitempty"`
}
