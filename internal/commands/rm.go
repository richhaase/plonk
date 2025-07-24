// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/state"
	"github.com/richhaase/plonk/internal/ui"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <files...>",
	Short: "Remove dotfiles from plonk management",
	Long: `Remove dotfiles from plonk management completely.

This command unlinks dotfiles from your home directory and removes them from
your plonk configuration directory. The dotfiles will no longer be managed by
plonk and cannot be re-linked without adding them again.

Examples:
  plonk rm ~/.zshrc                    # Remove single file
  plonk rm ~/.zshrc ~/.vimrc           # Remove multiple files
  plonk rm ~/.config/nvim/init.lua     # Remove specific file
  plonk rm --dry-run ~/.zshrc ~/.vimrc # Preview what would be removed`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRm,
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")
	rmCmd.Flags().BoolP("force", "f", false, "Force removal even if not managed")

	// Add file path completion
	rmCmd.ValidArgsFunction = CompleteDotfilePaths
}

func runRm(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "rm", "output-format", "invalid output format")
	}

	// Get flags
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return err
	}

	// Get directories
	homeDir := orchestrator.GetHomeDir()
	configDir := orchestrator.GetConfigDir()

	// Load config using LoadConfigWithDefaults for consistent zero-config behavior
	cfg := config.LoadConfigWithDefaults(configDir)

	// Process each dotfile directly
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var results []state.OperationResult

	// Show header for progress tracking
	reporter := ui.NewProgressReporterForOperation("remove", "dotfile", true)

	for _, dotfilePath := range args {
		// Check if context was canceled
		if ctx.Err() != nil {
			break
		}

		// Remove single dotfile directly
		result := RemoveSingleDotfile(homeDir, configDir, cfg, dotfilePath, flags.DryRun)

		// Show individual progress
		reporter.ShowItemProgress(result)

		// Collect result
		results = append(results, result)
	}

	// Show batch summary
	reporter.ShowBatchSummary(results)

	// Create output data
	summary := calculateRemovalSummary(results)
	outputData := DotfileRemovalOutput{
		TotalFiles: len(results),
		Results:    results,
		Summary:    summary,
	}

	// Render output
	if err := RenderOutput(outputData, format); err != nil {
		return err
	}

	// Determine exit code based on results
	exitErr := DetermineExitCode(results, errors.DomainDotfiles, "remove")
	if exitErr != nil {
		return exitErr
	}

	return nil
}

// DotfileRemovalOutput represents the output for dotfile removal
type DotfileRemovalOutput struct {
	TotalFiles int                     `json:"total_files" yaml:"total_files"`
	Results    []state.OperationResult `json:"results" yaml:"results"`
	Summary    DotfileRemovalSummary   `json:"summary" yaml:"summary"`
}

// DotfileRemovalSummary provides summary for dotfile removal
type DotfileRemovalSummary struct {
	Removed int `json:"removed" yaml:"removed"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculateRemovalSummary calculates summary from results using generic operations summary
func calculateRemovalSummary(results []state.OperationResult) DotfileRemovalSummary {
	genericSummary := state.CalculateSummary(results)
	return DotfileRemovalSummary{
		Removed: genericSummary.Added, // In removal context, "added" means "removed"
		Skipped: genericSummary.Skipped,
		Failed:  genericSummary.Failed,
	}
}

// TableOutput generates human-friendly output
func (d DotfileRemovalOutput) TableOutput() string {
	tb := NewTableBuilder()

	// For single file operations, show inline result
	if d.TotalFiles == 1 && len(d.Results) == 1 {
		result := d.Results[0]
		switch result.Status {
		case "removed":
			tb.AddLine("✅ Removed dotfile from plonk management")
			tb.AddLine("   File: %s", result.Name)
			if source, ok := result.Metadata["source"].(string); ok {
				tb.AddLine("   Source: %s (removed from config)", source)
			}
		case "would-remove":
			tb.AddLine("🔍 Would remove dotfile from plonk management (dry-run)")
			tb.AddLine("   File: %s", result.Name)
			if source, ok := result.Metadata["source"].(string); ok {
				tb.AddLine("   Source: %s", source)
			}
		case "skipped":
			tb.AddLine("⏭️ Skipped: %s", result.Name)
			if result.Error != nil {
				tb.AddLine("   Reason: %s", result.Error.Error())
			}
		case "failed":
			tb.AddLine("%s Failed: %s", IconUnhealthy, result.Name)
			if result.Error != nil {
				tb.AddLine("   Error: %s", result.Error.Error())
			}
		}
		return tb.Build()
	}

	// For batch operations, show summary
	tb.AddTitle("Dotfile Removal")
	tb.AddNewline()

	// Check if this is a dry run
	isDryRun := false
	wouldRemoveCount := 0
	for _, result := range d.Results {
		if result.Status == "would-remove" {
			isDryRun = true
			wouldRemoveCount++
		}
	}

	if isDryRun {
		if wouldRemoveCount > 0 {
			tb.AddLine("🔍 Would remove %d dotfiles (dry-run)", wouldRemoveCount)
		}
	} else {
		if d.Summary.Removed > 0 {
			tb.AddLine("📄 Removed %d dotfiles", d.Summary.Removed)
		}
	}

	if d.Summary.Skipped > 0 {
		tb.AddLine("⏭️ %d skipped", d.Summary.Skipped)
	}
	if d.Summary.Failed > 0 {
		tb.AddLine("%s %d failed", IconUnhealthy, d.Summary.Failed)
	}

	tb.AddNewline()

	// Show individual files
	for _, result := range d.Results {
		switch result.Status {
		case "removed":
			tb.AddLine("   ✓ %s", result.Name)
		case "would-remove":
			tb.AddLine("   - %s", result.Name)
		case "skipped":
			tb.AddLine("   ⏭ %s (not managed)", result.Name)
		case "failed":
			tb.AddLine("   ✗ %s", result.Name)
		}
	}

	tb.AddNewline()
	tb.AddLine("Total: %d dotfiles processed", d.TotalFiles)

	return tb.Build()
}

// StructuredData returns the structured data for serialization
func (d DotfileRemovalOutput) StructuredData() any {
	return d
}
