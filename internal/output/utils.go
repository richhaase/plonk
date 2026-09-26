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
	headers []string
	rows    [][]string
}

// NewStandardTableBuilder creates a new standardized table builder
func NewStandardTableBuilder() *StandardTableBuilder {
	return &StandardTableBuilder{}
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

// Build constructs the final table output
func (t *StandardTableBuilder) Build() string {
	if len(t.headers) == 0 && len(t.rows) == 0 {
		return ""
	}

	var output strings.Builder
	writer := tabwriter.NewWriter(&output, 0, 2, 2, ' ', 0)

	// Compute column widths (max of header and all row values)
	colWidths := make([]int, len(t.headers))
	for i, h := range t.headers {
		colWidths[i] = len(h)
	}
	for _, row := range t.rows {
		for i, val := range row {
			if i < len(colWidths) && len(val) > colWidths[i] {
				colWidths[i] = len(val)
			}
		}
	}

	// Headers with separator
	if len(t.headers) > 0 {
		fmt.Fprintln(writer, strings.Join(t.headers, "\t"))
		// Add dashed separator under headers (matching column widths)
		separators := make([]string, len(t.headers))
		for i := range t.headers {
			separators[i] = strings.Repeat("-", colWidths[i])
		}
		fmt.Fprintln(writer, strings.Join(separators, "\t"))
	}

	// Rows
	for _, row := range t.rows {
		fmt.Fprintln(writer, strings.Join(row, "\t"))
	}

	writer.Flush()
	output.WriteString("\n")
	return output.String()
}
