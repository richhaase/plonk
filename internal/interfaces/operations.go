// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package interfaces

import "context"

// BatchProcessor handles batch operations across domains
type BatchProcessor interface {
	ProcessItems(ctx context.Context, items []string) ([]OperationResult, error)
}

// ProgressReporter provides progress feedback for operations
type ProgressReporter interface {
	ShowItemProgress(result OperationResult)
	ShowBatchSummary(results []OperationResult)
}

// OutputRenderer handles multiple output formats for command results
type OutputRenderer interface {
	TableOutput() string
	StructuredData() any
}

// OperationResult represents the result of a single operation
type OperationResult struct {
	Name           string                 `json:"name" yaml:"name"`
	Status         string                 `json:"status" yaml:"status"` // "added", "updated", "failed", "would-add", etc.
	Error          error                  `json:"error,omitempty" yaml:"error,omitempty"`
	FilesProcessed int                    `json:"files_processed,omitempty" yaml:"files_processed,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// BatchOperationResult represents the result of a batch operation
type BatchOperationResult struct {
	TotalItems   int               `json:"total_items" yaml:"total_items"`
	Successful   int               `json:"successful" yaml:"successful"`
	Failed       int               `json:"failed" yaml:"failed"`
	Skipped      int               `json:"skipped" yaml:"skipped"`
	Results      []OperationResult `json:"results" yaml:"results"`
	ErrorSummary []string          `json:"error_summary,omitempty" yaml:"error_summary,omitempty"`
}
