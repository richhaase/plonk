// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package commands provides a unified pipeline abstraction for command execution
// that eliminates duplicated flag parsing → processing → rendering patterns.
package commands

import (
	"context"
	"time"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/spf13/cobra"
)

// CommandPipeline provides a unified abstraction for command execution
// that handles the common pattern: Parse flags → Process → Render
type CommandPipeline struct {
	cmd      *cobra.Command
	itemType string
	flags    *SimpleFlags
	format   OutputFormat
	reporter *operations.DefaultProgressReporter
}

// ProcessorFunc defines the business logic processor function signature
// Context is provided for cancellation, args are the command arguments
// Returns operation results that can be rendered in multiple formats
type ProcessorFunc func(ctx context.Context, args []string, flags *SimpleFlags) ([]operations.OperationResult, error)

// SimpleProcessorFunc defines a simpler processor that doesn't need flags
type SimpleProcessorFunc func(ctx context.Context, args []string) (OutputData, error)

// NewCommandPipeline creates a new command pipeline for the given command
func NewCommandPipeline(cmd *cobra.Command, itemType string) (*CommandPipeline, error) {
	// Parse flags first
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands,
			cmd.Name(), "flags", "invalid flag combination")
	}

	// Parse output format
	format, err := ParseOutputFormat(flags.Output)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands,
			cmd.Name(), "output-format", "invalid output format")
	}

	// Create progress reporter
	reporter := operations.NewProgressReporter(itemType, format == OutputTable)

	return &CommandPipeline{
		cmd:      cmd,
		itemType: itemType,
		flags:    flags,
		format:   format,
		reporter: reporter,
	}, nil
}

// NewSimpleCommandPipeline creates a pipeline that doesn't use SimpleFlags
func NewSimpleCommandPipeline(cmd *cobra.Command, itemType string) (*CommandPipeline, error) {
	// Parse output format directly from command
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands,
			cmd.Name(), "output-format", "invalid output format")
	}

	// Create progress reporter
	reporter := operations.NewProgressReporter(itemType, format == OutputTable)

	return &CommandPipeline{
		cmd:      cmd,
		itemType: itemType,
		flags:    nil, // No flags for simple pipeline
		format:   format,
		reporter: reporter,
	}, nil
}

// ExecuteWithResults executes a processor function and handles the result rendering
// This is for processors that return []operations.OperationResult
func (p *CommandPipeline) ExecuteWithResults(ctx context.Context, processor ProcessorFunc, args []string) error {
	// Execute processor with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	results, err := processor(ctxWithTimeout, args, p.flags)
	if err != nil {
		return err
	}

	// Show progress for each result
	for _, result := range results {
		p.reporter.ShowItemProgress(result)
	}

	// Handle output based on format
	if p.format == OutputTable {
		p.reporter.ShowBatchSummary(results)
	} else {
		// Render structured output
		return p.renderOperationResults(results)
	}

	// Determine exit code based on results
	return operations.DetermineExitCode(results, errors.DomainCommands, p.cmd.Name())
}

// ExecuteWithData executes a processor function that returns OutputData directly
// This is for processors that handle their own result aggregation
func (p *CommandPipeline) ExecuteWithData(ctx context.Context, processor SimpleProcessorFunc, args []string) error {
	// Execute processor with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	data, err := processor(ctxWithTimeout, args)
	if err != nil {
		return err
	}

	// Render the output data
	return RenderOutput(data, p.format)
}

// GetFlags returns the parsed flags for processors that need them
func (p *CommandPipeline) GetFlags() *SimpleFlags {
	return p.flags
}

// GetFormat returns the output format
func (p *CommandPipeline) GetFormat() OutputFormat {
	return p.format
}

// GetCommand returns the cobra command for processors that need access to other flags
func (p *CommandPipeline) GetCommand() *cobra.Command {
	return p.cmd
}

