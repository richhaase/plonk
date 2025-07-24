// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/runtime"
	"github.com/richhaase/plonk/internal/state"
	"github.com/richhaase/plonk/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <files...>",
	Short: "Add dotfiles to plonk management",
	Long: `Add dotfiles to plonk configuration and import them.

This command adds dotfiles to your plonk configuration directory and manages them.
It will copy the dotfiles from their current locations to your plonk dotfiles
directory and preserve the original files in case you need to revert.

For directories, plonk will recursively add all files individually, respecting
ignore patterns configured in your plonk.yaml.

Path Resolution:
- Absolute paths: /home/user/.vimrc
- Tilde paths: ~/.vimrc
- Relative paths: First tries current directory, then home directory

Examples:
  plonk add ~/.zshrc                    # Add single file
  plonk add ~/.zshrc ~/.vimrc           # Add multiple files
  plonk add .zshrc .vimrc               # Finds files in home directory
  plonk add ~/.config/nvim/ ~/.tmux.conf # Add directory and file
  plonk add --dry-run ~/.zshrc ~/.vimrc # Preview what would be added`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")
	addCmd.Flags().BoolP("force", "f", false, "Force addition even if already managed")

	// Add file path completion
	addCmd.ValidArgsFunction = CompleteDotfilePaths
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "add", "output-format", "invalid output format")
	}

	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	// TODO: force flag is defined but not currently used in core.AddSingleDotfile
	// force, _ := cmd.Flags().GetBool("force")

	// Get directories from shared context
	sharedCtx := runtime.GetSharedContext()
	homeDir := sharedCtx.HomeDir()
	configDir := sharedCtx.ConfigDir()

	// Load config for ignore patterns
	manager := config.NewConfigManager(configDir)
	cfg, err := manager.LoadOrCreate()
	if err != nil {
		return err
	}

	// Process each dotfile directly
	ctx := context.Background()
	var results []state.OperationResult

	for _, dotfilePath := range args {
		// Call core business logic directly
		dotfileResults := AddSingleDotfile(ctx, cfg, homeDir, configDir, dotfilePath, dryRun)
		results = append(results, dotfileResults...)
	}

	// Create output data based on number of results
	var outputData OutputData
	if len(results) == 1 {
		// Single file output
		result := results[0]
		output := &DotfileAddOutput{
			Source:      GetMetadataString(result, "source"),
			Destination: GetMetadataString(result, "destination"),
			Action:      ui.MapStatusToAction(result.Status),
			Path:        result.Name,
		}
		if result.Error != nil {
			output.Error = result.Error.Error()
		}
		outputData = output
	} else {
		// Batch output
		outputData = &DotfileBatchAddOutput{
			TotalFiles: len(results),
			AddedFiles: ui.ConvertToDotfileAddOutput(results),
			Errors:     ui.ExtractErrorMessages(results),
		}
	}

	// Render output
	if err := RenderOutput(outputData, format); err != nil {
		return err
	}

	// Check if all operations failed and return appropriate error
	allFailed := true
	for _, result := range results {
		if result.Status != "failed" {
			allFailed = false
			break
		}
	}

	if allFailed && len(results) > 0 {
		return errors.NewError(
			errors.ErrFileNotFound,
			errors.DomainDotfiles,
			"add-multiple",
			fmt.Sprintf("failed to process %d item(s)", len(results)),
		)
	}

	return nil
}
