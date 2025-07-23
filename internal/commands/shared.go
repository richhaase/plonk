// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/core"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/richhaase/plonk/internal/runtime"
	"github.com/richhaase/plonk/internal/services"
	"github.com/richhaase/plonk/internal/state"
	"github.com/richhaase/plonk/internal/ui"
	"github.com/spf13/cobra"
)

// Type aliases for UI types (these have been moved to internal/ui/formatters.go)
type ApplyOutput = ui.ApplyOutput
type ManagerApplyResult = ui.ManagerApplyResult
type PackageApplyResult = ui.PackageApplyResult

type DotfileApplyOutput = ui.DotfileApplyOutput
type DotfileAction = ui.DotfileAction

// TableOutput and StructuredData methods have been moved to internal/ui/formatters.go

type DotfileListOutput = ui.DotfileListOutput
type DotfileListSummary = ui.DotfileListSummary
type DotfileInfo = ui.DotfileInfo

// TableOutput and StructuredData methods moved to internal/ui/formatters.go

// Shared functions from the original commands

// applyPackages applies package configuration and returns the result (refactored to use business module)
func applyPackages(configDir string, cfg *config.Config, dryRun bool, format OutputFormat) (ApplyOutput, error) {
	ctx := context.Background()

	// Use business module for package operations
	options := services.PackageApplyOptions{
		ConfigDir: configDir,
		Config:    cfg,
		DryRun:    dryRun,
	}

	result, err := services.ApplyPackages(ctx, options)
	if err != nil {
		return ApplyOutput{}, err
	}

	// Convert business result to command output format
	outputData := ApplyOutput{
		DryRun:            result.DryRun,
		TotalMissing:      result.TotalMissing,
		TotalInstalled:    result.TotalInstalled,
		TotalFailed:       result.TotalFailed,
		TotalWouldInstall: result.TotalWouldInstall,
		Managers:          make([]ManagerApplyResult, len(result.Managers)),
	}

	// Convert manager results
	for i, mgr := range result.Managers {
		packages := make([]PackageApplyResult, len(mgr.Packages))
		for j, pkg := range mgr.Packages {
			packages[j] = PackageApplyResult{
				Name:   pkg.Name,
				Status: pkg.Status,
				Error:  pkg.Error,
			}
		}
		outputData.Managers[i] = ManagerApplyResult{
			Name:         mgr.Name,
			MissingCount: mgr.MissingCount,
			Packages:     packages,
		}
	}

	// Output summary for table format
	if format == OutputTable {
		if result.TotalMissing == 0 {
			fmt.Println("📦 All packages up to date")
		} else {
			if dryRun {
				fmt.Printf("📦 Package summary: %d packages would be installed\n", outputData.TotalWouldInstall)
			} else {
				fmt.Printf("📦 Package summary: %d installed, %d failed\n", outputData.TotalInstalled, outputData.TotalFailed)
			}
		}
		fmt.Println()
	}

	return outputData, nil
}

// applyDotfiles applies dotfile configuration and returns the result (refactored to use business module)
func applyDotfiles(configDir, homeDir string, cfg *config.Config, dryRun, backup bool, format OutputFormat) (DotfileApplyOutput, error) {
	ctx := context.Background()

	// Use business module for dotfile operations
	options := services.DotfileApplyOptions{
		ConfigDir: configDir,
		HomeDir:   homeDir,
		Config:    cfg,
		DryRun:    dryRun,
		Backup:    backup,
	}

	result, err := services.ApplyDotfiles(ctx, options)
	if err != nil {
		return DotfileApplyOutput{}, err
	}

	// Convert business result to command output format
	actions := make([]DotfileAction, len(result.Actions))
	for i, action := range result.Actions {
		actions[i] = DotfileAction{
			Source:      action.Source,
			Destination: action.Destination,
			Status:      action.Status,
			Reason:      "", // Business module uses Action field differently
		}
	}

	outputData := DotfileApplyOutput{
		DryRun:   result.DryRun,
		Deployed: result.Summary.Added + result.Summary.Updated,
		Skipped:  result.Summary.Unchanged,
		Actions:  actions,
	}

	// Output summary for table format
	if format == OutputTable {
		if result.TotalFiles == 0 {
			fmt.Println("📄 No dotfiles configured")
		} else {
			if dryRun {
				fmt.Printf("📄 Dotfile summary: %d dotfiles would be deployed, %d would be skipped\n", outputData.Deployed, outputData.Skipped)
			} else {
				fmt.Printf("📄 Dotfile summary: %d deployed, %d skipped\n", outputData.Deployed, outputData.Skipped)
			}
		}
	}

	return outputData, nil
}

