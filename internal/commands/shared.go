// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
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

// convertOperationResults converts resources.OperationResult to output.SerializableOperationResult
func convertOperationResults(results []resources.OperationResult) []output.SerializableOperationResult {
	converted := make([]output.SerializableOperationResult, len(results))
	for i, result := range results {
		converted[i] = output.SerializableOperationResult{
			Name:     result.Name,
			Manager:  result.Manager,
			Status:   result.Status,
			Error:    result.Error,
			Metadata: result.Metadata,
		}
	}
	return converted
}

// calculatePackageOperationSummary calculates summary from operation results
func calculatePackageOperationSummary(results []resources.OperationResult) output.PackageOperationSummary {
	summary := output.PackageOperationSummary{}
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
