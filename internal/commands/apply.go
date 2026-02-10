// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/output"
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

	// Get directories
	homeDir, err := config.GetHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configDir := config.GetDefaultConfigDirectory()

	// Load configuration
	cfg := config.LoadWithDefaults(configDir)

	ctx := context.Background()

	// If specific files are provided, apply only those dotfiles
	if len(args) > 0 {
		if packagesOnly || dotfilesOnly {
			return fmt.Errorf("cannot specify files with --packages or --dotfiles flags")
		}
		return runSelectiveApply(ctx, args, cfg, configDir, homeDir, dryRun)
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
	output.RenderOutput(result)

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

// normalizePathWithHome normalizes a path using the specified home directory
// This is used for selective apply to ensure consistent path normalization
func normalizePathWithHome(path, homeDir string) (string, error) {
	// First expand any environment variables (e.g., $HOME, $ZSHPATH)
	path = os.ExpandEnv(path)

	// Then expand tilde using the provided homeDir
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	// Finally, convert to absolute path (handles relative paths)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %s: %w", path, err)
	}

	// Clean the path to remove any redundant elements
	return filepath.Clean(absPath), nil
}

// runSelectiveApply applies only specific dotfiles
func runSelectiveApply(ctx context.Context, paths []string, cfg *config.Config, configDir, homeDir string, dryRun bool) error {
	// First, get all managed dotfiles to validate the requested files
	dm := dotfiles.NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)
	statuses, err := dm.Reconcile()
	if err != nil {
		return fmt.Errorf("failed to get dotfile status: %w", err)
	}

	managedDests := make(map[string]bool) // normalized dest path -> exists

	// Build map of managed files (by deployed path)
	for _, s := range statuses {
		if s.Target != "" {
			normalizedDest, err := normalizePathWithHome(s.Target, homeDir)
			if err == nil {
				managedDests[normalizedDest] = true
			}
		}
	}

	// Build filter set from requested paths, validating each one
	filterSet := make(map[string]bool)
	for _, path := range paths {
		normalizedPath, err := normalizePathWithHome(path, homeDir)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", path, err)
		}

		if !managedDests[normalizedPath] {
			return fmt.Errorf("file not managed by plonk: %s", path)
		}

		filterSet[normalizedPath] = true
	}

	// Apply only the selected dotfiles
	opts := dotfiles.ApplyFilterOptions{
		DryRun: dryRun,
		Filter: filterSet,
	}
	applyResult, applyErr := dotfiles.ApplySelective(ctx, configDir, homeDir, cfg, opts)

	// Wrap in ApplyResult for consistent output formatting
	result := output.ApplyResult{
		DryRun:   dryRun,
		Success:  applyErr == nil,
		Scope:    "dotfiles (selective)",
		Dotfiles: &applyResult,
	}

	// Always render output so users see per-file diagnostics on partial failure
	output.RenderOutput(result)

	if applyErr != nil {
		return fmt.Errorf("failed to apply dotfiles: %w", applyErr)
	}

	return nil
}
