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
  plonk apply --backup    # Create backups before overwriting existing dotfiles
  plonk apply --packages  # Apply packages only
  plonk apply --dotfiles  # Apply dotfiles only`,
	RunE: runApply,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Scope flags (mutually exclusive)
	applyCmd.Flags().Bool("packages", false, "Apply packages only")
	applyCmd.Flags().Bool("dotfiles", false, "Apply dotfiles only")
	applyCmd.MarkFlagsMutuallyExclusive("packages", "dotfiles")

	// Behavior flags
	applyCmd.Flags().BoolP("dry-run", "n", false, "Show what would be applied without making changes")
	applyCmd.Flags().Bool("backup", false, "Create backups before overwriting existing dotfiles")
}

func runApply(cmd *cobra.Command, args []string) error {
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

	// For now, maintain backward compatibility by using the old apply functions
	// TODO: After Task 6.2, update to use scope flags (packages-only, dotfiles-only)
	if packagesOnly || dotfilesOnly {
		// Handle scoped apply using legacy functions
		return runScopedApply(ctx, configDir, homeDir, cfg, dryRun, backup, packagesOnly, dotfilesOnly, format)
	}

	// Create new orchestrator
	orch := orchestrator.New(
		orchestrator.WithConfig(cfg),
		orchestrator.WithConfigDir(configDir),
		orchestrator.WithHomeDir(homeDir),
		orchestrator.WithDryRun(dryRun),
	)

	// Run apply with hooks and v2 lock
	result, err := orch.Apply(ctx)
	if err != nil {
		return fmt.Errorf("apply failed: %w", err)
	}

	// Convert to legacy output format for now
	outputData := convertApplyResult(result, getApplyScope(packagesOnly, dotfilesOnly))

	return RenderOutput(outputData, format)
}

// runScopedApply handles packages-only or dotfiles-only apply using legacy functions
func runScopedApply(ctx context.Context, configDir, homeDir string, cfg *config.Config, dryRun, backup, packagesOnly, dotfilesOnly bool, format OutputFormat) error {
	var packageOutput *ApplyOutput
	var dotfileOutput *DotfileApplyOutput

	// Apply packages (unless dotfiles-only)
	if !dotfilesOnly {
		result, err := orchestrator.ApplyPackages(ctx, configDir, cfg, dryRun)
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

	// Apply dotfiles (unless packages-only)
	if !packagesOnly {
		result, err := orchestrator.ApplyDotfiles(ctx, configDir, homeDir, cfg, dryRun, backup)
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

	outputData := CombinedApplyOutput{
		DryRun:   dryRun,
		Packages: pkgInterface,
		Dotfiles: dotInterface,
		Scope:    getApplyScope(packagesOnly, dotfilesOnly),
	}

	return RenderOutput(outputData, format)
}

// convertApplyResult converts new ApplyResult to legacy CombinedApplyOutput
func convertApplyResult(result orchestrator.ApplyResult, scope string) CombinedApplyOutput {
	return CombinedApplyOutput{
		DryRun:   result.DryRun,
		Packages: result.Packages,
		Dotfiles: result.Dotfiles,
		Scope:    scope,
	}
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

// CombinedApplyOutput represents the output structure for the apply command
type CombinedApplyOutput struct {
	DryRun   bool        `json:"dry_run" yaml:"dry_run"`
	Scope    string      `json:"scope" yaml:"scope"`
	Packages interface{} `json:"packages,omitempty" yaml:"packages,omitempty"`
	Dotfiles interface{} `json:"dotfiles,omitempty" yaml:"dotfiles,omitempty"`
}

// TableOutput generates human-friendly table output for apply
func (c CombinedApplyOutput) TableOutput() string {
	output := ""

	if c.DryRun {
		output += "Plonk Apply (Dry Run)\n"
		output += "=====================\n\n"
	} else {
		output += "Plonk Apply\n"
		output += "===========\n\n"
	}

	// Show scope
	switch c.Scope {
	case "packages":
		output += "📦 Applying packages only\n\n"
	case "dotfiles":
		output += "📄 Applying dotfiles only\n\n"
	default:
		output += "📦📄 Applying packages and dotfiles\n\n"
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
		output += "\nUse 'plonk apply' without --dry-run to apply these changes\n"
	}

	return output
}

// StructuredData returns the structured data for serialization
func (c CombinedApplyOutput) StructuredData() any {
	return c
}
