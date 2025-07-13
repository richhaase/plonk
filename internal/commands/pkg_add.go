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
	"github.com/richhaase/plonk/internal/operations"
	"github.com/richhaase/plonk/internal/state"
	"github.com/spf13/cobra"
)

var (
	manager string
)

var pkgAddCmd = &cobra.Command{
	Use:   "add [package1] [package2] ...",
	Short: "Add package(s) to plonk configuration and install them",
	Long: `Add one or more packages to your plonk.lock file and install them.

With package names:
  plonk pkg add htop                        # Add htop using default manager
  plonk pkg add git neovim ripgrep          # Add multiple packages
  plonk pkg add git --manager homebrew      # Add git specifically to homebrew
  plonk pkg add lodash --manager npm        # Add lodash to npm global packages
  plonk pkg add ripgrep --manager cargo     # Add ripgrep to cargo packages
  plonk pkg add --dry-run git neovim        # Preview what would be added

Without arguments:
  plonk pkg add                             # Add all untracked packages
  plonk pkg add --dry-run                   # Preview all untracked packages`,
	Args: cobra.ArbitraryArgs,
	RunE: runPkgAdd,
}

func init() {
	pkgCmd.AddCommand(pkgAddCmd)
	pkgAddCmd.Flags().StringVar(&manager, "manager", "", "Package manager to use (homebrew|npm|cargo)")
	pkgAddCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")

	// Add package name completion
	pkgAddCmd.ValidArgsFunction = completePackageNames

	// Add manager flag completion
	pkgAddCmd.RegisterFlagCompletionFunc("manager", completeManagerNames)
}

func runPkgAdd(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if len(args) == 0 {
		// No package specified - add all untracked packages
		return addAllUntrackedPackages(cmd, dryRun)
	}

	// Handle single or multiple packages
	return addPackages(cmd, args, dryRun)
}

// addPackages handles adding one or more specific packages
func addPackages(cmd *cobra.Command, packageNames []string, dryRun bool) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "pkg-add", "output-format", "invalid output format")
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Load config for default manager
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Determine which manager to use
	targetManager := manager
	if targetManager == "" {
		targetManager = cfg.Resolve().GetDefaultManager()
	}

	// Validate manager
	if targetManager != "homebrew" && targetManager != "npm" && targetManager != "cargo" {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainPackages, "validate", fmt.Sprintf("unsupported manager '%s'. Use: homebrew, npm, cargo", targetManager))
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
		return errors.Wrap(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "failed to check manager availability")
	}
	if !available {
		return errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "check", fmt.Sprintf("manager '%s' is not available", targetManager)).
			WithSuggestionCommand("plonk doctor")
	}

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)

	// Create operation context with timeout
	opCtx, cancel := operations.CreateOperationContext(5 * time.Minute)
	defer cancel()

	// Process packages sequentially
	results := make([]operations.OperationResult, 0, len(packageNames))
	reporter := operations.NewProgressReporter("package", format == OutputTable)

	for _, packageName := range packageNames {
		// Check for cancellation
		if err := operations.CheckCancellation(opCtx, errors.DomainPackages, "add-multiple"); err != nil {
			return err
		}

		result := addSinglePackage(opCtx, cfg, lockService, mgr, targetManager, packageName, dryRun)
		results = append(results, result)

		// Show progress immediately
		reporter.ShowItemProgress(result)
	}

	// Handle output based on format
	if format == OutputTable {
		// Show summary for table output
		reporter.ShowBatchSummary(results)
	} else {
		// For structured output, create appropriate response
		if len(packageNames) == 1 {
			// Single package - use existing EnhancedAddOutput format for compatibility
			result := results[0]
			output := EnhancedAddOutput{
				Package:          result.Name,
				Manager:          result.Manager,
				ConfigAdded:      result.Status == "added",
				AlreadyInConfig:  result.AlreadyManaged,
				Installed:        result.Status == "added",
				AlreadyInstalled: result.Status == "skipped",
				Actions:          createActionsFromResult(result),
			}
			if result.Error != nil {
				output.Error = result.Error.Error()
			}
			return RenderOutput(output, format)
		} else {
			// Multiple packages - use BatchAddOutput
			summary := operations.CalculateSummary(results)
			batchOutput := BatchAddOutput{
				TotalPackages:     summary.Total,
				AddedToConfig:     summary.Added,
				Installed:         summary.Added,
				AlreadyConfigured: summary.Skipped,
				AlreadyInstalled:  0, // We don't track this separately in our new model
				Errors:            summary.Failed,
				Packages:          convertResultsToEnhancedAdd(results),
			}
			return RenderOutput(batchOutput, format)
		}
	}

	// Determine exit code
	return operations.DetermineExitCode(results, errors.DomainPackages, "add-multiple")
}

