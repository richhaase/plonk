// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources

import "fmt"

// ItemState represents the reconciliation state of any managed item
type ItemState int

const (
	StateManaged   ItemState = iota // In config AND present/installed
	StateMissing                    // In config BUT not present/installed
	StateUntracked                  // Present/installed BUT not in config
	StateDegraded                   // Reserved for future use
)

// State string constants for consistency
const (
	StateManagedStr   = "managed"
	StateMissingStr   = "missing"
	StateUntrackedStr = "untracked"
	StateDegradedStr  = "degraded" // Reserved for future use
)

// String returns a human-readable representation of the item state
func (s ItemState) String() string {
	switch s {
	case StateManaged:
		return "managed"
	case StateMissing:
		return "missing"
	case StateUntracked:
		return "untracked"
	case StateDegraded:
		return "degraded"
	default:
		return "unknown"
	}
}

// Item represents any manageable item (package, dotfile, etc.) with its current state
type Item struct {
	Name     string                 `json:"name" yaml:"name"`
	Type     string                 `json:"type,omitempty" yaml:"type,omitempty"` // Type of item (e.g., "file", "directory", "cask", "formula")
	State    ItemState              `json:"state" yaml:"state"`
	Domain   string                 `json:"domain" yaml:"domain"`                         // "package", "dotfile", etc.
	Manager  string                 `json:"manager,omitempty" yaml:"manager,omitempty"`   // "homebrew", "npm", etc.
	Path     string                 `json:"path,omitempty" yaml:"path,omitempty"`         // For dotfiles
	Error    string                 `json:"error,omitempty" yaml:"error,omitempty"`       // Error message if any
	Meta     map[string]string      `json:"meta,omitempty" yaml:"meta,omitempty"`         // Additional string metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"` // Additional data
}

// Result contains the results of state reconciliation for a domain
type Result struct {
	Domain    string `json:"domain" yaml:"domain"`
	Manager   string `json:"manager,omitempty" yaml:"manager,omitempty"`
	Managed   []Item `json:"managed" yaml:"managed"`
	Missing   []Item `json:"missing" yaml:"missing"`
	Untracked []Item `json:"untracked" yaml:"untracked"`
}

// Summary provides aggregate counts across all states
type Summary struct {
	TotalManaged   int      `json:"total_managed" yaml:"total_managed"`
	TotalMissing   int      `json:"total_missing" yaml:"total_missing"`
	TotalUntracked int      `json:"total_untracked" yaml:"total_untracked"`
	Results        []Result `json:"results" yaml:"results"`
}

// Count returns the total number of items in this result
func (r *Result) Count() int {
	return len(r.Managed) + len(r.Missing) + len(r.Untracked)
}

// IsEmpty returns true if this result contains no items
func (r *Result) IsEmpty() bool {
	return r.Count() == 0
}

// AddToSummary adds this result's counts to the provided summary
func (r *Result) AddToSummary(summary *Summary) {
	summary.TotalManaged += len(r.Managed)
	summary.TotalMissing += len(r.Missing)
	summary.TotalUntracked += len(r.Untracked)
	summary.Results = append(summary.Results, *r)
}

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

// Status display utilities for CLI commands

// DomainSummary represents counts for a specific domain/manager
type DomainSummary struct {
	Domain         string `json:"domain" yaml:"domain"`
	Manager        string `json:"manager,omitempty" yaml:"manager,omitempty"`
	ManagedCount   int    `json:"managed_count" yaml:"managed_count"`
	MissingCount   int    `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int    `json:"untracked_count" yaml:"untracked_count"`
}

// ManagedItem represents a currently managed item for status display
type ManagedItem struct {
	Name    string `json:"name" yaml:"name"`
	Domain  string `json:"domain" yaml:"domain"`
	Manager string `json:"manager,omitempty" yaml:"manager,omitempty"`
}

// ConvertResultsToSummary converts reconciliation results to Summary for output compatibility
func ConvertResultsToSummary(results map[string]Result) Summary {
	summary := Summary{
		TotalManaged:   0,
		TotalMissing:   0,
		TotalUntracked: 0,
		Results:        make([]Result, 0, len(results)),
	}

	for _, result := range results {
		summary.TotalManaged += len(result.Managed)
		summary.TotalMissing += len(result.Missing)
		summary.TotalUntracked += len(result.Untracked)
		summary.Results = append(summary.Results, result)
	}

	return summary
}

// CreateDomainSummary creates domain summaries with counts only
func CreateDomainSummary(results []Result) []DomainSummary {
	var domains []DomainSummary
	for _, result := range results {
		if result.IsEmpty() {
			continue
		}
		domains = append(domains, DomainSummary{
			Domain:         result.Domain,
			Manager:        result.Manager,
			ManagedCount:   len(result.Managed),
			MissingCount:   len(result.Missing),
			UntrackedCount: len(result.Untracked),
		})
	}
	return domains
}

// ExtractManagedItems extracts only the managed items without full metadata
func ExtractManagedItems(results []Result) []ManagedItem {
	var items []ManagedItem
	for _, result := range results {
		for _, managed := range result.Managed {
			items = append(items, ManagedItem{
				Name:    managed.Name,
				Domain:  managed.Domain,
				Manager: managed.Manager,
			})
		}
	}
	return items
}

// Operation validation utilities for CLI commands

// ValidateOperationResults checks if all operations failed and returns appropriate error
func ValidateOperationResults(results []OperationResult, operationType string) error {
	if len(results) == 0 {
		return nil
	}

	allFailed := true
	for _, result := range results {
		if result.Status != "failed" {
			allFailed = false
			break
		}
	}

	if allFailed {
		return fmt.Errorf("%s operation failed: all %d item(s) failed to process", operationType, len(results))
	}

	return nil
}
