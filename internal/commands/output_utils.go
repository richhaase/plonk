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
