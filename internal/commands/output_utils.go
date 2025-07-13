// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"strings"
)

// Common status icons used across all commands
const (
	IconSuccess      = "✓"
	IconWarning      = "⚠"
	IconError        = "✗"
	IconInfo         = "•"
	IconUnknown      = "?"
	IconSkipped      = "-"
	IconHealthy      = "✅"
	IconUnhealthy    = "❌"
	IconSearch       = "🔍"
	IconPackage      = "📦"
	IconWarningEmoji = "⚠️"
)

// GetStatusIcon returns the appropriate icon for a given status
func GetStatusIcon(status string) string {
	switch status {
	case "managed", "added", "installed", "removed", "success", "completed":
		return IconSuccess
	case "missing", "warn", "warning":
		return IconWarning
	case "failed", "error", "fail":
		return IconError
	case "untracked", "unknown":
		return IconUnknown
	case "skipped", "already-configured", "already-installed":
		return IconInfo
	default:
		return IconSkipped
	}
}

// GetActionIcon returns the appropriate icon for an action message
func GetActionIcon(action string) string {
	actionLower := strings.ToLower(action)
	if strings.Contains(actionLower, "error") || strings.Contains(actionLower, "failed") {
		return IconError
	} else if strings.Contains(actionLower, "already") || strings.Contains(actionLower, "not found") {
		return IconInfo
	} else {
		return IconSuccess
	}
}

// TableBuilder helps construct consistent table outputs
type TableBuilder struct {
	output strings.Builder
}

// NewTableBuilder creates a new TableBuilder
func NewTableBuilder() *TableBuilder {
	return &TableBuilder{}
}

// AddTitle adds a title with underline
func (t *TableBuilder) AddTitle(title string) *TableBuilder {
	t.output.WriteString(title + "\n")
	t.output.WriteString(strings.Repeat("=", len(title)) + "\n")
	return t
}

// AddSeparator adds a separator line
func (t *TableBuilder) AddSeparator(char string, length int) *TableBuilder {
	t.output.WriteString(strings.Repeat(char, length) + "\n")
	return t
}

// AddLine adds a single line
func (t *TableBuilder) AddLine(format string, args ...interface{}) *TableBuilder {
	t.output.WriteString(fmt.Sprintf(format, args...) + "\n")
	return t
}

// AddNewline adds an empty line
func (t *TableBuilder) AddNewline() *TableBuilder {
	t.output.WriteString("\n")
	return t
}

// AddActionList adds a list of actions with appropriate icons
func (t *TableBuilder) AddActionList(actions []string) *TableBuilder {
	for _, action := range actions {
		icon := GetActionIcon(action)
		t.output.WriteString(fmt.Sprintf("%s %s\n", icon, action))
	}
	return t
}

// AddSummaryLine adds a summary line with optional counts
func (t *TableBuilder) AddSummaryLine(prefix string, counts map[string]int) *TableBuilder {
	parts := []string{prefix}
	for label, count := range counts {
		if count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", count, label))
		}
	}
	t.output.WriteString(strings.Join(parts, " | ") + "\n")
	return t
}

// Build returns the constructed output
func (t *TableBuilder) Build() string {
	return t.output.String()
}

// OperationSummary is a generic summary for operations
type OperationSummary struct {
	Total     int `json:"total" yaml:"total"`
	Succeeded int `json:"succeeded" yaml:"succeeded"`
	Skipped   int `json:"skipped" yaml:"skipped"`
	Failed    int `json:"failed" yaml:"failed"`
}

// CommonSummary provides fields common to many operations
type CommonSummary struct {
	Added   int `json:"added,omitempty" yaml:"added,omitempty"`
	Updated int `json:"updated,omitempty" yaml:"updated,omitempty"`
	Removed int `json:"removed,omitempty" yaml:"removed,omitempty"`
	Skipped int `json:"skipped,omitempty" yaml:"skipped,omitempty"`
	Failed  int `json:"failed,omitempty" yaml:"failed,omitempty"`
}

// StateSummary provides fields for state-based operations
type StateSummary struct {
	Total     int `json:"total" yaml:"total"`
	Managed   int `json:"managed" yaml:"managed"`
	Missing   int `json:"missing" yaml:"missing"`
	Untracked int `json:"untracked,omitempty" yaml:"untracked,omitempty"`
}

// CalculateOperationSummary calculates a generic summary from status counts
func CalculateOperationSummary(statusCounts map[string]int) OperationSummary {
	summary := OperationSummary{}

	// Calculate total
	for _, count := range statusCounts {
		summary.Total += count
	}

	// Map common statuses to summary fields
	successStatuses := []string{"added", "installed", "removed", "updated", "deployed", "would-add", "would-remove"}
	skipStatuses := []string{"skipped", "already-configured", "already-installed", "already-managed"}
	failStatuses := []string{"failed", "error"}

	for status, count := range statusCounts {
		for _, s := range successStatuses {
			if status == s {
				summary.Succeeded += count
				break
			}
		}
		for _, s := range skipStatuses {
			if status == s {
				summary.Skipped += count
				break
			}
		}
		for _, s := range failStatuses {
			if status == s {
				summary.Failed += count
				break
			}
		}
	}

	return summary
}

// FormatTableHeader creates a consistent table header
func FormatTableHeader(columns ...string) string {
	header := strings.Join(columns, " ")
	separator := ""
	for _, col := range columns {
		if separator != "" {
			separator += " "
		}
		separator += strings.Repeat("-", len(col))
	}
	return header + "\n" + separator + "\n"
}

// TruncateString truncates a string to a specified length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
