// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <packages...>",
	Short: "Install packages to plonk management",
	Long: `Install packages and add them to your lock file for management.

This command adds packages to your lock file so they can be managed by plonk.
Use specific manager flags to control which package manager to use.

Examples:
  plonk install htop                      # Install htop using default manager
  plonk install git neovim ripgrep        # Install multiple packages
  plonk install git --brew                # Install git specifically with Homebrew
  plonk install lodash --npm              # Install lodash with npm global packages
  plonk install ripgrep --cargo           # Install ripgrep with cargo packages
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
	installCmd.MarkFlagsMutuallyExclusive("brew", "npm", "cargo")

	// Common flags
	installCmd.Flags().BoolP("dry-run", "n", false, "Show what would be installed without making changes")
	installCmd.Flags().BoolP("force", "f", false, "Force installation even if already managed")
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Parse flags
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "install", "flags", "invalid flag combination")
	}

	// Parse output format
	format, err := ParseOutputFormat(flags.Output)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "install", "output-format", "invalid output format")
	}

	// Process packages
	results, err := installPackages(cmd, args, flags)
	if err != nil {
		return err
	}

	// Show progress and summary
	reporter := operations.NewProgressReporter("package", format == OutputTable)
	for _, result := range results {
		reporter.ShowItemProgress(result)
	}

	// Handle output based on format
	if format == OutputTable {
		reporter.ShowBatchSummary(results)
	} else {
		return renderPackageResults(results, format)
	}

	// Determine exit code
	return operations.DetermineExitCode(results, errors.DomainCommands, "install")
}

// installPackages handles package installations
func installPackages(cmd *cobra.Command, packageNames []string, flags *SimpleFlags) ([]operations.OperationResult, error) {
	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)

	// Get manager - default to configured default or homebrew
	manager := flags.Manager
	if manager == "" {
		cfg, err := config.LoadConfig(configDir)
		if err == nil && cfg.DefaultManager != nil && *cfg.DefaultManager != "" {
			manager = *cfg.DefaultManager
		} else {
			manager = "homebrew" // fallback default
		}
	}

	// Process packages sequentially
	results := make([]operations.OperationResult, 0, len(packageNames))

	for _, packageName := range packageNames {
		result := installSinglePackage(configDir, lockService, packageName, manager, flags.DryRun, flags.Force)
		results = append(results, result)
	}

	return results, nil
}

// installSinglePackage installs a single package
func installSinglePackage(configDir string, lockService *lock.YAMLLockService, packageName, manager string, dryRun, force bool) operations.OperationResult {
	result := operations.OperationResult{
		Name:    packageName,
		Manager: manager,
	}

	// Check if already managed
	if lockService.HasPackage(manager, packageName) {
		if !force {
			result.Status = "skipped"
			result.AlreadyManaged = true
			return result
		}
	}

	if dryRun {
		result.Status = "would-add"
		return result
	}

	// Get package manager instance
	pkgManager, err := getPackageManager(manager)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrManagerUnavailable, errors.DomainPackages, "install", packageName, "failed to get package manager")
		return result
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check if manager is available
	available, err := pkgManager.IsAvailable(ctx)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrManagerUnavailable, errors.DomainPackages, "install", packageName, "failed to check manager availability")
		return result
	}
	if !available {
		result.Status = "failed"
		result.Error = errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "install", fmt.Sprintf("package manager '%s' is not available", manager))
		return result
	}

	// Get package version if installed
	version, err := pkgManager.GetInstalledVersion(ctx, packageName)
	if err == nil && version != "" {
		result.Version = version
	}

	// Add to lock file
	err = lockService.AddPackage(manager, packageName, version)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainPackages, "install", packageName, "failed to add package to lock file")
		return result
	}

	result.Status = "added"
	return result
}

// getPackageManager returns the appropriate package manager instance
func getPackageManager(manager string) (managers.PackageManager, error) {
	switch manager {
	case "homebrew":
		return managers.NewHomebrewManager(), nil
	case "npm":
		return managers.NewNpmManager(), nil
	case "cargo":
		return managers.NewCargoManager(), nil
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", manager)
	}
}

// renderPackageResults renders package results in structured format
func renderPackageResults(results []operations.OperationResult, format OutputFormat) error {
	output := PackageInstallOutput{
		TotalPackages: len(results),
		Results:       results,
		Summary:       calculatePackageSummary(results),
	}
	return RenderOutput(output, format)
}

// PackageInstallOutput represents the output for package installation
type PackageInstallOutput struct {
	TotalPackages int                          `json:"total_packages" yaml:"total_packages"`
	Results       []operations.OperationResult `json:"results" yaml:"results"`
	Summary       PackageInstallSummary        `json:"summary" yaml:"summary"`
}

// PackageInstallSummary provides summary for package installation
type PackageInstallSummary struct {
	Added   int `json:"added" yaml:"added"`
	Skipped int `json:"skipped" yaml:"skipped"`
	Failed  int `json:"failed" yaml:"failed"`
}

// calculatePackageSummary calculates summary from results
func calculatePackageSummary(results []operations.OperationResult) PackageInstallSummary {
	summary := PackageInstallSummary{}
	for _, result := range results {
		switch result.Status {
		case "added", "would-add":
			summary.Added++
		case "skipped":
			summary.Skipped++
		case "failed":
			summary.Failed++
		}
	}
	return summary
}

// TableOutput generates human-friendly output
func (p PackageInstallOutput) TableOutput() string {
	output := "Package Installation\n===================\n\n"

	if p.Summary.Added > 0 {
		output += fmt.Sprintf("📦 Added %d packages\n", p.Summary.Added)
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
func (p PackageInstallOutput) StructuredData() any {
	return p
}
