// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/richhaase/plonk/internal/state"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <items...>",
	Short: "Add packages or dotfiles to plonk management",
	Long: `Intelligently add packages or dotfiles based on argument format.

Packages (detected automatically):
  plonk add htop                        # Add htop using default manager
  plonk add git neovim ripgrep          # Add multiple packages
  plonk add git --brew                  # Add git specifically to homebrew
  plonk add lodash --npm                # Add lodash to npm global packages
  plonk add ripgrep --cargo             # Add ripgrep to cargo packages

Dotfiles (detected automatically):
  plonk add ~/.zshrc                    # Add single dotfile
  plonk add ~/.zshrc ~/.vimrc           # Add multiple dotfiles
  plonk add ~/.config/nvim/             # Add directory of dotfiles

Mixed operations:
  plonk add git ~/.vimrc                # Add package and dotfile together
  plonk add --dry-run git neovim ~/.zshrc # Preview mixed additions

Force type interpretation:
  plonk add config --package            # Force 'config' to be treated as package
  plonk add config --dotfile            # Force 'config' to be treated as dotfile

Without arguments:
  plonk add                             # Add all untracked packages and dotfiles
  plonk add --dry-run                   # Preview all untracked items`,
	Args: cobra.ArbitraryArgs,
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Manager-specific flags (mutually exclusive)
	addCmd.Flags().Bool("brew", false, "Use Homebrew package manager")
	addCmd.Flags().Bool("npm", false, "Use NPM package manager")
	addCmd.Flags().Bool("cargo", false, "Use Cargo package manager")
	addCmd.MarkFlagsMutuallyExclusive("brew", "npm", "cargo")

	// Type override flags (mutually exclusive)
	addCmd.Flags().Bool("package", false, "Force all items to be treated as packages")
	addCmd.Flags().Bool("dotfile", false, "Force all items to be treated as dotfiles")
	addCmd.MarkFlagsMutuallyExclusive("package", "dotfile")

	// Common flags
	addCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")
	addCmd.Flags().BoolP("force", "f", false, "Force addition even if already managed")

	// Add intelligent completion
	addCmd.ValidArgsFunction = completeAddItems
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// No items specified - add all untracked items
		return addAllUntrackedItems(cmd)
	}

	// Parse flags
	flags, err := ParseUnifiedFlags(cmd)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "add", "flags", "invalid flag combination")
	}

	// Intelligently separate packages and dotfiles
	packages, dotfiles, ambiguous := ProcessMixedItemsWithFlags(args, flags)

	// Handle ambiguous items
	if len(ambiguous) > 0 {
		return handleAmbiguousItems(cmd, ambiguous, flags)
	}

	return processMixedItems(cmd, packages, dotfiles, flags)
}

// processMixedItems handles both packages and dotfiles in a single operation
func processMixedItems(cmd *cobra.Command, packages []string, dotfiles []string, flags *CommandFlags) error {
	// Parse output format
	format, err := ParseOutputFormat(flags.Output)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "add", "output-format", "invalid output format")
	}

	var allResults []operations.OperationResult
	reporter := operations.NewProgressReporter("add", format == OutputTable)

	// Process packages if any
	if len(packages) > 0 {
		pkgResults, err := processPackages(cmd, packages, flags)
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
		dotResults, err := processDotfiles(cmd, dotfiles, flags)
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
		return handleMixedResults(allResults, format)
	}

	// Determine exit code
	return operations.DetermineExitCode(allResults, errors.DomainCommands, "add-mixed")
}

// processPackages handles package additions using existing logic
func processPackages(cmd *cobra.Command, packageNames []string, flags *CommandFlags) ([]operations.OperationResult, error) {
	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Load config for default manager
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Determine which manager to use
	targetManager := flags.Manager
	if targetManager == "" {
		targetManager = cfg.Resolve().GetDefaultManager()
	}

	// Validate manager
	if targetManager != "homebrew" && targetManager != "npm" && targetManager != "cargo" {
		return nil, errors.NewError(errors.ErrInvalidInput, errors.DomainPackages, "validate", fmt.Sprintf("unsupported manager '%s'. Use: homebrew, npm, cargo", targetManager))
	}

	// Initialize package manager
	packageManagers := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"npm":      managers.NewNpmManager(),
		"cargo":    managers.NewCargoManager(),
	}
	mgr := packageManagers[targetManager]

	// Check if manager is available
	ctx := context.Background()
	available, err := mgr.IsAvailable(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "failed to check manager availability")
	}
	if !available {
		return nil, errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "check", fmt.Sprintf("manager '%s' is not available", targetManager)).
			WithSuggestionCommand("plonk doctor")
	}

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)

	// Create operation context with timeout
	opCtx, cancel := operations.CreateOperationContext(5 * time.Minute)
	defer cancel()

	// Process packages sequentially
	results := make([]operations.OperationResult, 0, len(packageNames))

	for _, packageName := range packageNames {
		// Check for cancellation
		if err := operations.CheckCancellation(opCtx, errors.DomainPackages, "add-multiple"); err != nil {
			return results, err
		}

		result := addSinglePackage(opCtx, cfg, lockService, mgr, targetManager, packageName, flags.DryRun)
		results = append(results, result)
	}

	return results, nil
}

