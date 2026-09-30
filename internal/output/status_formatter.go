// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"sort"
	"strings"
)

// Local types to avoid import cycles

// ItemState represents resource item state
type ItemState string

const (
	StateManaged   ItemState = "managed"
	StateMissing   ItemState = "missing"
	StateDegraded  ItemState = "drifted"
	StateUntracked ItemState = "untracked"
	StateError     ItemState = "error"
)

// Item represents a resource item
type Item struct {
	Name     string                 `json:"name"`
	Manager  string                 `json:"manager,omitempty"`
	Path     string                 `json:"path,omitempty"`
	State    ItemState              `json:"state"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Result represents domain result
type Result struct {
	Domain    string `json:"domain"`
	Managed   []Item `json:"managed"`
	Missing   []Item `json:"missing"`
	Untracked []Item `json:"untracked"`
	Errors    []Item `json:"errors,omitempty"`
}

// Summary represents resource summary
type Summary struct {
	TotalManaged   int      `json:"total_managed"`
	TotalMissing   int      `json:"total_missing"`
	TotalUntracked int      `json:"total_untracked"`
	TotalErrors    int      `json:"total_errors,omitempty"`
	Results        []Result `json:"results"`
}

// StatusOutput represents the output structure for status command
type StatusOutput struct {
	RemoteSync   string
	ShowAll      bool
	StateSummary Summary
	HomeDir      string
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
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
}

// sortItemsByManager returns sorted manager names from the map
func sortItemsByManager(itemsByManager map[string][]Item) []string {
	managers := make([]string, 0, len(itemsByManager))
	for manager := range itemsByManager {
		managers = append(managers, manager)
	}
	sort.Strings(managers)
	return managers
}

// tildeShorthand replaces the home directory prefix with ~ for display
func tildeShorthand(path, homeDir string) string {
	if homeDir == "" {
		return path
	}
	if strings.HasPrefix(path, homeDir) {
		return "~" + strings.TrimPrefix(path, homeDir)
	}
	return path
}

// TableOutput generates human-friendly table output for status
func (f StatusFormatter) TableOutput() string {
	s := f.Data
	var output strings.Builder

	WriteTitle(&output, "Plonk Status")
	WriteRemoteSync(&output, s.RemoteSync)

	if packageResult := findResultByDomain(s.StateSummary.Results, "package"); packageResult != nil {
		result := *packageResult
		if !s.ShowAll {
			result.Managed = nil
		}
		writePackagesTable(&output, result)
	}
	if dotfileResult := findResultByDomain(s.StateSummary.Results, "dotfile"); dotfileResult != nil {
		result := *dotfileResult
		if !s.ShowAll {
			result.Managed = nil
			for _, item := range dotfileResult.Managed {
				if item.State == StateDegraded {
					result.Managed = append(result.Managed, item)
				}
			}
		}
		writeDotfilesTable(&output, result, s.HomeDir)
	}

	driftedCount := countDriftedDotfiles(s.StateSummary.Results)
	if s.ShowAll {
		writeSummaryLine(&output, s.StateSummary, driftedCount)
	}
	writeDomainErrors(&output, s.StateSummary.Results)

	if s.StateSummary.TotalManaged == 0 && s.StateSummary.TotalMissing == 0 && s.StateSummary.TotalErrors == 0 {
		output.Reset()
		WriteTitle(&output, "Plonk Status")
		WriteRemoteSync(&output, s.RemoteSync)
		output.WriteString("No managed items.\n")
	}

	return output.String()
}

func findResultByDomain(results []Result, domain string) *Result {
	for i := range results {
		if results[i].Domain == domain {
			return &results[i]
		}
	}
	return nil
}

func writePackagesTable(output *strings.Builder, result Result) {
	packagesByManager := make(map[string][]Item)
	for _, item := range result.Managed {
		packagesByManager[item.Manager] = append(packagesByManager[item.Manager], item)
	}

	missingPackages := append([]Item(nil), result.Missing...)
	sortItems(missingPackages)

	if len(packagesByManager) == 0 && len(missingPackages) == 0 {
		return
	}

	pkgBuilder := NewStandardTableBuilder()
	pkgBuilder.SetHeaders("PACKAGE", "MANAGER", "STATUS")

	for _, manager := range sortItemsByManager(packagesByManager) {
		packages := append([]Item(nil), packagesByManager[manager]...)
		sortItems(packages)
		for _, pkg := range packages {
			pkgBuilder.AddRow(pkg.Name, manager, "managed")
		}
	}

	for _, pkg := range missingPackages {
		pkgBuilder.AddRow(pkg.Name, pkg.Manager, "missing")
	}

	output.WriteString(pkgBuilder.Build())
	output.WriteString("\n")
}

func writeDotfilesTable(output *strings.Builder, result Result, homeDir string) {
	itemsToShow := len(result.Managed) + len(result.Missing)
	if itemsToShow == 0 {
		return
	}

	dotBuilder := NewStandardTableBuilder()
	dotBuilder.SetHeaders("DOTFILE", "TYPE", "STATUS")

	managed := append([]Item(nil), result.Managed...)
	missing := append([]Item(nil), result.Missing...)
	sortItems(managed)
	sortItems(missing)

	for _, item := range managed {
		dotBuilder.AddRow(dotfileTarget(item, homeDir), sourceType(item), dotfileStatus(item))
	}
	for _, item := range missing {
		dotBuilder.AddRow(dotfileTarget(item, homeDir), sourceType(item), "missing")
	}

	output.WriteString(dotBuilder.Build())
	output.WriteString("\n")
}

func dotfileTarget(item Item, homeDir string) string {
	target := item.Name
	if dest, ok := item.Metadata["destination"].(string); ok {
		target = tildeShorthand(dest, homeDir)
	}
	return target
}

func dotfileStatus(item Item) string {
	if item.State == StateDegraded {
		return "drifted"
	}
	return "deployed"
}

func countDriftedDotfiles(results []Result) int {
	drifted := 0
	for _, result := range results {
		if result.Domain != "dotfile" {
			continue
		}
		for _, item := range result.Managed {
			if item.State == StateDegraded {
				drifted++
			}
		}
	}
	return drifted
}

func writeSummaryLine(output *strings.Builder, summary Summary, driftedCount int) {
	managedCount := summary.TotalManaged - driftedCount
	output.WriteString("Summary: ")
	fmt.Fprintf(output, "%d managed", managedCount)
	if summary.TotalMissing > 0 {
		fmt.Fprintf(output, ", %d missing", summary.TotalMissing)
	}
	if driftedCount > 0 {
		fmt.Fprintf(output, ", %d drifted", driftedCount)
	}
	if summary.TotalErrors > 0 {
		fmt.Fprintf(output, ", %d errors", summary.TotalErrors)
	}
	output.WriteString("\n")
}

func writeDomainErrors(output *strings.Builder, results []Result) {
	for _, result := range results {
		WriteErrors(output, result.Domain, result.Errors)
	}
}
