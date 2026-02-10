// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
)

// DotfilesStatusOutput represents the output structure for dotfiles status command
type DotfilesStatusOutput struct {
	Result    Result `json:"result" yaml:"result"`
	ConfigDir string `json:"-" yaml:"-"` // Not included in JSON/YAML output
	HomeDir   string `json:"-" yaml:"-"` // Not included in JSON/YAML output
}

// DotfilesStatusFormatter formats dotfiles status output
type DotfilesStatusFormatter struct {
	Data DotfilesStatusOutput
}

// NewDotfilesStatusFormatter creates a new formatter
func NewDotfilesStatusFormatter(data DotfilesStatusOutput) DotfilesStatusFormatter {
	return DotfilesStatusFormatter{Data: data}
}

// TableOutput generates human-friendly table output for dotfiles status
func (f DotfilesStatusFormatter) TableOutput() string {
	var output strings.Builder
	result := f.Data.Result

	output.WriteString("Dotfiles Status\n")
	output.WriteString("===============\n\n")

	// Include managed, missing, and error items
	// Drifted files are already in Managed with State==StateDegraded
	itemsToShow := append(result.Managed, result.Missing...)
	itemsToShow = append(itemsToShow, result.Errors...)

	if len(itemsToShow) > 0 {
		// Create a table for dotfiles
		dotBuilder := NewStandardTableBuilder("")

		// Simple two-column format: target path and status
		dotBuilder.SetHeaders("DOTFILE", "STATUS")

		// Sort managed and missing dotfiles
		sortItems(result.Managed)
		sortItems(result.Missing)

		// Show managed dotfiles
		for _, item := range result.Managed {
			// Use destination (target) from metadata - this is where the dotfile is deployed
			target := item.Name
			if dest, ok := item.Metadata["destination"].(string); ok {
				target = tildeShorthand(dest, f.Data.HomeDir)
			}
			// Check if this is actually a drifted file or has an error
			status := "deployed"
			if item.State == StateDegraded {
				if driftStatus, ok := item.Metadata["drift_status"].(string); ok && driftStatus == "error" {
					status = "error"
				} else {
					status = "drifted"
				}
			}
			dotBuilder.AddRow(target, status)
		}

		// Show missing dotfiles
		for _, item := range result.Missing {
			// Use destination (target) from metadata
			target := item.Name
			if dest, ok := item.Metadata["destination"].(string); ok {
				target = tildeShorthand(dest, f.Data.HomeDir)
			}
			dotBuilder.AddRow(target, "missing")
		}

		// Show error dotfiles
		for _, item := range result.Errors {
			target := item.Name
			if dest, ok := item.Metadata["destination"].(string); ok {
				target = tildeShorthand(dest, f.Data.HomeDir)
			}
			dotBuilder.AddRow(target, "error")
		}

		output.WriteString(dotBuilder.Build())
		output.WriteString("\n")
	}

	// Add summary
	// Count drifted items separately
	driftedCount := 0
	for _, item := range result.Managed {
		if item.State == StateDegraded {
			driftedCount++
		}
	}

	// Adjust managed count to exclude drifted
	managedCount := len(result.Managed) - driftedCount

	output.WriteString("Summary: ")
	output.WriteString(fmt.Sprintf("%d managed", managedCount))
	if len(result.Missing) > 0 {
		output.WriteString(fmt.Sprintf(", %d missing", len(result.Missing)))
	}
	if driftedCount > 0 {
		output.WriteString(fmt.Sprintf(", %d drifted", driftedCount))
	}
	if len(result.Errors) > 0 {
		output.WriteString(fmt.Sprintf(", %d error(s)", len(result.Errors)))
	}
	output.WriteString("\n")

	// If no output was generated (except for title), show helpful message
	outputStr := output.String()
	if outputStr == "Dotfiles Status\n===============\n\n" || outputStr == "" {
		output.Reset()
		output.WriteString("Dotfiles Status\n")
		output.WriteString("===============\n\n")
		output.WriteString("No managed dotfiles.\n")
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (f DotfilesStatusFormatter) StructuredData() any {
	result := f.Data.Result

	var items []ManagedItem

	// Add managed items
	for _, item := range result.Managed {
		managedItem := ManagedItem{
			Name:     item.Name,
			Domain:   "dotfile",
			State:    string(item.State),
			Manager:  item.Manager,
			Path:     item.Path,
			Metadata: sanitizeMetadata(item.Metadata),
		}
		// Add target for dotfiles
		if target, ok := item.Metadata["destination"].(string); ok {
			managedItem.Target = target
		}
		items = append(items, managedItem)
	}

	// Add missing items
	for _, item := range result.Missing {
		managedItem := ManagedItem{
			Name:     item.Name,
			Domain:   "dotfile",
			State:    string(item.State),
			Manager:  item.Manager,
			Path:     item.Path,
			Metadata: sanitizeMetadata(item.Metadata),
		}
		// Add target for dotfiles
		if target, ok := item.Metadata["destination"].(string); ok {
			managedItem.Target = target
		}
		items = append(items, managedItem)
	}

	// Add error items
	for _, item := range result.Errors {
		managedItem := ManagedItem{
			Name:     item.Name,
			Domain:   "dotfile",
			State:    string(item.State),
			Manager:  item.Manager,
			Path:     item.Path,
			Error:    item.Error,
			Metadata: sanitizeMetadata(item.Metadata),
		}
		if target, ok := item.Metadata["destination"].(string); ok {
			managedItem.Target = target
		}
		items = append(items, managedItem)
	}

	summary := Summary{
		TotalManaged:   len(result.Managed),
		TotalMissing:   len(result.Missing),
		TotalUntracked: len(result.Untracked),
		Results:        []Result{result},
	}

	// For backward compatibility with tests, add lowercase field aliases
	// The test expects "missing", "managed", "untracked" fields
	errorCount := len(result.Errors)
	return map[string]interface{}{
		"summary": map[string]interface{}{
			"managed":         summary.TotalManaged,
			"missing":         summary.TotalMissing,
			"untracked":       summary.TotalUntracked,
			"errors":          errorCount,
			"total_managed":   summary.TotalManaged,
			"total_missing":   summary.TotalMissing,
			"total_untracked": summary.TotalUntracked,
			"total_errors":    errorCount,
		},
		"items": items,
	}
}

// DotfilesStatusOutputSummary represents the structured output format
type DotfilesStatusOutputSummary struct {
	Summary Summary       `json:"summary" yaml:"summary"`
	Items   []ManagedItem `json:"items" yaml:"items"`
}
