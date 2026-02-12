// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/gitops"
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
- Plain names: vimrc → Resolves to ~/.vimrc (dot prefix added automatically)

Security:
- All paths must resolve to dotfiles (first path component starts with '.')
- All paths must resolve to locations under your home directory ($HOME)
- Paths outside $HOME are rejected to prevent unintended file operations

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
	homeDir, err := config.GetHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configDir := config.GetDefaultConfigDirectory()

	// Load config for ignore patterns with defaults
	cfg := config.LoadWithDefaults(configDir)

	// Handle sync-drifted flag
	if syncDrifted {
		return runSyncDrifted(cmd.Context(), cfg, configDir, homeDir, dryRun)
	}

	// Require at least one file argument if not syncing drifted
	if len(args) == 0 {
		return cmd.Usage()
	}

	// Create DotfileManager directly
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)

	// Configure options
	opts := AddOptions{
		DryRun: dryRun,
	}

	// Process dotfiles using helper function
	results := addDotfiles(dm, configDir, homeDir, args, opts)

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
			AddedFiles: convertAddResultsToAddOutput(results),
			Errors:     extractAddErrors(results),
		}
	}

	// Render output
	output.RenderOutput(outputData)

	// Auto-commit if any files were actually added/updated
	if !opts.DryRun && validateAddResultsErr(results) == nil {
		gitops.AutoCommit(cmd.Context(), configDir, "add", args)
	}

	// Check if all operations failed and return appropriate error
	return validateAddResultsErr(results)
}

// runSyncDrifted syncs all drifted files from $HOME back to $PLONKDIR
func runSyncDrifted(ctx context.Context, cfg *config.Config, configDir, homeDir string, dryRun bool) error {
	// Get drifted dotfiles from reconciliation
	driftedFiles, err := getDriftedDotfileStatuses(cfg, configDir, homeDir)
	if err != nil {
		return fmt.Errorf("failed to get drifted files: %w", err)
	}

	if len(driftedFiles) == 0 {
		output.Println("No drifted dotfiles found")
		return nil
	}

	// Build list of paths to sync (use deployed paths from $HOME)
	var paths []string
	for _, s := range driftedFiles {
		if s.Target != "" {
			paths = append(paths, s.Target)
		}
	}

	if len(paths) == 0 {
		output.Println("No drifted files to sync")
		return nil
	}

	// Create DotfileManager directly
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)

	// Configure options
	opts := AddOptions{
		DryRun: dryRun,
	}

	// Process the drifted files
	results := addDotfiles(dm, configDir, homeDir, paths, opts)

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
			AddedFiles: convertAddResultsToAddOutput(results),
			Errors:     extractAddErrors(results),
		}
	}

	// Render output
	output.RenderOutput(outputData)

	// Auto-commit synced drifted files
	if !dryRun && validateAddResultsErr(results) == nil {
		gitops.AutoCommit(ctx, configDir, "add --sync-drifted", paths)
	}

	// Check if all operations failed and return appropriate error
	return validateAddResultsErr(results)
}

// extractAddErrors extracts error messages from failed add results
func extractAddErrors(results []AddResult) []string {
	var errors []string
	for _, result := range results {
		if result.Status == AddStatusFailed && result.Error != nil {
			errors = append(errors, fmt.Sprintf("failed to add %s: %v", result.Path, result.Error))
		}
	}
	return errors
}

// convertAddResultsToAddOutput converts AddResult to DotfileAddOutput for structured output
func convertAddResultsToAddOutput(results []AddResult) []output.DotfileAddOutput {
	outputs := make([]output.DotfileAddOutput, 0, len(results))
	for _, result := range results {
		if result.Status == AddStatusFailed {
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

// validateAddResultsErr checks if all add operations failed and returns appropriate error
func validateAddResultsErr(results []AddResult) error {
	return ValidateBatchResults(len(results), "add dotfiles", func(i int) bool {
		return results[i].Status == AddStatusFailed
	})
}
