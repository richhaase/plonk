// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
)

// DotfilesStatusOutput represents the output structure for dotfiles status command
type DotfilesStatusOutput struct {
	Result     Result `json:"result" yaml:"result"`
	RemoteSync string `json:"remote_sync,omitempty" yaml:"remote_sync,omitempty"`
	HomeDir    string `json:"-" yaml:"-"` // Not included in JSON/YAML output
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

	WriteTitle(&output, "Dotfiles Status")
	WriteRemoteSync(&output, f.Data.RemoteSync)

	// Include managed, missing, and error items
	// Drifted files are already in Managed with State==StateDegraded
	itemsToShow := len(result.Managed) + len(result.Missing) + len(result.Errors)

	if itemsToShow > 0 {
		// Create a table for dotfiles
		dotBuilder := NewStandardTableBuilder()

		// Show the deployed target, its source type, and current status.
		dotBuilder.SetHeaders("DOTFILE", "TYPE", "STATUS")

		// Sort managed and missing dotfiles
		sortItems(result.Managed)
		sortItems(result.Missing)

		// Show managed dotfiles
		for _, item := range result.Managed {
			// Use destination (target) from metadata - this is where the dotfile is deployed
			target := dotfileTarget(item, f.Data.HomeDir)
			// Check if this is actually a drifted file or has an error
			status := "deployed"
			if item.State == StateDegraded {
				if driftStatus, ok := item.Metadata["drift_status"].(string); ok && driftStatus == "error" {
					status = "error"
				} else {
					status = "drifted"
				}
			}
			dotBuilder.AddRow(target, sourceType(item), status)
		}

		// Show missing dotfiles
		for _, item := range result.Missing {
			// Use destination (target) from metadata
			target := dotfileTarget(item, f.Data.HomeDir)
			dotBuilder.AddRow(target, sourceType(item), "missing")
		}

		// Show error dotfiles
		for _, item := range result.Errors {
			target := dotfileTarget(item, f.Data.HomeDir)
			dotBuilder.AddRow(target, sourceType(item), "error")
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
	fmt.Fprintf(&output, "%d managed", managedCount)
	if len(result.Missing) > 0 {
		fmt.Fprintf(&output, ", %d missing", len(result.Missing))
	}
	if driftedCount > 0 {
		fmt.Fprintf(&output, ", %d drifted", driftedCount)
	}
	if len(result.Errors) > 0 {
		fmt.Fprintf(&output, ", %d error(s)", len(result.Errors))
	}
	output.WriteString("\n")

	if len(result.Managed) == 0 && len(result.Missing) == 0 && len(result.Errors) == 0 {
		output.Reset()
		WriteTitle(&output, "Dotfiles Status")
		WriteRemoteSync(&output, f.Data.RemoteSync)
		output.WriteString("No managed dotfiles.\n")
	}

	return output.String()
}

func sourceType(item Item) string {
	if value, ok := item.Metadata["source_type"].(string); ok && value != "" {
		return value
	}
	if source, ok := item.Metadata["source"].(string); ok && strings.HasSuffix(source, ".tmpl") {
		return "template"
	}
	return "file"
}
