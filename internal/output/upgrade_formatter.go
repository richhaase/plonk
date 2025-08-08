// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
)

// UpgradeFormatter formats upgrade operation output
type UpgradeFormatter struct {
	Data UpgradeOutput
}

// NewUpgradeFormatter creates a new formatter
func NewUpgradeFormatter(data UpgradeOutput) UpgradeFormatter {
	return UpgradeFormatter{Data: data}
}

// TableOutput generates human-friendly output for upgrade operations
func (f UpgradeFormatter) TableOutput() string {
	data := f.Data
	output := fmt.Sprintf("Package Upgrade Results\n")
	output += strings.Repeat("=", 23) + "\n\n"

	if len(data.Results) == 0 {
		output += "No packages to upgrade\n"
		return output
	}

	// Group results by manager for better organization
	managerResults := make(map[string][]UpgradeResult)
	for _, result := range data.Results {
		managerResults[result.Manager] = append(managerResults[result.Manager], result)
	}

	// Display results grouped by package manager
	for manager, results := range managerResults {
		output += fmt.Sprintf("%s:\n", strings.Title(manager))

		for _, result := range results {
			var statusIcon string
			var statusText string

			switch result.Status {
			case "upgraded":
				statusIcon = "✓"
				if result.FromVersion != "" && result.ToVersion != "" {
					statusText = fmt.Sprintf("upgraded %s → %s", result.FromVersion, result.ToVersion)
				} else {
					statusText = "upgraded"
				}
			case "skipped":
				statusIcon = "-"
				statusText = "already up-to-date"
			case "failed":
				statusIcon = "✗"
				statusText = "failed"
				if result.Error != "" {
					statusText += fmt.Sprintf(": %s", result.Error)
				}
			default:
				statusIcon = "?"
				statusText = result.Status
			}

			output += fmt.Sprintf("  %s %s (%s)\n", statusIcon, result.Package, statusText)
		}
		output += "\n"
	}

	// Summary
	output += "Summary:\n"
	output += fmt.Sprintf("  Total: %d packages\n", data.Summary.Total)
	output += fmt.Sprintf("  Upgraded: %d\n", data.Summary.Upgraded)
	output += fmt.Sprintf("  Skipped: %d\n", data.Summary.Skipped)
	output += fmt.Sprintf("  Failed: %d\n", data.Summary.Failed)

	return output
}

// StructuredData returns the data structure for JSON/YAML serialization
func (f UpgradeFormatter) StructuredData() any {
	return f.Data
}
