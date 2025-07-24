// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/state"
	"github.com/richhaase/plonk/internal/ui"
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

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)

	// Get manager - default to configured default or homebrew
	manager := flags.Manager
	if manager == "" {
		cfg := config.LoadConfigWithDefaults(orchestrator.GetConfigDir())
		if cfg.DefaultManager != "" {
			manager = cfg.DefaultManager
		} else {
			manager = managers.DefaultManager // fallback default
		}
	}

	// Process each package directly
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var results []state.OperationResult

	// Show header for progress tracking
	reporter := ui.NewProgressReporterForOperation("install", "package", true)

	for _, packageName := range args {
		// Check if context was canceled
		if ctx.Err() != nil {
			break
		}

		// Install single package directly
		result := installSinglePackage(configDir, lockService, packageName, manager, flags.DryRun, flags.Force)

		// Show individual progress
		reporter.ShowItemProgress(result)

		// Collect result
		results = append(results, result)
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

// installSinglePackage installs a single package
func installSinglePackage(configDir string, lockService *lock.YAMLLockService, packageName, manager string, dryRun, force bool) state.OperationResult {
	result := state.OperationResult{
		Name:    packageName,
		Manager: manager,
	}

	// For Go packages, we need to check with the binary name
	checkPackageName := packageName
	if manager == "go" {
		checkPackageName = managers.ExtractBinaryNameFromPath(packageName)
	}

	// Check if already managed
	if lockService.HasPackage(manager, checkPackageName) {
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
		result.Error = fmt.Errorf("install %s: failed to get package manager: %w", packageName, err)
		return result
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check if manager is available
	available, err := pkgManager.IsAvailable(ctx)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("install %s: failed to check %s availability: %w", packageName, manager, err)
		return result
	}
	if !available {
		result.Status = "failed"
		result.Error = fmt.Errorf("install %s: %s manager not available (%s)", packageName, manager, getManagerInstallSuggestion(manager))
		return result
	}

	// Install the package
	err = pkgManager.Install(ctx, packageName)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("install %s via %s: %w", packageName, manager, err)
		return result
	}

	// For Go packages, we need to determine the actual binary name
	lockPackageName := packageName
	if manager == "go" {
		// Extract binary name from module path
		lockPackageName = managers.ExtractBinaryNameFromPath(packageName)
	}

	// Get package version after installation
	version, err := pkgManager.GetInstalledVersion(ctx, lockPackageName)
	if err == nil && version != "" {
		result.Version = version
	}

	// Add to lock file
	err = lockService.AddPackage(manager, lockPackageName, version)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("install %s: failed to add to lock file (manager: %s, version: %s): %w", packageName, manager, version, err)
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
