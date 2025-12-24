// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [files...]",
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
  plonk add --dry-run ~/.zshrc ~/.vimrc # Preview what would be added
  plonk add -y                          # Sync all drifted files back to $PLONKDIR`,
	RunE:         runAdd,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")
	addCmd.Flags().BoolP("sync-drifted", "y", false, "Sync all drifted files from $HOME back to $PLONKDIR")

	// Add file path completion
	addCmd.ValidArgsFunction = CompleteDotfilePaths
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	syncDrifted, _ := cmd.Flags().GetBool("sync-drifted")

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetDefaultConfigDirectory()

	// Load config for ignore patterns with defaults
	cfg := config.LoadWithDefaults(configDir)

	ctx := context.Background()

	// Handle sync-drifted flag
	if syncDrifted {
		return runSyncDrifted(ctx, cmd, cfg, configDir, homeDir, dryRun)
	}

	// Require at least one file argument if not syncing drifted
	if len(args) == 0 {
		return cmd.Usage()
	}

	// Create dotfile manager with injected config
	manager := dotfiles.NewManagerWithConfig(homeDir, configDir, cfg)

	// Configure options
	opts := dotfiles.AddOptions{
		DryRun: dryRun,
	}

	// Process dotfiles using domain package
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
			Source:      result.Source,
			Destination: result.Destination,
			Action:      output.MapStatusToAction(result.Status.String()),
			Path:        result.Path,
		}
		if result.Error != nil {
			dotfileOutput.Error = result.Error.Error()
		}
		outputData = dotfileOutput
	} else {
		// Batch output
		outputData = &output.DotfileBatchAddOutput{
			TotalFiles: len(results),
			AddedFiles: convertAddResultsToOutput(results),
			Errors:     extractAddErrorMessages(results),
		}
	}

	// Render output
	output.RenderOutput(outputData)

	// Check if all operations failed and return appropriate error
	return validateAddResults(results)
}

// runSyncDrifted syncs all drifted files from $HOME back to $PLONKDIR
func runSyncDrifted(ctx context.Context, cmd *cobra.Command, cfg *config.Config, configDir, homeDir string, dryRun bool) error {
	// Get drifted dotfiles from reconciliation
	driftedFiles, err := getDriftedDotfiles(ctx, cfg, configDir, homeDir)
	if err != nil {
		return fmt.Errorf("failed to get drifted files: %w", err)
	}

	if len(driftedFiles) == 0 {
		output.Println("No drifted dotfiles found")
		return nil
	}

	// Build list of paths to sync (use deployed paths from $HOME)
	var paths []string
	for _, item := range driftedFiles {
		if item.Destination != "" {
			paths = append(paths, item.Destination)
		}
	}

	if len(paths) == 0 {
		output.Println("No drifted files to sync")
		return nil
	}

	// Create dotfile manager
	manager := dotfiles.NewManagerWithConfig(homeDir, configDir, cfg)

	// Configure options
	opts := dotfiles.AddOptions{
		DryRun: dryRun,
	}

	// Process the drifted files
	results, err := manager.AddFiles(ctx, cfg, paths, opts)
	if err != nil {
		return err
	}

	// Create output data
	var outputData output.OutputData
	if len(results) == 1 {
		// Single file output
		result := results[0]
		dotfileOutput := &output.DotfileAddOutput{
			Source:      result.Source,
			Destination: result.Destination,
			Action:      output.MapStatusToAction(result.Status.String()),
			Path:        result.Path,
		}
		if result.Error != nil {
			dotfileOutput.Error = result.Error.Error()
		}
		outputData = dotfileOutput
	} else {
		// Batch output
		outputData = &output.DotfileBatchAddOutput{
			TotalFiles: len(results),
			AddedFiles: convertAddResultsToOutput(results),
			Errors:     extractAddErrorMessages(results),
		}
	}

	// Render output
	output.RenderOutput(outputData)

	// Check if all operations failed and return appropriate error
	return validateAddResults(results)
}

// extractAddErrorMessages extracts error messages from failed add results
func extractAddErrorMessages(results []dotfiles.AddResult) []string {
	var errors []string
	for _, result := range results {
		if result.Status == dotfiles.AddStatusFailed && result.Error != nil {
			errors = append(errors, fmt.Sprintf("failed to add %s: %v", result.Path, result.Error))
		}
	}
	return errors
}

// convertAddResultsToOutput converts dotfiles.AddResult to DotfileAddOutput for structured output
func convertAddResultsToOutput(results []dotfiles.AddResult) []output.DotfileAddOutput {
	outputs := make([]output.DotfileAddOutput, 0, len(results))
	for _, result := range results {
		if result.Status == dotfiles.AddStatusFailed {
			continue // Skip failed results, they're handled in errors
		}

		outputs = append(outputs, output.DotfileAddOutput{
			Source:      result.Source,
			Destination: result.Destination,
			Action:      output.MapStatusToAction(result.Status.String()),
			Path:        result.Path,
		})
	}
	return outputs
}

// validateAddResults checks if all add operations failed and returns appropriate error
func validateAddResults(results []dotfiles.AddResult) error {
	return ValidateBatchResults(len(results), "add dotfiles", func(i int) bool {
		return results[i].Status == dotfiles.AddStatusFailed
	})
}
