// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/runtime"
	"github.com/richhaase/plonk/internal/state"
	"github.com/richhaase/plonk/internal/ui"
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
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "uninstall", "output-format", "invalid output format")
	}

	// Get flags
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return err
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)

	// Process each package directly
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var results []state.OperationResult

	// Show header for progress tracking
	reporter := ui.NewProgressReporterForOperation("uninstall", "package", true)

	for _, packageName := range args {
		// Check if context was canceled
		if ctx.Err() != nil {
			break
		}

		// Uninstall single package directly
		result := uninstallSinglePackage(configDir, lockService, packageName, flags.DryRun, flags.Manager)

		// Show individual progress
		reporter.ShowItemProgress(result)

		// Collect result
		results = append(results, result)
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
	exitErr := DetermineExitCode(results, errors.DomainPackages, "uninstall")
	if exitErr != nil {
		return exitErr
	}

	return nil
}

// uninstallSinglePackage removes a single package
func uninstallSinglePackage(configDir string, lockService *lock.YAMLLockService, packageName string, dryRun bool, managerFlag string) state.OperationResult {
	result := state.OperationResult{
		Name: packageName,
	}

	// For Go packages, we need to check with the binary name
	checkPackageName := packageName
	if managerFlag == "go" {
		checkPackageName = managers.ExtractBinaryNameFromPath(packageName)
	}

	// Find package in lock file
	managerName, found := findPackageInLockFile(lockService, checkPackageName)
	wasManaged := found

	// If we found it and it's a go package, we might need to check with binary name
	if !found && managerFlag == "" && strings.Contains(packageName, "/") {
		// This might be a Go module path, try with binary name
		checkPackageName = managers.ExtractBinaryNameFromPath(packageName)
		managerName, found = findPackageInLockFile(lockService, checkPackageName)
		wasManaged = found
	}

	// If not in lock file, we need to detect which manager to use
	if !found {
		// If manager flag is provided, use it
		if managerFlag != "" {
			managerName = managerFlag
		} else {
			// Try to detect which manager has the package installed
			detectedManager, err := detectInstalledPackageManager(packageName)
			if err != nil {
				result.Status = "skipped"
				result.Error = errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "detect", fmt.Sprintf("package '%s' not found in any package manager", packageName))
				return result
			}
			managerName = detectedManager
		}
	}

	result.Manager = managerName

	if dryRun {
		result.Status = "would-remove"
		return result
	}

	// Attempt to uninstall from system
	err := uninstallPackageFromSystem(managerName, packageName)
	if err != nil {
		// If package wasn't installed but was in lock file, we should still remove it from lock
		if wasManaged {
			lockErr := lockService.RemovePackage(managerName, checkPackageName)
			if lockErr == nil {
				result.Status = "removed"
				result.Error = errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", packageName, "package not installed, removed from lock file")
				return result
			}
		}
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", packageName, "failed to uninstall package").WithMetadata("manager", managerName)
		return result
	}

	// Remove from lock file if it was managed
	if wasManaged {
		err = lockService.RemovePackage(managerName, checkPackageName)
		if err != nil {
			result.Status = "partially-removed"
			result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainPackages, "remove-lock", packageName, "uninstalled but failed to remove from lock file").WithMetadata("manager", managerName)
			return result
		}
	}

	result.Status = "removed"
	return result
}

// findPackageInLockFile finds which manager manages a package
func findPackageInLockFile(lockService *lock.YAMLLockService, packageName string) (string, bool) {
	for _, manager := range managers.SupportedManagers {
		if lockService.HasPackage(manager, packageName) {
			return manager, true
		}
	}

	return "", false
}

// detectInstalledPackageManager tries to detect which package manager has the package installed
func detectInstalledPackageManager(packageName string) (string, error) {
	sharedCtx := runtime.GetSharedContext()
	registry := sharedCtx.ManagerRegistry()
	ctx := context.Background()

	// Try each manager to see if package is installed
	for _, managerName := range managers.SupportedManagers {
		mgr, err := registry.GetManager(managerName)
		if err != nil {
			continue
		}

		// Check if manager is available
		available, err := mgr.IsAvailable(ctx)
		if err != nil || !available {
			continue
		}

		// Check if package is installed
		installed, err := mgr.IsInstalled(ctx, packageName)
		if err != nil {
			continue
		}

		if installed {
			return managerName, nil
		}
	}

	return "", errors.NewError(errors.ErrPackageNotFound, errors.DomainPackages, "detect", "package not found in any available package manager")
}

// uninstallPackageFromSystem uninstalls a package using the appropriate manager
func uninstallPackageFromSystem(managerName, packageName string) error {
	sharedCtx := runtime.GetSharedContext()
	registry := sharedCtx.ManagerRegistry()
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
			"manager '"+managerName+"' is not available").WithSuggestionMessage(getManagerInstallSuggestion(managerName))
	}

	// Uninstall the package
	return mgr.Uninstall(ctx, packageName)
}

// PackageUninstallOutput represents the output for package uninstallation
type PackageUninstallOutput struct {
	TotalPackages int                     `json:"total_packages" yaml:"total_packages"`
	Results       []state.OperationResult `json:"results" yaml:"results"`
	Summary       PackageUninstallSummary `json:"summary" yaml:"summary"`
}

// PackageUninstallSummary provides summary for package uninstallation
type PackageUninstallSummary struct {
	Removed int `json:"removed" yaml:"removed"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculateUninstallSummary calculates summary from uninstall results using generic operations summary
func calculateUninstallSummary(results []state.OperationResult) PackageUninstallSummary {
	genericSummary := state.CalculateSummary(results)
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
