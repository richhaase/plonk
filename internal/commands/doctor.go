// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/richhaase/plonk/internal/diagnostics"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

// No flags needed for doctor command anymore

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

Doctor reports issues with suggestions on how to fix them.
To automatically install missing package managers, use 'plonk clone'.

Examples:
  plonk doctor           # Run health checks
  plonk doctor -o json   # Show as JSON
  plonk doctor -o yaml   # Show as YAML`,
	RunE:         runDoctor,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
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

	// Convert to output package type and create formatter
	formatterData := output.DoctorOutput{
		Overall: output.HealthStatus{
			Status:  doctorOutput.Overall.Status,
			Message: doctorOutput.Overall.Message,
		},
		Checks: convertHealthChecks(doctorOutput.Checks),
	}
	formatter := output.NewDoctorFormatter(formatterData)
	return output.RenderOutput(formatter, format)
}

// convertHealthChecks converts from diagnostics types to output types
func convertHealthChecks(checks []diagnostics.HealthCheck) []output.HealthCheck {
	converted := make([]output.HealthCheck, len(checks))
	for i, check := range checks {
		converted[i] = output.HealthCheck{
			Name:        check.Name,
			Category:    check.Category,
			Status:      check.Status,
			Message:     check.Message,
			Details:     check.Details,
			Issues:      check.Issues,
			Suggestions: check.Suggestions,
		}
	}
	return converted
}

// DoctorOutput represents the output of the doctor command (health checks)
type DoctorOutput struct {
	Overall diagnostics.HealthStatus  `json:"overall" yaml:"overall"`
	Checks  []diagnostics.HealthCheck `json:"checks" yaml:"checks"`
}
