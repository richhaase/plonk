// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <packages...>",
	Short: "Install packages and add them to plonk management",
	Long: `Install packages on your system and add them to your lock file for management.

This command installs packages using the specified package manager and adds them
to your lock file so they can be managed by plonk. Use prefix syntax to specify
the package manager, or omit the prefix to use your default manager.

Examples:
  plonk install htop                      # Install htop using default manager
  plonk install git neovim ripgrep        # Install multiple packages
  plonk install brew:git                  # Install git specifically with Homebrew
  plonk install npm:lodash                # Install lodash with npm global packages
  plonk install cargo:ripgrep             # Install ripgrep with cargo packages
  plonk install pip:black pip:flake8      # Install Python tools with pip
  plonk install gem:bundler gem:rubocop   # Install Ruby tools with gem
  plonk install go:golang.org/x/tools/cmd/gopls  # Install Go tools with go install
  plonk install --dry-run htop neovim     # Preview what would be installed`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Common flags
	installCmd.Flags().BoolP("dry-run", "n", false, "Show what would be installed without making changes")
	installCmd.Flags().BoolP("force", "f", false, "Force installation even if already managed")
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get flags (only common flags now)
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")

	// Get directories and config
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)

	// Process each package with prefix parsing
	var allResults []resources.OperationResult

	for _, packageSpec := range args {
		// Parse the package specification
		manager, packageName := ParsePackageSpec(packageSpec)

		// Validate package specification
		if packageName == "" {
			allResults = append(allResults, resources.OperationResult{
				Name:   packageSpec,
				Status: "failed",
				Error:  fmt.Errorf("invalid package specification %q: empty package name", packageSpec),
			})
			continue
		}

		if manager == "" && packageSpec != packageName {
			// This means there was a colon but empty prefix
			allResults = append(allResults, resources.OperationResult{
				Name:   packageSpec,
				Status: "failed",
				Error:  fmt.Errorf("invalid package specification %q: empty manager prefix", packageSpec),
			})
			continue
		}

		// Use default manager if no prefix specified
		if manager == "" {
			if cfg.DefaultManager != "" {
				manager = cfg.DefaultManager
			} else {
				manager = packages.DefaultManager
			}
		}

		// Validate manager
		if !IsValidManager(manager) {
			allResults = append(allResults, resources.OperationResult{
				Name:    packageSpec,
				Manager: manager,
				Status:  "failed",
				Error:   fmt.Errorf("unknown package manager %q. Valid managers: %s", manager, strings.Join(GetValidManagers(), ", ")),
			})
			continue
		}

		// Configure installation options for this package
		opts := packages.InstallOptions{
			Manager: manager,
			DryRun:  dryRun,
			Force:   force,
		}

		// Process this package
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		results, err := packages.InstallPackages(ctx, configDir, []string{packageName}, opts)
		cancel()

		if err != nil {
			allResults = append(allResults, resources.OperationResult{
				Name:    packageSpec,
				Manager: manager,
				Status:  "failed",
				Error:   err,
			})
			continue
		}

		allResults = append(allResults, results...)
	}

	// Show progress for each result
	for _, result := range allResults {
		icon := GetStatusIcon(result.Status)
		fmt.Printf("%s %s %s\n", icon, result.Status, result.Name)
	}

	// Create output data
	summary := calculatePackageSummary(allResults)
	outputData := PackageInstallOutput{
		TotalPackages: len(allResults),
		Results:       allResults,
		Summary:       summary,
	}

	// Render output
	if err := RenderOutput(outputData, format); err != nil {
		return err
	}

	// Check if all operations failed and return appropriate error
	return resources.ValidateOperationResults(allResults, "install packages")
}

// PackageInstallOutput represents the output for package installation
type PackageInstallOutput struct {
	TotalPackages int                         `json:"total_packages" yaml:"total_packages"`
	Results       []resources.OperationResult `json:"results" yaml:"results"`
	Summary       PackageInstallSummary       `json:"summary" yaml:"summary"`
}

// PackageInstallSummary provides summary for package installation
type PackageInstallSummary struct {
	Added   int `json:"added" yaml:"added"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculatePackageSummary calculates summary from results using generic operations summary
func calculatePackageSummary(results []resources.OperationResult) PackageInstallSummary {
	genericSummary := resources.CalculateSummary(results)
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
