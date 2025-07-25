// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
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
		return err
	}

	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	// TODO: force flag is defined but not currently used in core.AddSingleDotfile
	// force, _ := cmd.Flags().GetBool("force")

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load config for ignore patterns
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return err
	}

	// Create dotfile manager
	manager := dotfiles.NewManager(homeDir, configDir)

	// Configure options
	opts := dotfiles.AddOptions{
		DryRun: dryRun,
		Force:  false, // TODO: implement force flag
	}

	// Process dotfiles using domain package
	ctx := context.Background()
	results, err := manager.AddFiles(ctx, cfg, args, opts)
	if err != nil {
		return err
	}

	// Create output data based on number of results
	var outputData OutputData
	if len(results) == 1 {
		// Single file output
		result := results[0]
		output := &DotfileAddOutput{
			Source:      GetMetadataString(result, "source"),
			Destination: GetMetadataString(result, "destination"),
			Action:      output.MapStatusToAction(result.Status),
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
			AddedFiles: output.ConvertToDotfileAddOutput(results),
			Errors:     output.ExtractErrorMessages(results),
		}
	}

	// Render output
	if err := RenderOutput(outputData, format); err != nil {
		return err
	}

	// Check if all operations failed and return appropriate error
	return resources.ValidateOperationResults(results, "add dotfiles")
}
