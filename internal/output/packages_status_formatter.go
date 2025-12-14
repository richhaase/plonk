// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
)

// PackagesStatusOutput represents the output structure for packages status command
type PackagesStatusOutput struct {
	Result      Result `json:"result" yaml:"result"`
	ShowMissing bool   `json:"-" yaml:"-"` // Not included in JSON/YAML output
}

// PackagesStatusFormatter formats packages status output
type PackagesStatusFormatter struct {
	Data PackagesStatusOutput
}

// NewPackagesStatusFormatter creates a new formatter
func NewPackagesStatusFormatter(data PackagesStatusOutput) PackagesStatusFormatter {
	return PackagesStatusFormatter{Data: data}
}

// TableOutput generates human-friendly table output for packages status
func (f PackagesStatusFormatter) TableOutput() string {
	var output strings.Builder
	result := f.Data.Result

	output.WriteString("Packages Status\n")
	output.WriteString("===============\n\n")

	// Group packages by manager
	packagesByManager := make(map[string][]Item)
	missingPackages := []Item{}

	if f.Data.ShowMissing {
		// Show only missing items
		missingPackages = append(missingPackages, result.Missing...)
	} else {
		// Show managed and missing items
		for _, item := range result.Managed {
			packagesByManager[item.Manager] = append(packagesByManager[item.Manager], item)
		}
		missingPackages = append(missingPackages, result.Missing...)
	}

	// Sort missing packages
	sortItems(missingPackages)

	// Build packages table
	if len(packagesByManager) > 0 || len(missingPackages) > 0 {
		// Create a table for packages
		pkgBuilder := NewStandardTableBuilder("")
		pkgBuilder.SetHeaders("NAME", "MANAGER", "STATUS")

		// Show managed packages by manager (unless showing only missing)
		if !f.Data.ShowMissing {
			sortedManagers := sortItemsByManager(packagesByManager)
			for _, manager := range sortedManagers {
				packages := packagesByManager[manager]
				for _, pkg := range packages {
					pkgBuilder.AddRow(pkg.Name, manager, "managed")
				}
			}
		}

		// Show missing packages
		for _, pkg := range missingPackages {
			pkgBuilder.AddRow(pkg.Name, pkg.Manager, "missing")
		}

		output.WriteString(pkgBuilder.Build())
		output.WriteString("\n")
	}

	// Add summary (skip for missing-only to avoid misleading counts)
	if !f.Data.ShowMissing {
		managedCount := len(result.Managed)
		missingCount := len(result.Missing)

		output.WriteString("Summary: ")
		output.WriteString(fmt.Sprintf("%d managed", managedCount))
		if missingCount > 0 {
			output.WriteString(fmt.Sprintf(", %d missing", missingCount))
		}
		output.WriteString("\n")
	}

	// If no output was generated (except for title), show helpful message
	outputStr := output.String()
	if outputStr == "Packages Status\n===============\n\n" || outputStr == "" {
		output.Reset()
		output.WriteString("Packages Status\n")
		output.WriteString("===============\n\n")
		output.WriteString("No packages match the specified filters.\n")
		if f.Data.ShowMissing {
			output.WriteString("(Great! All tracked packages are installed)\n")
		}
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (f PackagesStatusFormatter) StructuredData() any {
	result := f.Data.Result

	// Filter items based on flags
	var items []ManagedItem

	// Add managed items unless we're only showing missing
	if !f.Data.ShowMissing {
		for _, item := range result.Managed {
			items = append(items, ManagedItem{
				Name:     item.Name,
				Domain:   "package",
				State:    string(item.State),
				Manager:  item.Manager,
				Path:     item.Path,
				Metadata: sanitizeMetadata(item.Metadata),
			})
		}
	}

	// Add missing items
	for _, item := range result.Missing {
		items = append(items, ManagedItem{
			Name:     item.Name,
			Domain:   "package",
			State:    string(item.State),
			Manager:  item.Manager,
			Path:     item.Path,
			Metadata: sanitizeMetadata(item.Metadata),
		})
	}

	// Adjust summary counts based on filter flags
	summary := Summary{
		TotalManaged:   len(result.Managed),
		TotalMissing:   len(result.Missing),
		TotalUntracked: len(result.Untracked),
		Results:        []Result{result},
	}

	// If filtering by missing only, adjust counts
	if f.Data.ShowMissing {
		summary.TotalManaged = 0
		summary.TotalUntracked = 0
	}

	return PackagesStatusOutputSummary{
		Summary: summary,
		Items:   items,
	}
}

// PackagesStatusOutputSummary represents the structured output format
type PackagesStatusOutputSummary struct {
	Summary Summary       `json:"summary" yaml:"summary"`
	Items   []ManagedItem `json:"items" yaml:"items"`
}
