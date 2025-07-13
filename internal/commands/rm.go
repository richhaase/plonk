// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <items...>",
	Short: "Remove packages or dotfiles",
	Long: `Intelligently remove packages or dotfiles based on argument format.

Packages (detected automatically):
  plonk rm htop                         # Remove from lock file only
  plonk rm htop --uninstall             # Remove from lock file and uninstall
  plonk rm git neovim --uninstall       # Remove multiple packages

Dotfiles (detected automatically):
  plonk rm ~/.zshrc                     # Unlink single dotfile
  plonk rm ~/.zshrc ~/.vimrc            # Unlink multiple dotfiles
  plonk rm ~/.config/nvim/init.lua      # Unlink specific file

Mixed operations:
  plonk rm git ~/.vimrc                 # Remove package and unlink dotfile
  plonk rm --dry-run git ~/.zshrc       # Preview mixed removals

Force type interpretation:
  plonk rm config --package             # Force 'config' to be treated as package
  plonk rm config --dotfile             # Force 'config' to be treated as dotfile

Note: Package removal only removes from configuration by default.
Use --uninstall to also remove the package from your system.
Dotfile removal unlinks the file (removes the symlink, keeps source).`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRm,
}

func init() {
	rootCmd.AddCommand(rmCmd)

	// Manager-specific flags (mutually exclusive)
	rmCmd.Flags().Bool("brew", false, "Use Homebrew package manager")
	rmCmd.Flags().Bool("npm", false, "Use NPM package manager")
	rmCmd.Flags().Bool("cargo", false, "Use Cargo package manager")
	rmCmd.MarkFlagsMutuallyExclusive("brew", "npm", "cargo")

	// Type override flags (mutually exclusive)
	rmCmd.Flags().Bool("package", false, "Force all items to be treated as packages")
	rmCmd.Flags().Bool("dotfile", false, "Force all items to be treated as dotfiles")
	rmCmd.MarkFlagsMutuallyExclusive("package", "dotfile")

	// Common flags
	rmCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")
	rmCmd.Flags().Bool("uninstall", false, "Also uninstall packages from the system (packages only)")
	rmCmd.Flags().BoolP("force", "f", false, "Force removal even if not managed")
}

func runRm(cmd *cobra.Command, args []string) error {
	// Parse flags
	flags, err := ParseUnifiedFlags(cmd)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "rm", "flags", "invalid flag combination")
	}

	// Get additional flags specific to rm
	uninstallFlag, _ := cmd.Flags().GetBool("uninstall")

	// Intelligently separate packages and dotfiles
	packages, dotfiles, ambiguous := ProcessMixedItemsWithFlags(args, flags)

	// Handle ambiguous items
	if len(ambiguous) > 0 {
		return handleAmbiguousRemovalItems(cmd, ambiguous, flags)
	}

	return processMixedRemovals(cmd, packages, dotfiles, flags, uninstallFlag)
}

// processMixedRemovals handles both packages and dotfiles in a single removal operation
func processMixedRemovals(cmd *cobra.Command, packages []string, dotfiles []string, flags *CommandFlags, uninstallFlag bool) error {
	// Parse output format
	format, err := ParseOutputFormat(flags.Output)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "rm", "output-format", "invalid output format")
	}

	var allResults []operations.OperationResult
	reporter := operations.NewProgressReporter("remove", format == OutputTable)

	// Process packages if any
	if len(packages) > 0 {
		pkgResults, err := removePackages(cmd, packages, flags, uninstallFlag)
		if err != nil && !isPartialFailure(err) {
			return err
		}
		allResults = append(allResults, pkgResults...)

		// Show progress immediately for packages
		for _, result := range pkgResults {
			reporter.ShowItemProgress(result)
		}
	}

	// Process dotfiles if any
	if len(dotfiles) > 0 {
		dotResults, err := removeDotfiles(cmd, dotfiles, flags)
		if err != nil && !isPartialFailure(err) {
			return err
		}
		allResults = append(allResults, dotResults...)

		// Show progress immediately for dotfiles
		for _, result := range dotResults {
			reporter.ShowItemProgress(result)
		}
	}

	// Handle output based on format
	if format == OutputTable {
		// Show summary for table output
		reporter.ShowBatchSummary(allResults)
	} else {
		// For structured output, create mixed response
		return handleMixedRemovalResults(allResults, format)
	}

	// Determine exit code
	return operations.DetermineExitCode(allResults, errors.DomainCommands, "rm-mixed")
}

