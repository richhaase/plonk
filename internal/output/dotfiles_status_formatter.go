// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"os"
	"strings"
)

// DotfilesStatusOutput represents the output structure for dotfiles status command
type DotfilesStatusOutput struct {
	Result        Result `json:"result" yaml:"result"`
	ShowMissing   bool   `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowUnmanaged bool   `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ConfigDir     string `json:"-" yaml:"-"` // Not included in JSON/YAML output
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

	// Determine which items to show based on flags
	var itemsToShow []Item
	if f.Data.ShowUnmanaged {
		itemsToShow = result.Untracked
	} else if f.Data.ShowMissing {
		itemsToShow = result.Missing
	} else {
		// Include managed and missing items
		// Drifted files are already in Managed with State==StateDegraded
		itemsToShow = append(result.Managed, result.Missing...)
	}

	if len(itemsToShow) > 0 {
		// Create a table for dotfiles
		dotBuilder := NewStandardTableBuilder("")

		if f.Data.ShowUnmanaged {
			// For unmanaged, use single column showing just the path
			dotBuilder.SetHeaders("UNMANAGED DOTFILES")

			// Sort untracked dotfiles
			sortItems(result.Untracked)

			// Show untracked dotfiles
			for _, item := range result.Untracked {
				// Show the dotfile path with ~ notation
				path := "~/" + item.Name

				// Add trailing slash for directories
				if itemPath, ok := item.Metadata["path"].(string); ok {
					if info, err := os.Stat(itemPath); err == nil && info.IsDir() {
						path += "/"
					}
				}

				dotBuilder.AddRow(path)
			}
		} else {
			// For managed/missing, use the three-column format
			dotBuilder.SetHeaders("$HOME", "$PLONK_DIR", "STATUS")

			// Sort managed and missing dotfiles
			sortItems(result.Managed)
			sortItems(result.Missing)

			// Show managed dotfiles (unless showing only missing)
			if !f.Data.ShowMissing {
				for _, item := range result.Managed {
					// Use source from metadata if available, otherwise fall back to Name
					source := item.Name
					if src, ok := item.Metadata["source"].(string); ok {
						source = src
					}
					target := ""
					if dest, ok := item.Metadata["destination"].(string); ok {
						target = dest
					}
					// Check if this is actually a drifted file
					status := "deployed"
					if item.State == StateDegraded {
						status = "drifted"
					}
					// Swap column order: target ($HOME), source ($PLONK_DIR), status
					dotBuilder.AddRow(target, source, status)
				}
			}

			// Show missing dotfiles
			for _, item := range result.Missing {
				// Use source from metadata if available, otherwise fall back to Name
				source := item.Name
				if src, ok := item.Metadata["source"].(string); ok {
					source = src
				}
				target := ""
				if dest, ok := item.Metadata["destination"].(string); ok {
					target = dest
				}
				// Swap column order: target ($HOME), source ($PLONK_DIR), status
				dotBuilder.AddRow(target, source, "missing")
			}
		}

		output.WriteString(dotBuilder.Build())
		output.WriteString("\n")
	}

	// Add summary (skip for unmanaged or missing to avoid misleading counts)
	if !f.Data.ShowUnmanaged && !f.Data.ShowMissing {
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
		output.WriteString("\n")
	}

	// If no output was generated (except for title), show helpful message
	outputStr := output.String()
	if outputStr == "Dotfiles Status\n===============\n\n" || outputStr == "" {
		output.Reset()
		output.WriteString("Dotfiles Status\n")
		output.WriteString("===============\n\n")
		output.WriteString("No dotfiles match the specified filters.\n")
		if f.Data.ShowMissing {
			output.WriteString("(Great! All tracked dotfiles are deployed)\n")
		}
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (f DotfilesStatusFormatter) StructuredData() any {
	result := f.Data.Result

	// Filter items based on flags
	var items []ManagedItem

	// Add managed items unless we're only showing missing or untracked
	if !f.Data.ShowMissing && !f.Data.ShowUnmanaged {
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
	}

	// Add missing items unless we're only showing untracked
	if !f.Data.ShowUnmanaged {
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
	}

	// Add untracked items if we're showing unmanaged
	if f.Data.ShowUnmanaged {
		for _, item := range result.Untracked {
			managedItem := ManagedItem{
				Name:     item.Name,
				Domain:   "dotfile",
				State:    string(item.State),
				Manager:  item.Manager,
				Path:     item.Path,
				Metadata: sanitizeMetadata(item.Metadata),
			}
			items = append(items, managedItem)
		}
	}

	// Adjust summary counts based on filter flags
	summary := Summary{
		TotalManaged:   len(result.Managed),
		TotalMissing:   len(result.Missing),
		TotalUntracked: len(result.Untracked),
		Results:        []Result{result},
	}

	// If filtering by a specific state, adjust counts to reflect only that state
	if f.Data.ShowMissing {
		summary.TotalManaged = 0
		summary.TotalUntracked = 0
	} else if f.Data.ShowUnmanaged {
		summary.TotalManaged = 0
		summary.TotalMissing = 0
	}

	// For backward compatibility with tests, add lowercase field aliases
	// The test expects "missing", "managed", "untracked" fields
	return map[string]interface{}{
		"summary": map[string]interface{}{
			"managed":         summary.TotalManaged,
			"missing":         summary.TotalMissing,
			"untracked":       summary.TotalUntracked,
			"total_managed":   summary.TotalManaged,
			"total_missing":   summary.TotalMissing,
			"total_untracked": summary.TotalUntracked,
		},
		"items": items,
	}
}

// DotfilesStatusOutputSummary represents the structured output format
type DotfilesStatusOutputSummary struct {
	Summary Summary       `json:"summary" yaml:"summary"`
	Items   []ManagedItem `json:"items" yaml:"items"`
}
