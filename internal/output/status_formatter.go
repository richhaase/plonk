// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"reflect"
	"sort"
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
	ConfigPath   string  `json:"config_path" yaml:"config_path"`
	LockPath     string  `json:"lock_path" yaml:"lock_path"`
	ConfigExists bool    `json:"config_exists" yaml:"config_exists"`
	ConfigValid  bool    `json:"config_valid" yaml:"config_valid"`
	LockExists   bool    `json:"lock_exists" yaml:"lock_exists"`
	StateSummary Summary `json:"state_summary" yaml:"state_summary"`
	ConfigDir    string  `json:"-" yaml:"-"` // Not included in JSON/YAML output
	HomeDir      string  `json:"-" yaml:"-"` // Not included in JSON/YAML output
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
	Error    string                 `json:"error,omitempty" yaml:"error,omitempty"`
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

	writeStatusTitle(&output)

	if packageResult := findResultByDomain(s.StateSummary.Results, "package"); packageResult != nil {
		writePackagesTable(&output, *packageResult)
	}
	if dotfileResult := findResultByDomain(s.StateSummary.Results, "dotfile"); dotfileResult != nil {
		writeDotfilesTable(&output, *dotfileResult, s.HomeDir)
	}

	driftedCount := countDriftedDotfiles(s.StateSummary.Results)
	writeSummaryLine(&output, s.StateSummary, driftedCount)
	writeDomainErrors(&output, s.StateSummary.Results)

	if output.String() == "Plonk Status\n============\n\n" {
		output.WriteString("No managed items.\n")
	}

	return output.String()
}

func writeStatusTitle(output *strings.Builder) {
	output.WriteString("Plonk Status\n")
	output.WriteString("============\n\n")
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

	pkgBuilder := NewStandardTableBuilder("")
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

	dotBuilder := NewStandardTableBuilder("")
	dotBuilder.SetHeaders("DOTFILE", "STATUS")

	managed := append([]Item(nil), result.Managed...)
	missing := append([]Item(nil), result.Missing...)
	sortItems(managed)
	sortItems(missing)

	for _, item := range managed {
		dotBuilder.AddRow(dotfileTarget(item, homeDir), dotfileStatus(item))
	}
	for _, item := range missing {
		dotBuilder.AddRow(dotfileTarget(item, homeDir), "missing")
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
		if len(result.Errors) == 0 {
			continue
		}
		fmt.Fprintf(output, "\n%s errors:\n", result.Domain)
		for _, item := range result.Errors {
			if item.Error != "" {
				fmt.Fprintf(output, "  ✗ %s: %s\n", item.Name, item.Error)
				continue
			}
			fmt.Fprintf(output, "  ✗ %s\n", item.Name)
		}
	}
}

// StructuredData returns the structured data for serialization
func (f StatusFormatter) StructuredData() any {
	s := f.Data
	return StatusOutputSummary{
		ConfigPath:   s.ConfigPath,
		LockPath:     s.LockPath,
		ConfigExists: s.ConfigExists,
		ConfigValid:  s.ConfigValid,
		LockExists:   s.LockExists,
		StateSummary: sanitizeSummary(s.StateSummary),
	}
}

// sanitizeMetadata returns a shallow copy of metadata without function-typed values
func sanitizeMetadata(meta map[string]interface{}) map[string]interface{} {
	if meta == nil {
		return nil
	}
	cleaned := make(map[string]interface{}, len(meta))
	for k, v := range meta {
		if reflect.ValueOf(v).Kind() == reflect.Func {
			continue
		}
		cleaned[k] = v
	}
	return cleaned
}

// sanitizeSummary removes function-typed metadata values from summary items
func sanitizeSummary(sum Summary) Summary {
	cleaned := Summary{
		TotalManaged:   sum.TotalManaged,
		TotalMissing:   sum.TotalMissing,
		TotalUntracked: sum.TotalUntracked,
		TotalErrors:    sum.TotalErrors,
		Results:        make([]Result, len(sum.Results)),
	}
	for i, r := range sum.Results {
		cr := Result{Domain: r.Domain}
		if len(r.Managed) > 0 {
			cr.Managed = make([]Item, len(r.Managed))
			for j, it := range r.Managed {
				it.Metadata = sanitizeMetadata(it.Metadata)
				cr.Managed[j] = it
			}
		}
		if len(r.Missing) > 0 {
			cr.Missing = make([]Item, len(r.Missing))
			for j, it := range r.Missing {
				it.Metadata = sanitizeMetadata(it.Metadata)
				cr.Missing[j] = it
			}
		}
		if len(r.Untracked) > 0 {
			cr.Untracked = make([]Item, len(r.Untracked))
			for j, it := range r.Untracked {
				it.Metadata = sanitizeMetadata(it.Metadata)
				cr.Untracked[j] = it
			}
		}
		if len(r.Errors) > 0 {
			cr.Errors = make([]Item, len(r.Errors))
			for j, it := range r.Errors {
				it.Metadata = sanitizeMetadata(it.Metadata)
				cr.Errors[j] = it
			}
		}
		cleaned.Results[i] = cr
	}
	return cleaned
}