// Shared functions from pkg_add.go and dot_add.go

// completeDotfilePaths provides file path completion for dotfiles
func completeDotfilePaths(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get home directory from shared context (no error handling needed)
	sharedCtx := runtime.GetSharedContext()
	_ = sharedCtx.HomeDir()

	// Define common dotfile suggestions
	commonDotfiles := []string{
		"~/.zshrc", "~/.bashrc", "~/.bash_profile", "~/.profile",
		"~/.vimrc", "~/.vim/", "~/.nvim/",
		"~/.gitconfig", "~/.gitignore_global",
		"~/.tmux.conf", "~/.tmux/",
		"~/.ssh/config", "~/.ssh/",
		"~/.aws/config", "~/.aws/credentials",
		"~/.config/", "~/.config/nvim/", "~/.config/fish/", "~/.config/alacritty/",
		"~/.docker/config.json",
		"~/.zprofile", "~/.zshenv",
		"~/.inputrc", "~/.editorconfig",
	}

	// If no input yet, return all common suggestions
	if toComplete == "" {
		return commonDotfiles, cobra.ShellCompDirectiveNoSpace
	}

	// If starts with tilde, filter common dotfiles
	if strings.HasPrefix(toComplete, "~/") {
		var filtered []string
		for _, suggestion := range commonDotfiles {
			if strings.HasPrefix(suggestion, toComplete) {
				filtered = append(filtered, suggestion)
			}
		}

		if len(filtered) > 0 {
			return filtered, cobra.ShellCompDirectiveNoSpace
		}

		// Fall back to file completion for ~/.config/ style paths
		return nil, cobra.ShellCompDirectiveDefault
	}

	// For relative paths, try to suggest based on common dotfile names
	if !strings.HasPrefix(toComplete, "/") {
		relativeSuggestions := []string{
			".zshrc", ".bashrc", ".bash_profile", ".profile",
			".vimrc", ".gitconfig", ".tmux.conf", ".inputrc",
			".editorconfig", ".zprofile", ".zshenv",
		}

		var filtered []string
		for _, suggestion := range relativeSuggestions {
			if strings.HasPrefix(suggestion, toComplete) {
				filtered = append(filtered, suggestion)
			}
		}

		if len(filtered) > 0 {
			return filtered, cobra.ShellCompDirectiveNoSpace
		}
	}

	// Fall back to default file completion for absolute paths and other cases
	return nil, cobra.ShellCompDirectiveDefault
}

// loadOrCreateConfig loads existing config or creates a new one
func loadOrCreateConfig(configDir string) (*config.Config, error) {
	manager := config.NewConfigManager(configDir)
	return manager.LoadOrCreate()
}

// createPackageProvider creates a multi-manager package provider using lock file
func createPackageProvider(ctx context.Context, configDir string) (*state.MultiManagerPackageProvider, error) {
	// Use SharedContext to create provider
	sharedCtx := runtime.GetSharedContext()
	return sharedCtx.CreatePackageProvider(ctx)
}

// createDotfileProvider creates a dotfile provider
func createDotfileProvider(homeDir string, configDir string, cfg *config.Config) *state.DotfileProvider {
	// Use SharedContext to create provider
	sharedCtx := runtime.GetSharedContext()
	provider, _ := sharedCtx.CreateDotfileProvider()
	return provider
}

// addSingleDotfile processes a single dotfile path and returns results for all files processed
func addSingleDotfile(ctx context.Context, cfg *config.Config, homeDir, configDir, dotfilePath string, dryRun bool) []operations.OperationResult {
	return core.AddSingleDotfile(ctx, cfg, homeDir, configDir, dotfilePath, dryRun)
}

// Shared output types from dot_add.go (moved to internal/ui/formatters.go)
type DotfileAddOutput = ui.DotfileAddOutput
type DotfileBatchAddOutput = ui.DotfileBatchAddOutput

// TableOutput and StructuredData methods moved to internal/ui/formatters.go

