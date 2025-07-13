// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/richhaase/plonk/internal/errors"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the available output formats
type OutputFormat string

const (
	OutputTable OutputFormat = "table"
	OutputJSON  OutputFormat = "json"
	OutputYAML  OutputFormat = "yaml"
)

// OutputData defines the interface for command output data
type OutputData interface {
	TableOutput() string // Human-friendly table format
	StructuredData() any // Data structure for json/yaml/toml
}

// RenderOutput renders data in the specified format
func RenderOutput(data OutputData, format OutputFormat) error {
	switch format {
	case OutputTable:
		fmt.Print(data.TableOutput())
		return nil
	case OutputJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data.StructuredData())
	case OutputYAML:
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(data.StructuredData())
	default:
		return errors.NewError(errors.ErrInvalidInput, errors.DomainCommands, "output",
			fmt.Sprintf("unsupported output format: %s", format)).WithSuggestionMessage("Use: table, json, or yaml")
	}
}

// ParseOutputFormat converts string to OutputFormat
func ParseOutputFormat(format string) (OutputFormat, error) {
	switch format {
	case "table":
		return OutputTable, nil
	case "json":
		return OutputJSON, nil
	case "yaml":
		return OutputYAML, nil
	default:
		return OutputTable, errors.NewError(errors.ErrInvalidInput, errors.DomainCommands, "parse",
			fmt.Sprintf("unsupported format '%s'", format)).WithSuggestionMessage("Use: table, json, or yaml")
	}
}

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

// Enhanced Add/Remove Output structures
type EnhancedAddOutput struct {
	Package          string   `json:"package" yaml:"package"`
	Manager          string   `json:"manager" yaml:"manager"`
	ConfigAdded      bool     `json:"config_added" yaml:"config_added"`
	AlreadyInConfig  bool     `json:"already_in_config" yaml:"already_in_config"`
	Installed        bool     `json:"installed" yaml:"installed"`
	AlreadyInstalled bool     `json:"already_installed" yaml:"already_installed"`
	Error            string   `json:"error,omitempty" yaml:"error,omitempty"`
	Actions          []string `json:"actions" yaml:"actions"`
}

type EnhancedRemoveOutput struct {
	Package       string   `json:"package" yaml:"package"`
	Manager       string   `json:"manager" yaml:"manager"`
	ConfigRemoved bool     `json:"config_removed" yaml:"config_removed"`
	Uninstalled   bool     `json:"uninstalled" yaml:"uninstalled"`
	WasInConfig   bool     `json:"was_in_config" yaml:"was_in_config"`
	WasInstalled  bool     `json:"was_installed" yaml:"was_installed"`
	Error         string   `json:"error,omitempty" yaml:"error,omitempty"`
	Actions       []string `json:"actions" yaml:"actions"`
}

type BatchAddOutput struct {
	TotalPackages     int                 `json:"total_packages" yaml:"total_packages"`
	AddedToConfig     int                 `json:"added_to_config" yaml:"added_to_config"`
	Installed         int                 `json:"installed" yaml:"installed"`
	AlreadyConfigured int                 `json:"already_configured" yaml:"already_configured"`
	AlreadyInstalled  int                 `json:"already_installed" yaml:"already_installed"`
	Errors            int                 `json:"errors" yaml:"errors"`
	Packages          []EnhancedAddOutput `json:"packages" yaml:"packages"`
}

// Enhanced table output methods
func (a EnhancedAddOutput) TableOutput() string {
	tb := NewTableBuilder()

	// Header
	tb.AddTitle("Package Add")

	// Actions
	tb.AddActionList(a.Actions)

	// Error if present
	if a.Error != "" {
		tb.AddLine("%s Error: %s", IconError, a.Error)
	}

	// Summary
	tb.AddNewline()
	if a.Error != "" {
		tb.AddLine("Summary: Failed to add %s", a.Package)
	} else {
		summary := "Summary: "
		if a.ConfigAdded {
			summary += "Added to configuration"
		} else if a.AlreadyInConfig {
			summary += "Already in configuration"
		}

		if a.Installed {
			if a.ConfigAdded {
				summary += " and installed"
			} else {
				summary += " and installed"
			}
		} else if a.AlreadyInstalled {
			summary += " (already installed)"
		}

		tb.AddLine("%s", summary)
	}

	return tb.Build()
}

func (r EnhancedRemoveOutput) TableOutput() string {
	tb := NewTableBuilder()

	// Header
	tb.AddTitle("Package Remove")

	// Actions
	tb.AddActionList(r.Actions)

	// Error if present
	if r.Error != "" {
		tb.AddLine("%s Error: %s", IconError, r.Error)
	}

	// Summary
	tb.AddNewline()
	if r.Error != "" {
		tb.AddLine("Summary: Failed to remove %s", r.Package)
	} else {
		summary := "Summary: "
		parts := []string{}
		if r.ConfigRemoved {
			parts = append(parts, "removed from configuration")
		}
		if r.Uninstalled {
			parts = append(parts, "uninstalled from system")
		}
		if len(parts) == 0 {
			summary += "No changes made"
		} else {
			summary += strings.Join(parts, " and ")
		}

		tb.AddLine("%s", summary)
	}

	return tb.Build()
}

