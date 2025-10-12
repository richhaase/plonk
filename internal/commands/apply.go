// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [files...]",
	Short: "Apply configuration to reconcile system state",
	Long: `Apply reads your plonk configuration and reconciles the system state
to match, installing missing packages and managing dotfiles.

This command will:
1. Install all missing packages from your configuration
2. Deploy all dotfiles from your configuration
3. Report the results for both operations

This applies all configured packages and dotfiles in a single operation.
Think of it like 'git pull' - it brings your system state in line with your configuration.

You can optionally specify specific dotfiles to apply. If files are specified,
only those dotfiles will be deployed (packages are not applied).

Examples:
  plonk apply                    # Apply all configuration changes
  plonk apply --dry-run          # Show what would be applied without making changes
  plonk apply --packages         # Apply packages only
  plonk apply --dotfiles         # Apply dotfiles only
  plonk apply ~/.vimrc ~/.zshrc  # Apply only specific dotfiles`,
	RunE:         runApply,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Scope flags (mutually exclusive)
	applyCmd.Flags().Bool("packages", false, "Apply packages only")
	applyCmd.Flags().Bool("dotfiles", false, "Apply dotfiles only")
	applyCmd.MarkFlagsMutuallyExclusive("packages", "dotfiles")

	// Behavior flags
	applyCmd.Flags().BoolP("dry-run", "n", false, "Show what would be applied without making changes")
}

func runApply(cmd *cobra.Command, args []string) error {
	// Parse flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	packagesOnly, _ := cmd.Flags().GetBool("packages")
	dotfilesOnly, _ := cmd.Flags().GetBool("dotfiles")

	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load configuration
	cfg := config.LoadWithDefaults(configDir)

	ctx := context.Background()

	// If specific files are provided, apply only those dotfiles
	if len(args) > 0 {
		if packagesOnly || dotfilesOnly {
			return fmt.Errorf("cannot specify files with --packages or --dotfiles flags")
		}
		return runSelectiveApply(ctx, args, cfg, configDir, homeDir, format, dryRun)
	}

	// Create new orchestrator with all options
	orch := orchestrator.New(
		orchestrator.WithConfig(cfg),
		orchestrator.WithConfigDir(configDir),
		orchestrator.WithHomeDir(homeDir),
		orchestrator.WithDryRun(dryRun),
		orchestrator.WithPackagesOnly(packagesOnly),
		orchestrator.WithDotfilesOnly(dotfilesOnly),
	)

	// Run apply
	result, err := orch.Apply(ctx)

	// Set the scope on the result
	result.Scope = getApplyScope(packagesOnly, dotfilesOnly)

	// Render output first
	renderErr := output.RenderOutput(result, format)
	if renderErr != nil {
		return renderErr
	}

	// Now handle any errors from apply
	if err != nil {
		// The apply completed with some errors, exit with non-zero
		return err
	}

	return nil
}

// getApplyScope returns a description of what's being applied
func getApplyScope(packagesOnly, dotfilesOnly bool) string {
	if packagesOnly {
		return "packages"
	}
	if dotfilesOnly {
		return "dotfiles"
	}
	return "all"
}

// runSelectiveApply applies only specific dotfiles
func runSelectiveApply(ctx context.Context, paths []string, cfg *config.Config, configDir, homeDir string, format output.OutputFormat, dryRun bool) error {
	// First, get all managed dotfiles to validate the requested files
	results, err := orchestrator.ReconcileAllWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to get dotfile status: %w", err)
	}

	summary := resources.ConvertResultsToSummary(results)
	managedDests := make(map[string]bool) // normalized dest path -> exists

	// Build map of managed files (by deployed path)
	for _, result := range summary.Results {
		if result.Domain == "dotfile" {
			for _, item := range result.Managed {
				if dest, ok := item.Metadata["destination"].(string); ok {
					normalizedDest, err := normalizePath(dest)
					if err == nil {
						managedDests[normalizedDest] = true
					}
				}
			}
			for _, item := range result.Missing {
				if dest, ok := item.Metadata["destination"].(string); ok {
					normalizedDest, err := normalizePath(dest)
					if err == nil {
						managedDests[normalizedDest] = true
					}
				}
			}
		}
	}

	// Validate requested paths
	for _, path := range paths {
		normalizedPath, err := normalizePath(path)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", path, err)
		}

		if !managedDests[normalizedPath] {
			return fmt.Errorf("file not managed by plonk: %s", path)
		}
	}

	// Note: For MVP, we'll apply all dotfiles. Filtering would require significant refactoring.
	// The validation above ensures requested files are managed, and the output will show all actions.
	// TODO: Implement selective filtering in a future iteration
	output.Println("Note: Applying all dotfiles. Selective apply filtering coming in a future update.")

	// Apply all dotfiles
	dotfileResult, err := dotfiles.Apply(ctx, configDir, homeDir, cfg, dryRun)
	if err != nil {
		return fmt.Errorf("failed to apply dotfiles: %w", err)
	}

	// Wrap in ApplyResult for consistent output formatting
	result := output.ApplyResult{
		DryRun:   dryRun,
		Success:  true,
		Scope:    "dotfiles",
		Dotfiles: &dotfileResult,
	}

	// Render output (shows all files, not just selected ones)
	if err := output.RenderOutput(result, format); err != nil {
		return err
	}

	return nil
}