// addSinglePackage processes a single package and returns the result
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

// createActionsFromResult creates action strings for backward compatibility
func createActionsFromResult(result operations.OperationResult) []string {
	var actions []string

	switch result.Status {
	case "added":
		if result.Version != "" && result.Version != "unknown" {
			actions = append(actions, fmt.Sprintf("Successfully installed %s@%s", result.Name, result.Version))
		} else {
			actions = append(actions, fmt.Sprintf("Successfully installed %s", result.Name))
		}
		actions = append(actions, fmt.Sprintf("Added %s to lock file", result.Name))
	case "skipped":
		actions = append(actions, fmt.Sprintf("%s already managed by %s", result.Name, result.Manager))
	case "would-add":
		actions = append(actions, fmt.Sprintf("Would install %s", result.Name))
		actions = append(actions, fmt.Sprintf("Would add %s to lock file", result.Name))
	case "failed":
		actions = append(actions, fmt.Sprintf("Failed to process %s", result.Name))
	}

	return actions
}

// convertResultsToEnhancedAdd converts OperationResult to EnhancedAddOutput for structured output
func convertResultsToEnhancedAdd(results []operations.OperationResult) []EnhancedAddOutput {
	outputs := make([]EnhancedAddOutput, len(results))
	for i, result := range results {
		outputs[i] = EnhancedAddOutput{
			Package:          result.Name,
			Manager:          result.Manager,
			ConfigAdded:      result.Status == "added",
			AlreadyInConfig:  result.AlreadyManaged,
			Installed:        result.Status == "added",
			AlreadyInstalled: result.Status == "skipped",
			Actions:          createActionsFromResult(result),
		}
		if result.Error != nil {
			outputs[i].Error = result.Error.Error()
		}
	}
	return outputs
}

