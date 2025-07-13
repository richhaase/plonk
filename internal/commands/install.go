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
	// Create command pipeline
	pipeline, err := NewCommandPipeline(cmd, "package")
	if err != nil {
		return err
	}

	// Define the processor function
	processor := func(ctx context.Context, args []string, flags *SimpleFlags) ([]operations.OperationResult, error) {
		return installPackages(cmd, args, flags)
	}

	// Execute the pipeline
	return pipeline.ExecuteWithResults(context.Background(), processor, args)
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
		cfg := config.LoadConfigWithDefaults(configDir)
		if cfg.DefaultManager != nil && *cfg.DefaultManager != "" {
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
	registry := managers.NewManagerRegistry()
	return registry.GetManager(manager)
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
