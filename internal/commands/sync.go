// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/orchestrator"
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
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load configuration
	cfg := config.LoadWithDefaults(configDir)

	ctx := context.Background()

	// For now, maintain backward compatibility by using the old sync functions
	// TODO: After Task 6.2, update to use scope flags (packages-only, dotfiles-only)
	if packagesOnly || dotfilesOnly {
		// Handle scoped sync using legacy functions
		return runScopedSync(ctx, configDir, homeDir, cfg, dryRun, backup, packagesOnly, dotfilesOnly, format)
	}

	// Create new orchestrator
	orch := orchestrator.New(
		orchestrator.WithConfig(cfg),
		orchestrator.WithConfigDir(configDir),
		orchestrator.WithHomeDir(homeDir),
		orchestrator.WithDryRun(dryRun),
	)

	// Run sync with hooks and v2 lock
	result, err := orch.Sync(ctx)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Convert to legacy output format for now
	outputData := convertSyncResult(result, getSyncScope(packagesOnly, dotfilesOnly))

	return RenderOutput(outputData, format)
}

// runScopedSync handles packages-only or dotfiles-only sync using legacy functions
func runScopedSync(ctx context.Context, configDir, homeDir string, cfg *config.Config, dryRun, backup, packagesOnly, dotfilesOnly bool, format OutputFormat) error {
	var packageOutput *ApplyOutput
	var dotfileOutput *DotfileApplyOutput

	// Sync packages (unless dotfiles-only)
	if !dotfilesOnly {
		result, err := orchestrator.SyncPackages(ctx, configDir, cfg, dryRun)
		if err != nil {
			return err
		}

		// Convert to command output format
		pkgOutput := ApplyOutput{
			DryRun:            result.DryRun,
			TotalMissing:      result.TotalMissing,
			TotalInstalled:    result.TotalInstalled,
			TotalFailed:       result.TotalFailed,
			TotalWouldInstall: result.TotalWouldInstall,
			Managers:          make([]ManagerApplyResult, len(result.Managers)),
		}

		// Convert manager results
		for i, mgr := range result.Managers {
			packages := make([]PackageApplyResult, len(mgr.Packages))
			for j, pkg := range mgr.Packages {
				packages[j] = PackageApplyResult{
					Name:   pkg.Name,
					Status: pkg.Status,
					Error:  pkg.Error,
				}
			}
			pkgOutput.Managers[i] = ManagerApplyResult{
				Name:         mgr.Name,
				MissingCount: mgr.MissingCount,
				Packages:     packages,
			}
		}

		packageOutput = &pkgOutput

		// Print summary for table format
		if format == OutputTable {
			if result.TotalMissing == 0 {
				fmt.Println("📦 All packages up to date")
			} else {
				if dryRun {
					fmt.Printf("📦 Package summary: %d packages would be installed\n", pkgOutput.TotalWouldInstall)
				} else {
					fmt.Printf("📦 Package summary: %d installed, %d failed\n", pkgOutput.TotalInstalled, pkgOutput.TotalFailed)
				}
			}
			fmt.Println()
		}
	}

	// Sync dotfiles (unless packages-only)
	if !packagesOnly {
		result, err := orchestrator.SyncDotfiles(ctx, configDir, homeDir, cfg, dryRun, backup)
		if err != nil {
			return err
		}

		// Convert to command output format
		actions := make([]DotfileAction, len(result.Actions))
		for i, action := range result.Actions {
			actions[i] = DotfileAction{
				Source:      action.Source,
				Destination: action.Destination,
				Status:      action.Status,
				Reason:      "", // Not used in this context
			}
		}

		dotOutput := DotfileApplyOutput{
			DryRun:   result.DryRun,
			Deployed: result.Summary.Added + result.Summary.Updated,
			Skipped:  result.Summary.Unchanged,
			Actions:  actions,
		}

		dotfileOutput = &dotOutput

		// Print summary for table format
		if format == OutputTable {
			if result.TotalFiles == 0 {
				fmt.Println("📄 No dotfiles configured")
			} else {
				if dryRun {
					fmt.Printf("📄 Dotfile summary: %d dotfiles would be deployed, %d would be skipped\n", dotOutput.Deployed, dotOutput.Skipped)
				} else {
					fmt.Printf("📄 Dotfile summary: %d deployed, %d skipped\n", dotOutput.Deployed, dotOutput.Skipped)
				}
			}
		}
	}

	// Prepare combined output - handle nil pointers
	var pkgInterface, dotInterface interface{}
	if packageOutput != nil {
		pkgInterface = *packageOutput
	}
	if dotfileOutput != nil {
		dotInterface = *dotfileOutput
	}

	outputData := CombinedSyncOutput{
		DryRun:   dryRun,
		Packages: pkgInterface,
		Dotfiles: dotInterface,
		Scope:    getSyncScope(packagesOnly, dotfilesOnly),
	}

	return RenderOutput(outputData, format)
}

// convertSyncResult converts new SyncResult to legacy CombinedSyncOutput
func convertSyncResult(result orchestrator.SyncResult, scope string) CombinedSyncOutput {
	return CombinedSyncOutput{
		DryRun:   result.DryRun,
		Packages: result.Packages,
		Dotfiles: result.Dotfiles,
		Scope:    scope,
	}
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
