// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"path/filepath"
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

// tildeShorthand replaces the home directory prefix with ~ for display
func tildeShorthand(path, homeDir string) string {
	if homeDir == "" {
		return path
	}
	if path == homeDir || strings.HasPrefix(path, homeDir+string(filepath.Separator)) {
		return "~" + strings.TrimPrefix(path, homeDir)
	}
	return path
}

// TableOutput generates human-friendly table output for status
func (f StatusFormatter) TableOutput() string {
	s := f.Data
	var output strings.Builder

	WriteRemoteSync(&output, s.RemoteSync)

	if packageResult := findResultByDomain(s.StateSummary.Results, "package"); packageResult != nil {
		result := *packageResult
		if !s.ShowAll {
			result.Managed = nil
		}
		writePackages(&output, result, s.ShowAll)
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
		writeDotfiles(&output, result, s.HomeDir, s.ShowAll)
	}

	driftedCount := countDriftedDotfiles(s.StateSummary.Results)
	if s.ShowAll {
		writeSummaryLine(&output, s.StateSummary, driftedCount)
	}

	if s.StateSummary.TotalManaged == 0 && s.StateSummary.TotalMissing == 0 && s.StateSummary.TotalErrors == 0 {
		output.Reset()
		WriteRemoteSync(&output, s.RemoteSync)
		output.WriteString("No managed items.\n")
	}

	if !s.ShowAll && s.StateSummary.TotalManaged > 0 && s.StateSummary.TotalMissing == 0 && s.StateSummary.TotalErrors == 0 && driftedCount == 0 {
		output.WriteString("All managed items are in sync.\n")
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

// statusItems normalizes states from their result buckets without changing input.
func statusItems(result Result) []Item {
	items := append([]Item(nil), result.Managed...)
	for i := range items {
		if items[i].State == "" {
			items[i].State = StateManaged
		}
	}
	for _, item := range result.Missing {
		item.State = StateMissing
		items = append(items, item)
	}
	for _, item := range result.Errors {
		item.State = StateError
		items = append(items, item)
	}
	return items
}

func writePackagesTable(w *strings.Builder, result Result) { writePackages(w, result, true) }

func writePackages(w *strings.Builder, result Result, ledger bool) {
	items := statusItems(result)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Manager != items[j].Manager {
			return items[i].Manager < items[j].Manager
		}
		return items[i].Name < items[j].Name
	})
	for _, item := range items {
		state := string(item.State)
		if item.Error != "" {
			state = "error"
		}
		name := item.Name
		if item.Manager != "" {
			name = item.Manager + ":" + name
		}
		WriteAction(w, state, name, item.Error, ledger)
	}
	if len(items) > 0 {
		w.WriteString("\n")
	}
}

func writeDotfilesTable(w *strings.Builder, result Result, homeDir string) {
	writeDotfiles(w, result, homeDir, true)
}

func writeDotfiles(w *strings.Builder, result Result, homeDir string, ledger bool) {
	items := statusItems(result)
	sortItems(items)
	for _, item := range items {
		state := dotfileStatus(item)
		if item.State == StateMissing {
			state = "missing"
		}
		if item.State == StateError || item.Error != "" {
			state = "error"
		}
		detail := sourceType(item)
		if item.Error != "" {
			detail += ": " + item.Error
		}
		WriteAction(w, state, dotfileTarget(item, homeDir), detail, ledger)
	}
	if len(items) > 0 {
		w.WriteString("\n")
	}
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
		if item.Metadata["drift_status"] == "error" {
			return "error"
		}
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
