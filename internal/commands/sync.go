// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/runtime"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all pending changes",
	Long: `Apply all pending changes from your plonk configuration to your system.

This command will:
1. Install all missing packages from your configuration
2. Deploy all dotfiles from your configuration
3. Report the results for both operations

This syncs all configured packages and dotfiles in a single operation.
Think of it like 'git pull' - it brings your system in sync with your configuration.

Examples:
  plonk sync             # Sync all configuration changes
  plonk sync --dry-run   # Show what would be synced without making changes
  plonk sync --backup    # Create backups before overwriting existing dotfiles
  plonk sync --packages  # Sync packages only
  plonk sync --dotfiles  # Sync dotfiles only`,
	RunE: runSync,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Scope flags (mutually exclusive)
	syncCmd.Flags().Bool("packages", false, "Sync packages only")
	syncCmd.Flags().Bool("dotfiles", false, "Sync dotfiles only")
	syncCmd.MarkFlagsMutuallyExclusive("packages", "dotfiles")

	// Behavior flags
	syncCmd.Flags().BoolP("dry-run", "n", false, "Show what would be synced without making changes")
	syncCmd.Flags().Bool("backup", false, "Create backups before overwriting existing dotfiles")
}

func runSync(cmd *cobra.Command, args []string) error {
	// Parse flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	backup, _ := cmd.Flags().GetBool("backup")
	packagesOnly, _ := cmd.Flags().GetBool("packages")
	dotfilesOnly, _ := cmd.Flags().GetBool("dotfiles")

	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "sync", "output-format", "invalid output format")
	}

	// Get directories from shared context
	sharedCtx := runtime.GetSharedContext()
	homeDir := sharedCtx.HomeDir()
	configDir := sharedCtx.ConfigDir()

	// Load configuration
	cfg := config.LoadConfigWithDefaults(configDir)

	var packageResult, dotfileResult interface{}

	// Sync packages (unless dotfiles-only)
	if !dotfilesOnly {
		packageResult, err = syncPackages(configDir, cfg, dryRun, format)
		if err != nil {
			return err
		}
	}

	// Sync dotfiles (unless packages-only)
	if !packagesOnly {
		dotfileResult, err = syncDotfiles(configDir, homeDir, cfg, dryRun, backup, format)
		if err != nil {
			return err
		}
	}

	// Prepare combined output
	outputData := CombinedSyncOutput{
		DryRun:   dryRun,
		Packages: packageResult,
		Dotfiles: dotfileResult,
		Scope:    getSyncScope(packagesOnly, dotfilesOnly),
	}

	return RenderOutput(outputData, format)
}

// syncPackages handles package synchronization (reuses apply logic)
func syncPackages(configDir string, cfg *config.Config, dryRun bool, format OutputFormat) (ApplyOutput, error) {
	// Reuse the existing package apply logic from apply.go
	return applyPackages(configDir, cfg, dryRun, format)
}

// syncDotfiles handles dotfile synchronization (reuses apply logic)
func syncDotfiles(configDir, homeDir string, cfg *config.Config, dryRun, backup bool, format OutputFormat) (DotfileApplyOutput, error) {
	// Reuse the existing dotfile apply logic from apply.go
	return applyDotfiles(configDir, homeDir, cfg, dryRun, backup, format)
}

// getSyncScope returns a description of what's being synced
func getSyncScope(packagesOnly, dotfilesOnly bool) string {
	if packagesOnly {
		return "packages"
	}
	if dotfilesOnly {
		return "dotfiles"
	}
	return "all"
}

// CombinedSyncOutput represents the output structure for the sync command
type CombinedSyncOutput struct {
	DryRun   bool        `json:"dry_run" yaml:"dry_run"`
	Scope    string      `json:"scope" yaml:"scope"`
	Packages interface{} `json:"packages,omitempty" yaml:"packages,omitempty"`
	Dotfiles interface{} `json:"dotfiles,omitempty" yaml:"dotfiles,omitempty"`
}

// TableOutput generates human-friendly table output for sync
func (c CombinedSyncOutput) TableOutput() string {
	output := ""

	if c.DryRun {
		output += "Plonk Sync (Dry Run)\n"
		output += "====================\n\n"
	} else {
		output += "Plonk Sync\n"
		output += "==========\n\n"
	}

	// Show scope
	switch c.Scope {
	case "packages":
		output += "📦 Syncing packages only\n\n"
	case "dotfiles":
		output += "📄 Syncing dotfiles only\n\n"
	default:
		output += "📦📄 Syncing packages and dotfiles\n\n"
	}

	// Package summary
	if c.Packages != nil {
		if pkgResult, ok := c.Packages.(ApplyOutput); ok {
			if c.DryRun {
				output += fmt.Sprintf("📦 Packages: %d would be installed\n", pkgResult.TotalWouldInstall)
			} else {
				output += fmt.Sprintf("📦 Packages: %d installed, %d failed\n", pkgResult.TotalInstalled, pkgResult.TotalFailed)
			}
		}
	}

	// Dotfile summary
	if c.Dotfiles != nil {
		if dotResult, ok := c.Dotfiles.(DotfileApplyOutput); ok {
			if c.DryRun {
				output += fmt.Sprintf("📄 Dotfiles: %d would be deployed, %d would be skipped\n", dotResult.Deployed, dotResult.Skipped)
			} else {
				output += fmt.Sprintf("📄 Dotfiles: %d deployed, %d skipped\n", dotResult.Deployed, dotResult.Skipped)
			}
		}
	}

	if c.DryRun {
		output += "\nUse 'plonk sync' without --dry-run to apply these changes\n"
	}

	return output
}

// StructuredData returns the structured data for serialization
func (c CombinedSyncOutput) StructuredData() any {
	return c
}
