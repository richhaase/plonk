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

	writePackagesTable(&output, result)

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
