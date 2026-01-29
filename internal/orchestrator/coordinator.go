// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/packages"
)

// Orchestrator manages resources and coordinates apply operations
type Orchestrator struct {
	ctx          context.Context
	config       *config.Config
	configDir    string
	homeDir      string
	dryRun       bool
	packagesOnly bool
	dotfilesOnly bool
}

// New creates a new orchestrator instance with options
func New(opts ...Option) *Orchestrator {
	o := &Orchestrator{}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// Apply orchestrates the application of all resources
func (o *Orchestrator) Apply(ctx context.Context) (output.ApplyResult, error) {
	result := output.ApplyResult{
		DryRun:  o.dryRun,
		Success: false,
	}

	// Store context
	o.ctx = ctx

	// Derive per-domain timeouts
	t := config.GetTimeouts(o.config)

	// Apply packages (unless dotfiles-only)
	if !o.dotfilesOnly {
		pctx, pcancel := context.WithTimeout(ctx, t.Package)
		simpleResult, err := packages.SimpleApply(pctx, o.configDir, o.dryRun)
		pcancel()
		if simpleResult != nil {
			packageResult := convertSimpleApplyResult(simpleResult, o.dryRun)
			result.Packages = &packageResult
		}
		if err != nil {
			result.AddPackageError(fmt.Errorf("package apply failed: %w", err))
		}
	}

	// Apply dotfiles (unless packages-only)
	if !o.packagesOnly {
		dctx, dcancel := context.WithTimeout(ctx, t.Dotfile)
		dotfileResult, err := dotfiles.Apply(dctx, o.configDir, o.homeDir, o.config, o.dryRun)
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

	// Determine if any changes were made (useful for reporting)
	changed := false
	if result.Packages != nil {
		if !o.dryRun && result.Packages.TotalInstalled > 0 {
			changed = true
		} else if o.dryRun && result.Packages.TotalWouldInstall > 0 {
			changed = true
		}
	}
	if result.Dotfiles != nil {
		if !o.dryRun && result.Dotfiles.Summary.Added > 0 {
			changed = true
		} else if o.dryRun && result.Dotfiles.Summary.Added > 0 {
			changed = true
		}
	}
	result.Changed = changed

	// If we had any failures, return an error even if some operations succeeded
	if result.HasErrors() {
		return result, result.GetCombinedError()
	}

	return result, nil
}

// convertSimpleApplyResult converts packages.SimpleApplyResult to output.PackageResults
func convertSimpleApplyResult(r *packages.SimpleApplyResult, dryRun bool) output.PackageResults {
	result := output.PackageResults{
		DryRun: dryRun,
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
		result.TotalInstalled++
	}

	// Handle would-install packages (dry-run)
	for _, spec := range r.WouldInstall {
		manager, pkg := splitSpec(spec)
		managerPackages[manager] = append(managerPackages[manager], output.PackageOperation{
			Name:   pkg,
			Status: "would-install",
		})
		result.TotalWouldInstall++
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
		result.TotalFailed++
	}

	// TotalMissing = packages that were not installed at reconciliation time
	// In dry-run: WouldInstall (packages that need installation)
	// In real run: Installed + Failed (packages that were missing - some fixed, some still missing)
	if dryRun {
		result.TotalMissing = result.TotalWouldInstall
	} else {
		result.TotalMissing = result.TotalInstalled + result.TotalFailed
	}

	// Build manager results
	for manager, pkgs := range managerPackages {
		result.Managers = append(result.Managers, output.ManagerResults{
			Name:     manager,
			Packages: pkgs,
		})
	}

	return result
}

// splitSpec splits "manager:package" into manager and package
func splitSpec(spec string) (string, string) {
	for i, c := range spec {
		if c == ':' {
			return spec[:i], spec[i+1:]
		}
	}
	return "", spec
}
