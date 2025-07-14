// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package operations

import (
	"context"
	"time"

	"github.com/richhaase/plonk/internal/errors"
)

// ItemProcessor defines a function that processes a single item and returns a result
type ItemProcessor func(ctx context.Context, item string) OperationResult

// BatchProcessorOptions configures how batch processing should behave
type BatchProcessorOptions struct {
	ItemType               string        // "package", "dotfile", etc.
	Operation              string        // "add", "remove", "install", etc.
	ShowIndividualProgress bool          // Whether to show progress for each item
	Timeout                time.Duration // Timeout for the entire operation
	ContinueOnError        *bool         // Whether to continue if individual items fail (nil = default true)
}

// GenericBatchProcessor processes multiple items using a provided ItemProcessor function
type GenericBatchProcessor struct {
	processor       ItemProcessor
	reporter        ProgressReporter
	options         BatchProcessorOptions
	continueOnError bool // resolved from options
}

// NewBatchProcessor creates a new generic batch processor
func NewBatchProcessor(processor ItemProcessor, options BatchProcessorOptions) *GenericBatchProcessor {
	// Set default timeout if not specified
	if options.Timeout == 0 {
		options.Timeout = 5 * time.Minute // default timeout
	}

	// Resolve continue on error setting
	continueOnError := true // default
	if options.ContinueOnError != nil {
		continueOnError = *options.ContinueOnError
	}

	// Create progress reporter
	reporter := NewProgressReporterForOperation(options.Operation, options.ItemType, options.ShowIndividualProgress)

	return &GenericBatchProcessor{
		processor:       processor,
		reporter:        reporter,
		options:         options,
		continueOnError: continueOnError,
	}
}

// ProcessItems processes multiple items in sequence, collecting results
func (b *GenericBatchProcessor) ProcessItems(ctx context.Context, items []string) ([]OperationResult, error) {
	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, b.options.Timeout)
	defer cancel()

	results := make([]OperationResult, 0, len(items))

	for _, item := range items {
		// Check if context was canceled
		if ctxWithTimeout.Err() != nil {
			break
		}

		// Process individual item
		result := b.processor(ctxWithTimeout, item)

		// Show individual progress if enabled
		b.reporter.ShowItemProgress(result)

		// Collect result
		results = append(results, result)

		// If not continuing on error and this item failed, stop processing
		if !b.continueOnError && result.Status == "failed" {
			break
		}
	}

	// Show batch summary only if individual progress was shown
	if b.options.ShowIndividualProgress {
		b.reporter.ShowBatchSummary(results)
	}

	return results, ctxWithTimeout.Err()
}

// SimpleProcessor creates an ItemProcessor from a simple processing function
// This helps convert existing per-item processing logic into ItemProcessor functions
func SimpleProcessor(processFn func(ctx context.Context, item string) OperationResult) ItemProcessor {
	return func(ctx context.Context, item string) OperationResult {
		return processFn(ctx, item)
	}
}

// PackageProcessor creates an ItemProcessor for package operations
// This is a specialized helper for package-related operations that need additional context
func PackageProcessor(processFn func(ctx context.Context, packageName, manager string) OperationResult, manager string) ItemProcessor {
	return func(ctx context.Context, item string) OperationResult {
		return processFn(ctx, item, manager)
	}
}

// StandardBatchWorkflow provides a complete workflow for batch operations
// This combines batch processing with standardized error handling and exit codes
func StandardBatchWorkflow(ctx context.Context, items []string, processor ItemProcessor, options BatchProcessorOptions) ([]OperationResult, error) {
	batchProcessor := NewBatchProcessor(processor, options)

	results, err := batchProcessor.ProcessItems(ctx, items)
	if err != nil {
		return results, err
	}

	// Use standardized exit code determination
	// Convert operation and itemType to appropriate error domain
	domain := getErrorDomainForItemType(options.ItemType)
	exitErr := DetermineExitCode(results, domain, options.Operation)
	if exitErr != nil {
		return results, exitErr
	}

	return results, nil
}

// getErrorDomainForItemType converts item type to error domain
func getErrorDomainForItemType(itemType string) errors.Domain {
	switch itemType {
	case "package":
		return errors.DomainPackages
	case "dotfile":
		return errors.DomainDotfiles
	default:
		return errors.DomainCommands
	}
}
