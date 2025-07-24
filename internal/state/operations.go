// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

// OperationResult represents the result of a single operation (package install, dotfile add, etc.)
type OperationResult struct {
	Name           string                 `json:"name"`                      // Package name or file path
	Manager        string                 `json:"manager,omitempty"`         // Package manager (for packages only)
	Version        string                 `json:"version,omitempty"`         // Package version (for packages only)
	Status         string                 `json:"status"`                    // "added", "updated", "skipped", "failed", "would-add", "would-update"
	Error          error                  `json:"error,omitempty"`           // Error if operation failed
	AlreadyManaged bool                   `json:"already_managed,omitempty"` // Whether item was already managed
	FilesProcessed int                    `json:"files_processed,omitempty"` // Number of files processed (for directories)
	Metadata       map[string]interface{} `json:"metadata,omitempty"`        // Additional operation-specific data
}

// ResultSummary provides aggregate information about a batch operation
type ResultSummary struct {
	Total          int `json:"total"`
	Added          int `json:"added"`
	Updated        int `json:"updated"`
	Removed        int `json:"removed"`
	Unlinked       int `json:"unlinked"`
	Skipped        int `json:"skipped"`
	Failed         int `json:"failed"`
	FilesProcessed int `json:"files_processed,omitempty"` // Total files processed (for dotfiles)
}

// CalculateSummary generates a summary from operation results
func CalculateSummary(results []OperationResult) ResultSummary {
	summary := ResultSummary{Total: len(results)}

	for _, result := range results {
		switch result.Status {
		case "added", "would-add":
			summary.Added++
		case "updated", "would-update":
			summary.Updated++
		case "removed", "would-remove":
			summary.Removed++
		case "unlinked", "would-unlink":
			summary.Unlinked++
		case "skipped":
			summary.Skipped++
		case "failed":
			summary.Failed++
		}
		summary.FilesProcessed += result.FilesProcessed
	}

	return summary
}

// CountByStatus counts results with a specific status
func CountByStatus(results []OperationResult, status string) int {
	count := 0
	for _, result := range results {
		if result.Status == status {
			count++
		}
	}
	return count
}
