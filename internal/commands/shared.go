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

// Type aliases for UI types (these have been moved to internal/output/types.go)
type ManagerResults = output.ManagerResults
type PackageOperation = output.PackageOperation
type PackageResults = output.PackageResults
type ApplyResult = output.ApplyResult

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

	// Convert to dotfile items
	convertToDotfileItems := func(items []resources.Item) []output.DotfileItem {
		converted := make([]output.DotfileItem, len(items))
		for i, item := range items {
			converted[i] = output.DotfileItem{
				Name:     item.Name,
				Path:     item.Path,
				State:    item.State.String(),
				Metadata: item.Metadata,
			}
		}
		return converted
	}

	// Convert to output package type and create formatter
	formatterData := output.DotfileStatusOutput{
		Managed:   convertToDotfileItems(filteredResult.Managed),
		Missing:   convertToDotfileItems(filteredResult.Missing),
		Untracked: convertToDotfileItems(filteredResult.Untracked),
		Verbose:   verbose,
	}
	formatter := output.NewDotfileListFormatter(formatterData)
	return RenderOutput(formatter, format)
}
