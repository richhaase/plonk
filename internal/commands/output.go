// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/richhaase/plonk/internal/output"
)

// Re-export types from output package for backward compatibility
type OutputFormat = output.OutputFormat
type OutputData = output.OutputData

var (
	OutputTable = output.OutputTable
	OutputJSON  = output.OutputJSON
	OutputYAML  = output.OutputYAML
)

// Re-export functions from output package
var RenderOutput = output.RenderOutput
var ParseOutputFormat = output.ParseOutputFormat

// Re-export formatter types
type DotfileRemovalFormatter = output.DotfileRemovalFormatter
type ConfigShowFormatter = output.ConfigShowFormatter
type PackageOperationFormatter = output.PackageOperationFormatter
type StatusFormatter = output.StatusFormatter
type InfoFormatter = output.InfoFormatter
type SearchFormatter = output.SearchFormatter
type DoctorFormatter = output.DoctorFormatter

// Re-export formatter constructors
var NewDotfileRemovalFormatter = output.NewDotfileRemovalFormatter
var NewConfigShowFormatter = output.NewConfigShowFormatter
var NewPackageOperationFormatter = output.NewPackageOperationFormatter
var NewStatusFormatter = output.NewStatusFormatter
var NewInfoFormatter = output.NewInfoFormatter
var NewSearchFormatter = output.NewSearchFormatter
var NewDoctorFormatter = output.NewDoctorFormatter

// Re-export utility functions and types
var GetStatusIcon = output.GetStatusIcon
var NewTableBuilder = output.NewTableBuilder
var NewStandardTableBuilder = output.NewStandardTableBuilder

type TableBuilder = output.TableBuilder
type StandardTableBuilder = output.StandardTableBuilder

// Re-export output structures from output package (for formatters that were moved)
// Note: Commands may need to create these structures differently to avoid import cycles

// PackageListOutput represents the output structure for package list commands
type PackageListOutput struct {
	ManagedCount   int                     `json:"managed_count" yaml:"managed_count"`
	MissingCount   int                     `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int                     `json:"untracked_count" yaml:"untracked_count"`
	TotalCount     int                     `json:"total_count" yaml:"total_count"`
	Managers       []EnhancedManagerOutput `json:"managers" yaml:"managers"`
	Verbose        bool                    `json:"verbose" yaml:"verbose"`
	Items          []EnhancedPackageOutput `json:"items" yaml:"items"`
}

// EnhancedManagerOutput represents a package manager's enhanced output
type EnhancedManagerOutput struct {
	Name           string                  `json:"name" yaml:"name"`
	ManagedCount   int                     `json:"managed_count" yaml:"managed_count"`
	MissingCount   int                     `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int                     `json:"untracked_count" yaml:"untracked_count"`
	Packages       []EnhancedPackageOutput `json:"packages" yaml:"packages"`
}

// EnhancedPackageOutput represents a package in the enhanced output
type EnhancedPackageOutput struct {
	Name    string `json:"name" yaml:"name"`
	State   string `json:"state" yaml:"state"`
	Manager string `json:"manager" yaml:"manager"`
}

// Legacy types for backward compatibility
type ManagerOutput struct {
	Name     string          `json:"name" yaml:"name"`
	Count    int             `json:"count" yaml:"count"`
	Packages []PackageOutput `json:"packages" yaml:"packages"`
}

type PackageOutput struct {
	Name  string `json:"name" yaml:"name"`
	State string `json:"state,omitempty" yaml:"state,omitempty"`
}

// PackageStatusOutput represents the output structure for package status command
type PackageStatusOutput struct {
	Summary StatusSummary   `json:"summary" yaml:"summary"`
	Details []ManagerStatus `json:"details" yaml:"details"`
}

// StatusSummary represents the overall status summary
type StatusSummary struct {
	Managed   int `json:"managed" yaml:"managed"`
	Missing   int `json:"missing" yaml:"missing"`
	Untracked int `json:"untracked" yaml:"untracked"`
}

// ManagerStatus represents status for a specific manager
type ManagerStatus struct {
	Name    string `json:"name" yaml:"name"`
	Managed int    `json:"managed" yaml:"managed"`
	Missing int    `json:"missing" yaml:"missing"`
}