// removePackages handles package removals using existing logic
func removePackages(cmd *cobra.Command, packageNames []string, flags *CommandFlags, uninstallFlag bool) ([]operations.OperationResult, error) {
	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)

	// Process packages sequentially
	results := make([]operations.OperationResult, 0, len(packageNames))

	for _, packageName := range packageNames {
		result := removeSinglePackage(configDir, lockService, packageName, flags.DryRun, uninstallFlag)
		results = append(results, result)
	}

	return results, nil
}

// removeDotfiles handles dotfile removals (unlinking)
func removeDotfiles(cmd *cobra.Command, dotfilePaths []string, flags *CommandFlags) ([]operations.OperationResult, error) {
	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "rm", "failed to get home directory")
	}

	configDir := config.GetDefaultConfigDirectory()

	// Load config
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// If config doesn't exist, we can't unlink dotfiles
		return nil, errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Process dotfiles sequentially
	results := make([]operations.OperationResult, 0, len(dotfilePaths))

	for _, dotfilePath := range dotfilePaths {
		result := removeSingleDotfile(homeDir, configDir, cfg, dotfilePath, flags.DryRun)
		results = append(results, result)
	}

	return results, nil
}

// removeSinglePackage removes a single package using existing logic
func removeSinglePackage(configDir string, lockService *lock.YAMLLockService, packageName string, dryRun bool, uninstall bool) operations.OperationResult {
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
		err := uninstallPackage(managerName, packageName)
		if err != nil {
			result.Status = "partially-removed"
			result.Error = errors.WrapWithItem(err, errors.ErrPackageUninstall, errors.DomainPackages, "uninstall", packageName, "removed from config but failed to uninstall")
			return result
		}
	}

	result.Status = "removed"
	return result
}

// removeSingleDotfile removes (unlinks) a single dotfile
func removeSingleDotfile(homeDir, configDir string, cfg *config.Config, dotfilePath string, dryRun bool) operations.OperationResult {
	result := operations.OperationResult{
		Name: dotfilePath,
	}

	// Resolve dotfile path
	resolvedPath, err := resolveDotfilePath(dotfilePath, homeDir)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainDotfiles, "resolve", dotfilePath, "failed to resolve dotfile path")
		return result
	}

	// Check if file is managed (has a symlink)
	if !isSymlink(resolvedPath) {
		result.Status = "skipped"
		result.Error = errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "check", fmt.Sprintf("dotfile '%s' is not a managed symlink", dotfilePath))
		return result
	}

	if dryRun {
		result.Status = "would-unlink"
		return result
	}

	// Remove the symlink
	err = os.Remove(resolvedPath)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "unlink", dotfilePath, "failed to remove symlink")
		return result
	}

	result.Status = "unlinked"
	result.Metadata = map[string]interface{}{
		"source":      dotfilePath,
		"destination": resolvedPath,
	}
	return result
}

// Helper functions

// findPackageInLockFile finds which manager manages a package (reuse existing logic)
func findPackageInLockFile(lockService *lock.YAMLLockService, packageName string) (string, bool) {
	managers := []string{"homebrew", "npm", "cargo"}

	for _, manager := range managers {
		if lockService.HasPackage(manager, packageName) {
			return manager, true
		}
	}

	return "", false
}

// uninstallPackage uninstalls a package using the appropriate manager
func uninstallPackage(managerName, packageName string) error {
	var mgr managers.PackageManager

	switch managerName {
	case "homebrew":
		mgr = managers.NewHomebrewManager()
	case "npm":
		mgr = managers.NewNpmManager()
	case "cargo":
		mgr = managers.NewCargoManager()
	default:
		return fmt.Errorf("unsupported package manager: %s", managerName)
	}

	ctx := context.Background()

	// Check if manager is available
	available, err := mgr.IsAvailable(ctx)
	if err != nil {
		return fmt.Errorf("failed to check manager availability: %w", err)
	}
	if !available {
		return fmt.Errorf("manager '%s' is not available", managerName)
	}

	// Uninstall the package
	return mgr.Uninstall(ctx, packageName)
}

