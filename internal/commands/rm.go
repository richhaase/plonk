// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"errors"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/gitops"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	GroupID: "manage",
	Use:     "rm <files|manager:package...>",
	Short:   "Remove files or packages from management",
	Long: `Remove managed dotfile sources or package entries from $PLONK_DIR.

By default, deployed files and installed packages are kept. With --force (-f),
also delete the deployed file or uninstall the package before removing it from
management. Failed removals stay managed. Unmanaged items are skipped.

Files and packages can be mixed. Explicit file paths (./ or /) disambiguate names
containing a colon. File removal stays under $HOME and $PLONK_DIR.

Examples:
  plonk rm ~/.vimrc brew:ripgrep         # Stop managing; keep installed items
  plonk rm -f ~/.vimrc brew:ripgrep      # Delete/uninstall and stop managing
  plonk rm --dry-run -f cargo:bat        # Preview both steps`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runRm,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolP("force", "f", false, "Also delete deployed files or uninstall packages")
	rmCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")

	// Add file path completion
	rmCmd.ValidArgsFunction = CompleteResourceArgs
}

func runRm(cmd *cobra.Command, args []string) error {
	files, specs := splitResourceArgs(args)
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	var errs []error
	if len(specs) > 0 {
		errs = append(errs, mutatePackages(cmd.Context(), config.GetDefaultConfigDirectory(), specs, false, force, dryRun))
	}
	if len(files) > 0 {
		errs = append(errs, runRmFiles(cmd, files))
	}
	return errors.Join(errs...)
}

func runRmFiles(cmd *cobra.Command, args []string) error {
	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	force, _ := cmd.Flags().GetBool("force")

	// Get directories
	homeDir, err := config.GetHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configDir := config.GetDefaultConfigDirectory()

	// Load config using LoadWithDefaults for consistent zero-config behavior
	cfg := config.LoadWithDefaults(configDir)

	// Create DotfileManager directly
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)

	// Configure options
	opts := RemoveOptions{
		DryRun: dryRun,
		Force:  force,
	}

	// Process dotfiles using helper function
	results := removeDotfiles(dm, configDir, homeDir, args, opts)

	// Create output data
	summary := calculateRemovalSummary(results)

	// Convert results to serializable format
	formatterData := output.DotfileRemovalOutput{
		TotalFiles: len(results),
		Results:    convertRemoveResultsToSerializable(results),
		Summary: output.DotfileRemovalSummary{
			Removed: summary.Removed,
			Skipped: summary.Skipped,
			Failed:  summary.Failed,
		},
	}
	formatter := output.NewDotfileRemovalFormatter(formatterData)
	output.RenderOutput(formatter)

	// Auto-commit if any files were actually removed
	if !dryRun && summary.Removed > 0 {
		gitops.AutoCommit(cmd.Context(), configDir, "rm", args)
	}

	// Check if all operations failed and return appropriate error
	return validateRemoveResultsErr(results)
}

// DotfileRemovalSummary provides summary for dotfile removal
type DotfileRemovalSummary struct {
	Removed int `json:"removed" yaml:"removed"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculateRemovalSummary calculates summary from remove results
func calculateRemovalSummary(results []RemoveResult) DotfileRemovalSummary {
	summary := DotfileRemovalSummary{}
	for _, result := range results {
		switch result.Status {
		case RemoveStatusRemoved:
			summary.Removed++
		case RemoveStatusSkipped:
			summary.Skipped++
		case RemoveStatusFailed:
			summary.Failed++
		}
	}
	return summary
}

// convertRemoveResultsToSerializable converts RemoveResult to SerializableRemovalResult
func convertRemoveResultsToSerializable(results []RemoveResult) []output.SerializableRemovalResult {
	converted := make([]output.SerializableRemovalResult, len(results))
	for i, result := range results {
		errorStr := ""
		if result.Error != nil {
			errorStr = result.Error.Error()
		}
		converted[i] = output.SerializableRemovalResult{
			Name:   result.Path,
			Status: result.Status.String(),
			Error:  errorStr,
			Metadata: map[string]interface{}{
				"source":      result.Source,
				"destination": result.Destination,
				"force":       result.Force,
			},
		}
	}
	return converted
}

// validateRemoveResultsErr returns an error if any remove operation failed
// (partial failure is a non-zero exit per the documented batch policy)
func validateRemoveResultsErr(results []RemoveResult) error {
	return ValidateBatchResults(len(results), "remove dotfiles", func(i int) bool {
		return results[i].Status == RemoveStatusFailed
	})
}
