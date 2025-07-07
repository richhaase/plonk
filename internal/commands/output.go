// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"encoding/json"
	"fmt"
	"os"

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
	TableOutput() string   // Human-friendly table format
	StructuredData() any   // Data structure for json/yaml/toml
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
	Filter   string          `json:"filter" yaml:"filter"`
	Managers []ManagerOutput `json:"managers" yaml:"managers"`
}

// ManagerOutput represents a package manager's output
type ManagerOutput struct {
	Name     string          `json:"name" yaml:"name"`
	Count    int             `json:"count" yaml:"count"`
	Packages []PackageOutput `json:"packages" yaml:"packages"`
}

// PackageOutput represents a package in the output
type PackageOutput struct {
	Name            string `json:"name" yaml:"name"`
	Version         string `json:"version,omitempty" yaml:"version,omitempty"`
	State           string `json:"state,omitempty" yaml:"state,omitempty"`
	ExpectedVersion string `json:"expected_version,omitempty" yaml:"expected_version,omitempty"`
}

// TableOutput generates human-friendly table output
func (p PackageListOutput) TableOutput() string {
	var output string
	for _, mgr := range p.Managers {
		if mgr.Count == 0 {
			continue
		}
		output += fmt.Sprintf("# %s (%d packages)\n", mgr.Name, mgr.Count)
		for _, pkg := range mgr.Packages {
			output += pkg.Name + "\n"
		}
		output += "\n"
	}
	if output == "" {
		output = "No packages found\n"
	}
	return output
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