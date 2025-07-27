// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/richhaase/plonk/internal/resources"
)

// StandardOutput provides consistent output formatting across all commands
type StandardOutput struct {
	Command string      `json:"command" yaml:"command"`
	Summary interface{} `json:"summary" yaml:"summary"`
	Items   interface{} `json:"items,omitempty" yaml:"items,omitempty"`
	Errors  []string    `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// TableOutput generates standardized table output
func (s StandardOutput) TableOutput() string {
	// This will be implemented by specific command outputs
	return ""
}

// StructuredData returns the structured data for serialization
func (s StandardOutput) StructuredData() any {
	return s
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

// AddError adds an error message to be displayed
func (t *StandardTableBuilder) AddError(error string) *StandardTableBuilder {
	t.errors = append(t.errors, error)
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
	title := fmt.Sprintf("Package %s", strings.Title(p.Command))
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

// StatusOperationOutput standardized output for status-type commands
type StatusOperationOutput struct {
	Command    string             `json:"command" yaml:"command"`
	Summary    StatusSummaryInfo  `json:"summary" yaml:"summary"`
	Details    []StatusDetailInfo `json:"details,omitempty" yaml:"details,omitempty"`
	ActionHint string             `json:"action_hint,omitempty" yaml:"action_hint,omitempty"`
}

// StatusSummaryInfo provides summary for status operations
type StatusSummaryInfo struct {
	Total     int `json:"total" yaml:"total"`
	Managed   int `json:"managed" yaml:"managed"`
	Missing   int `json:"missing" yaml:"missing"`
	Untracked int `json:"untracked,omitempty" yaml:"untracked,omitempty"`
}

// StatusDetailInfo provides detail for status operations
type StatusDetailInfo struct {
	Category string           `json:"category" yaml:"category"`
	Items    []StatusItemInfo `json:"items" yaml:"items"`
	Summary  map[string]int   `json:"summary" yaml:"summary"`
}

// StatusItemInfo represents individual status items
type StatusItemInfo struct {
	Name    string `json:"name" yaml:"name"`
	Manager string `json:"manager,omitempty" yaml:"manager,omitempty"`
	Status  string `json:"status" yaml:"status"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Target  string `json:"target,omitempty" yaml:"target,omitempty"`
}

// TableOutput generates human-friendly output for status operations
func (s StatusOperationOutput) TableOutput() string {
	builder := NewStandardTableBuilder("Plonk Status")

	for _, detail := range s.Details {
		if len(detail.Items) == 0 {
			continue
		}

		// Category header with summary
		categoryTitle := strings.ToUpper(detail.Category)
		if summary := detail.Summary; len(summary) > 0 {
			var parts []string
			for key, count := range summary {
				if count > 0 {
					parts = append(parts, fmt.Sprintf("%d %s", count, key))
				}
			}
			if len(parts) > 0 {
				categoryTitle += fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
			}
		}

		builder.AddRow(categoryTitle, "", "")

		// Set appropriate headers based on category
		if detail.Category == "packages" {
			builder.SetHeaders("NAME", "MANAGER", "STATUS")
			for _, item := range detail.Items {
				icon := GetStatusIcon(item.Status)
				statusText := fmt.Sprintf("%s %s", icon, item.Status)
				builder.AddRow(item.Name, item.Manager, statusText)
			}
		} else if detail.Category == "dotfiles" {
			builder.SetHeaders("PATH", "TARGET", "STATUS")
			for _, item := range detail.Items {
				icon := GetStatusIcon(item.Status)
				statusText := fmt.Sprintf("%s %s", icon, item.Status)
				builder.AddRow(item.Name, item.Target, statusText)
			}
		}

		builder.AddRow("", "", "") // Empty row between categories
	}

	// Summary
	summaryText := fmt.Sprintf("Total: %d managed", s.Summary.Managed)
	if s.Summary.Missing > 0 {
		summaryText += fmt.Sprintf(", %d missing", s.Summary.Missing)
	}
	if s.Summary.Untracked > 0 {
		summaryText += fmt.Sprintf(", %d untracked", s.Summary.Untracked)
	}

	builder.SetSummary(summaryText)

	// Action hint
	if s.ActionHint != "" {
		builder.AddRow("", "", "")
		builder.AddRow(s.ActionHint, "", "")
	}

	return builder.Build()
}

// StructuredData returns the structured data for serialization
func (s StatusOperationOutput) StructuredData() any {
	return s
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

// StandardizeErrorMessage creates consistent error messages
func StandardizeErrorMessage(operation, input, issue, suggestion string) string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("Error: %s", issue))
	if input != "" {
		msg.WriteString(fmt.Sprintf(" '%s'", input))
	}
	msg.WriteString("\n")

	if suggestion != "" {
		msg.WriteString(suggestion)
		msg.WriteString("\n")
	}

	return msg.String()
}

// CreateDidYouMeanSuggestion creates "did you mean" suggestions for typos
func CreateDidYouMeanSuggestion(input string, validOptions []string) string {
	// Simple similarity check - find closest match
	var closest string
	minDistance := len(input) + 1

	for _, option := range validOptions {
		distance := calculateLevenshteinDistance(input, option)
		if distance < minDistance && distance <= 2 { // Max 2 character difference
			minDistance = distance
			closest = option
		}
	}

	if closest != "" {
		return fmt.Sprintf("Did you mean: %s", closest)
	}

	return fmt.Sprintf("Valid options: %s", strings.Join(validOptions, ", "))
}

// Simple Levenshtein distance calculation for typo suggestions
func calculateLevenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}

	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
