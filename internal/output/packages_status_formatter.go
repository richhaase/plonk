// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
)

// PackagesStatusOutput represents the output structure for packages status command
type PackagesStatusOutput struct {
	Result     Result `json:"result" yaml:"result"`
	RemoteSync string `json:"remote_sync,omitempty" yaml:"remote_sync,omitempty"`
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

	WriteTitle(&output, "Packages Status")
	WriteRemoteSync(&output, f.Data.RemoteSync)

	// Group packages by manager
	packagesByManager := make(map[string][]Item)
	missingPackages := []Item{}

	// Show managed and missing items
	for _, item := range result.Managed {
		packagesByManager[item.Manager] = append(packagesByManager[item.Manager], item)
	}
	missingPackages = append(missingPackages, result.Missing...)

	// Sort missing packages
	sortItems(missingPackages)

	// Build packages table
	if len(packagesByManager) > 0 || len(missingPackages) > 0 {
		// Create a table for packages
		pkgBuilder := NewStandardTableBuilder("")
		pkgBuilder.SetHeaders("PACKAGE", "MANAGER", "STATUS")

		// Show managed packages by manager (sorted alphabetically)
		sortedManagers := sortItemsByManager(packagesByManager)
		for _, manager := range sortedManagers {
			packages := packagesByManager[manager]
			sortItems(packages) // Sort packages alphabetically within each manager
			for _, pkg := range packages {
				pkgBuilder.AddRow(pkg.Name, manager, "managed")
			}
		}

		// Show missing packages
		for _, pkg := range missingPackages {
			pkgBuilder.AddRow(pkg.Name, pkg.Manager, "missing")
		}

		output.WriteString(pkgBuilder.Build())
		output.WriteString("\n")
	}

	// Add summary
	managedCount := len(result.Managed)
	missingCount := len(result.Missing)
	errorCount := len(result.Errors)

	output.WriteString("Summary: ")
	fmt.Fprintf(&output, "%d managed", managedCount)
	if missingCount > 0 {
		fmt.Fprintf(&output, ", %d missing", missingCount)
	}
	if errorCount > 0 {
		fmt.Fprintf(&output, ", %d errors", errorCount)
	}
	output.WriteString("\n")

	WriteErrors(&output, "package", result.Errors)

	if len(result.Managed) == 0 && len(result.Missing) == 0 && len(result.Errors) == 0 {
		output.Reset()
		WriteTitle(&output, "Packages Status")
		WriteRemoteSync(&output, f.Data.RemoteSync)
		output.WriteString("No managed packages.\n")
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (f PackagesStatusFormatter) StructuredData() any {
	result := f.Data.Result

	var items []ManagedItem

	// Add managed items
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

	// Add error items
	for _, item := range result.Errors {
		mi := ManagedItem{
			Name:    item.Name,
			Domain:  "package",
			State:   string(item.State),
			Manager: item.Manager,
			Error:   item.Error,
		}
		items = append(items, mi)
	}

	summary := Summary{
		TotalManaged:   len(result.Managed),
		TotalMissing:   len(result.Missing),
		TotalUntracked: len(result.Untracked),
		TotalErrors:    len(result.Errors),
		Results:        []Result{result},
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
