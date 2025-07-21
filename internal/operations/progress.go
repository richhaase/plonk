// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package operations

import (
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
)

// DefaultProgressReporter provides a standard implementation of progress reporting
type DefaultProgressReporter struct {
	ShowIndividualProgress bool
	ItemType               string // "package" or "dotfile"
	Operation              string // "add", "remove", etc.
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(itemType string, showIndividual bool) *DefaultProgressReporter {
	return &DefaultProgressReporter{
		ShowIndividualProgress: showIndividual,
		ItemType:               itemType,
		Operation:              "add", // default to add for backward compatibility
	}
}

// NewProgressReporterForOperation creates a new progress reporter for a specific operation
func NewProgressReporterForOperation(operation, itemType string, showIndividual bool) *DefaultProgressReporter {
	return &DefaultProgressReporter{
		ShowIndividualProgress: showIndividual,
		ItemType:               itemType,
		Operation:              operation,
	}
}

// ShowItemProgress displays progress for an individual item
func (r *DefaultProgressReporter) ShowItemProgress(result OperationResult) {
	if !r.ShowIndividualProgress {
		return
	}

	switch result.Status {
	case "added":
		if result.Version != "" {
			fmt.Printf("✓ %s@%s (%s)\n", result.Name, result.Version, result.Manager)
		} else {
			if result.FilesProcessed > 0 {
				fmt.Printf("✓ %s\n", result.Name)
			} else {
				fmt.Printf("✓ %s\n", result.Name)
			}
		}
	case "updated":
		if result.Version != "" {
			fmt.Printf("↻ %s@%s (%s) (updated)\n", result.Name, result.Version, result.Manager)
		} else {
			fmt.Printf("↻ %s (updated)\n", result.Name)
		}
	case "removed":
		if result.Manager != "" {
			fmt.Printf("✓ %s (%s) - removed from configuration\n", result.Name, result.Manager)
		} else {
			fmt.Printf("✓ %s - removed\n", result.Name)
		}
	case "unlinked":
		fmt.Printf("✓ %s - unlinked\n", result.Name)
	case "skipped":
		if r.Operation == "remove" || r.Operation == "uninstall" {
			fmt.Printf("⚠ %s - not managed\n", result.Name)
		} else {
			fmt.Printf("ℹ %s - already managed\n", result.Name)
		}
	case "failed":
		fmt.Printf("✗ %s - %s\n", result.Name, FormatErrorWithSuggestion(result.Error, result.Name, r.ItemType))
	case "would-add":
		if result.Version != "" {
			fmt.Printf("+ %s (%s) - would add\n", result.Name, result.Manager)
		} else {
			fmt.Printf("+ %s - would add\n", result.Name)
		}
	case "would-update":
		fmt.Printf("+ %s - would update\n", result.Name)
	case "would-remove":
		if result.Manager != "" {
			fmt.Printf("- %s (%s) - would remove from configuration\n", result.Name, result.Manager)
		} else {
			fmt.Printf("- %s - would remove\n", result.Name)
		}
	case "would-unlink":
		fmt.Printf("- %s - would unlink\n", result.Name)
	}
}

// ShowBatchSummary displays a summary of the batch operation
func (r *DefaultProgressReporter) ShowBatchSummary(results []OperationResult) {
	summary := CalculateSummary(results)

	// Generate operation-appropriate summary message
	var summaryMsg string
	switch r.Operation {
	case "remove", "uninstall", "rm":
		if r.ItemType == "dotfile" && summary.FilesProcessed > 0 {
			summaryMsg = fmt.Sprintf("\nSummary: %d removed, %d unlinked, %d skipped, %d failed (%d total files)\n",
				summary.Removed, summary.Unlinked, summary.Skipped, summary.Failed, summary.FilesProcessed)
		} else {
			summaryMsg = fmt.Sprintf("\nSummary: %d removed, %d unlinked, %d skipped, %d failed\n",
				summary.Removed, summary.Unlinked, summary.Skipped, summary.Failed)
		}
	default: // "add", "install" or other operations
		if r.ItemType == "dotfile" && summary.FilesProcessed > 0 {
			summaryMsg = fmt.Sprintf("\nSummary: %d added, %d updated, %d skipped, %d failed (%d total files)\n",
				summary.Added, summary.Updated, summary.Skipped, summary.Failed, summary.FilesProcessed)
		} else {
			summaryMsg = fmt.Sprintf("\nSummary: %d added, %d updated, %d skipped, %d failed\n",
				summary.Added, summary.Updated, summary.Skipped, summary.Failed)
		}
	}

	fmt.Print(summaryMsg)

	// Show failed items with suggestions
	if summary.Failed > 0 {
		fmt.Printf("\nFailed %ss:\n", r.ItemType)
		for _, result := range results {
			if result.Status == "failed" {
				fmt.Printf("  %s: %v\n", result.Name, result.Error)
			}
		}
		if r.ItemType == "package" {
			fmt.Println("\nTry running 'plonk doctor' to check system health")
		}
	}
}

// FormatErrorWithSuggestion formats an error message with helpful suggestions
func FormatErrorWithSuggestion(err error, itemName string, itemType string) string {
	if err == nil {
		return ""
	}

	// Check if it's a PlonkError with suggestions
	if plonkErr, ok := err.(*errors.PlonkError); ok {
		msg := plonkErr.UserMessage()
		// UserMessage already includes suggestions
		return msg
	}

	msg := err.Error()

	// Add suggestions based on error type and item type
	if strings.Contains(msg, "not found") {
		if itemType == "package" {
			return fmt.Sprintf("%s\n     Try: plonk search %s", msg, itemName)
		} else {
			return fmt.Sprintf("%s\n     Check if path exists: ls -la %s", msg, itemName)
		}
	}

	if strings.Contains(msg, "manager unavailable") || strings.Contains(msg, "command not found") {
		return fmt.Sprintf("%s\n     Try: plonk doctor", msg)
	}

	if strings.Contains(msg, "network") || strings.Contains(msg, "timeout") {
		return fmt.Sprintf("%s\n     Check network connectivity", msg)
	}

	if strings.Contains(msg, "permission") {
		if itemType == "package" {
			return fmt.Sprintf("%s\n     Check file permissions and try again", msg)
		} else {
			return fmt.Sprintf("%s\n     Try: chmod +r %s", msg, itemName)
		}
	}

	if strings.Contains(msg, "already exists") {
		return fmt.Sprintf("%s\n     Use --force to overwrite", msg)
	}

	return msg
}
