// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
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
	configDir := config.GetDefaultConfigDirectory()

	// Load configuration (may fail if config is invalid, but we handle this gracefully)
	_, configLoadErr := config.Load(configDir)

	// Reconcile dotfiles with injected config
	cfg := config.LoadWithDefaults(configDir)
	ctx := context.Background()

	// Reconcile dotfiles only (packages no longer have reconcile concept)
	dotfileResult, err := dotfiles.ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return err
	}

	// Convert dotfiles result to output summary
	summary := convertDotfileResultToSummary(dotfileResult)

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

	// Create formatter data directly
	formatterData := output.StatusOutput{
		ConfigPath:   configPath,
		LockPath:     lockPath,
		ConfigExists: configExists,
		ConfigValid:  configValid,
		LockExists:   lockExists,
		StateSummary: summary,
		ConfigDir:    configDir,
		HomeDir:      homeDir,
	}
	formatter := output.NewStatusFormatter(formatterData)
	output.RenderOutput(formatter)
	return nil
}

// convertDotfileResultToSummary converts dotfiles.Result to output.Summary
func convertDotfileResultToSummary(result dotfiles.Result) output.Summary {
	// Convert dotfiles result to output.Result
	dotfileResult := output.Result{
		Domain:    "dotfile",
		Managed:   convertDotfileItemsToOutput(result.Managed),
		Missing:   convertDotfileItemsToOutput(result.Missing),
		Untracked: convertDotfileItemsToOutput(result.Untracked),
	}

	return output.Summary{
		TotalManaged:   len(result.Managed),
		TotalMissing:   len(result.Missing),
		TotalUntracked: len(result.Untracked),
		Results:        []output.Result{dotfileResult},
	}
}
