// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/spf13/cobra"
)

// Status command implementation using unified state management system

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display overall plonk status",
	Long: `Display a compact overview of your plonk-managed environment.

Shows:
- Overall health status
- Configuration and lock file status
- Summary of managed and untracked items

For detailed lists, use 'plonk dot list' or 'plonk pkg list'.

Examples:
  plonk status           # Show compact status
  plonk status --health  # Include comprehensive health checks (doctor mode)
  plonk status -o json   # Show as JSON
  plonk status -o yaml   # Show as YAML`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("health", false, "Include comprehensive health checks (doctor mode)")
	statusCmd.Flags().Bool("check", false, "Alias for --health")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Check if health/check flag is set
	checkHealth, _ := cmd.Flags().GetBool("health")
	if !checkHealth {
		checkHealth, _ = cmd.Flags().GetBool("check")
	}

	// If health check requested, run doctor functionality
	if checkHealth {
		return runHealthChecks(format)
	}

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load configuration (may fail if config is invalid, but we handle this gracefully)
	_, configLoadErr := config.LoadConfig(configDir)

	// Reconcile all domains
	ctx := context.Background()
	results, err := orchestrator.ReconcileAll(ctx, homeDir, configDir)
	if err != nil {
		return err
	}

	// Convert results to summary for compatibility with existing output logic
	summary := resources.ConvertResultsToSummary(results)

	// Check file existence and validity
	configPath := filepath.Join(configDir, "plonk.yaml")
	lockPath := filepath.Join(configDir, "plonk.lock")

	configExists := false
	configValid := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
		// Config is valid only if it loaded without error
		configValid = (configLoadErr == nil)
	}

	lockExists := false
	if _, err := os.Stat(lockPath); err == nil {
		lockExists = true
	}

	// Prepare output
	outputData := StatusOutput{
		ConfigPath:   configPath,
		LockPath:     lockPath,
		ConfigExists: configExists,
		ConfigValid:  configValid,
		LockExists:   lockExists,
		StateSummary: summary,
	}

	return RenderOutput(outputData, format)
}

// runHealthChecks runs the doctor functionality when --health flag is set
func runHealthChecks(format OutputFormat) error {
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

// convertManagedItems converts resources.ManagedItem to command-specific ManagedItem
func convertManagedItems(items []resources.ManagedItem) []ManagedItem {
	result := make([]ManagedItem, len(items))
	for i, item := range items {
		result[i] = ManagedItem{
			Name:    item.Name,
			Domain:  item.Domain,
			Manager: item.Manager,
		}
	}
	return result
}

// Removed - using config.ConfigAdapter instead

// StatusOutput represents the output structure for status command
type StatusOutput struct {
	ConfigPath   string            `json:"config_path" yaml:"config_path"`
	LockPath     string            `json:"lock_path" yaml:"lock_path"`
	ConfigExists bool              `json:"config_exists" yaml:"config_exists"`
	ConfigValid  bool              `json:"config_valid" yaml:"config_valid"`
	LockExists   bool              `json:"lock_exists" yaml:"lock_exists"`
	StateSummary resources.Summary `json:"state_summary" yaml:"state_summary"`
}

// StatusOutputSummary represents a summary-focused version for JSON/YAML output
type StatusOutputSummary struct {
	ConfigPath   string            `json:"config_path" yaml:"config_path"`
	LockPath     string            `json:"lock_path" yaml:"lock_path"`
	ConfigExists bool              `json:"config_exists" yaml:"config_exists"`
	ConfigValid  bool              `json:"config_valid" yaml:"config_valid"`
	LockExists   bool              `json:"lock_exists" yaml:"lock_exists"`
	Summary      StatusSummaryData `json:"summary" yaml:"summary"`
	ManagedItems []ManagedItem     `json:"managed_items" yaml:"managed_items"`
}

// StatusSummaryData represents aggregate counts and domain summaries
type StatusSummaryData struct {
	TotalManaged   int                       `json:"total_managed" yaml:"total_managed"`
	TotalMissing   int                       `json:"total_missing" yaml:"total_missing"`
	TotalUntracked int                       `json:"total_untracked" yaml:"total_untracked"`
	Domains        []resources.DomainSummary `json:"domains" yaml:"domains"`
}

// ManagedItem represents a currently managed item
type ManagedItem struct {
	Name    string `json:"name" yaml:"name"`
	Domain  string `json:"domain" yaml:"domain"`
	Manager string `json:"manager,omitempty" yaml:"manager,omitempty"`
}

// TableOutput generates human-friendly table output for status
func (s StatusOutput) TableOutput() string {
	// Determine overall health status
	healthStatus := "✅ Healthy"
	if s.StateSummary.TotalMissing > 0 {
		healthStatus = "⚠️ Issues"
	}
	if !s.ConfigValid && s.ConfigExists {
		healthStatus = "❌ Error"
	}

	// Configuration status
	configStatus := "ℹ️ defaults"
	if s.ConfigExists {
		if s.ConfigValid {
			configStatus = "✅ valid"
		} else {
			configStatus = "❌ invalid"
		}
	}

	// Lock status
	lockStatus := "ℹ️ defaults"
	if s.LockExists {
		lockStatus = "✅ exists"
	}

	// Build compact output
	summary := s.StateSummary
	output := fmt.Sprintf("Plonk Status: %s\n", healthStatus)
	output += fmt.Sprintf("Config: %s | Lock: %s | Managing: %d items |\n",
		configStatus, lockStatus, summary.TotalManaged)
	output += fmt.Sprintf("Available: %d untracked\n", summary.TotalUntracked)

	return output
}

// StructuredData returns the structured data for serialization
func (s StatusOutput) StructuredData() any {
	// Create a summary-focused version for structured output
	return StatusOutputSummary{
		ConfigPath:   s.ConfigPath,
		LockPath:     s.LockPath,
		ConfigExists: s.ConfigExists,
		ConfigValid:  s.ConfigValid,
		LockExists:   s.LockExists,
		Summary: StatusSummaryData{
			TotalManaged:   s.StateSummary.TotalManaged,
			TotalMissing:   s.StateSummary.TotalMissing,
			TotalUntracked: s.StateSummary.TotalUntracked,
			Domains:        resources.CreateDomainSummary(s.StateSummary.Results),
		},
		ManagedItems: convertManagedItems(resources.ExtractManagedItems(s.StateSummary.Results)),
	}
}

// DoctorOutput represents the output of the doctor command (health checks)
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
	output.WriteString("# Plonk Health Report\n\n")

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
