// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/richhaase/plonk/internal/config"
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
		// Ensure deterministic display
		results := append([]SearchResultEntry(nil), s.Results...)
		sort.Slice(results, func(i, j int) bool { return results[i].Manager < results[j].Manager })
		for _, result := range results {
			output.WriteString(fmt.Sprintf("\n%s:\n", result.Manager))
			pkgs := append([]string(nil), result.Packages...)
			sort.Strings(pkgs)
			for _, pkg := range pkgs {
				output.WriteString(fmt.Sprintf("  • %s\n", pkg))
			}
		}
		output.WriteString(fmt.Sprintf("\nInstall examples:\n"))
		for _, result := range results {
			output.WriteString(fmt.Sprintf("  • plonk install %s:%s\n", result.Manager, s.Package))
		}

	case "not-found":
		output.WriteString(fmt.Sprintf("%s\n", s.Message))

	case "no-managers":
		output.WriteString(fmt.Sprintf("%s\n", s.Message))
		output.WriteString("\n")
		output.WriteString(buildManagerInstallHint())

	case "manager-unavailable":
		output.WriteString(fmt.Sprintf("%s\n", s.Message))

	default:
		output.WriteString(fmt.Sprintf("%s\n", s.Message))
	}

	return output.String()
}

// buildManagerInstallHint builds a hint message listing example managers from
// the current configuration, falling back to a generic message when none are
// configured.
func buildManagerInstallHint() string {
	cfg := config.LoadWithDefaults(config.GetDefaultConfigDirectory())
	if cfg == nil || cfg.Managers == nil || len(cfg.Managers) == 0 {
		return "Please install a supported package manager to search for packages.\n"
	}

	var names []string
	for name := range cfg.Managers {
		names = append(names, name)
	}
	sort.Strings(names)

	max := 2
	if len(names) < max {
		max = len(names)
	}
	examples := strings.Join(names[:max], " or ")

	return fmt.Sprintf("Please install a supported package manager (e.g. %s) to search for packages.\n", examples)
}

// StructuredData returns the structured data for serialization
func (f SearchFormatter) StructuredData() any {
	return f.Data
}
