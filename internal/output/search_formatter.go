// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"
)

// SearchResultEntry represents a search result from a specific manager
type SearchResultEntry struct {
	Manager  string   `json:"manager" yaml:"manager"`
	Packages []string `json:"packages" yaml:"packages"`
}

// SearchOutput represents the output structure for search command
type SearchOutput struct {
	Package string              `json:"package" yaml:"package"`
	Status  string              `json:"status" yaml:"status"`
	Message string              `json:"message" yaml:"message"`
	Results []SearchResultEntry `json:"results,omitempty" yaml:"results,omitempty"`
}

// SearchFormatter formats search output
type SearchFormatter struct {
	Data SearchOutput
}

// NewSearchFormatter creates a new formatter
func NewSearchFormatter(data SearchOutput) SearchFormatter {
	return SearchFormatter{Data: data}
}

// TableOutput generates human-friendly table output for search command
func (f SearchFormatter) TableOutput() string {
	s := f.Data
	var output strings.Builder

	switch s.Status {
	case "found":
		output.WriteString(fmt.Sprintf("%s\n", s.Message))
		if len(s.Results) > 0 && len(s.Results[0].Packages) > 0 {
			output.WriteString("\nMatching packages:\n")
			for _, pkg := range s.Results[0].Packages {
				output.WriteString(fmt.Sprintf("  • %s\n", pkg))
			}
			output.WriteString(fmt.Sprintf("\nInstall with: plonk install %s:%s\n", s.Results[0].Manager, s.Package))
		}

	case "found-multiple":
		output.WriteString(fmt.Sprintf("%s\n", s.Message))
		output.WriteString("\nResults by manager:\n")
		for _, result := range s.Results {
			output.WriteString(fmt.Sprintf("\n%s:\n", result.Manager))
			for _, pkg := range result.Packages {
				output.WriteString(fmt.Sprintf("  • %s\n", pkg))
			}
		}
		output.WriteString(fmt.Sprintf("\nInstall examples:\n"))
		for _, result := range s.Results {
			output.WriteString(fmt.Sprintf("  • plonk install %s:%s\n", result.Manager, s.Package))
		}

	case "not-found":
		output.WriteString(fmt.Sprintf("%s\n", s.Message))

	case "no-managers":
		output.WriteString(fmt.Sprintf("%s\n", s.Message))
		output.WriteString("\nPlease install a package manager (Homebrew or NPM) to search for packages.\n")

	case "manager-unavailable":
		output.WriteString(fmt.Sprintf("%s\n", s.Message))

	default:
		output.WriteString(fmt.Sprintf("%s\n", s.Message))
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (f SearchFormatter) StructuredData() any {
	return f.Data
}
