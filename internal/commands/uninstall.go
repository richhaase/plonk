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

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <packages...>",
	Short: "Uninstall packages and remove from plonk management",
	Long: `Uninstall packages from your system and remove them from your lock file.

This command uninstalls packages from your system using the appropriate package
manager and removes them from plonk management. Use prefix syntax to specify
the package manager, or omit the prefix to use your default manager.

Examples:
  plonk uninstall htop                    # Uninstall htop using default manager
  plonk uninstall git neovim              # Uninstall multiple packages
  plonk uninstall brew:git                # Uninstall git specifically with Homebrew
  plonk uninstall npm:lodash              # Uninstall lodash from npm global packages
  plonk uninstall --dry-run htop          # Preview what would be uninstalled`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

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

	// Get flags (only common flags now)
	dryRun, _ := cmd.Flags().GetBool("dry-run")

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

		// Configure uninstallation options for this package
		opts := packages.UninstallOptions{
			Manager: manager,
			DryRun:  dryRun,
		}

		// Process this package
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		results, err := packages.UninstallPackages(ctx, configDir, []string{packageName}, opts)
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
	summary := calculateUninstallSummary(allResults)
	outputData := PackageUninstallOutput{
		TotalPackages: len(allResults),
		Results:       allResults,
		Summary:       summary,
	}

	// Render output
	if err := RenderOutput(outputData, format); err != nil {
		return err
	}

	// Determine exit code based on results
	exitErr := DetermineExitCode(allResults, "packages", "uninstall")
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
