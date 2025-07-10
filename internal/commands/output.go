// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
		return fmt.Errorf("unsupported output format: %s", format)
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
		return OutputTable, fmt.Errorf("unsupported format '%s'. Use: table, json, yaml", format)
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
	var output strings.Builder

	// Header
	output.WriteString("Package Add\n")
	output.WriteString("===========\n")

	// Actions
	for _, action := range a.Actions {
		if strings.Contains(action, "error") || strings.Contains(action, "failed") {
			output.WriteString("✗ " + action + "\n")
		} else if strings.Contains(action, "already") {
			output.WriteString("• " + action + "\n")
		} else {
			output.WriteString("✓ " + action + "\n")
		}
	}

	// Error if present
	if a.Error != "" {
		output.WriteString("✗ Error: " + a.Error + "\n")
	}

	// Summary
	output.WriteString("\n")
	if a.Error != "" {
		output.WriteString("Summary: Failed to add " + a.Package + "\n")
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

		output.WriteString(summary + "\n")
	}

	return output.String()
}

func (r EnhancedRemoveOutput) TableOutput() string {
	var output strings.Builder

	// Header
	output.WriteString("Package Remove\n")
	output.WriteString("==============\n")

	// Actions
	for _, action := range r.Actions {
		if strings.Contains(action, "error") || strings.Contains(action, "failed") {
			output.WriteString("✗ " + action + "\n")
		} else if strings.Contains(action, "not found") || strings.Contains(action, "already") {
			output.WriteString("• " + action + "\n")
		} else {
			output.WriteString("✓ " + action + "\n")
		}
	}

	// Error if present
	if r.Error != "" {
		output.WriteString("✗ Error: " + r.Error + "\n")
	}

	// Summary
	output.WriteString("\n")
	if r.Error != "" {
		output.WriteString("Summary: Failed to remove " + r.Package + "\n")
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

		output.WriteString(summary + "\n")
	}

	return output.String()
}

func (b BatchAddOutput) TableOutput() string {
	var output strings.Builder

	// Header
	output.WriteString("Package Add\n")
	output.WriteString("===========\n")

	// Individual package actions
	for _, pkg := range b.Packages {
		for _, action := range pkg.Actions {
			if strings.Contains(action, "error") || strings.Contains(action, "failed") {
				output.WriteString("✗ " + action + "\n")
			} else if strings.Contains(action, "already") {
				output.WriteString("• " + action + "\n")
			} else {
				output.WriteString("✓ " + action + "\n")
			}
		}
	}

	// Summary
	output.WriteString("\n")
	summary := fmt.Sprintf("Summary: %d packages processed", b.TotalPackages)
	if b.AddedToConfig > 0 {
		summary += fmt.Sprintf(" | %d added to config", b.AddedToConfig)
	}
	if b.Installed > 0 {
		summary += fmt.Sprintf(" | %d installed", b.Installed)
	}
	if b.AlreadyConfigured > 0 {
		summary += fmt.Sprintf(" | %d already configured", b.AlreadyConfigured)
	}
	if b.Errors > 0 {
		summary += fmt.Sprintf(" | %d errors", b.Errors)
	}

	output.WriteString(summary + "\n")

	return output.String()
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
	var output strings.Builder

	// Header with summary
	output.WriteString("Package Summary\n")
	output.WriteString("===============\n")
	output.WriteString(fmt.Sprintf("Total: %d packages | ✓ Managed: %d | ⚠ Missing: %d | ? Untracked: %d\n\n",
		p.TotalCount, p.ManagedCount, p.MissingCount, p.UntrackedCount))

	// If no packages, show simple message
	if p.TotalCount == 0 {
		output.WriteString("No packages found\n")
		return output.String()
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
		output.WriteString("  Status Package                        Manager   \n")
		output.WriteString("  ------ ------------------------------ ----------\n")

		// Table rows
		for _, item := range itemsToShow {
			var statusIcon string
			switch item.State {
			case "managed":
				statusIcon = "✓"
			case "missing":
				statusIcon = "⚠"
			case "untracked":
				statusIcon = "?"
			default:
				statusIcon = "-"
			}

			output.WriteString(fmt.Sprintf("  %-6s %-30s %-10s\n",
				statusIcon, truncateString(item.Name, 30), item.Manager))
		}
		output.WriteString("\n")
	}

	// Show untracked count hint if not in verbose mode
	if !p.Verbose && p.UntrackedCount > 0 {
		output.WriteString(fmt.Sprintf("%d untracked packages (use --verbose to show details)\n", p.UntrackedCount))
	}

	return output.String()
}

// Helper function to truncate strings to a specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
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
