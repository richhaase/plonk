// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/richhaase/plonk/internal/runtime"
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
	addCmd.ValidArgsFunction = completeDotfilePaths
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Create command pipeline for dotfiles
	pipeline, err := NewSimpleCommandPipeline(cmd, "dotfile")
	if err != nil {
		return err
	}

	// Define the processor function
	processor := func(ctx context.Context, args []string) (OutputData, error) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")

		// Process dotfiles and return results that can be rendered
		results, err := addDotfilesProcessor(args, dryRun, force)
		if err != nil {
			return nil, err
		}

		// Convert to appropriate output format
		if len(results) == 1 {
			result := results[0]
			return &DotfileAddOutput{
				Source:      getMetadataString(result, "source"),
				Destination: getMetadataString(result, "destination"),
				Action:      mapStatusToAction(result.Status),
				Path:        result.Name,
			}, nil
		} else {
			return &DotfileBatchAddOutput{
				TotalFiles: len(results),
				AddedFiles: convertToDotfileAddOutput(results),
				Errors:     extractErrorMessages(results),
			}, nil
		}
	}

	// Execute the pipeline
	return pipeline.ExecuteWithData(context.Background(), processor, args)
}

// addDotfilesProcessor processes dotfile addition and returns operation results
func addDotfilesProcessor(dotfilePaths []string, dryRun, force bool) ([]operations.OperationResult, error) {
	// Get directories from shared context
	sharedCtx := runtime.GetSharedContext()
	homeDir := sharedCtx.HomeDir()
	configDir := sharedCtx.ConfigDir()

	// Load config for ignore patterns
	cfg, err := loadOrCreateConfig(configDir)
	if err != nil {
		return nil, err
	}

	// Process dotfiles and collect results
	return addSingleDotfiles(dotfilePaths, homeDir, configDir, cfg, dryRun, force)
}

// addSingleDotfiles processes multiple dotfile paths and returns results
func addSingleDotfiles(dotfilePaths []string, homeDir, configDir string, cfg *config.Config, dryRun, force bool) ([]operations.OperationResult, error) {
	var results []operations.OperationResult

	for _, dotfilePath := range dotfilePaths {
		// Process each dotfile (can result in multiple files for directories)
		dotfileResults := addSingleDotfile(context.Background(), cfg, homeDir, configDir, dotfilePath, dryRun)
		results = append(results, dotfileResults...)
	}

	return results, nil
}
