// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/interfaces"
	"github.com/richhaase/plonk/internal/runtime"
	"github.com/richhaase/plonk/internal/types"
	"github.com/richhaase/plonk/internal/ui"
	"github.com/spf13/cobra"
)

// Type aliases for UI types (these have been moved to internal/ui/formatters.go)
type ApplyOutput = ui.ApplyOutput
type ManagerApplyResult = ui.ManagerApplyResult
type PackageApplyResult = ui.PackageApplyResult

type DotfileApplyOutput = ui.DotfileApplyOutput
type DotfileAction = ui.DotfileAction

// TableOutput and StructuredData methods have been moved to internal/ui/formatters.go

type DotfileListOutput = ui.DotfileListOutput
type DotfileListSummary = ui.DotfileListSummary
type DotfileInfo = ui.DotfileInfo

// TableOutput and StructuredData methods moved to internal/ui/formatters.go

// Shared output types from dot_add.go (moved to internal/ui/formatters.go)
type DotfileAddOutput = ui.DotfileAddOutput
type DotfileBatchAddOutput = ui.DotfileBatchAddOutput

// TableOutput and StructuredData methods moved to internal/ui/formatters.go

// Shared functions for pkg and dot list operations

// Simplified runPkgList that passes raw data to RenderOutput
func runPkgList(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "packages", "output-format", "invalid output format")
	}

	// Get shared context
	sharedCtx := runtime.GetSharedContext()

	// Get specific manager if flag is set
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return err
	}

	// Reconcile packages directly
	ctx := context.Background()
	domainResult, err := sharedCtx.ReconcilePackages(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile package state")
	}

	// If a specific manager is requested, filter results
	if flags.Manager != "" {
		filteredManaged := make([]interfaces.Item, 0)
		filteredMissing := make([]interfaces.Item, 0)
		filteredUntracked := make([]interfaces.Item, 0)

		for _, item := range domainResult.Managed {
			if item.Manager == flags.Manager {
				filteredManaged = append(filteredManaged, item)
			}
		}
		for _, item := range domainResult.Missing {
			if item.Manager == flags.Manager {
				filteredMissing = append(filteredMissing, item)
			}
		}
		for _, item := range domainResult.Untracked {
			if item.Manager == flags.Manager {
				filteredUntracked = append(filteredUntracked, item)
			}
		}

		domainResult.Managed = filteredManaged
		domainResult.Missing = filteredMissing
		domainResult.Untracked = filteredUntracked
	}

	// For non-verbose mode, clear untracked items
	if !flags.Verbose {
		domainResult.Untracked = []interfaces.Item{}
	}

	// Wrap result to implement OutputData interface
	outputWrapper := &packageListResultWrapper{
		Result: domainResult,
	}

	// Pass raw data directly to RenderOutput
	return RenderOutput(outputWrapper, format)
}

// packageListResultWrapper wraps types.Result to implement OutputData
type packageListResultWrapper struct {
	Result types.Result
}

// TableOutput generates human-friendly table output
func (w *packageListResultWrapper) TableOutput() string {
	tb := NewTableBuilder()

	// Header with summary
	totalCount := len(w.Result.Managed) + len(w.Result.Missing) + len(w.Result.Untracked)
	tb.AddTitle("Package Summary")
	tb.AddLine("Total: %d packages | %s Managed: %d | %s Missing: %d | %s Untracked: %d",
		totalCount, IconSuccess, len(w.Result.Managed),
		IconWarning, len(w.Result.Missing),
		IconUnknown, len(w.Result.Untracked))
	tb.AddNewline()

	// If no packages, show simple message
	if totalCount == 0 {
		tb.AddLine("No packages found")
		return tb.Build()
	}

	// Collect all items with their states
	type itemWithState struct {
		item  interfaces.Item
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
	// Add untracked items if they exist (verbose mode)
	for _, item := range w.Result.Untracked {
		items = append(items, itemWithState{item: item, state: "untracked"})
	}

	// If we have items to show, create the table
	if len(items) > 0 {
		// Table header
		tb.AddLine("  Status Package                        Manager   ")
		tb.AddLine("  ------ ------------------------------ ----------")

		// Table rows
		for _, i := range items {
			statusIcon := GetStatusIcon(i.state)
			tb.AddLine("  %-6s %-30s %-10s",
				statusIcon, TruncateString(i.item.Name, 30), i.item.Manager)
		}
		tb.AddNewline()
	}

	// Show untracked count hint if untracked items were hidden
	if len(w.Result.Untracked) == 0 && totalCount > len(w.Result.Managed)+len(w.Result.Missing) {
		tb.AddLine("Untracked packages hidden (use --verbose to show)")
	}

	return tb.Build()
}

// StructuredData returns the raw result for JSON/YAML output
func (w *packageListResultWrapper) StructuredData() any {
	return w.Result
}

// Simplified runDotList that passes raw data to RenderOutput
func runDotList(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "dotfiles", "output-format", "invalid output format")
	}

	// Get shared context
	sharedCtx := runtime.GetSharedContext()

	// Reconcile dotfiles directly
	ctx := context.Background()
	domainResult, err := sharedCtx.ReconcileDotfiles(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile dotfiles")
	}

	// Parse filter flags
	showManaged, _ := cmd.Flags().GetBool("managed")
	showMissing, _ := cmd.Flags().GetBool("missing")
	showUntracked, _ := cmd.Flags().GetBool("untracked")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Apply filters to the result
	filteredResult := types.Result{
		Domain:    domainResult.Domain,
		Manager:   domainResult.Manager,
		Managed:   domainResult.Managed,
		Missing:   domainResult.Missing,
		Untracked: domainResult.Untracked,
	}

	// Filter based on flags
	if showManaged {
		filteredResult.Missing = []interfaces.Item{}
		filteredResult.Untracked = []interfaces.Item{}
	} else if showMissing {
		filteredResult.Managed = []interfaces.Item{}
		filteredResult.Untracked = []interfaces.Item{}
	} else if showUntracked {
		filteredResult.Managed = []interfaces.Item{}
		filteredResult.Missing = []interfaces.Item{}
	} else if !verbose {
		// Default: show managed + missing, hide untracked unless verbose
		filteredResult.Untracked = []interfaces.Item{}
	}

	// Wrap result to implement OutputData interface
	outputWrapper := &dotfileListResultWrapper{
		Result:  filteredResult,
		Verbose: verbose,
	}

	// Pass raw data directly to RenderOutput
	return RenderOutput(outputWrapper, format)
}

// dotfileListResultWrapper wraps types.Result to implement OutputData
type dotfileListResultWrapper struct {
	Result  types.Result
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
		item  interfaces.Item
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
	addItems := func(items []interfaces.Item, stateStr string) {
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