// isSymlink checks if a path is a symbolic link
func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// handleAmbiguousRemovalItems handles items that couldn't be automatically categorized
func handleAmbiguousRemovalItems(cmd *cobra.Command, ambiguous []string, flags *CommandFlags) error {
	// For now, default ambiguous items to packages with a warning
	format, _ := ParseOutputFormat(flags.Output)

	if format == OutputTable {
		fmt.Printf("Warning: Ambiguous items detected, treating as packages: %v\n", ambiguous)
		fmt.Printf("Use --package or --dotfile flags to force interpretation\n\n")
	}

	// Process as packages
	uninstallFlag, _ := cmd.Flags().GetBool("uninstall")
	return processMixedRemovals(cmd, ambiguous, []string{}, flags, uninstallFlag)
}

// handleMixedRemovalResults handles structured output for mixed removal operations
func handleMixedRemovalResults(results []operations.OperationResult, format OutputFormat) error {
	// Create a mixed removal summary
	packageResults := make([]operations.OperationResult, 0)
	dotfileResults := make([]operations.OperationResult, 0)

	for _, result := range results {
		if result.Manager != "" {
			// Has manager = package
			packageResults = append(packageResults, result)
		} else {
			// No manager = dotfile
			dotfileResults = append(dotfileResults, result)
		}
	}

	mixedOutput := MixedRemoveOutput{
		TotalItems:     len(results),
		PackageResults: packageResults,
		DotfileResults: dotfileResults,
		Summary: MixedRemovalSummary{
			PackagesRemoved:  CountByStatus(packageResults, "removed"),
			DotfilesUnlinked: CountByStatus(dotfileResults, "unlinked"),
			Failed:           CountByStatus(results, "failed"),
			Skipped:          CountByStatus(results, "skipped"),
		},
	}

	return RenderOutput(mixedOutput, format)
}

// MixedRemoveOutput represents output for mixed remove operations
type MixedRemoveOutput struct {
	TotalItems     int                          `json:"total_items" yaml:"total_items"`
	PackageResults []operations.OperationResult `json:"packages" yaml:"packages"`
	DotfileResults []operations.OperationResult `json:"dotfiles" yaml:"dotfiles"`
	Summary        MixedRemovalSummary          `json:"summary" yaml:"summary"`
}

// MixedRemovalSummary provides summary for mixed removal operations
type MixedRemovalSummary struct {
	PackagesRemoved  int `json:"packages_removed" yaml:"packages_removed"`
	DotfilesUnlinked int `json:"dotfiles_unlinked" yaml:"dotfiles_unlinked"`
	Failed           int `json:"failed" yaml:"failed"`
	Skipped          int `json:"skipped" yaml:"skipped"`
}

// TableOutput generates human-friendly table output for mixed remove
func (m MixedRemoveOutput) TableOutput() string {
	output := "Mixed Remove Operation\n====================\n\n"

	if m.Summary.PackagesRemoved > 0 {
		output += fmt.Sprintf("📦 Removed %d packages\n", m.Summary.PackagesRemoved)
	}
	if m.Summary.DotfilesUnlinked > 0 {
		output += fmt.Sprintf("📄 Unlinked %d dotfiles\n", m.Summary.DotfilesUnlinked)
	}
	if m.Summary.Failed > 0 {
		output += fmt.Sprintf("❌ %d failed\n", m.Summary.Failed)
	}
	if m.Summary.Skipped > 0 {
		output += fmt.Sprintf("⏭️ %d skipped\n", m.Summary.Skipped)
	}

	output += fmt.Sprintf("\nTotal: %d items processed\n", m.TotalItems)
	return output
}

// StructuredData returns the structured data for serialization
func (m MixedRemoveOutput) StructuredData() any {
	return m
}
