// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/richhaase/plonk/internal/diagnostics"
	"github.com/richhaase/plonk/internal/setup"
	"github.com/spf13/cobra"
)

var (
	doctorFix bool
	doctorYes bool
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

With --fix flag, offers to install missing package managers.

Examples:
  plonk doctor           # Run health checks
  plonk doctor --fix     # Run checks and offer to fix issues
  plonk doctor --fix --yes  # Auto-fix issues without prompts
  plonk doctor -o json   # Show as JSON
  plonk doctor -o yaml   # Show as YAML`,
	RunE: runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Offer to install missing package managers")
	doctorCmd.Flags().BoolVar(&doctorYes, "yes", false, "Auto-install missing tools without prompts (requires --fix)")
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

	// If --fix flag is set, try to fix issues
	if doctorFix {
		ctx := context.Background()
		setupConfig := setup.Config{
			Interactive: !doctorYes,
			Verbose:     false,
		}

		if err := setup.CheckAndInstallToolsFromReport(ctx, healthReport, setupConfig); err != nil {
			return fmt.Errorf("failed to fix issues: %w", err)
		}

		// Re-run health checks to show updated status
		fmt.Println("\nRe-running health checks after fixes...")
		healthReport = diagnostics.RunHealthChecks()
	}

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
func (d DoctorOutput) StructuredData() any {
	return d
}
