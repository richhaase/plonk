// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply configuration to reconcile system state",
	Long: `Apply reads your plonk configuration and reconciles the system state
to match, installing missing packages and managing dotfiles.

This command will:
1. Install all missing packages from your configuration
2. Deploy all dotfiles from your configuration
3. Report the results for both operations

This applies all configured packages and dotfiles in a single operation.
Think of it like 'git pull' - it brings your system state in line with your configuration.

Examples:
  plonk apply             # Apply all configuration changes
  plonk apply --dry-run   # Show what would be applied without making changes
  plonk apply --packages  # Apply packages only
  plonk apply --dotfiles  # Apply dotfiles only`,
	RunE:         runApply,
	SilenceUsage: true,
	Args:         cobra.NoArgs,
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

	// Create new orchestrator with all options
	orch := orchestrator.New(
		orchestrator.WithConfig(cfg),
		orchestrator.WithConfigDir(configDir),
		orchestrator.WithHomeDir(homeDir),
		orchestrator.WithDryRun(dryRun),
		orchestrator.WithPackagesOnly(packagesOnly),
		orchestrator.WithDotfilesOnly(dotfilesOnly),
	)

	// Run apply with hooks and v2 lock
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
