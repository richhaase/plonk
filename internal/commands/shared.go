// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/spf13/cobra"
)

// Type aliases for UI types (these have been moved to internal/output/formatters.go)
type ManagerApplyResult = output.ManagerApplyResult
type PackageApplyResult = output.PackageApplyResult

type DotfileAction = output.DotfileAction

// TableOutput and StructuredData methods have been moved to internal/output/formatters.go

type DotfileListOutput = output.DotfileListOutput
type DotfileListSummary = output.DotfileListSummary
type DotfileInfo = output.DotfileInfo

// TableOutput and StructuredData methods moved to internal/output/formatters.go

// Shared output types from dot_add.go (moved to internal/output/formatters.go)
type DotfileAddOutput = output.DotfileAddOutput
type DotfileBatchAddOutput = output.DotfileBatchAddOutput

// TableOutput and StructuredData methods moved to internal/output/formatters.go

// Shared functions for dot list operations

// Simplified runDotList that passes raw data to RenderOutput
func runDotList(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Reconcile dotfiles
	ctx := context.Background()
	domainResult, err := dotfiles.Reconcile(ctx, homeDir, configDir)
	if err != nil {
		return err
	}

	// Parse filter flags
	showManaged, _ := cmd.Flags().GetBool("managed")
	showMissing, _ := cmd.Flags().GetBool("missing")
	showUntracked, _ := cmd.Flags().GetBool("untracked")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Apply filters to the result
	filteredResult := resources.Result{
		Domain:    domainResult.Domain,
		Manager:   domainResult.Manager,
		Managed:   domainResult.Managed,
		Missing:   domainResult.Missing,
		Untracked: domainResult.Untracked,
	}

	// Filter based on flags
	if showManaged {
		filteredResult.Missing = []resources.Item{}
		filteredResult.Untracked = []resources.Item{}
	} else if showMissing {
		filteredResult.Managed = []resources.Item{}
		filteredResult.Untracked = []resources.Item{}
	} else if showUntracked {
		filteredResult.Managed = []resources.Item{}
		filteredResult.Missing = []resources.Item{}
	} else if !verbose {
		// Default: show managed + missing, hide untracked unless verbose
		filteredResult.Untracked = []resources.Item{}
	}

	// Wrap result to implement OutputData interface
	outputWrapper := &dotfileListResultWrapper{
		Result:  filteredResult,
		Verbose: verbose,
	}

	// Pass raw data directly to RenderOutput
	return RenderOutput(outputWrapper, format)
}

// dotfileListResultWrapper wraps resources.Result to implement OutputData
type dotfileListResultWrapper struct {
	Result  resources.Result
	Verbose bool
}

// TableOutput generates human-friendly table output
func (w *dotfileListResultWrapper) TableOutput() string {
	tb := NewTableBuilder()

	// Calculate totals
	totalItems := len(w.Result.Managed) + len(w.Result.Missing)
	if w.Verbose {
		totalItems += len(w.Result.Untracked)
	}

	// Header
	tb.AddTitle("Dotfiles Summary")
	summaryLine := fmt.Sprintf("Total: %d files", totalItems)
	if len(w.Result.Managed) > 0 {
		summaryLine += fmt.Sprintf(" | %s Managed: %d", IconSuccess, len(w.Result.Managed))
	}
	if len(w.Result.Missing) > 0 {
		summaryLine += fmt.Sprintf(" | %s Missing: %d", IconWarning, len(w.Result.Missing))
	}
	if len(w.Result.Untracked) > 0 && w.Verbose {
		summaryLine += fmt.Sprintf(" | %s Untracked: %d", IconUnknown, len(w.Result.Untracked))
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
		item  resources.Item
		state string
	}
	var items []itemWithState

	// Add managed items
	for _, item := range w.Result.Managed {
		items = append(items, itemWithState{item: item, state: "managed"})
	}
	// Add missing items
	for _, item := range w.Result.Missing {
		items = append(items, itemWithState{item: item, state: "missing"})
	}
	// Add untracked items if verbose
	if w.Verbose {
		for _, item := range w.Result.Untracked {
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
			target := i.item.Path
			source := i.item.Name

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
	if !w.Verbose && len(w.Result.Untracked) > 0 {
		tb.AddLine("%d untracked files (use --verbose to show details)", len(w.Result.Untracked))
	}

	return tb.Build()
}

// StructuredData returns the structured format that matches the old output
func (w *dotfileListResultWrapper) StructuredData() any {
	// Calculate total based on what's being shown
	total := len(w.Result.Managed) + len(w.Result.Missing)
	if w.Verbose {
		total += len(w.Result.Untracked)
	}

	// Build dotfiles list
	var dotfiles []map[string]string

	// Helper to add items
	addItems := func(items []resources.Item, stateStr string) {
		for _, item := range items {
			target := item.Path
			source := item.Name

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
	addItems(w.Result.Managed, "managed")
	addItems(w.Result.Missing, "missing")
	if w.Verbose {
		addItems(w.Result.Untracked, "untracked")
	}

	// Return in the expected format
	return map[string]any{
		"summary": map[string]any{
			"total":     total,
			"managed":   len(w.Result.Managed),
			"missing":   len(w.Result.Missing),
			"untracked": len(w.Result.Untracked),
			"verbose":   w.Verbose,
		},
		"dotfiles": dotfiles,
	}
}
