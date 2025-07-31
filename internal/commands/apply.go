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

	// Convert to legacy output format
	outputData := convertApplyResult(result, getApplyScope(packagesOnly, dotfilesOnly))

	// Render output first
	renderErr := RenderOutput(outputData, format)
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

// convertApplyResult converts new ApplyResult to legacy CombinedApplyOutput
func convertApplyResult(result orchestrator.ApplyResult, scope string) CombinedApplyOutput {
	return CombinedApplyOutput{
		DryRun:        result.DryRun,
		Packages:      result.Packages,
		Dotfiles:      result.Dotfiles,
		Scope:         scope,
		PackageErrors: result.PackageErrors,
		DotfileErrors: result.DotfileErrors,
		Success:       result.Success,
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
	DryRun        bool        `json:"dry_run" yaml:"dry_run"`
	Scope         string      `json:"scope" yaml:"scope"`
	Packages      interface{} `json:"packages,omitempty" yaml:"packages,omitempty"`
	Dotfiles      interface{} `json:"dotfiles,omitempty" yaml:"dotfiles,omitempty"`
	PackageErrors []string    `json:"package_errors,omitempty" yaml:"package_errors,omitempty"`
	DotfileErrors []string    `json:"dotfile_errors,omitempty" yaml:"dotfile_errors,omitempty"`
	Success       bool        `json:"success" yaml:"success"`
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

	// Show detailed results if available

	// Package details
	if c.Packages != nil {
		if pkgResult, ok := c.Packages.(orchestrator.PackageApplyResult); ok && len(pkgResult.Managers) > 0 {
			for _, mgr := range pkgResult.Managers {
				if len(mgr.Packages) > 0 {
					output += fmt.Sprintf("%s:\n", mgr.Name)
					for _, pkg := range mgr.Packages {
						switch pkg.Status {
						case "installed":
							output += fmt.Sprintf("  ✓ %s\n", pkg.Name)
						case "would-install":
							output += fmt.Sprintf("  → %s (would install)\n", pkg.Name)
						case "failed":
							output += fmt.Sprintf("  ✗ %s: %s\n", pkg.Name, pkg.Error)
						}
					}
					output += "\n"
				}
			}
		}
	}

	// Dotfile details
	if c.Dotfiles != nil {
		if dotResult, ok := c.Dotfiles.(orchestrator.DotfileApplyResult); ok && len(dotResult.Actions) > 0 {
			output += "Dotfiles:\n"
			for _, action := range dotResult.Actions {
				switch action.Status {
				case "added":
					output += fmt.Sprintf("  ✓ %s\n", action.Destination)
				case "would-add":
					output += fmt.Sprintf("  → %s (would deploy)\n", action.Destination)
				case "failed":
					output += fmt.Sprintf("  ✗ %s: %s\n", action.Destination, action.Error)
				}
			}
			output += "\n"
		}
	}

	// Summary section
	output += "Summary:\n"
	output += "--------\n"

	totalSucceeded := 0
	totalFailed := 0

	// Package summary
	if c.Packages != nil {
		if pkgResult, ok := c.Packages.(orchestrator.PackageApplyResult); ok {
			if c.DryRun {
				output += fmt.Sprintf("Packages: %d would be installed\n", pkgResult.TotalWouldInstall)
			} else {
				if pkgResult.TotalInstalled > 0 || pkgResult.TotalFailed > 0 {
					output += fmt.Sprintf("Packages: %d installed, %d failed\n", pkgResult.TotalInstalled, pkgResult.TotalFailed)
					totalSucceeded += pkgResult.TotalInstalled
					totalFailed += pkgResult.TotalFailed
				} else if pkgResult.TotalMissing == 0 {
					output += "Packages: All up to date\n"
				}
			}
		}
	}

	// Dotfile summary
	if c.Dotfiles != nil {
		if dotResult, ok := c.Dotfiles.(orchestrator.DotfileApplyResult); ok {
			if c.DryRun {
				output += fmt.Sprintf("📄 Dotfiles: %d would be deployed\n", dotResult.Summary.Added)
			} else {
				if dotResult.Summary.Added > 0 || dotResult.Summary.Failed > 0 {
					output += fmt.Sprintf("📄 Dotfiles: %d deployed, %d failed\n", dotResult.Summary.Added, dotResult.Summary.Failed)
					totalSucceeded += dotResult.Summary.Added
					totalFailed += dotResult.Summary.Failed
				} else if dotResult.TotalFiles == 0 {
					output += "📄 Dotfiles: None configured\n"
				} else {
					output += "📄 Dotfiles: All up to date\n"
				}
			}
		}
	}

	// Overall result
	if !c.DryRun && (totalSucceeded > 0 || totalFailed > 0) {
		output += fmt.Sprintf("\nTotal: %d succeeded, %d failed\n", totalSucceeded, totalFailed)
		if totalFailed > 0 {
			output += "\nSome operations failed. Check the errors above.\n"
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
