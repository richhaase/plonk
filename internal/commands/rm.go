// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <files...>",
	Short: "Remove dotfiles from plonk management",
	Long: `Remove dotfiles from plonk management by deleting them from the configuration directory.

This command removes dotfiles from plonk management by deleting them from your
plonk configuration directory ($PLONK_DIR). The original files in your home
directory are NOT affected and remain in place unchanged.

After removal, the dotfiles will no longer be managed by plonk and won't be
affected by 'plonk apply' operations. Use 'plonk status' to see which files
are currently managed.

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
- Only removes from $PLONK_DIR, never touches files in $HOME
- Cannot remove directories (individual files only)
- Dotfiles within $PLONK_DIR (like .git) are protected

File Mapping (what gets removed):
- ~/.zshrc removes → $PLONK_DIR/zshrc
- ~/.config/nvim/init.lua removes → $PLONK_DIR/config/nvim/init.lua

Examples:
  plonk rm ~/.zshrc                    # Remove single file from management
  plonk rm ~/.zshrc ~/.vimrc           # Remove multiple files from management
  plonk rm vimrc                       # Finds ~/.vimrc automatically
  plonk rm .config/nvim/init.lua       # Remove specific nested file
  plonk rm --dry-run ~/.zshrc ~/.vimrc # Preview what would be removed`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runRm,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")

	// Add file path completion
	rmCmd.ValidArgsFunction = CompleteDotfilePaths
}

func runRm(cmd *cobra.Command, args []string) error {
	// Get flags
	flags, err := parseSimpleFlags(cmd)
	if err != nil {
		return err
	}

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load config using LoadWithDefaults for consistent zero-config behavior
	cfg := config.LoadWithDefaults(configDir)

	// Create dotfile manager with injected config
	manager := dotfiles.NewManagerWithConfig(homeDir, configDir, cfg)

	// Configure options
	opts := dotfiles.RemoveOptions{
		DryRun: flags.DryRun,
	}

	// Process dotfiles using domain package
	results, err := manager.RemoveFiles(cfg, args, opts)
	if err != nil {
		return err
	}

	// Create output data
	summary := calculateRemovalSummary(results)

	// Convert results to serializable format
	formatterData := output.DotfileRemovalOutput{
		TotalFiles: len(results),
		Results:    convertRemoveResultsToOutput(results),
		Summary: output.DotfileRemovalSummary{
			Removed: summary.Removed,
			Skipped: summary.Skipped,
			Failed:  summary.Failed,
		},
	}
	formatter := output.NewDotfileRemovalFormatter(formatterData)
	output.RenderOutput(formatter)

	// Check if all operations failed and return appropriate error
	return validateRemoveResults(results)
}

// DotfileRemovalSummary provides summary for dotfile removal
type DotfileRemovalSummary struct {
	Removed int `json:"removed" yaml:"removed"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculateRemovalSummary calculates summary from remove results
func calculateRemovalSummary(results []dotfiles.RemoveResult) DotfileRemovalSummary {
	summary := DotfileRemovalSummary{}
	for _, result := range results {
		switch result.Status {
		case dotfiles.RemoveStatusRemoved, dotfiles.RemoveStatusWouldRemove:
			summary.Removed++
		case dotfiles.RemoveStatusSkipped:
			summary.Skipped++
		case dotfiles.RemoveStatusFailed:
			summary.Failed++
		}
	}
	return summary
}

// convertRemoveResultsToOutput converts dotfiles.RemoveResult to SerializableRemovalResult
func convertRemoveResultsToOutput(results []dotfiles.RemoveResult) []output.SerializableRemovalResult {
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
			},
		}
	}
	return converted
}

// validateRemoveResults checks if all remove operations failed and returns appropriate error
func validateRemoveResults(results []dotfiles.RemoveResult) error {
	if len(results) == 0 {
		return nil
	}

	allFailed := true
	for _, result := range results {
		if result.Status != dotfiles.RemoveStatusFailed {
			allFailed = false
			break
		}
	}

	if allFailed {
		return fmt.Errorf("remove dotfiles operation failed: all %d item(s) failed to process", len(results))
	}

	return nil
}
