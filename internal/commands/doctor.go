// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and diagnose issues",
	Long: `Run comprehensive health checks to diagnose potential issues with plonk.

This command checks:
- System requirements and environment
- Package manager availability and functionality
- Configuration file validity and permissions
- Common path and permission issues
- Suggested fixes for detected problems

The doctor command helps troubleshoot installation and configuration issues,
making it easier to get plonk working correctly on your system.

Examples:
  plonk doctor              # Run all health checks
  plonk doctor -o json      # Output results as JSON for scripting`,
	RunE: runDoctor,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Run comprehensive health checks using orchestrator
	healthReport := orchestrator.RunHealthChecks()

	// Convert to command output type
	doctorOutput := DoctorOutput{
		Overall: HealthStatus{
			Status:  healthReport.Overall.Status,
			Message: healthReport.Overall.Message,
		},
		Checks: make([]HealthCheck, len(healthReport.Checks)),
	}

	for i, check := range healthReport.Checks {
		doctorOutput.Checks[i] = HealthCheck{
			Name:        check.Name,
			Category:    check.Category,
			Status:      check.Status,
			Message:     check.Message,
			Details:     check.Details,
			Issues:      check.Issues,
			Suggestions: check.Suggestions,
		}
	}

	return RenderOutput(doctorOutput, format)
}

// DoctorOutput represents the output of the doctor command
type DoctorOutput struct {
	Overall HealthStatus  `json:"overall" yaml:"overall"`
	Checks  []HealthCheck `json:"checks" yaml:"checks"`
}

type HealthStatus struct {
	Status  string `json:"status" yaml:"status"`
	Message string `json:"message" yaml:"message"`
}

type HealthCheck struct {
	Name        string   `json:"name" yaml:"name"`
	Category    string   `json:"category" yaml:"category"`
	Status      string   `json:"status" yaml:"status"`
	Message     string   `json:"message" yaml:"message"`
	Details     []string `json:"details,omitempty" yaml:"details,omitempty"`
	Issues      []string `json:"issues,omitempty" yaml:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}

// TableOutput generates human-friendly table output for doctor command
func (d DoctorOutput) TableOutput() string {
	var output strings.Builder

	// Overall status
	output.WriteString("# Plonk Doctor Report\n\n")

	switch d.Overall.Status {
	case "healthy":
		output.WriteString("🟢 Overall Status: HEALTHY\n")
	case "warning":
		output.WriteString("🟡 Overall Status: WARNING\n")
	case "unhealthy":
		output.WriteString("🔴 Overall Status: UNHEALTHY\n")
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
				// Status icon
				var icon string
				switch check.Status {
				case "pass":
					icon = "✅"
				case "warn":
					icon = "⚠️"
				case "fail":
					icon = "❌"
				case "info":
					icon = "ℹ️"
				default:
					icon = "❓"
				}

				output.WriteString(fmt.Sprintf("### %s %s\n", icon, check.Name))
				output.WriteString(fmt.Sprintf("**Status**: %s\n", strings.ToUpper(check.Status)))
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
						output.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
					}
				}

				if len(check.Suggestions) > 0 {
					output.WriteString("\n**Suggestions:**\n")
					for _, suggestion := range check.Suggestions {
						output.WriteString(fmt.Sprintf("- 💡 %s\n", suggestion))
					}
				}

				output.WriteString("\n")
			}
		}
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (d DoctorOutput) StructuredData() any {
	return d
}