// processDotfiles handles dotfile additions using existing logic
func processDotfiles(cmd *cobra.Command, dotfilePaths []string, flags *CommandFlags) ([]operations.OperationResult, error) {
	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "add", "failed to get home directory")
	}

	configDir := config.GetDefaultConfigDirectory()

	// Load config for ignore patterns
	cfg, err := loadOrCreateConfig(configDir)
	if err != nil {
		return nil, err
	}

	// Create operation context with timeout
	opCtx, cancel := operations.CreateOperationContext(5 * time.Minute)
	defer cancel()

	// Process dotfiles sequentially
	var results []operations.OperationResult

	for _, dotfilePath := range dotfilePaths {
		// Check for cancellation
		if err := operations.CheckCancellation(opCtx, errors.DomainDotfiles, "add-multiple"); err != nil {
			return results, err
		}

		// Process each dotfile (can result in multiple files for directories)
		dotfileResults := addSingleDotfile(opCtx, cfg, homeDir, configDir, dotfilePath, flags.DryRun)
		results = append(results, dotfileResults...)
	}

	return results, nil
}

// addSinglePackage reuses the existing package addition logic
func addSinglePackage(ctx context.Context, cfg *config.Config, lockService *lock.YAMLLockService, mgr managers.PackageManager, targetManager string, packageName string, dryRun bool) operations.OperationResult {
	result := operations.OperationResult{
		Name:    packageName,
		Manager: targetManager,
	}

	// Check if package is already in lock file
	if lockService.HasPackage(targetManager, packageName) {
		result.Status = "skipped"
		result.AlreadyManaged = true
		return result
	}

	// If dry run, just report what would happen
	if dryRun {
		result.Status = "would-add"
		return result
	}

	// Check if package is already installed
	installed, err := mgr.IsInstalled(ctx, packageName)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.Wrap(err, errors.ErrCommandExecution, errors.DomainPackages, "check-installed", "failed to check if package is installed").
			WithItem(packageName)
		return result
	}

	// Install package if not already installed
	if !installed {
		err = mgr.Install(ctx, packageName)
		if err != nil {
			result.Status = "failed"
			result.Error = errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", packageName, "failed to install package").
				WithSuggestionCommand("plonk search " + packageName)
			return result
		}
	}

	// Get installed version for reporting
	version, err := mgr.GetInstalledVersion(ctx, packageName)
	if err != nil {
		// Version lookup failed, but installation succeeded - use "unknown"
		result.Version = "unknown"
	} else {
		result.Version = version
	}

	// Add package to lock file
	err = lockService.AddPackage(targetManager, packageName, result.Version)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainPackages, "update-lock", packageName, "failed to add package to lock file")
		return result
	}

	result.Status = "added"
	return result
}

// handleAmbiguousItems handles items that couldn't be automatically categorized
func handleAmbiguousItems(cmd *cobra.Command, ambiguous []string, flags *CommandFlags) error {
	// For now, default ambiguous items to packages with a warning
	format, _ := ParseOutputFormat(flags.Output)

	if format == OutputTable {
		fmt.Printf("Warning: Ambiguous items detected, treating as packages: %v\n", ambiguous)
		fmt.Printf("Use --package or --dotfile flags to force interpretation\n\n")
	}

	// Process as packages
	return processMixedItems(cmd, ambiguous, []string{}, flags)
}

