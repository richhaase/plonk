// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"os"
	"strings"
)

// Local types to avoid import cycles

// ItemState represents resource item state
type ItemState string

const (
	StateManaged ItemState = "managed"
	StateMissing ItemState = "missing"
	// Align with resources.StateDegraded.String() which returns "drifted"
	StateDegraded  ItemState = "drifted"
	StateUntracked ItemState = "untracked"
)

// Item represents a resource item
type Item struct {
	Name     string                 `json:"name"`
	Manager  string                 `json:"manager,omitempty"`
	Path     string                 `json:"path,omitempty"`
	State    ItemState              `json:"state"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Result represents domain result
type Result struct {
	Domain    string `json:"domain"`
	Managed   []Item `json:"managed"`
	Missing   []Item `json:"missing"`
	Untracked []Item `json:"untracked"`
}

// Summary represents resource summary
type Summary struct {
	TotalManaged   int      `json:"total_managed"`
	TotalMissing   int      `json:"total_missing"`
	TotalUntracked int      `json:"total_untracked"`
	Results        []Result `json:"results"`
}

// StatusOutput represents the output structure for status command
type StatusOutput struct {
	ConfigPath    string  `json:"config_path" yaml:"config_path"`
	LockPath      string  `json:"lock_path" yaml:"lock_path"`
	ConfigExists  bool    `json:"config_exists" yaml:"config_exists"`
	ConfigValid   bool    `json:"config_valid" yaml:"config_valid"`
	LockExists    bool    `json:"lock_exists" yaml:"lock_exists"`
	StateSummary  Summary `json:"state_summary" yaml:"state_summary"`
	ShowPackages  bool    `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowDotfiles  bool    `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowUnmanaged bool    `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowMissing   bool    `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ConfigDir     string  `json:"-" yaml:"-"` // Not included in JSON/YAML output
}

// StatusOutputSummary represents a summary-focused version for JSON/YAML output
type StatusOutputSummary struct {
	ConfigPath   string  `json:"config_path" yaml:"config_path"`
	LockPath     string  `json:"lock_path" yaml:"lock_path"`
	ConfigExists bool    `json:"config_exists" yaml:"config_exists"`
	ConfigValid  bool    `json:"config_valid" yaml:"config_valid"`
	LockExists   bool    `json:"lock_exists" yaml:"lock_exists"`
	StateSummary Summary `json:"state_summary" yaml:"state_summary"`
}

// ManagedItem represents an item under management with its details
type ManagedItem struct {
	Name     string                 `json:"name" yaml:"name"`
	Domain   string                 `json:"domain" yaml:"domain"`
	State    string                 `json:"state" yaml:"state"`
	Manager  string                 `json:"manager,omitempty" yaml:"manager,omitempty"`
	Path     string                 `json:"path,omitempty" yaml:"path,omitempty"`
	Target   string                 `json:"target,omitempty" yaml:"target,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// StatusFormatter formats status output
type StatusFormatter struct {
	Data StatusOutput
}

// NewStatusFormatter creates a new formatter
func NewStatusFormatter(data StatusOutput) StatusFormatter {
	return StatusFormatter{Data: data}
}

// sortItems sorts items by name in-place
func sortItems(items []Item) {
	// Simple bubble sort for stability
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Name > items[j].Name {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// sortItemsByManager returns sorted manager names from the map
func sortItemsByManager(itemsByManager map[string][]Item) []string {
	var managers []string
	for manager := range itemsByManager {
		managers = append(managers, manager)
	}
	// Simple sort
	for i := 0; i < len(managers); i++ {
		for j := i + 1; j < len(managers); j++ {
			if managers[i] > managers[j] {
				managers[i], managers[j] = managers[j], managers[i]
			}
		}
	}
	return managers
}

// TableOutput generates human-friendly table output for status
func (f StatusFormatter) TableOutput() string {
	s := f.Data
	var output strings.Builder

	// Title
	output.WriteString("Plonk Status\n")
	output.WriteString("============\n\n")

	// Process results by domain
	var packageResult, dotfileResult *Result
	for i := range s.StateSummary.Results {
		if s.StateSummary.Results[i].Domain == "package" {
			packageResult = &s.StateSummary.Results[i]
		} else if s.StateSummary.Results[i].Domain == "dotfile" {
			dotfileResult = &s.StateSummary.Results[i]
		}
	}

	// Show packages table if requested
	if s.ShowPackages && packageResult != nil {
		// Group packages by manager
		packagesByManager := make(map[string][]Item)
		missingPackages := []Item{}
		untrackedPackages := make(map[string][]Item)

		if s.ShowUnmanaged {
			// Show only untracked items
			for _, item := range packageResult.Untracked {
				untrackedPackages[item.Manager] = append(untrackedPackages[item.Manager], item)
			}
		} else if s.ShowMissing {
			// Show only missing items
			for _, item := range packageResult.Missing {
				missingPackages = append(missingPackages, item)
			}
		} else {
			// Show managed and missing items
			for _, item := range packageResult.Managed {
				packagesByManager[item.Manager] = append(packagesByManager[item.Manager], item)
			}
			for _, item := range packageResult.Missing {
				missingPackages = append(missingPackages, item)
			}
		}

		// Sort missing packages
		sortItems(missingPackages)

		// Build packages table
		if len(packagesByManager) > 0 || len(missingPackages) > 0 || len(untrackedPackages) > 0 {
			output.WriteString("PACKAGES\n")
			output.WriteString("--------\n")

			// Create a table for packages
			pkgBuilder := NewStandardTableBuilder("")
			pkgBuilder.SetHeaders("NAME", "MANAGER", "STATUS")

			// Show managed packages by manager (unless showing only missing)
			if !s.ShowMissing {
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

			// Show untracked packages when --unmanaged flag is set
			sortedUntrackedManagers := sortItemsByManager(untrackedPackages)
			for _, manager := range sortedUntrackedManagers {
				packages := untrackedPackages[manager]
				for _, pkg := range packages {
					pkgBuilder.AddRow(pkg.Name, manager, "untracked")
				}
			}

			output.WriteString(pkgBuilder.Build())
			output.WriteString("\n")
		}
	}

	// Show dotfiles table if requested
	if s.ShowDotfiles && dotfileResult != nil {
		// Determine which items to show based on flags
		var itemsToShow []Item
		if s.ShowUnmanaged {
			itemsToShow = dotfileResult.Untracked
		} else if s.ShowMissing {
			itemsToShow = dotfileResult.Missing
		} else {
			// Include managed and missing items
			itemsToShow = append(dotfileResult.Managed, dotfileResult.Missing...)
			// Also need to check for drifted items (they won't be in Managed due to GroupItemsByState)
			// We'll handle drifted items separately below
		}

		if len(itemsToShow) > 0 {
			output.WriteString("DOTFILES\n")
			output.WriteString("--------\n")

			// Create a table for dotfiles
			dotBuilder := NewStandardTableBuilder("")

			if s.ShowUnmanaged {
				// For unmanaged, use single column showing just the path
				dotBuilder.SetHeaders("UNMANAGED DOTFILES")

				// Load config to check expand directories - simplified without config import
				// cfg := config.LoadWithDefaults(s.ConfigDir)

				// Sort untracked dotfiles
				sortItems(dotfileResult.Untracked)

				// Show untracked dotfiles
				for _, item := range dotfileResult.Untracked {
					// Show the dotfile path with ~ notation
					path := "~/" + item.Name

					// Add trailing slash for directories (simplified without config)
					if itemPath, ok := item.Metadata["path"].(string); ok {
						if info, err := os.Stat(itemPath); err == nil && info.IsDir() {
							path += "/"
						}
					}

					dotBuilder.AddRow(path)
				}
			} else {
				// For managed/missing, use the three-column format
				dotBuilder.SetHeaders("SOURCE", "TARGET", "STATUS")

				// Sort managed and missing dotfiles
				sortItems(dotfileResult.Managed)
				sortItems(dotfileResult.Missing)

				// Show managed dotfiles (unless showing only missing)
				if !s.ShowMissing {
					for _, item := range dotfileResult.Managed {
						// Use source from metadata if available, otherwise fall back to Name
						source := item.Name
						if src, ok := item.Metadata["source"].(string); ok {
							source = src
						}
						target := ""
						if dest, ok := item.Metadata["destination"].(string); ok {
							target = dest
						}
						// Check if this is actually a drifted file
						status := "deployed"
						if item.State == StateDegraded {
							status = "drifted"
						}
						dotBuilder.AddRow(source, target, status)
					}
				}

				// Show missing dotfiles
				for _, item := range dotfileResult.Missing {
					// Use source from metadata if available, otherwise fall back to Name
					source := item.Name
					if src, ok := item.Metadata["source"].(string); ok {
						source = src
					}
					target := ""
					if dest, ok := item.Metadata["destination"].(string); ok {
						target = dest
					}
					dotBuilder.AddRow(source, target, "missing")
				}
			}

			output.WriteString(dotBuilder.Build())
			output.WriteString("\n")
		}
	}

	// Add summary (skip for unmanaged or missing to avoid misleading counts)
	if !s.ShowUnmanaged && !s.ShowMissing {
		summary := s.StateSummary

		// Count drifted items separately
		driftedCount := 0
		for _, result := range s.StateSummary.Results {
			if result.Domain == "dotfile" {
				for _, item := range result.Managed {
					if item.State == StateDegraded {
						driftedCount++
					}
				}
			}
		}

		// Adjust managed count to exclude drifted
		managedCount := summary.TotalManaged - driftedCount

		output.WriteString("Summary: ")
		output.WriteString(fmt.Sprintf("%d managed", managedCount))
		if summary.TotalMissing > 0 {
			output.WriteString(fmt.Sprintf(", %d missing", summary.TotalMissing))
		}
		if driftedCount > 0 {
			output.WriteString(fmt.Sprintf(", %d drifted", driftedCount))
		}
		output.WriteString("\n")
	}

	// If no output was generated (except for title), show helpful message
	outputStr := output.String()
	if outputStr == "Plonk Status\n============\n\n" || outputStr == "" {
		output.Reset()
		output.WriteString("Plonk Status\n")
		output.WriteString("============\n\n")
		output.WriteString("No items match the specified filters.\n")
		if s.ShowMissing {
			output.WriteString("(Great! Everything tracked is installed/deployed)\n")
		}
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (f StatusFormatter) StructuredData() any {
	s := f.Data
	// Filter items based on flags
	var items []ManagedItem

	// Process each result domain
	for _, result := range s.StateSummary.Results {
		// Add managed items unless we're only showing missing or untracked
		if !s.ShowMissing && !s.ShowUnmanaged {
			for _, item := range result.Managed {
				managedItem := ManagedItem{
					Name:     item.Name,
					Domain:   result.Domain,
					State:    string(item.State),
					Manager:  item.Manager,
					Path:     item.Path,
					Metadata: item.Metadata,
				}
				// Add target for dotfiles
				if target, ok := item.Metadata["destination"].(string); ok {
					managedItem.Target = target
				}
				items = append(items, managedItem)
			}
		}

		// Add missing items unless we're only showing untracked
		if !s.ShowUnmanaged {
			for _, item := range result.Missing {
				managedItem := ManagedItem{
					Name:     item.Name,
					Domain:   result.Domain,
					State:    string(item.State),
					Manager:  item.Manager,
					Path:     item.Path,
					Metadata: item.Metadata,
				}
				// Add target for dotfiles
				if target, ok := item.Metadata["destination"].(string); ok {
					managedItem.Target = target
				}
				items = append(items, managedItem)
			}
		}

		// Add untracked items if we're showing unmanaged or showing all
		if s.ShowUnmanaged || (!s.ShowMissing && !s.ShowPackages && !s.ShowDotfiles) {
			for _, item := range result.Untracked {
				managedItem := ManagedItem{
					Name:     item.Name,
					Domain:   result.Domain,
					State:    string(item.State),
					Manager:  item.Manager,
					Path:     item.Path,
					Metadata: item.Metadata,
				}
				items = append(items, managedItem)
			}
		}
	}

	// Return summary format for structured output
	return StatusOutputSummary{
		ConfigPath:   s.ConfigPath,
		LockPath:     s.LockPath,
		ConfigExists: s.ConfigExists,
		ConfigValid:  s.ConfigValid,
		LockExists:   s.LockExists,
		StateSummary: s.StateSummary,
	}
}