// addAllUntrackedPackages adds all untracked packages to the lock file
func addAllUntrackedPackages(cmd *cobra.Command, dryRun bool) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "pkg-add-all", "output-format", "invalid output format")
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Initialize lock file service
	lockService := lock.NewYAMLLockService(configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	// Create reconciler to get untracked packages
	reconciler := state.NewReconciler()

	// Create package provider using lock file adapter
	ctx := context.Background()
	packageProvider := state.NewMultiManagerPackageProvider()

	// Add all available managers
	managers := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"npm":      managers.NewNpmManager(),
		"cargo":    managers.NewCargoManager(),
	}

	for managerName, mgr := range managers {
		available, err := mgr.IsAvailable(ctx)
		if err != nil {
			return fmt.Errorf("failed to check %s availability: %w", managerName, err)
		}
		if available {
			managerAdapter := state.NewManagerAdapter(mgr)
			packageProvider.AddManager(managerName, managerAdapter, lockAdapter)
		}
	}

	reconciler.RegisterProvider("package", packageProvider)

	// Reconcile to get package states
	result, err := reconciler.ReconcileProvider(ctx, "package")
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile package states")
	}

	untrackedPackages := result.Untracked

	if len(untrackedPackages) == 0 {
		if format == OutputTable {
			fmt.Println("No untracked packages found")
		}
		return nil
	}

	if dryRun {
		if format == OutputTable {
			fmt.Printf("Would add %d untracked packages:\n\n", len(untrackedPackages))
			for _, pkg := range untrackedPackages {
				fmt.Printf("  %s (%s)\n", pkg.Name, pkg.Manager)
			}
		}
		return nil
	}

	// Add packages to lock file
	addedCount := 0
	for _, pkg := range untrackedPackages {
		if !lockService.HasPackage(pkg.Manager, pkg.Name) {
			err = lockService.AddPackage(pkg.Manager, pkg.Name, "latest")
			if err != nil {
				return errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainCommands, "update", pkg.Name, "failed to add package to lock file")
			}
			addedCount++
		}
	}

	if addedCount == 0 {
		if format == OutputTable {
			fmt.Println("No packages were added (all were already managed)")
		}
		return nil
	}

	if format == OutputTable {
		fmt.Printf("Successfully added %d packages to lock file\n", addedCount)
	}

	// Prepare structured output
	addAllResult := AddAllOutput{
		Added:  addedCount,
		Total:  len(untrackedPackages),
		Action: "added-all",
	}

	return RenderOutput(addAllResult, format)
}

// completePackageNames provides package name completion based on available managers
func completePackageNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()

	// Get manager preference from flag or config
	targetManager, _ := cmd.Flags().GetString("manager")
	if targetManager == "" {
		configDir := config.GetDefaultConfigDirectory()
		cfg, err := config.LoadConfig(configDir)
		if err == nil {
			targetManager = cfg.Resolve().GetDefaultManager()
		} else {
			targetManager = "homebrew" // fallback
		}
	}

	// Get manager instance
	mgr := getManagerInstance(targetManager)
	if mgr == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Check if manager is available
	available, err := mgr.IsAvailable(ctx)
	if err != nil || !available {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// For now, return some common packages based on manager type
	// This could be enhanced to use actual search functionality
	suggestions := getCommonPackages(targetManager, toComplete)

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// completeManagerNames provides completion for manager flag
func completeManagerNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	managers := []string{"homebrew", "npm", "cargo"}
	return managers, cobra.ShellCompDirectiveNoFileComp
}

// getManagerInstance returns a manager instance for the given name
func getManagerInstance(managerName string) managers.PackageManager {
	switch managerName {
	case "homebrew":
		return managers.NewHomebrewManager()
	case "npm":
		return managers.NewNpmManager()
	case "cargo":
		return managers.NewCargoManager()
	default:
		return nil
	}
}

// getCommonPackages returns common package suggestions for the given manager
func getCommonPackages(managerName, prefix string) []string {
	var packages []string

	switch managerName {
	case "homebrew":
		packages = []string{
			"git", "curl", "wget", "htop", "ripgrep", "fzf", "neovim", "tmux",
			"jq", "tree", "bat", "exa", "fd", "zsh", "fish", "nodejs", "python",
			"go", "rust", "docker", "kubectl", "helm", "terraform", "awscli",
		}
	case "npm":
		packages = []string{
			"typescript", "eslint", "prettier", "jest", "webpack", "babel",
			"react", "vue", "angular", "express", "lodash", "axios", "moment",
			"chalk", "commander", "inquirer", "yargs", "cross-env", "nodemon",
		}
	case "cargo":
		packages = []string{
			"ripgrep", "bat", "exa", "fd-find", "tokei", "hyperfine", "dust",
			"bandwhich", "bottom", "starship", "zoxide", "delta", "gitui",
		}
	}

	// Filter packages that start with the prefix
	if prefix == "" {
		return packages
	}

	var filtered []string
	for _, pkg := range packages {
		if strings.HasPrefix(pkg, prefix) {
			filtered = append(filtered, pkg)
		}
	}

	return filtered
}
