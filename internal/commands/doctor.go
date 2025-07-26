// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/diagnostics"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system readiness for using plonk",
	Long: `Perform health checks to ensure your system is properly configured
for plonk. This includes checking for required package managers,
configuration files, and system compatibility.

Shows:
- System information (OS, arch, etc.)
- Package manager availability
- Configuration file status and location
- Environment variables (PLONK_DIR, etc.)
- Any issues that would prevent plonk from working

Examples:
  plonk doctor         # Run health checks
  plonk doctor -o json # Show as JSON
  plonk doctor -o yaml # Show as YAML`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Run comprehensive health checks using diagnostics
	healthReport := diagnostics.RunHealthChecks()

	// Convert to command output type
	doctorOutput := DoctorOutput{
		Overall: healthReport.Overall,
		Checks:  healthReport.Checks,
	}

	return RenderOutput(doctorOutput, format)
}

// DoctorOutput represents the output of the doctor command (health checks)
type DoctorOutput struct {
	Overall diagnostics.HealthStatus  `json:"overall" yaml:"overall"`
	Checks  []diagnostics.HealthCheck `json:"checks" yaml:"checks"`
}

// TableOutput generates human-friendly table output for doctor command
func (d DoctorOutput) TableOutput() string {
	var output strings.Builder

	// Overall status
	output.WriteString("Plonk Doctor Report\n\n")

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
	categories := make(map[string][]diagnostics.HealthCheck)
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
