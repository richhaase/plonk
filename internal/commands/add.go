// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/gitops"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	GroupID: "manage",
	Use:     "add [files|manager:package...]",
	Short:   "Add dotfiles or install and track packages",
	Long: `Copy dotfiles from $HOME into $PLONK_DIR, or add manager:package specifications.

Installed packages are tracked; missing packages are installed before tracking.
Files and packages can be mixed in one invocation. Use an explicit path (./ or /)
for filenames containing a colon. Package managers must already be available.

File paths must stay under $HOME. Directories are added recursively with configured
ignore patterns. Templates should be edited directly rather than overwritten.

Examples:
  plonk add ~/.zshrc brew:ripgrep
  plonk add cargo:bat go:golang.org/x/tools/gopls
  plonk add --dry-run ~/.vimrc pnpm:typescript
  plonk add -y   # Sync drifted files back to $PLONK_DIR`,
	RunE:         runAdd,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")
	addCmd.Flags().BoolP("sync-drifted", "y", false, "Sync all drifted files from $HOME back to $PLONKDIR")

	// Add file path completion
	addCmd.ValidArgsFunction = CompleteResourceArgs
}

func runAdd(cmd *cobra.Command, args []string) error {
	syncDrifted, _ := cmd.Flags().GetBool("sync-drifted")
	if syncDrifted {
		if len(args) != 0 {
			return fmt.Errorf("--sync-drifted cannot be combined with file or package arguments")
		}
		return runAddFiles(cmd, args)
	}
	if len(args) == 0 {
		return fmt.Errorf("specify a file or manager:package")
	}
	files, specs := splitResourceArgs(args)
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	var errs []error
	if len(specs) > 0 {
		errs = append(errs, mutatePackages(cmd.Context(), config.GetDefaultConfigDirectory(), specs, true, false, dryRun))
	}
	if len(files) > 0 {
		errs = append(errs, runAddFiles(cmd, files))
	}
	return errors.Join(errs...)
}

func runAddFiles(cmd *cobra.Command, args []string) error {
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

	// Auto-commit if any files were actually added/updated (even on partial
	// failure: successful mutations are committed, matching rm behavior;
	// the exit status below still reflects any failure)
	if !opts.DryRun && anyAddSucceeded(results) {
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

	// Auto-commit synced drifted files (including on partial failure —
	// successful mutations are committed regardless of the exit status)
	if !dryRun && anyAddSucceeded(results) {
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

// anyAddSucceeded reports whether at least one add result represents an
// actual mutation (added or updated), used to decide whether to auto-commit.
// This is independent of the exit-status policy: successful partial batches
// are still committed, even if other items failed.
func anyAddSucceeded(results []AddResult) bool {
	for _, result := range results {
		if result.Status == AddStatusAdded || result.Status == AddStatusUpdated {
			return true
		}
	}
	return false
}

// validateAddResultsErr returns an error if any add operation failed
// (partial failure is a non-zero exit per the documented batch policy)
func validateAddResultsErr(results []AddResult) error {
	return ValidateBatchResults(len(results), "add dotfiles", func(i int) bool {
		return results[i].Status == AddStatusFailed
	})
}
