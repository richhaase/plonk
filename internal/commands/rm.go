// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"time"

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
	// Create command pipeline for dotfile removal
	pipeline, err := NewCommandPipeline(cmd, "dotfile-remove")
	if err != nil {
		return err
	}

	// Define the processor function
	processor := func(ctx context.Context, args []string, flags *SimpleFlags) ([]operations.OperationResult, error) {
		// Get directories
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "rm", "failed to get home directory")
		}

		configDir := config.GetDefaultConfigDirectory()

		// Load config using LoadConfigWithDefaults for consistent zero-config behavior
		cfg := config.LoadConfigWithDefaults(configDir)

		// Create item processor for dotfile removal
		processor := operations.SimpleProcessor(
			func(ctx context.Context, dotfilePath string) operations.OperationResult {
				return removeSingleDotfile(homeDir, configDir, cfg, dotfilePath, flags.DryRun)
			},
		)

		// Configure batch processing options
		options := operations.BatchProcessorOptions{
			ItemType:               "dotfile",
			Operation:              "remove",
			ShowIndividualProgress: flags.Verbose || flags.DryRun, // Show progress in verbose or dry-run mode
			Timeout:                2 * time.Minute,               // Dotfile removal timeout
			ContinueOnError:        nil,                           // Use default (true) - continue on individual failures
		}

		// Use standard batch workflow
		return operations.StandardBatchWorkflow(context.Background(), args, processor, options)
	}

	// Execute the pipeline
	return pipeline.ExecuteWithResults(context.Background(), processor, args)
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

// TableOutput generates human-friendly output
func (d DotfileRemovalOutput) TableOutput() string {
	tb := NewTableBuilder()

	tb.AddTitle("Dotfile Removal")
	tb.AddNewline()

	if d.Summary.Removed > 0 {
		tb.AddLine("📄 Removed %d dotfiles", d.Summary.Removed)
	}
	if d.Summary.Skipped > 0 {
		tb.AddLine("⏭️ %d skipped", d.Summary.Skipped)
	}
	if d.Summary.Failed > 0 {
		tb.AddLine("%s %d failed", IconUnhealthy, d.Summary.Failed)
	}

	tb.AddNewline()
	tb.AddLine("Total: %d dotfiles processed", d.TotalFiles)

	return tb.Build()
}

// StructuredData returns the structured data for serialization
func (d DotfileRemovalOutput) StructuredData() any {
	return d
}
