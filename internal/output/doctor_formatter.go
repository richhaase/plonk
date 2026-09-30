// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"

	"sort"
)

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status  string `json:"status" yaml:"status"`
	Message string `json:"message" yaml:"message"`
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Name        string   `json:"name" yaml:"name"`
	Category    string   `json:"category" yaml:"category"`
	Status      string   `json:"status" yaml:"status"`
	Message     string   `json:"message" yaml:"message"`
	Details     []string `json:"details,omitempty" yaml:"details,omitempty"`
	Issues      []string `json:"issues,omitempty" yaml:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}

// DoctorOutput represents the output of the doctor command (health checks)
type DoctorOutput struct {
	Overall HealthStatus  `json:"overall" yaml:"overall"`
	Checks  []HealthCheck `json:"checks" yaml:"checks"`
}

// DoctorFormatter formats doctor output
type DoctorFormatter struct {
	Data DoctorOutput
}

// NewDoctorFormatter creates a new formatter
func NewDoctorFormatter(data DoctorOutput) DoctorFormatter {
	return DoctorFormatter{Data: data}
}

// TableOutput groups diagnostic checks without printing Markdown syntax.
func (f DoctorFormatter) TableOutput() string {
	d := f.Data
	var w strings.Builder
	WriteAction(&w, d.Overall.Status, "System readiness", d.Overall.Message, false)
	categories := make(map[string][]HealthCheck)
	for _, check := range d.Checks {
		categories[check.Category] = append(categories[check.Category], check)
	}
	order := []string{"system", "environment", "permissions", "configuration", "package-managers", "installation", "dotfiles"}
	var extra []string
	for category := range categories {
		found := false
		for _, known := range order {
			if category == known {
				found = true
				break
			}
		}
		if !found {
			extra = append(extra, category)
		}
	}
	sort.Strings(extra)
	order = append(order, extra...)
	for _, category := range order {
		checks := categories[category]
		if len(checks) == 0 {
			continue
		}
		fmt.Fprintf(&w, "\n%s\n", titleCase(strings.ReplaceAll(category, "-", " ")))
		for _, check := range checks {
			WriteAction(&w, check.Status, check.Name, check.Message, true)
			for _, detail := range check.Details {
				fmt.Fprintf(&w, "  %s\n", detail)
			}
			for _, issue := range check.Issues {
				fmt.Fprintf(&w, "  %s\n", issue)
			}
			for _, suggestion := range check.Suggestions {
				fmt.Fprintf(&w, "  Next: %s\n", suggestion)
			}
		}
	}
	return w.String()
}

// titleCase converts a string to title case (first letter of each word uppercase)
// This is a simple replacement for the deprecated strings.Title
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
