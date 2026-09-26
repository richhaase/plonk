// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/packages"
)

// Options selects the configuration and resources to apply.
type Options struct {
	Config       *config.Config
	ConfigDir    string
	HomeDir      string
	DryRun       bool
	PackagesOnly bool
	DotfilesOnly bool
}

// Apply installs missing packages and deploys dotfiles, collecting partial failures.
func Apply(ctx context.Context, opts Options) (output.ApplyResult, error) {
	result := output.ApplyResult{DryRun: opts.DryRun}

	// Derive per-domain timeouts
	t := config.GetTimeouts(opts.Config)

	// Apply packages (unless dotfiles-only).
	// Per-package timeouts live inside packages.SimpleApply; we no longer wrap
	// the whole batch in one budget — a single slow Homebrew download used to
	// burn the entire phase's deadline.
	if !opts.DotfilesOnly {
		simpleResult, err := packages.SimpleApply(ctx, opts.ConfigDir, opts.DryRun)
		if simpleResult != nil {
			packageResult := convertSimpleApplyResult(simpleResult, opts.DryRun)
			result.Packages = &packageResult
		}
		if err != nil {
			result.AddPackageError(fmt.Errorf("package apply failed: %w", err))
		}
	}

	// Apply dotfiles (unless packages-only)
	if !opts.PackagesOnly {
		dctx, dcancel := context.WithTimeout(ctx, t.Dotfile)
		dotfileResult, err := dotfiles.Apply(dctx, opts.ConfigDir, opts.HomeDir, opts.Config, opts.DryRun)
		dcancel()
		result.Dotfiles = &dotfileResult
		if err != nil {
			result.AddDotfileError(fmt.Errorf("dotfile apply failed: %w", err))
		}
	}

	// Determine overall success
	// Success means no errors occurred. A clean no-op is considered success.
	// This supports idempotent operations - running apply multiple times is safe.
	result.Success = !result.HasErrors()

	if result.Packages != nil {
		result.Changed = (!opts.DryRun && result.Packages.TotalInstalled > 0) ||
			(opts.DryRun && result.Packages.TotalWouldInstall > 0)
	}
	if result.Dotfiles != nil {
		result.Changed = result.Changed || result.Dotfiles.Summary.Added > 0 || result.Dotfiles.Summary.Updated > 0
	}

	return result, result.GetCombinedError()
}

// convertSimpleApplyResult converts packages.SimpleApplyResult to output.PackageResults
func convertSimpleApplyResult(r *packages.SimpleApplyResult, dryRun bool) output.PackageResults {
	result := output.PackageResults{
		DryRun:            dryRun,
		TotalInstalled:    len(r.Installed),
		TotalWouldInstall: len(r.WouldInstall),
		TotalFailed:       len(r.Failed),
	}

	// Group by manager
	managerPackages := make(map[string][]output.PackageOperation)

	// Handle actually installed packages
	for _, spec := range r.Installed {
		manager, pkg := splitSpec(spec)
		managerPackages[manager] = append(managerPackages[manager], output.PackageOperation{
			Name:   pkg,
			Status: "installed",
		})
	}

	// Handle would-install packages (dry-run)
	for _, spec := range r.WouldInstall {
		manager, pkg := splitSpec(spec)
		managerPackages[manager] = append(managerPackages[manager], output.PackageOperation{
			Name:   pkg,
			Status: "would-install",
		})
	}

	// Build error map for failed packages
	errorMap := make(map[string]string)
	for i, spec := range r.Failed {
		if i < len(r.Errors) && r.Errors[i] != nil {
			errorMap[spec] = r.Errors[i].Error()
		}
	}

	// Handle failed packages with error details
	for _, spec := range r.Failed {
		manager, pkg := splitSpec(spec)
		op := output.PackageOperation{
			Name:   pkg,
			Status: "failed",
		}
		if errMsg, ok := errorMap[spec]; ok {
			op.Error = errMsg
		}
		managerPackages[manager] = append(managerPackages[manager], op)
	}

	// TotalMissing = packages that were not installed at reconciliation time
	// In dry-run: WouldInstall + Failed (packages that need installation or couldn't be evaluated)
	// In real run: Installed + Failed (packages that were missing - some fixed, some still missing)
	if dryRun {
		result.TotalMissing = result.TotalWouldInstall + result.TotalFailed
	} else {
		result.TotalMissing = result.TotalInstalled + result.TotalFailed
	}

	// Build manager results with per-manager missing counts (sorted for deterministic output)
	for _, manager := range slices.Sorted(maps.Keys(managerPackages)) {
		pkgs := managerPackages[manager]
		result.Managers = append(result.Managers, output.ManagerResults{
			Name:         manager,
			MissingCount: len(pkgs),
			Packages:     pkgs,
		})
	}

	return result
}

// splitSpec splits "manager:package" into manager and package
func splitSpec(spec string) (string, string) {
	manager, pkg, found := strings.Cut(spec, ":")
	if !found {
		return "", spec
	}
	return manager, pkg
}
