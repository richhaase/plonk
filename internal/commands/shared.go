// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/cli"
	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/core"
	"github.com/richhaase/plonk/internal/errors"
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
	packageProvider, err := core.CreatePackageProvider(ctx, configDir)
	if err != nil {
		return err
	}
	reconciler.RegisterProvider("package", packageProvider)

	// Get specific manager if flag is set
	flags, err := cli.ParseSimpleFlags(cmd)
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
	dotfileProvider := core.CreateDotfileProvider(homeDir, configDir, cfg)
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