func (b BatchAddOutput) TableOutput() string {
	tb := NewTableBuilder()

	// Header
	tb.AddTitle("Package Add")

	// Individual package actions
	for _, pkg := range b.Packages {
		tb.AddActionList(pkg.Actions)
	}

	// Summary
	tb.AddNewline()
	counts := map[string]int{
		"packages processed": b.TotalPackages,
	}
	if b.AddedToConfig > 0 {
		counts["added to config"] = b.AddedToConfig
	}
	if b.Installed > 0 {
		counts["installed"] = b.Installed
	}
	if b.AlreadyConfigured > 0 {
		counts["already configured"] = b.AlreadyConfigured
	}
	if b.Errors > 0 {
		counts["errors"] = b.Errors
	}

	tb.AddSummaryLine("Summary:", counts)

	return tb.Build()
}

func (a EnhancedAddOutput) StructuredData() any {
	return a
}

func (r EnhancedRemoveOutput) StructuredData() any {
	return r
}

func (b BatchAddOutput) StructuredData() any {
	return b
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

// Legacy add/remove output types (keeping for compatibility)
type AddOutput struct {
	Package string `json:"package" yaml:"package"`
	Manager string `json:"manager" yaml:"manager"`
	Action  string `json:"action" yaml:"action"`
}

type AddAllOutput struct {
	Added  int    `json:"added" yaml:"added"`
	Total  int    `json:"total" yaml:"total"`
	Action string `json:"action" yaml:"action"`
}

type RemoveOutput struct {
	Package string `json:"package" yaml:"package"`
	Manager string `json:"manager" yaml:"manager"`
	Action  string `json:"action" yaml:"action"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Legacy table output methods (minimal output, handled in command logic)
func (a AddOutput) TableOutput() string {
	return "" // Table output is handled in the command logic
}

func (a AddAllOutput) TableOutput() string {
	return "" // Table output is handled in the command logic
}

func (r RemoveOutput) TableOutput() string {
	return "" // Table output is handled in the command logic
}

func (a AddOutput) StructuredData() any {
	return a
}

func (a AddAllOutput) StructuredData() any {
	return a
}

func (r RemoveOutput) StructuredData() any {
	return r
}

// TableOutput generates human-friendly table output
func (p PackageListOutput) TableOutput() string {
	tb := NewTableBuilder()

	// Header with summary
	tb.AddTitle("Package Summary")
	tb.AddLine("Total: %d packages | %s Managed: %d | %s Missing: %d | %s Untracked: %d",
		p.TotalCount, IconSuccess, p.ManagedCount, IconWarning, p.MissingCount, IconUnknown, p.UntrackedCount)
	tb.AddNewline()

	// If no packages, show simple message
	if p.TotalCount == 0 {
		tb.AddLine("No packages found")
		return tb.Build()
	}

	// Collect items to show based on verbose mode
	var itemsToShow []EnhancedPackageOutput
	if p.Verbose {
		itemsToShow = p.Items
	} else {
		// Show only managed and missing packages
		for _, item := range p.Items {
			if item.State == "managed" || item.State == "missing" {
				itemsToShow = append(itemsToShow, item)
			}
		}
	}

	// If we have items to show, create the table
	if len(itemsToShow) > 0 {
		// Table header
		tb.AddLine("  Status Package                        Manager   ")
		tb.AddLine("  ------ ------------------------------ ----------")

		// Table rows
		for _, item := range itemsToShow {
			statusIcon := GetStatusIcon(item.State)
			tb.AddLine("  %-6s %-30s %-10s",
				statusIcon, TruncateString(item.Name, 30), item.Manager)
		}
		tb.AddNewline()
	}

	// Show untracked count hint if not in verbose mode
	if !p.Verbose && p.UntrackedCount > 0 {
		tb.AddLine("%d untracked packages (use --verbose to show details)", p.UntrackedCount)
	}

	return tb.Build()
}

// StructuredData returns the structured data for serialization
func (p PackageListOutput) StructuredData() any {
	return p
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

// TableOutput generates human-friendly table output for status
func (p PackageStatusOutput) TableOutput() string {
	output := "Package Status\n==============\n\n"

	if p.Summary.Managed > 0 {
		output += fmt.Sprintf("✅ %d managed packages\n", p.Summary.Managed)
	} else {
		output += "📦 No managed packages\n"
	}

	if p.Summary.Missing > 0 {
		output += fmt.Sprintf("❌ %d missing packages\n", p.Summary.Missing)
	}

	if p.Summary.Untracked > 0 {
		output += fmt.Sprintf("🔍 %d untracked packages\n", p.Summary.Untracked)
	}

	// Show details if there are any managed or missing packages
	if p.Summary.Missing > 0 || p.Summary.Managed > 0 {
		output += "\nDetails:\n"
		for _, mgr := range p.Details {
			if mgr.Managed == 0 && mgr.Missing == 0 {
				continue
			}

			output += fmt.Sprintf("  %s: ", mgr.Name)
			parts := []string{}
			if mgr.Managed > 0 {
				parts = append(parts, fmt.Sprintf("%d managed", mgr.Managed))
			}
			if mgr.Missing > 0 {
				parts = append(parts, fmt.Sprintf("%d missing", mgr.Missing))
			}

			for i, part := range parts {
				if i > 0 {
					output += ", "
				}
				output += part
			}
			output += "\n"
		}
	}

	return output
}

// StructuredData returns the structured data for serialization
func (p PackageStatusOutput) StructuredData() any {
	return p
}
