// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/packages"
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

	// Reconcile all domains with injected config
	cfg := config.LoadWithDefaults(configDir)
	ctx := context.Background()

	// Reconcile all domains using orchestrator
	result, err := orchestrator.ReconcileAllWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return err
	}

	// Convert domain results directly to output summary
	summary := convertReconcileResultToSummary(result)

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
	}
	formatter := output.NewStatusFormatter(formatterData)
	output.RenderOutput(formatter)
	return nil
}

// convertReconcileResultToSummary converts orchestrator.ReconcileAllResult to output.Summary
func convertReconcileResultToSummary(result orchestrator.ReconcileAllResult) output.Summary {
	// Convert dotfiles result to output.Result
	dotfileResult := output.Result{
		Domain:    "dotfile",
		Managed:   convertDotfileItemsToOutput(result.Dotfiles.Managed),
		Missing:   convertDotfileItemsToOutput(result.Dotfiles.Missing),
		Untracked: convertDotfileItemsToOutput(result.Dotfiles.Untracked),
	}

	// Convert packages result to output.Result
	packageResult := output.Result{
		Domain:    "package",
		Managed:   convertPackageSpecsToOutputWithState(result.Packages.Managed, "managed"),
		Missing:   convertPackageSpecsToOutputWithState(result.Packages.Missing, "missing"),
		Untracked: convertPackageSpecsToOutputWithState(result.Packages.Untracked, "untracked"),
	}

	// Calculate totals
	totalManaged := len(result.Dotfiles.Managed) + len(result.Packages.Managed)
	totalMissing := len(result.Dotfiles.Missing) + len(result.Packages.Missing)
	totalUntracked := len(result.Dotfiles.Untracked) + len(result.Packages.Untracked)

	return output.Summary{
		TotalManaged:   totalManaged,
		TotalMissing:   totalMissing,
		TotalUntracked: totalUntracked,
		Results:        []output.Result{dotfileResult, packageResult},
	}
}

// convertPackageSpecsToOutputWithState converts packages.PackageSpec slice to output.Item slice with state
func convertPackageSpecsToOutputWithState(specs []packages.PackageSpec, state string) []output.Item {
	converted := make([]output.Item, len(specs))
	for i, spec := range specs {
		converted[i] = output.Item{
			Name:    spec.Name,
			Manager: spec.Manager,
			State:   output.ItemState(state),
		}
	}
	return converted
}
