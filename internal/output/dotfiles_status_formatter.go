// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
)

// DotfilesStatusOutput represents the output structure for dotfiles status command
type DotfilesStatusOutput struct {
	Result     Result `json:"result" yaml:"result"`
	RemoteSync string `json:"remote_sync,omitempty" yaml:"remote_sync,omitempty"`
	HomeDir    string `json:"-" yaml:"-"` // Not included in JSON/YAML output
}

// DotfilesStatusFormatter formats dotfiles status output
type DotfilesStatusFormatter struct {
	Data DotfilesStatusOutput
}

// NewDotfilesStatusFormatter creates a new formatter
func NewDotfilesStatusFormatter(data DotfilesStatusOutput) DotfilesStatusFormatter {
	return DotfilesStatusFormatter{Data: data}
}

// TableOutput generates human-friendly table output for dotfiles status
func (f DotfilesStatusFormatter) TableOutput() string {
	var output strings.Builder
	result := f.Data.Result

	WriteTitle(&output, "Dotfiles Status")
	WriteRemoteSync(&output, f.Data.RemoteSync)

	writeDotfilesTable(&output, result, f.Data.HomeDir)

	// Add summary
	// Count drifted items separately
	driftedCount := 0
	for _, item := range result.Managed {
		if item.State == StateDegraded {
			driftedCount++
		}
	}

	// Adjust managed count to exclude drifted
	managedCount := len(result.Managed) - driftedCount

	output.WriteString("Summary: ")
	fmt.Fprintf(&output, "%d managed", managedCount)
	if len(result.Missing) > 0 {
		fmt.Fprintf(&output, ", %d missing", len(result.Missing))
	}
	if driftedCount > 0 {
		fmt.Fprintf(&output, ", %d drifted", driftedCount)
	}
	if len(result.Errors) > 0 {
		fmt.Fprintf(&output, ", %d error(s)", len(result.Errors))
	}
	output.WriteString("\n")

	if len(result.Managed) == 0 && len(result.Missing) == 0 && len(result.Errors) == 0 {
		output.Reset()
		WriteTitle(&output, "Dotfiles Status")
		WriteRemoteSync(&output, f.Data.RemoteSync)
		output.WriteString("No managed dotfiles.\n")
	}

	return output.String()
}

func sourceType(item Item) string {
	if value, ok := item.Metadata["source_type"].(string); ok && value != "" {
		return value
	}
	if source, ok := item.Metadata["source"].(string); ok && strings.HasSuffix(source, ".tmpl") {
		return "template"
	}
	return "file"
}
