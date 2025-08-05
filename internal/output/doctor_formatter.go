// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
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

// TableOutput generates human-friendly table output for doctor command
func (f DoctorFormatter) TableOutput() string {
	d := f.Data
	var output strings.Builder

	// Overall status
	output.WriteString("Plonk Doctor Report\n\n")

	switch d.Overall.Status {
	case "healthy":
		green := color.New(color.FgGreen, color.Bold)
		output.WriteString(green.Sprintf("Overall Status: HEALTHY\n"))
	case "warning":
		yellow := color.New(color.FgYellow, color.Bold)
		output.WriteString(yellow.Sprintf("Overall Status: WARNING\n"))
	case "unhealthy":
		red := color.New(color.FgRed, color.Bold)
		output.WriteString(red.Sprintf("Overall Status: UNHEALTHY\n"))
	}
	output.WriteString(fmt.Sprintf("   %s\n\n", d.Overall.Message))

	// Group checks by category
	categories := make(map[string][]HealthCheck)
	for _, check := range d.Checks {
		categories[check.Category] = append(categories[check.Category], check)
	}

	// Display each category
	categoryOrder := []string{"system", "environment", "permissions", "configuration", "package-managers", "installation"}
	for _, category := range categoryOrder {
		if checks, exists := categories[category]; exists {
			output.WriteString(fmt.Sprintf("## %s\n", strings.Title(strings.ReplaceAll(category, "-", " "))))

			for _, check := range checks {
				// Color-coded status
				var statusColor *color.Color
				var statusText string
				switch check.Status {
				case "pass":
					statusColor = color.New(color.FgGreen)
					statusText = "PASS"
				case "warn":
					statusColor = color.New(color.FgYellow)
					statusText = "WARN"
				case "fail":
					statusColor = color.New(color.FgRed)
					statusText = "FAIL"
				case "info":
					statusColor = color.New(color.FgBlue)
					statusText = "INFO"
				default:
					statusColor = color.New(color.FgWhite)
					statusText = "UNKNOWN"
				}

				coloredName := statusColor.Sprintf("### %s", check.Name)
				coloredStatus := statusColor.Sprintf("**Status**: %s", statusText)

				output.WriteString(fmt.Sprintf("%s\n", coloredName))
				output.WriteString(fmt.Sprintf("%s\n", coloredStatus))
				output.WriteString(fmt.Sprintf("**Message**: %s\n", check.Message))

				if len(check.Details) > 0 {
					output.WriteString("\n**Details:**\n")
					for _, detail := range check.Details {
						output.WriteString(fmt.Sprintf("- %s\n", detail))
					}
				}

				if len(check.Issues) > 0 {
					output.WriteString("\n**Issues:**\n")
					for _, issue := range check.Issues {
						output.WriteString(fmt.Sprintf("- %s\n", issue))
					}
				}

				if len(check.Suggestions) > 0 {
					output.WriteString("\n**Suggestions:**\n")
					for _, suggestion := range check.Suggestions {
						output.WriteString(fmt.Sprintf("- %s\n", suggestion))
					}
				}

				output.WriteString("\n")
			}
		}
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (f DoctorFormatter) StructuredData() any {
	return f.Data
}