// addAllUntrackedItems adds all untracked packages and dotfiles
func addAllUntrackedItems(cmd *cobra.Command) error {
	flags, err := ParseUnifiedFlags(cmd)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "add-all", "flags", "invalid flag combination")
	}

	// Parse output format
	format, err := ParseOutputFormat(flags.Output)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "add-all", "output-format", "invalid output format")
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "add-all", "failed to get home directory")
	}

	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register package provider
	ctx := context.Background()
	packageProvider, err := createPackageProvider(ctx, configDir)
	if err != nil {
		return err
	}
	reconciler.RegisterProvider("package", packageProvider)

	// Register dotfile provider
	cfg, _ := config.LoadConfig(configDir)
	if cfg == nil {
		cfg = &config.Config{}
	}
	dotfileProvider := createDotfileProvider(homeDir, configDir, cfg)
	reconciler.RegisterProvider("dotfile", dotfileProvider)

	// Reconcile all domains to find untracked items
	summary, err := reconciler.ReconcileAll(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile state")
	}

	// Extract untracked items from all domains
	var untrackedPackages []string
	var untrackedDotfiles []string

	for _, result := range summary.Results {
		for _, item := range result.Untracked {
			if item.Domain == "package" {
				untrackedPackages = append(untrackedPackages, item.Name)
			} else if item.Domain == "dotfile" {
				untrackedDotfiles = append(untrackedDotfiles, item.Name)
			}
		}
	}

	if len(untrackedPackages) == 0 && len(untrackedDotfiles) == 0 {
		if format == OutputTable {
			fmt.Println("No untracked items found")
		}
		return nil
	}

	if flags.DryRun {
		if format == OutputTable {
			fmt.Printf("Would add %d untracked packages and %d untracked dotfiles:\n\n",
				len(untrackedPackages), len(untrackedDotfiles))
			for _, pkg := range untrackedPackages {
				fmt.Printf("  📦 %s\n", pkg)
			}
			for _, dot := range untrackedDotfiles {
				fmt.Printf("  📄 %s\n", dot)
			}
		}
		return nil
	}

	// Process all untracked items
	return processMixedItems(cmd, untrackedPackages, untrackedDotfiles, flags)
}

// handleMixedResults handles structured output for mixed operations
func handleMixedResults(results []operations.OperationResult, format OutputFormat) error {
	// Create a mixed operation summary
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

	mixedOutput := MixedAddOutput{
		TotalItems:     len(results),
		PackageResults: packageResults,
		DotfileResults: dotfileResults,
		Summary: MixedOperationSummary{
			PackagesAdded: CountByStatus(packageResults, "added"),
			DotfilesAdded: CountByStatus(dotfileResults, "added"),
			Failed:        CountByStatus(results, "failed"),
			Skipped:       CountByStatus(results, "skipped"),
		},
	}

	return RenderOutput(mixedOutput, format)
}

// Helper functions
func isPartialFailure(err error) bool {
	// Check if this is a partial failure that should allow continuation
	if mixedErr, ok := err.(MixedOperationError); ok {
		return mixedErr.PartialSuccess
	}
	return false
}

// CountByStatus counts results with a specific status (reuse from operations if available)
func CountByStatus(results []operations.OperationResult, status string) int {
	count := 0
	for _, result := range results {
		if result.Status == status {
			count++
		}
	}
	return count
}

// completeAddItems provides intelligent completion for the add command
func completeAddItems(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Detect if completing package or dotfile based on input
	if DetectItemType(toComplete) == ItemTypeDotfile {
		return completeDotfilePaths(cmd, args, toComplete)
	}
	return completePackageNames(cmd, args, toComplete)
}

// MixedAddOutput represents output for mixed add operations
type MixedAddOutput struct {
	TotalItems     int                          `json:"total_items" yaml:"total_items"`
	PackageResults []operations.OperationResult `json:"packages" yaml:"packages"`
	DotfileResults []operations.OperationResult `json:"dotfiles" yaml:"dotfiles"`
	Summary        MixedOperationSummary        `json:"summary" yaml:"summary"`
}

// MixedOperationSummary provides summary for mixed operations
type MixedOperationSummary struct {
	PackagesAdded int `json:"packages_added" yaml:"packages_added"`
	DotfilesAdded int `json:"dotfiles_added" yaml:"dotfiles_added"`
	Failed        int `json:"failed" yaml:"failed"`
	Skipped       int `json:"skipped" yaml:"skipped"`
}

// TableOutput generates human-friendly table output for mixed add
func (m MixedAddOutput) TableOutput() string {
	output := "Mixed Add Operation\n==================\n\n"

	if m.Summary.PackagesAdded > 0 {
		output += fmt.Sprintf("📦 Added %d packages\n", m.Summary.PackagesAdded)
	}
	if m.Summary.DotfilesAdded > 0 {
		output += fmt.Sprintf("📄 Added %d dotfiles\n", m.Summary.DotfilesAdded)
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
func (m MixedAddOutput) StructuredData() any {
	return m
}
