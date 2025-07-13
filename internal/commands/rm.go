// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <files...>",
	Short: "Remove dotfiles from plonk management",
	Long: `Remove dotfiles from plonk management and unlink them.

This command removes dotfiles from your plonk configuration and unlinks them
from your home directory. The source files in your plonk configuration
directory will remain available for re-linking later.

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
	rmCmd.ValidArgsFunction = completeDotfilePaths
}

func runRm(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "rm", "failed to get home directory")
	}

	configDir := config.GetDefaultConfigDirectory()

	// Load config
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// If config doesn't exist, we can't remove dotfiles
		return errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "rm", "output-format", "invalid output format")
	}

	// Process dotfiles sequentially
	results := make([]operations.OperationResult, 0, len(args))
	reporter := operations.NewProgressReporterForOperation("remove", "dotfile", format == OutputTable)

	for _, dotfilePath := range args {
		result := removeSingleDotfile(homeDir, configDir, cfg, dotfilePath, dryRun)
		results = append(results, result)

		// Show progress immediately
		reporter.ShowItemProgress(result)
	}

	// Handle output based on format
	if format == OutputTable {
		// Show summary for table output
		reporter.ShowBatchSummary(results)
	} else {
		// For structured output, create appropriate response
		return renderDotfileRemovalResults(results, format)
	}

	// Determine exit code
	return operations.DetermineExitCode(results, errors.DomainDotfiles, "rm")
}

// renderDotfileRemovalResults renders results in structured format
func renderDotfileRemovalResults(results []operations.OperationResult, format OutputFormat) error {
	output := DotfileRemovalOutput{
		TotalFiles: len(results),
		Results:    results,
		Summary:    calculateDotfileRemovalSummary(results),
	}
	return RenderOutput(output, format)
}

// DotfileRemovalOutput represents the output for dotfile removal
type DotfileRemovalOutput struct {
	TotalFiles int                          `json:"total_files" yaml:"total_files"`
	Results    []operations.OperationResult `json:"results" yaml:"results"`
	Summary    DotfileRemovalSummary        `json:"summary" yaml:"summary"`
}

// DotfileRemovalSummary provides summary for dotfile removal
type DotfileRemovalSummary struct {
	Removed int `json:"removed" yaml:"removed"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculateDotfileRemovalSummary calculates summary from results
func calculateDotfileRemovalSummary(results []operations.OperationResult) DotfileRemovalSummary {
	summary := DotfileRemovalSummary{}
	for _, result := range results {
		switch result.Status {
		case "unlinked", "would-unlink":
			summary.Removed++
		case "skipped":
			summary.Skipped++
		case "failed":
			summary.Failed++
		}
	}
	return summary
}

// TableOutput generates human-friendly output
func (d DotfileRemovalOutput) TableOutput() string {
	output := "Dotfile Removal\n===============\n\n"

	if d.Summary.Removed > 0 {
		output += fmt.Sprintf("📄 Removed %d dotfiles\n", d.Summary.Removed)
	}
	if d.Summary.Skipped > 0 {
		output += fmt.Sprintf("⏭️ %d skipped\n", d.Summary.Skipped)
	}
	if d.Summary.Failed > 0 {
		output += fmt.Sprintf("❌ %d failed\n", d.Summary.Failed)
	}

	output += fmt.Sprintf("\nTotal: %d dotfiles processed\n", d.TotalFiles)
	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileRemovalOutput) StructuredData() any {
	return d
}
