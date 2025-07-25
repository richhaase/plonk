// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
  plonk status -o json   # Show as JSON
  plonk status -o yaml   # Show as YAML`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get directories
	homeDir := orchestrator.GetHomeDir()
	configDir := orchestrator.GetConfigDir()

	// Load configuration (may fail if config is invalid, but we handle this gracefully)
	_, configLoadErr := config.LoadConfig(configDir)

	// Reconcile all domains
	ctx := context.Background()
	results, err := orchestrator.ReconcileAll(ctx, homeDir, configDir)
	if err != nil {
		return fmt.Errorf("failed to reconcile state: %w", err)
	}

	// Convert results to summary for compatibility with existing output logic
	summary := convertResultsToSummary(results)

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

// convertResultsToSummary converts reconciliation results to resources.Summary for output compatibility
func convertResultsToSummary(results map[string]resources.Result) resources.Summary {
	summary := resources.Summary{
		TotalManaged:   0,
		TotalMissing:   0,
		TotalUntracked: 0,
		Results:        make([]resources.Result, 0, len(results)),
	}

	for _, result := range results {
		summary.TotalManaged += len(result.Managed)
		summary.TotalMissing += len(result.Missing)
		summary.TotalUntracked += len(result.Untracked)
		summary.Results = append(summary.Results, result)
	}

	return summary
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
	TotalManaged   int             `json:"total_managed" yaml:"total_managed"`
	TotalMissing   int             `json:"total_missing" yaml:"total_missing"`
	TotalUntracked int             `json:"total_untracked" yaml:"total_untracked"`
	Domains        []DomainSummary `json:"domains" yaml:"domains"`
}

// DomainSummary represents counts for a specific domain/manager
type DomainSummary struct {
	Domain         string `json:"domain" yaml:"domain"`
	Manager        string `json:"manager,omitempty" yaml:"manager,omitempty"`
	ManagedCount   int    `json:"managed_count" yaml:"managed_count"`
	MissingCount   int    `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int    `json:"untracked_count" yaml:"untracked_count"`
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
			Domains:        createDomainSummary(s.StateSummary.Results),
		},
		ManagedItems: extractManagedItems(s.StateSummary.Results),
	}
}

// createDomainSummary creates domain summaries with counts only
func createDomainSummary(results []resources.Result) []DomainSummary {
	var domains []DomainSummary
	for _, result := range results {
		if result.IsEmpty() {
			continue
		}
		domains = append(domains, DomainSummary{
			Domain:         result.Domain,
			Manager:        result.Manager,
			ManagedCount:   len(result.Managed),
			MissingCount:   len(result.Missing),
			UntrackedCount: len(result.Untracked),
		})
	}
	return domains
}

// extractManagedItems extracts only the managed items without full metadata
func extractManagedItems(results []resources.Result) []ManagedItem {
	var items []ManagedItem
	for _, result := range results {
		for _, managed := range result.Managed {
			items = append(items, ManagedItem{
				Name:    managed.Name,
				Domain:  managed.Domain,
				Manager: managed.Manager,
			})
		}
	}
	return items
}
