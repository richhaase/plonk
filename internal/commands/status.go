// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/spf13/cobra"
)

// Status command implementation using unified state management system

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Display overall plonk status",
	Long: `Display a detailed list of all plonk-managed items and their status.

Shows:
- All managed packages and dotfiles
- Missing items that need to be installed
- Configuration and lock file status
- Unmanaged items (with --unmanaged flag)

Flag Behavior:
- --packages and --dotfiles: Filter by resource type (both shown if neither specified)
- --unmanaged and --missing: Filter by state (mutually exclusive)
- Combinations work as expected: --packages --missing shows only missing packages

Examples:
  plonk status              # Show all managed items
  plonk status --packages   # Show only packages
  plonk status --dotfiles   # Show only dotfiles
  plonk status --unmanaged  # Show only unmanaged items
  plonk status --missing    # Show only missing resources
  plonk status --missing --packages  # Show only missing packages
  plonk status -o json      # Show as JSON
  plonk status -o yaml      # Show as YAML`,
	RunE:         runStatus,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("packages", false, "Show only package status")
	statusCmd.Flags().Bool("dotfiles", false, "Show only dotfile status")
	statusCmd.Flags().Bool("unmanaged", false, "Show only unmanaged items")
	statusCmd.Flags().Bool("missing", false, "Show only missing resources")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get filter flags
	showPackages, _ := cmd.Flags().GetBool("packages")
	showDotfiles, _ := cmd.Flags().GetBool("dotfiles")
	showUnmanaged, _ := cmd.Flags().GetBool("unmanaged")
	showMissing, _ := cmd.Flags().GetBool("missing")

	// Validate mutually exclusive flags
	if err := validateStatusFlags(showUnmanaged, showMissing); err != nil {
		return err
	}

	// Normalize display flags
	showPackages, showDotfiles = normalizeDisplayFlags(showPackages, showDotfiles)

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load configuration (may fail if config is invalid, but we handle this gracefully)
	_, configLoadErr := config.Load(configDir)

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
		ConfigPath:    configPath,
		LockPath:      lockPath,
		ConfigExists:  configExists,
		ConfigValid:   configValid,
		LockExists:    lockExists,
		StateSummary:  summary,
		ShowPackages:  showPackages,
		ShowDotfiles:  showDotfiles,
		ShowUnmanaged: showUnmanaged,
		ShowMissing:   showMissing,
		ConfigDir:     configDir,
	}

	// Convert to output package type and create formatter
	formatterData := output.StatusOutput{
		ConfigPath:    outputData.ConfigPath,
		LockPath:      outputData.LockPath,
		ConfigExists:  outputData.ConfigExists,
		ConfigValid:   outputData.ConfigValid,
		LockExists:    outputData.LockExists,
		StateSummary:  convertSummary(outputData.StateSummary),
		ShowPackages:  outputData.ShowPackages,
		ShowDotfiles:  outputData.ShowDotfiles,
		ShowUnmanaged: outputData.ShowUnmanaged,
		ShowMissing:   outputData.ShowMissing,
		ConfigDir:     outputData.ConfigDir,
	}
	formatter := output.NewStatusFormatter(formatterData)
	return RenderOutput(formatter, format)
}

// convertSummary converts from resources.Summary to output.Summary
func convertSummary(summary resources.Summary) output.Summary {
	converted := output.Summary{
		TotalManaged:   summary.TotalManaged,
		TotalMissing:   summary.TotalMissing,
		TotalUntracked: summary.TotalUntracked,
		Results:        make([]output.Result, len(summary.Results)),
	}
	for i, result := range summary.Results {
		converted.Results[i] = output.Result{
			Domain:    result.Domain,
			Managed:   convertItems(result.Managed),
			Missing:   convertItems(result.Missing),
			Untracked: convertItems(result.Untracked),
		}
	}
	return converted
}

// convertItems converts from resources.Item to output.Item
func convertItems(items []resources.Item) []output.Item {
	converted := make([]output.Item, len(items))
	for i, item := range items {
		converted[i] = output.Item{
			Name:     item.Name,
			Manager:  item.Manager,
			Path:     item.Path,
			State:    output.ItemState(item.State.String()),
			Metadata: item.Metadata,
		}
	}
	return converted
}

// sortItems sorts a slice of resources.Item alphabetically by name (case-insensitive)
func sortItems(items []resources.Item) {
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
}

// sortItemsByManager sorts items first by manager, then by name (case-insensitive)
func sortItemsByManager(items map[string][]resources.Item) []string {
	// Get sorted manager names
	managers := make([]string, 0, len(items))
	for manager := range items {
		managers = append(managers, manager)
	}
	sort.Strings(managers)

	// Sort items within each manager
	for _, manager := range managers {
		sortItems(items[manager])
	}

	return managers
}

// Removed - using config.ConfigAdapter instead

// StatusOutput represents the output structure for status command
type StatusOutput struct {
	ConfigPath    string            `json:"config_path" yaml:"config_path"`
	LockPath      string            `json:"lock_path" yaml:"lock_path"`
	ConfigExists  bool              `json:"config_exists" yaml:"config_exists"`
	ConfigValid   bool              `json:"config_valid" yaml:"config_valid"`
	LockExists    bool              `json:"lock_exists" yaml:"lock_exists"`
	StateSummary  resources.Summary `json:"state_summary" yaml:"state_summary"`
	ShowPackages  bool              `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowDotfiles  bool              `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowUnmanaged bool              `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowMissing   bool              `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ConfigDir     string            `json:"-" yaml:"-"` // Not included in JSON/YAML output
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
