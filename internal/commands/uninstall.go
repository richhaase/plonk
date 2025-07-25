// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <packages...>",
	Short: "Uninstall packages and remove from plonk management",
	Long: `Uninstall packages from your system and remove them from your lock file.

This command uninstalls packages from your system using the appropriate package
manager and removes them from plonk management.

Examples:
  plonk uninstall htop                    # Uninstall htop and remove from lock file
  plonk uninstall git neovim              # Uninstall multiple packages
  plonk uninstall --dry-run htop          # Preview what would be uninstalled`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	// Manager-specific flags (mutually exclusive)
	uninstallCmd.Flags().Bool("brew", false, "Use Homebrew package manager")
	uninstallCmd.Flags().Bool("npm", false, "Use NPM package manager")
	uninstallCmd.Flags().Bool("cargo", false, "Use Cargo package manager")
	uninstallCmd.Flags().Bool("pip", false, "Use pip package manager")
	uninstallCmd.Flags().Bool("gem", false, "Use gem package manager")
	uninstallCmd.Flags().Bool("go", false, "Use go install package manager")
	uninstallCmd.MarkFlagsMutuallyExclusive("brew", "npm", "cargo", "pip", "gem", "go")

	// Common flags
	uninstallCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")
	uninstallCmd.Flags().BoolP("force", "f", false, "Force removal even if not managed")
}

func runUninstall(cmd *cobra.Command, args []string) error {
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
	configDir := config.GetDefaultConfigDirectory()

	// Configure uninstallation options
	opts := packages.UninstallOptions{
		Manager: flags.Manager,
		DryRun:  flags.DryRun,
	}

	// Process packages using managers package
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	results, err := packages.UninstallPackages(ctx, configDir, args, opts)
	if err != nil {
		return err
	}

	// Show progress reporting
	reporter := output.NewProgressReporterForOperation("uninstall", "package", true)
	for _, result := range results {
		reporter.ShowItemProgress(result)
	}

	// Show batch summary
	reporter.ShowBatchSummary(results)

	// Create output data
	summary := calculateUninstallSummary(results)
	outputData := PackageUninstallOutput{
		TotalPackages: len(results),
		Results:       results,
		Summary:       summary,
	}

	// Render output
	if err := RenderOutput(outputData, format); err != nil {
		return err
	}

	// Determine exit code based on results
	exitErr := DetermineExitCode(results, "packages", "uninstall")
	if exitErr != nil {
		return exitErr
	}

	return nil
}

// PackageUninstallOutput represents the output for package uninstallation
type PackageUninstallOutput struct {
	TotalPackages int                         `json:"total_packages" yaml:"total_packages"`
	Results       []resources.OperationResult `json:"results" yaml:"results"`
	Summary       PackageUninstallSummary     `json:"summary" yaml:"summary"`
}

// PackageUninstallSummary provides summary for package uninstallation
type PackageUninstallSummary struct {
	Removed int `json:"removed" yaml:"removed"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculateUninstallSummary calculates summary from uninstall results using generic operations summary
func calculateUninstallSummary(results []resources.OperationResult) PackageUninstallSummary {
	genericSummary := resources.CalculateSummary(results)
	return PackageUninstallSummary{
		Removed: genericSummary.Removed,
		Skipped: genericSummary.Skipped,
		Failed:  genericSummary.Failed,
	}
}

// TableOutput generates human-friendly output
func (p PackageUninstallOutput) TableOutput() string {
	tb := NewTableBuilder()

	tb.AddTitle("Package Uninstallation")
	tb.AddNewline()

	if p.Summary.Removed > 0 {
		tb.AddLine("%s Removed %d packages", IconPackage, p.Summary.Removed)
	}
	if p.Summary.Skipped > 0 {
		tb.AddLine("⏭️ %d skipped", p.Summary.Skipped)
	}
	if p.Summary.Failed > 0 {
		tb.AddLine("%s %d failed", IconUnhealthy, p.Summary.Failed)
	}

	tb.AddNewline()
	tb.AddLine("Total: %d packages processed", p.TotalPackages)

	return tb.Build()
}

// StructuredData returns the structured data for serialization
func (p PackageUninstallOutput) StructuredData() any {
	return p
}
