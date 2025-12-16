// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/richhaase/plonk/internal/resources/packages"
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

Examples:
  plonk status    # Show all managed items
  plonk st        # Short alias`,
	RunE:         runStatus,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load configuration (may fail if config is invalid, but we handle this gracefully)
	_, configLoadErr := config.Load(configDir)

	// Reconcile all domains with injected config
	cfg := config.LoadWithDefaults(configDir)
	ctx := context.Background()

	// Reconcile packages and dotfiles
	packageResult, err := packages.ReconcileWithConfig(ctx, configDir, cfg)
	if err != nil {
		return err
	}

	dotfileResult, err := dotfiles.ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return err
	}

	// Compose results into unified summary
	results := map[string]resources.Result{
		"package": packageResult,
		"dotfile": dotfileResult,
	}

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
		ConfigDir:    configDir,
	}

	// Convert to output package type and create formatter
	formatterData := output.StatusOutput{
		ConfigPath:   outputData.ConfigPath,
		LockPath:     outputData.LockPath,
		ConfigExists: outputData.ConfigExists,
		ConfigValid:  outputData.ConfigValid,
		LockExists:   outputData.LockExists,
		StateSummary: convertSummary(outputData.StateSummary),
		ConfigDir:    outputData.ConfigDir,
	}
	formatter := output.NewStatusFormatter(formatterData)
	output.RenderOutput(formatter)
	return nil
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

// Removed - using config.ConfigAdapter instead

// StatusOutput represents the output structure for status command
type StatusOutput struct {
	ConfigPath   string            `json:"config_path" yaml:"config_path"`
	LockPath     string            `json:"lock_path" yaml:"lock_path"`
	ConfigExists bool              `json:"config_exists" yaml:"config_exists"`
	ConfigValid  bool              `json:"config_valid" yaml:"config_valid"`
	LockExists   bool              `json:"lock_exists" yaml:"lock_exists"`
	StateSummary resources.Summary `json:"state_summary" yaml:"state_summary"`
	ConfigDir    string            `json:"-" yaml:"-"` // Not included in JSON/YAML output
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
