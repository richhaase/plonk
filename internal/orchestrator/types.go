// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import "github.com/richhaase/plonk/internal/output"

// ApplyResult represents the result of an apply operation
// ApplyResult is now defined in output package
type ApplyResult = output.ApplyResult

// Legacy type aliases for backward compatibility
type PackageApplyResult = output.PackageResults
type ManagerApplyResult = output.ManagerResults
type PackageOperationApplyResult = output.PackageOperation
type DotfileApplyResult = output.DotfileResults
type DotfileActionApplyResult = output.DotfileOperation
type DotfileSummaryApplyResult = output.DotfileSummary
