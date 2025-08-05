// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/richhaase/plonk/internal/resources"
)

// StandardTableBuilder and NewStandardTableBuilder moved to output package - using re-export

// Package-specific output types

// SerializableOperationResult wraps OperationResult for proper JSON serialization
type SerializableOperationResult struct {
	Name           string                 `json:"name"`
	Manager        string                 `json:"manager,omitempty"`
	Version        string                 `json:"version,omitempty"`
	Status         string                 `json:"status"`
	Error          string                 `json:"error,omitempty"`
	AlreadyManaged bool                   `json:"already_managed,omitempty"`
	FilesProcessed int                    `json:"files_processed,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConvertOperationResults converts OperationResult to SerializableOperationResult
func ConvertOperationResults(results []resources.OperationResult) []SerializableOperationResult {
	converted := make([]SerializableOperationResult, len(results))
	for i, result := range results {
		errorMsg := ""
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		converted[i] = SerializableOperationResult{
			Name:           result.Name,
			Manager:        result.Manager,
			Version:        result.Version,
			Status:         result.Status,
			Error:          errorMsg,
			AlreadyManaged: result.AlreadyManaged,
			FilesProcessed: result.FilesProcessed,
			Metadata:       result.Metadata,
		}
	}
	return converted
}

// PackageOperationOutput standardized output for package operations (install/uninstall)
type PackageOperationOutput struct {
	Command    string                        `json:"command" yaml:"command"`
	TotalItems int                           `json:"total_items" yaml:"total_items"`
	Results    []SerializableOperationResult `json:"results" yaml:"results"`
	Summary    PackageOperationSummary       `json:"summary" yaml:"summary"`
	DryRun     bool                          `json:"dry_run,omitempty" yaml:"dry_run,omitempty"`
}

// PackageOperationSummary provides summary for package operations
type PackageOperationSummary struct {
	Succeeded int `json:"succeeded" yaml:"succeeded"`
	Skipped   int `json:"skipped" yaml:"skipped"`
	Failed    int `json:"failed" yaml:"failed"`
}

// Calculation helpers

// CalculatePackageOperationSummary calculates summary from operation results
func CalculatePackageOperationSummary(results []resources.OperationResult) PackageOperationSummary {
	summary := PackageOperationSummary{}
	for _, result := range results {
		switch result.Status {
		case "added", "removed", "installed", "uninstalled", "success":
			summary.Succeeded++
		case "skipped", "already-installed", "already-configured":
			summary.Skipped++
		case "failed", "error":
			summary.Failed++
		}
	}
	return summary
}