// Shared functions for pkg and dot list operations
// Note: runPkgList implementation deferred pending requirements clarification
func runPkgList(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "packages", "output-format", "invalid output format")
	}

	// Get directories from shared context
	sharedCtx := runtime.GetSharedContext()
	configDir := sharedCtx.ConfigDir()

	// Get reconciler from shared context
	reconciler := sharedCtx.Reconciler()

	// Register package provider
	ctx := context.Background()
	packageProvider, err := createPackageProvider(ctx, configDir)
	if err != nil {
		return err
	}
	reconciler.RegisterProvider("package", packageProvider)

	// Get specific manager if flag is set
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return err
	}

	// Reconcile packages
	domainResult, err := reconciler.ReconcileProvider(ctx, "package")
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile package state")
	}

	// If a specific manager is requested, filter results
	if flags.Manager != "" {
		filteredManaged := make([]state.Item, 0)
		filteredMissing := make([]state.Item, 0)
		filteredUntracked := make([]state.Item, 0)

		for _, item := range domainResult.Managed {
			if item.Manager == flags.Manager {
				filteredManaged = append(filteredManaged, item)
			}
		}
		for _, item := range domainResult.Missing {
			if item.Manager == flags.Manager {
				filteredMissing = append(filteredMissing, item)
			}
		}
		for _, item := range domainResult.Untracked {
			if item.Manager == flags.Manager {
				filteredUntracked = append(filteredUntracked, item)
			}
		}

		domainResult.Managed = filteredManaged
		domainResult.Missing = filteredMissing
		domainResult.Untracked = filteredUntracked
	}

	// Prepare manager groups
	managerGroups := make(map[string]*EnhancedManagerOutput)

	// Add managed packages
	for _, item := range domainResult.Managed {
		if _, exists := managerGroups[item.Manager]; !exists {
			managerGroups[item.Manager] = &EnhancedManagerOutput{
				Name:           item.Manager,
				ManagedCount:   0,
				MissingCount:   0,
				UntrackedCount: 0,
				Packages:       []EnhancedPackageOutput{},
			}
		}
		managerGroups[item.Manager].ManagedCount++
		managerGroups[item.Manager].Packages = append(managerGroups[item.Manager].Packages, EnhancedPackageOutput{
			Name:    item.Name,
			State:   "managed",
			Manager: item.Manager,
		})
	}

	// Add missing packages
	for _, item := range domainResult.Missing {
		if _, exists := managerGroups[item.Manager]; !exists {
			managerGroups[item.Manager] = &EnhancedManagerOutput{
				Name:           item.Manager,
				ManagedCount:   0,
				MissingCount:   0,
				UntrackedCount: 0,
				Packages:       []EnhancedPackageOutput{},
			}
		}
		managerGroups[item.Manager].MissingCount++
		managerGroups[item.Manager].Packages = append(managerGroups[item.Manager].Packages, EnhancedPackageOutput{
			Name:    item.Name,
			State:   "missing",
			Manager: item.Manager,
		})
	}

	// Add untracked packages if verbose
	if flags.Verbose {
		for _, item := range domainResult.Untracked {
			if _, exists := managerGroups[item.Manager]; !exists {
				managerGroups[item.Manager] = &EnhancedManagerOutput{
					Name:           item.Manager,
					ManagedCount:   0,
					MissingCount:   0,
					UntrackedCount: 0,
					Packages:       []EnhancedPackageOutput{},
				}
			}
			managerGroups[item.Manager].UntrackedCount++
			managerGroups[item.Manager].Packages = append(managerGroups[item.Manager].Packages, EnhancedPackageOutput{
				Name:    item.Name,
				State:   "untracked",
				Manager: item.Manager,
			})
		}
	}

	// Convert to slice
	managers := make([]EnhancedManagerOutput, 0, len(managerGroups))
	items := []EnhancedPackageOutput{}

	for _, mgr := range managerGroups {
		managers = append(managers, *mgr)
		items = append(items, mgr.Packages...)
	}

	// Create output structure
	output := PackageListOutput{
		ManagedCount:   len(domainResult.Managed),
		MissingCount:   len(domainResult.Missing),
		UntrackedCount: len(domainResult.Untracked),
		TotalCount:     len(domainResult.Managed) + len(domainResult.Missing) + len(domainResult.Untracked),
		Managers:       managers,
		Verbose:        flags.Verbose,
		Items:          items,
	}

	return RenderOutput(output, format)
}