// renderOperationResults renders operation results for structured output
func (p *CommandPipeline) renderOperationResults(results []operations.OperationResult) error {
	// For package operations, create a simple package output structure
	if p.itemType == "package" {
		output := PackageInstallOutput{
			TotalPackages: len(results),
			Results:       results,
			Summary:       calculatePackageSummary(results),
		}
		return RenderOutput(output, p.format)
	}

	// For uninstall operations, create an uninstall output structure
	if p.itemType == "uninstall" {
		output := PackageUninstallOutput{
			TotalPackages: len(results),
			Results:       results,
			Summary:       calculateUninstallSummary(results),
		}
		return RenderOutput(output, p.format)
	}

	// For dotfile operations, determine output format based on result count
	if p.itemType == "dotfile" {
		if len(results) == 1 {
			// Single dotfile/file - use existing DotfileAddOutput format
			result := results[0]
			output := DotfileAddOutput{
				Source:      getMetadataString(result, "source"),
				Destination: getMetadataString(result, "destination"),
				Action:      mapStatusToAction(result.Status),
				Path:        result.Name,
			}
			return RenderOutput(output, p.format)
		} else {
			// Multiple dotfiles/files - use batch output
			batchOutput := DotfileBatchAddOutput{
				TotalFiles: len(results),
				AddedFiles: convertToDotfileAddOutput(results),
				Errors:     extractErrorMessages(results),
			}
			return RenderOutput(batchOutput, p.format)
		}
	}

	// For dotfile removal operations
	if p.itemType == "dotfile-remove" {
		output := DotfileRemovalOutput{
			TotalFiles: len(results),
			Results:    results,
			Summary:    calculateDotfileRemovalSummary(results),
		}
		return RenderOutput(output, p.format)
	}

	// For other operations, use generic rendering
	return p.renderGenericResults(results)
}

// renderGenericResults provides generic rendering for other operation types
func (p *CommandPipeline) renderGenericResults(results []operations.OperationResult) error {
	// Create a generic output structure
	output := struct {
		TotalItems int                          `json:"total_items" yaml:"total_items"`
		Successful int                          `json:"successful" yaml:"successful"`
		Failed     int                          `json:"failed" yaml:"failed"`
		Results    []operations.OperationResult `json:"results" yaml:"results"`
	}{
		TotalItems: len(results),
		Successful: countStatus(results, "completed", "success", "added", "updated"),
		Failed:     countStatus(results, "failed", "error"),
		Results:    results,
	}

	// Create a wrapper that implements OutputData interface
	wrapper := &genericOutputWrapper{data: output}
	return RenderOutput(wrapper, p.format)
}

// genericOutputWrapper wraps generic output for the OutputData interface
type genericOutputWrapper struct {
	data interface{}
}

func (g *genericOutputWrapper) TableOutput() string {
	return "Generic output (table format not implemented)"
}

func (g *genericOutputWrapper) StructuredData() any {
	return g.data
}

// Helper functions

// countStatus counts results with any of the given statuses
func countStatus(results []operations.OperationResult, statuses ...string) int {
	count := 0
	for _, result := range results {
		for _, status := range statuses {
			if result.Status == status {
				count++
				break
			}
		}
	}
	return count
}

// getMetadataString safely extracts string metadata
func getMetadataString(result operations.OperationResult, key string) string {
	if result.Metadata == nil {
		return ""
	}
	if value, ok := result.Metadata[key].(string); ok {
		return value
	}
	return ""
}

// calculateUninstallSummary calculates summary from uninstall results
func calculateUninstallSummary(results []operations.OperationResult) PackageUninstallSummary {
	summary := PackageUninstallSummary{}
	for _, result := range results {
		switch result.Status {
		case "removed", "would-remove":
			summary.Removed++
		case "skipped":
			summary.Skipped++
		case "failed":
			summary.Failed++
		}
	}
	return summary
}

// calculateDotfileRemovalSummary calculates summary from dotfile removal results
func calculateDotfileRemovalSummary(results []operations.OperationResult) DotfileRemovalSummary {
	summary := DotfileRemovalSummary{}
	for _, result := range results {
		switch result.Status {
		case "unlinked", "would-unlink":
			summary.Removed++
		case "skipped":
			summary.Skipped++
		case "failed":
			summary.Failed++
		}
	}
	return summary
}

// generateActionMessage generates a human-readable action message
func generateActionMessage(result operations.OperationResult) string {
	if result.Error != nil {
		return "Failed: " + result.Error.Error()
	}

	switch result.Status {
	case "added":
		return "Added " + result.Name
	case "installed":
		return "Installed " + result.Name
	case "already-configured":
		return result.Name + " already configured"
	case "already-installed":
		return result.Name + " already installed"
	case "failed":
		return "Failed to process " + result.Name
	default:
		return "Processed " + result.Name
	}
}
