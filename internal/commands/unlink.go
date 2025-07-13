// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/spf13/cobra"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink <files...>",
	Short: "Explicitly unlink dotfiles",
	Long: `Explicitly remove dotfiles from plonk management and unlink them.

This command forces all arguments to be treated as dotfiles, regardless
of their format. Use this when the automatic detection in 'plonk rm'
doesn't work correctly, or when you want to be explicit.

This command will:
- Remove the symlinks from your home directory
- Keep the source files in your plonk configuration directory
- The dotfiles remain available for re-linking later

Examples:
  plonk unlink ~/.zshrc                    # Unlink single file
  plonk unlink ~/.zshrc ~/.vimrc           # Unlink multiple files
  plonk unlink ~/.config/nvim/init.lua     # Unlink specific file
  plonk unlink --dry-run ~/.zshrc ~/.vimrc # Preview what would be unlinked
  plonk unlink config                      # Force 'config' to be treated as dotfile`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUnlink,
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
	unlinkCmd.Flags().BoolP("dry-run", "n", false, "Show what would be unlinked without making changes")
	unlinkCmd.Flags().BoolP("force", "f", false, "Force removal even if not managed")

	// Add file path completion
	unlinkCmd.ValidArgsFunction = completeDotfilePaths
}

func runUnlink(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "unlink", "failed to get home directory")
	}

	configDir := config.GetDefaultConfigDirectory()

	// Load config
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// If config doesn't exist, we can't unlink dotfiles
		return errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "unlink", "output-format", "invalid output format")
	}

	// Process dotfiles sequentially
	results := make([]operations.OperationResult, 0, len(args))
	reporter := operations.NewProgressReporter("unlink", format == OutputTable)

	for _, dotfilePath := range args {
		result := removeSingleDotfile(homeDir, configDir, cfg, dotfilePath, dryRun)
		results = append(results, result)

		// Show progress immediately
		reporter.ShowItemProgress(result)
	}

	// Handle output based on format
	if format == OutputTable {
		// Show summary for table output
		reporter.ShowBatchSummary(results)
	} else {
		// For structured output, create appropriate response
		if len(args) == 1 {
			// Single dotfile - use simple output
			result := results[0]
			output := DotfileUnlinkOutput{
				Path:   result.Name,
				Status: result.Status,
			}
			if result.Error != nil {
				output.Error = result.Error.Error()
			}
			return RenderOutput(output, format)
		} else {
			// Multiple dotfiles - use batch output
			batchOutput := DotfileBatchUnlinkOutput{
				TotalFiles:    len(results),
				UnlinkedFiles: convertResultsToUnlink(results),
				Errors:        extractErrorMessages(results),
			}
			return RenderOutput(batchOutput, format)
		}
	}

	// Determine exit code
	return operations.DetermineExitCode(results, errors.DomainDotfiles, "unlink-multiple")
}

// convertResultsToUnlink converts OperationResult to DotfileUnlinkOutput for structured output
func convertResultsToUnlink(results []operations.OperationResult) []DotfileUnlinkOutput {
	outputs := make([]DotfileUnlinkOutput, 0, len(results))
	for _, result := range results {
		if result.Status == "failed" {
			continue // Skip failed results, they're handled in errors
		}

		outputs = append(outputs, DotfileUnlinkOutput{
			Path:   result.Name,
			Status: result.Status,
		})
	}
	return outputs
}

// DotfileUnlinkOutput represents the output structure for dotfile unlink command
type DotfileUnlinkOutput struct {
	Path   string `json:"path" yaml:"path"`
	Status string `json:"status" yaml:"status"`
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileBatchUnlinkOutput represents the output structure for batch dotfile unlink operations
type DotfileBatchUnlinkOutput struct {
	TotalFiles    int                   `json:"total_files" yaml:"total_files"`
	UnlinkedFiles []DotfileUnlinkOutput `json:"unlinked_files" yaml:"unlinked_files"`
	Errors        []string              `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// TableOutput generates human-friendly table output for dotfile unlink
func (d DotfileUnlinkOutput) TableOutput() string {
	switch d.Status {
	case "unlinked":
		return "✅ Unlinked: " + d.Path + "\n"
	case "would-unlink":
		return "🔗 Would unlink: " + d.Path + "\n"
	case "skipped":
		return "⏭️ Skipped: " + d.Path + " (not a managed symlink)\n"
	default:
		return "❌ Failed: " + d.Path + "\n"
	}
}

// StructuredData returns the structured data for serialization
func (d DotfileUnlinkOutput) StructuredData() any {
	return d
}

// TableOutput generates human-friendly table output for batch dotfile unlink
func (d DotfileBatchUnlinkOutput) TableOutput() string {
	output := "Dotfile Unlink\n==============\n\n"

	if len(d.UnlinkedFiles) > 0 {
		output += "✅ Unlinked files:\n"
		for _, file := range d.UnlinkedFiles {
			output += "  " + file.Path + "\n"
		}
		output += "\n"
	}

	if len(d.Errors) > 0 {
		output += "❌ Errors:\n"
		for _, err := range d.Errors {
			output += "  " + err + "\n"
		}
		output += "\n"
	}

	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileBatchUnlinkOutput) StructuredData() any {
	return d
}