func runDotList(cmd *cobra.Command, args []string) error {
	// Note: Full dotfiles layer integration deferred to maintain current functionality
	// Current implementation delegates to the state reconciliation system

	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "dotfiles", "output-format", "invalid output format")
	}

	// Get directories from shared context
	sharedCtx := runtime.GetSharedContext()
	configDir := sharedCtx.ConfigDir()
	homeDir := sharedCtx.HomeDir()

	// Load configuration using shared context cache
	cfg := sharedCtx.ConfigWithDefaults()

	// Use the shared reconciler
	reconciler := sharedCtx.Reconciler()
	dotfileProvider := createDotfileProvider(homeDir, configDir, cfg)
	reconciler.RegisterProvider("dotfile", dotfileProvider)

	ctx := context.Background()
	domainResult, err := reconciler.ReconcileProvider(ctx, "dotfile")
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile dotfiles")
	}

	// Parse filter flags
	showManaged, _ := cmd.Flags().GetBool("managed")
	showMissing, _ := cmd.Flags().GetBool("missing")
	showUntracked, _ := cmd.Flags().GetBool("untracked")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Filter based on flags
	var items []state.Item
	if showManaged {
		items = domainResult.Managed
	} else if showMissing {
		items = domainResult.Missing
	} else if showUntracked {
		items = domainResult.Untracked
	} else {
		// Default: show managed + missing, optionally untracked
		items = append(items, domainResult.Managed...)
		items = append(items, domainResult.Missing...)
		if verbose {
			items = append(items, domainResult.Untracked...)
		}
	}

	// Convert to output format using existing dotfiles types
	output := DotfileListOutput{
		Summary: DotfileListSummary{
			Total:     len(items),
			Managed:   len(domainResult.Managed),
			Missing:   len(domainResult.Missing),
			Untracked: len(domainResult.Untracked),
			Verbose:   verbose,
		},
		Dotfiles: convertToDotfileInfo(items),
	}

	return RenderOutput(output, format)
}

// convertToDotfileInfo converts state.Item to DotfileInfo for display
func convertToDotfileInfo(items []state.Item) []DotfileInfo {
	result := make([]DotfileInfo, len(items))
	for i, item := range items {
		// Map state.Item fields to DotfileInfo
		target := item.Path
		source := item.Name

		// Extract additional info from metadata if available
		if item.Metadata != nil {
			if t, ok := item.Metadata["target"].(string); ok && t != "" {
				target = t
			}
			if s, ok := item.Metadata["source"].(string); ok && s != "" {
				source = s
			}
		}

		result[i] = DotfileInfo{
			Name:   item.Name,
			State:  item.State.String(),
			Target: target,
			Source: source,
		}
	}
	return result
}

// removeSingleDotfile removes a single dotfile
func removeSingleDotfile(homeDir, configDir string, cfg *config.Config, dotfilePath string, dryRun bool) operations.OperationResult {
	return core.RemoveSingleDotfile(homeDir, configDir, cfg, dotfilePath, dryRun)
}

// SimpleFlags represents basic command flags without detection logic
type SimpleFlags struct {
	Manager string
	DryRun  bool
	Force   bool
	Verbose bool
	Output  string
}

// ParseSimpleFlags parses basic flags for package commands
func ParseSimpleFlags(cmd *cobra.Command) (*SimpleFlags, error) {
	flags := &SimpleFlags{}

	// Parse manager flags with precedence
	if brew, _ := cmd.Flags().GetBool("brew"); brew {
		flags.Manager = "homebrew"
	} else if npm, _ := cmd.Flags().GetBool("npm"); npm {
		flags.Manager = "npm"
	} else if cargo, _ := cmd.Flags().GetBool("cargo"); cargo {
		flags.Manager = "cargo"
	} else if pip, _ := cmd.Flags().GetBool("pip"); pip {
		flags.Manager = "pip"
	} else if gem, _ := cmd.Flags().GetBool("gem"); gem {
		flags.Manager = "gem"
	} else if goFlag, _ := cmd.Flags().GetBool("go"); goFlag {
		flags.Manager = "go"
	}

	// Parse common flags
	flags.DryRun, _ = cmd.Flags().GetBool("dry-run")
	flags.Force, _ = cmd.Flags().GetBool("force")
	flags.Verbose, _ = cmd.Flags().GetBool("verbose")
	flags.Output, _ = cmd.Flags().GetString("output")

	return flags, nil
}

// extractBinaryNameFromPath extracts the binary name from a Go module path
func extractBinaryNameFromPath(modulePath string) string {
	// Remove version specification if present
	modulePath = strings.Split(modulePath, "@")[0]

	// Extract the last component of the path
	parts := strings.Split(modulePath, "/")
	binaryName := parts[len(parts)-1]

	// Handle special case of .../cmd/toolname pattern
	if len(parts) >= 2 && parts[len(parts)-2] == "cmd" {
		return binaryName
	}

	// For simple cases, the binary name is usually the last component
	return binaryName
}
