// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <packages...>",
	Short: "Remove packages from plonk management",
	Long: `Remove packages from your lock file and optionally uninstall them from your system.

This command removes packages from plonk management. By default, it only removes
them from the lock file. Use --uninstall to also remove them from your system.

Examples:
  plonk uninstall htop                    # Remove from lock file only
  plonk uninstall htop --uninstall        # Remove from lock file and uninstall
  plonk uninstall git neovim --uninstall  # Remove multiple packages and uninstall
  plonk uninstall --dry-run htop          # Preview what would be removed`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	// Manager-specific flags (mutually exclusive)
	uninstallCmd.Flags().Bool("brew", false, "Use Homebrew package manager")
	uninstallCmd.Flags().Bool("npm", false, "Use NPM package manager")
	uninstallCmd.Flags().Bool("cargo", false, "Use Cargo package manager")
	uninstallCmd.MarkFlagsMutuallyExclusive("brew", "npm", "cargo")

	// Common flags
	uninstallCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")
	uninstallCmd.Flags().Bool("uninstall", false, "Also uninstall packages from the system")
	uninstallCmd.Flags().BoolP("force", "f", false, "Force removal even if not managed")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	// Parse flags
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "uninstall", "flags", "invalid flag combination")
	}

	// Get additional flags specific to uninstall
	uninstallFlag, _ := cmd.Flags().GetBool("uninstall")

	// Parse output format
	format, err := ParseOutputFormat(flags.Output)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "uninstall", "output-format", "invalid output format")
	}

	// Process packages
	results, err := uninstallPackages(cmd, args, flags, uninstallFlag)
	if err != nil {
		return err
	}

	// Show progress and summary
	reporter := operations.NewProgressReporterForOperation("remove", "package", format == OutputTable)
	for _, result := range results {
		reporter.ShowItemProgress(result)
	}

	// Handle output based on format
	if format == OutputTable {
		reporter.ShowBatchSummary(results)
	} else {
		return renderUninstallResults(results, format)
	}

	// Determine exit code
	return operations.DetermineExitCode(results, errors.DomainCommands, "uninstall")
}

// uninstallPackages handles package uninstallations
func uninstallPackages(cmd *cobra.Command, packageNames []string, flags *SimpleFlags, uninstallFlag bool) ([]operations.OperationResult, error) {
	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)

	// Process packages sequentially
	results := make([]operations.OperationResult, 0, len(packageNames))

	for _, packageName := range packageNames {
		result := uninstallSinglePackage(configDir, lockService, packageName, flags.DryRun, uninstallFlag)
		results = append(results, result)
	}

	return results, nil
}

// uninstallSinglePackage removes a single package
func uninstallSinglePackage(configDir string, lockService *lock.YAMLLockService, packageName string, dryRun bool, uninstall bool) operations.OperationResult {
	result := operations.OperationResult{
		Name: packageName,
	}

	// Find package in lock file
	managerName, found := findPackageInLockFile(lockService, packageName)
	result.Manager = managerName

	if !found {
		result.Status = "skipped"
		result.Error = errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "find", fmt.Sprintf("package '%s' not found in lock file", packageName))
		return result
	}

	if dryRun {
		result.Status = "would-remove"
		return result
	}

	// Remove from lock file
	err := lockService.RemovePackage(managerName, packageName)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainPackages, "remove-lock", packageName, "failed to remove package from lock file")
		return result
	}

	// Uninstall if requested
	if uninstall {
		err := uninstallPackageFromSystem(managerName, packageName)
		if err != nil {
			result.Status = "partially-removed"
			result.Error = errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", packageName, "removed from config but failed to uninstall")
			return result
		}
	}

	result.Status = "removed"
	return result
}

// findPackageInLockFile finds which manager manages a package
func findPackageInLockFile(lockService *lock.YAMLLockService, packageName string) (string, bool) {
	managers := []string{"homebrew", "npm", "cargo"}

	for _, manager := range managers {
		if lockService.HasPackage(manager, packageName) {
			return manager, true
		}
	}

	return "", false
}

// uninstallPackageFromSystem uninstalls a package using the appropriate manager
func uninstallPackageFromSystem(managerName, packageName string) error {
	registry := managers.NewManagerRegistry()
	mgr, err := registry.GetManager(managerName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Check if manager is available
	available, err := mgr.IsAvailable(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "uninstall",
			"failed to check manager availability")
	}
	if !available {
		return errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "uninstall",
			"manager '"+managerName+"' is not available")
	}

	// Uninstall the package
	return mgr.Uninstall(ctx, packageName)
}

// renderUninstallResults renders uninstall results in structured format
func renderUninstallResults(results []operations.OperationResult, format OutputFormat) error {
	output := PackageUninstallOutput{
		TotalPackages: len(results),
		Results:       results,
		Summary:       calculateUninstallSummary(results),
	}
	return RenderOutput(output, format)
}

// PackageUninstallOutput represents the output for package uninstallation
type PackageUninstallOutput struct {
	TotalPackages int                          `json:"total_packages" yaml:"total_packages"`
	Results       []operations.OperationResult `json:"results" yaml:"results"`
	Summary       PackageUninstallSummary      `json:"summary" yaml:"summary"`
}

// PackageUninstallSummary provides summary for package uninstallation
type PackageUninstallSummary struct {
	Removed int `json:"removed" yaml:"removed"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculateUninstallSummary calculates summary from results
func calculateUninstallSummary(results []operations.OperationResult) PackageUninstallSummary {
	summary := PackageUninstallSummary{}
	for _, result := range results {
		switch result.Status {
		case "removed", "would-remove":
			summary.Removed++
		case "skipped":
			summary.Skipped++
		case "failed":
			summary.Failed++
		}
	}
	return summary
}

// TableOutput generates human-friendly output
func (p PackageUninstallOutput) TableOutput() string {
	output := "Package Uninstallation\n=====================\n\n"

	if p.Summary.Removed > 0 {
		output += fmt.Sprintf("📦 Removed %d packages\n", p.Summary.Removed)
	}
	if p.Summary.Skipped > 0 {
		output += fmt.Sprintf("⏭️ %d skipped\n", p.Summary.Skipped)
	}
	if p.Summary.Failed > 0 {
		output += fmt.Sprintf("❌ %d failed\n", p.Summary.Failed)
	}

	output += fmt.Sprintf("\nTotal: %d packages processed\n", p.TotalPackages)
	return output
}

// StructuredData returns the structured data for serialization
func (p PackageUninstallOutput) StructuredData() any {
	return p
}
