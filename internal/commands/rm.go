// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
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
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get flags
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return err
	}

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load config using LoadWithDefaults for consistent zero-config behavior
	cfg := config.LoadWithDefaults(configDir)

	// Create dotfile manager
	manager := dotfiles.NewManager(homeDir, configDir)

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
	outputData := DotfileRemovalOutput{
		TotalFiles: len(results),
		Results:    results,
		Summary:    summary,
	}

	// Render output
	if err := RenderOutput(outputData, format); err != nil {
		return err
	}

	// Check if all operations failed and return appropriate error
	return resources.ValidateOperationResults(results, "remove dotfiles")
}

// DotfileRemovalOutput represents the output for dotfile removal
type DotfileRemovalOutput struct {
	TotalFiles int                         `json:"total_files" yaml:"total_files"`
	Results    []resources.OperationResult `json:"results" yaml:"results"`
	Summary    DotfileRemovalSummary       `json:"summary" yaml:"summary"`
}

// DotfileRemovalSummary provides summary for dotfile removal
type DotfileRemovalSummary struct {
	Removed int `json:"removed" yaml:"removed"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculateRemovalSummary calculates summary from results using generic operations summary
func calculateRemovalSummary(results []resources.OperationResult) DotfileRemovalSummary {
	genericSummary := resources.CalculateSummary(results)
	return DotfileRemovalSummary{
		Removed: genericSummary.Added, // In removal context, "added" means "removed"
		Skipped: genericSummary.Skipped,
		Failed:  genericSummary.Failed,
	}
}

// TableOutput generates human-friendly output
func (d DotfileRemovalOutput) TableOutput() string {
	tb := NewTableBuilder()

	// For single file operations, show inline result
	if d.TotalFiles == 1 && len(d.Results) == 1 {
		result := d.Results[0]
		switch result.Status {
		case "removed":
			tb.AddLine("Removed dotfile from plonk management")
			tb.AddLine("   File: %s", result.Name)
			if source, ok := result.Metadata["source"].(string); ok {
				tb.AddLine("   Source: %s (removed from config)", source)
			}
		case "would-remove":
			tb.AddLine("Would remove dotfile from plonk management (dry-run)")
			tb.AddLine("   File: %s", result.Name)
			if source, ok := result.Metadata["source"].(string); ok {
				tb.AddLine("   Source: %s", source)
			}
		case "skipped":
			tb.AddLine("Skipped: %s", result.Name)
			if result.Error != nil {
				tb.AddLine("   Reason: %s", result.Error.Error())
			}
		case "failed":
			tb.AddLine("Failed: %s", result.Name)
			if result.Error != nil {
				tb.AddLine("   Error: %s", result.Error.Error())
			}
		}
		return tb.Build()
	}

	// For batch operations, show summary
	tb.AddTitle("Dotfile Removal")
	tb.AddNewline()

	// Check if this is a dry run
	isDryRun := false
	wouldRemoveCount := 0
	for _, result := range d.Results {
		if result.Status == "would-remove" {
			isDryRun = true
			wouldRemoveCount++
		}
	}

	if isDryRun {
		if wouldRemoveCount > 0 {
			tb.AddLine("Would remove %d dotfiles (dry-run)", wouldRemoveCount)
		}
	} else {
		if d.Summary.Removed > 0 {
			tb.AddLine("📄 Removed %d dotfiles", d.Summary.Removed)
		}
	}

	if d.Summary.Skipped > 0 {
		tb.AddLine("%d skipped", d.Summary.Skipped)
	}
	if d.Summary.Failed > 0 {
		tb.AddLine("%d failed", d.Summary.Failed)
	}

	tb.AddNewline()

	// Show individual files
	for _, result := range d.Results {
		switch result.Status {
		case "removed":
			tb.AddLine("   ✓ %s", result.Name)
		case "would-remove":
			tb.AddLine("   - %s", result.Name)
		case "skipped":
			tb.AddLine("   %s (not managed)", result.Name)
		case "failed":
			tb.AddLine("   ✗ %s", result.Name)
		}
	}

	tb.AddNewline()
	tb.AddLine("Total: %d dotfiles processed", d.TotalFiles)

	return tb.Build()
}

// StructuredData returns the structured data for serialization
func (d DotfileRemovalOutput) StructuredData() any {
	return d
}
