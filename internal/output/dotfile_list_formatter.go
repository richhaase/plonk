// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
)

// DotfileStatusOutput represents the output data for dotfile list operations
type DotfileStatusOutput struct {
	Managed   []DotfileItem `json:"managed" yaml:"managed"`
	Missing   []DotfileItem `json:"missing" yaml:"missing"`
	Untracked []DotfileItem `json:"untracked" yaml:"untracked"`
	Verbose   bool          `json:"verbose" yaml:"verbose"`
}

// DotfileItem represents a single dotfile item
type DotfileItem struct {
	Name     string                 `json:"name" yaml:"name"`
	Path     string                 `json:"path" yaml:"path"`
	State    string                 `json:"state" yaml:"state"`
	Target   string                 `json:"target,omitempty" yaml:"target,omitempty"`
	Source   string                 `json:"source,omitempty" yaml:"source,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// DotfileListFormatter formats dotfile list output
type DotfileListFormatter struct {
	Data DotfileStatusOutput
}

// NewDotfileListFormatter creates a new DotfileListFormatter
func NewDotfileListFormatter(data DotfileStatusOutput) *DotfileListFormatter {
	return &DotfileListFormatter{Data: data}
}

// TableOutput generates human-friendly table output
func (f *DotfileListFormatter) TableOutput() string {
	tb := NewTableBuilder()

	// Calculate totals
	totalItems := len(f.Data.Managed) + len(f.Data.Missing)
	if f.Data.Verbose {
		totalItems += len(f.Data.Untracked)
	}

	// Header
	tb.AddTitle("Dotfiles Summary")
	summaryLine := fmt.Sprintf("Total: %d files", totalItems)
	if len(f.Data.Managed) > 0 {
		summaryLine += fmt.Sprintf(" | %s Managed: %d", IconSuccess, len(f.Data.Managed))
	}
	if len(f.Data.Missing) > 0 {
		summaryLine += fmt.Sprintf(" | %s Missing: %d", IconWarning, len(f.Data.Missing))
	}
	if len(f.Data.Untracked) > 0 && f.Data.Verbose {
		summaryLine += fmt.Sprintf(" | %s Untracked: %d", IconUnknown, len(f.Data.Untracked))
	}
	tb.AddLine("%s", summaryLine)
	tb.AddNewline()

	// If no dotfiles, show simple message
	if totalItems == 0 {
		tb.AddLine("No dotfiles found")
		return tb.Build()
	}

	// Collect all items with their states
	type itemWithState struct {
		item  DotfileItem
		state string
	}
	var items []itemWithState

	// Add managed items
	for _, item := range f.Data.Managed {
		items = append(items, itemWithState{item: item, state: "managed"})
	}
	// Add missing items
	for _, item := range f.Data.Missing {
		items = append(items, itemWithState{item: item, state: "missing"})
	}
	// Add untracked items if verbose
	if f.Data.Verbose {
		for _, item := range f.Data.Untracked {
			items = append(items, itemWithState{item: item, state: "untracked"})
		}
	}

	// If we have items to show, create the table
	if len(items) > 0 {
		// Table header
		tb.AddLine("  Status Target                                    Source")
		tb.AddLine("  ------ ----------------------------------------- --------------------------------------")

		// Table rows
		for _, i := range items {
			statusIcon := GetStatusIcon(i.state)

			// Extract target and source from item
			target := i.item.Target
			source := i.item.Source

			// Fall back to Path and Name if Target/Source are empty
			if target == "" {
				target = i.item.Path
			}
			if source == "" {
				source = i.item.Name
			}

			// Check metadata for more specific values
			if i.item.Metadata != nil {
				if t, ok := i.item.Metadata["target"].(string); ok && t != "" {
					target = t
				}
				if s, ok := i.item.Metadata["source"].(string); ok && s != "" {
					source = s
				}
			}

			// Default to dash if empty
			if target == "" {
				target = "-"
			}
			if source == "" {
				source = "-"
			}

			tb.AddLine("  %-6s %-41s %s",
				statusIcon, TruncateString(target, 41), TruncateString(source, 38))
		}
		tb.AddNewline()
	}

	// Show untracked hint if not verbose
	if !f.Data.Verbose && len(f.Data.Untracked) > 0 {
		tb.AddLine("%d untracked files (use --verbose to show details)", len(f.Data.Untracked))
	}

	return tb.Build()
}

// StructuredData returns the structured format that matches the old output
func (f *DotfileListFormatter) StructuredData() any {
	// Calculate total based on what's being shown
	total := len(f.Data.Managed) + len(f.Data.Missing)
	if f.Data.Verbose {
		total += len(f.Data.Untracked)
	}

	// Build dotfiles list
	var dotfiles []map[string]string

	// Helper to add items
	addItems := func(items []DotfileItem, stateStr string) {
		for _, item := range items {
			target := item.Target
			source := item.Source

			// Fall back to Path and Name if Target/Source are empty
			if target == "" {
				target = item.Path
			}
			if source == "" {
				source = item.Name
			}

			// Check metadata for more specific values
			if item.Metadata != nil {
				if t, ok := item.Metadata["target"].(string); ok && t != "" {
					target = t
				}
				if s, ok := item.Metadata["source"].(string); ok && s != "" {
					source = s
				}
			}

			dotfiles = append(dotfiles, map[string]string{
				"name":   item.Name,
				"state":  stateStr,
				"target": target,
				"source": source,
			})
		}
	}

	// Add all items
	addItems(f.Data.Managed, "managed")
	addItems(f.Data.Missing, "missing")
	if f.Data.Verbose {
		addItems(f.Data.Untracked, "untracked")
	}

	// Return in the expected format
	return map[string]any{
		"summary": map[string]any{
			"total":     total,
			"managed":   len(f.Data.Managed),
			"missing":   len(f.Data.Missing),
			"untracked": len(f.Data.Untracked),
			"verbose":   f.Data.Verbose,
		},
		"dotfiles": dotfiles,
	}
}
