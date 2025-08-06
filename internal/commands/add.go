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
	Long: `Add dotfiles to plonk management by copying them to the configuration directory.

This command copies dotfiles from their current locations to your plonk configuration
directory ($PLONK_DIR) for management. The original files remain unchanged in their
current locations.

For directories, plonk will recursively process all files individually, respecting
ignore patterns configured in your plonk.yaml. After adding files, use 'plonk apply'
to deploy them from the configuration directory to your home directory.

Path Resolution:
Plonk accepts paths in multiple formats and intelligently resolves them:

- Absolute paths: /home/user/.vimrc → Used as-is
- Tilde paths: ~/.vimrc → Expands to /home/user/.vimrc
- Relative paths: .vimrc → Tries:
  1. Current directory: /current/dir/.vimrc
  2. Home directory: /home/user/.vimrc
- Plain names: vimrc → Tries:
  1. Current directory: /current/dir/vimrc
  2. Home with dot: /home/user/.vimrc

Special Cases:
- Directories: Recursively processes all files (add only)
- Symlinks: Follows links and copies target file
- Hidden files: Automatically handled (dot removed in plonk dir)

File Mapping:
- ~/.zshrc → $PLONK_DIR/zshrc (leading dot removed)
- ~/.config/nvim/init.lua → $PLONK_DIR/config/nvim/init.lua

Examples:
  plonk add ~/.zshrc                    # Add single file
  plonk add ~/.zshrc ~/.vimrc           # Add multiple files
  plonk add vimrc                       # Finds ~/.vimrc automatically
  plonk add .config/nvim                # Adds entire nvim config directory
  plonk add ../myfile                   # Relative to current directory
  plonk add --dry-run ~/.zshrc ~/.vimrc # Preview what would be added`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runAdd,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")

	// Add file path completion
	addCmd.ValidArgsFunction = CompleteDotfilePaths
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load config for ignore patterns with defaults
	cfg := config.LoadWithDefaults(configDir)

	// Create dotfile manager
	manager := dotfiles.NewManager(homeDir, configDir)

	// Configure options
	opts := dotfiles.AddOptions{
		DryRun: dryRun,
	}

	// Process dotfiles using domain package
	ctx := context.Background()
	results, err := manager.AddFiles(ctx, cfg, args, opts)
	if err != nil {
		return err
	}

	// Create output data based on number of results
	var outputData output.OutputData
	if len(results) == 1 {
		// Single file output
		result := results[0]
		dotfileOutput := &output.DotfileAddOutput{
			Source:      getMetadataString(result, "source"),
			Destination: getMetadataString(result, "destination"),
			Action:      output.MapStatusToAction(result.Status),
			Path:        result.Name,
		}
		if result.Error != nil {
			dotfileOutput.Error = result.Error.Error()
		}
		outputData = dotfileOutput
	} else {
		// Batch output
		outputData = &output.DotfileBatchAddOutput{
			TotalFiles: len(results),
			AddedFiles: output.ConvertToDotfileAddOutput(results),
			Errors:     output.ExtractErrorMessages(results),
		}
	}

	// Render output
	if err := output.RenderOutput(outputData, format); err != nil {
		return err
	}

	// Check if all operations failed and return appropriate error
	return resources.ValidateOperationResults(results, "add dotfiles")
}
