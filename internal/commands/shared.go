// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/richhaase/plonk/internal/output"
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
