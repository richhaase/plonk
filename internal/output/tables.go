// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package output contains all user interface rendering logic including
// tables, formatters, and output presentation.
package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Table represents a table to be rendered
type Table struct {
	Headers []string
	Rows    [][]string
	Footer  []string
}

// WriteTable writes a formatted table to the writer using tabwriter
func WriteTable(w io.Writer, headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	// Write headers
	if len(headers) > 0 {
		fmt.Fprintln(tw, strings.Join(headers, "\t"))

		// Write separator line
		separators := make([]string, len(headers))
		for i, header := range headers {
			separators[i] = strings.Repeat("-", len(header))
		}
		fmt.Fprintln(tw, strings.Join(separators, "\t"))
	}

	// Write rows
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	return nil
}

// WriteTableWithFooter writes a formatted table with an optional footer
func WriteTableWithFooter(w io.Writer, table Table) error {
	// Write main table
	if err := WriteTable(w, table.Headers, table.Rows); err != nil {
		return err
	}

	// Write footer if present
	if len(table.Footer) > 0 {
		fmt.Fprintln(w)
		for _, line := range table.Footer {
			fmt.Fprintln(w, line)
		}
	}

	return nil
}

// FormatPackageTable formats package list results as a table
func FormatPackageTable(items []PackageListItem) Table {
	headers := []string{"Status", "Package", "Manager"}
	rows := make([][]string, 0, len(items))

	for _, item := range items {
		status := getStatusIcon(item.State)
		rows = append(rows, []string{status, item.Name, item.Manager})
	}

	return Table{
		Headers: headers,
		Rows:    rows,
	}
}

// FormatDotfileTable formats dotfile list results as a table
func FormatDotfileTable(items []DotfileListItem) Table {
	headers := []string{"Status", "Target", "Source"}
	rows := make([][]string, 0, len(items))

	for _, item := range items {
		status := getStatusIcon(item.State)
		rows = append(rows, []string{status, item.Target, item.Source})
	}

	return Table{
		Headers: headers,
		Rows:    rows,
	}
}

// PackageListItem represents a package in the list output
type PackageListItem struct {
	Name    string
	Manager string
	State   string
}

// DotfileListItem represents a dotfile in the list output
type DotfileListItem struct {
	Source string
	Target string
	State  string
}

// getStatusIcon returns a simple icon for the state
func getStatusIcon(state string) string {
	switch state {
	case "managed":
		return "✓"
	case "missing":
		return "✗"
	case "untracked":
		return "?"
	default:
		return " "
	}
}
