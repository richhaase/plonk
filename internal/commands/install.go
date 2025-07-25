// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/richhaase/plonk/internal/state"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <packages...>",
	Short: "Install packages and add them to plonk management",
	Long: `Install packages on your system and add them to your lock file for management.

This command installs packages using the specified package manager and adds them
to your lock file so they can be managed by plonk. Use specific manager flags
to control which package manager to use.

Examples:
  plonk install htop                      # Install htop using default manager
  plonk install git neovim ripgrep        # Install multiple packages
  plonk install git --brew                # Install git specifically with Homebrew
  plonk install lodash --npm              # Install lodash with npm global packages
  plonk install ripgrep --cargo           # Install ripgrep with cargo packages
  plonk install black flake8 --pip        # Install Python tools with pip
  plonk install bundler rubocop --gem     # Install Ruby tools with gem
  plonk install gopls --go                # Install Go tools with go install
  plonk install --dry-run htop neovim     # Preview what would be installed`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Manager-specific flags (mutually exclusive)
	installCmd.Flags().Bool("brew", false, "Use Homebrew package manager")
	installCmd.Flags().Bool("npm", false, "Use NPM package manager")
	installCmd.Flags().Bool("cargo", false, "Use Cargo package manager")
	installCmd.Flags().Bool("pip", false, "Use pip package manager")
	installCmd.Flags().Bool("gem", false, "Use gem package manager")
	installCmd.Flags().Bool("go", false, "Use go install package manager")
	installCmd.MarkFlagsMutuallyExclusive("brew", "npm", "cargo", "pip", "gem", "go")

	// Common flags
	installCmd.Flags().BoolP("dry-run", "n", false, "Show what would be installed without making changes")
	installCmd.Flags().BoolP("force", "f", false, "Force installation even if already managed")
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return fmt.Errorf("install: invalid output format %s: %w", outputFormat, err)
	}

	// Get flags
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return err
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Configure installation options
	opts := packages.InstallOptions{
		Manager: flags.Manager,
		DryRun:  flags.DryRun,
		Force:   flags.Force,
	}

	// Process packages using managers package
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	results, err := packages.InstallPackages(ctx, configDir, args, opts)
	if err != nil {
		return fmt.Errorf("install: failed to process packages: %w", err)
	}

	// Show progress reporting
	reporter := output.NewProgressReporterForOperation("install", "package", true)
	for _, result := range results {
		reporter.ShowItemProgress(result)
	}

	// Show batch summary
	reporter.ShowBatchSummary(results)

	// Create output data
	summary := calculatePackageSummary(results)
	outputData := PackageInstallOutput{
		TotalPackages: len(results),
		Results:       results,
		Summary:       summary,
	}

	// Render output
	if err := RenderOutput(outputData, format); err != nil {
		return err
	}

	// Determine exit code based on results
	exitErr := DetermineExitCode(results, "packages", "install")
	if exitErr != nil {
		return exitErr
	}

	return nil
}

// PackageInstallOutput represents the output for package installation
type PackageInstallOutput struct {
	TotalPackages int                     `json:"total_packages" yaml:"total_packages"`
	Results       []state.OperationResult `json:"results" yaml:"results"`
	Summary       PackageInstallSummary   `json:"summary" yaml:"summary"`
}

// PackageInstallSummary provides summary for package installation
type PackageInstallSummary struct {
	Added   int `json:"added" yaml:"added"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculatePackageSummary calculates summary from results using generic operations summary
func calculatePackageSummary(results []state.OperationResult) PackageInstallSummary {
	genericSummary := state.CalculateSummary(results)
	return PackageInstallSummary{
		Added:   genericSummary.Added,
		Skipped: genericSummary.Skipped,
		Failed:  genericSummary.Failed,
	}
}

// TableOutput generates human-friendly output
func (p PackageInstallOutput) TableOutput() string {
	tb := NewTableBuilder()

	tb.AddTitle("Package Installation")
	tb.AddNewline()

	if p.Summary.Added > 0 {
		tb.AddLine("%s Added %d packages", IconPackage, p.Summary.Added)
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
func (p PackageInstallOutput) StructuredData() any {
	return p
}
