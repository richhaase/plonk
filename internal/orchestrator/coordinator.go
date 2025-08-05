// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// Orchestrator manages resources and their lock file state
type Orchestrator struct {
	ctx          context.Context
	config       *config.Config
	lock         lock.LockService
	configDir    string
	homeDir      string
	hookRunner   *HookRunner
	dryRun       bool
	packagesOnly bool
	dotfilesOnly bool
}

// New creates a new orchestrator instance with options
func New(opts ...Option) *Orchestrator {
	o := &Orchestrator{
		hookRunner: NewHookRunner(),
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.configDir != "" {
		o.lock = lock.NewYAMLLockService(o.configDir)
	}

	return o
}

// Apply orchestrates the application of all resources
func (o *Orchestrator) Apply(ctx context.Context) (ApplyResult, error) {
	result := ApplyResult{
		DryRun:  o.dryRun,
		Success: false,
	}

	// Store context
	o.ctx = ctx

	// Run pre-apply hooks
	if o.config != nil && len(o.config.Hooks.PreApply) > 0 {
		if err := o.hookRunner.RunPreApply(ctx, o.config.Hooks.PreApply); err != nil {
			result.Error = fmt.Sprintf("pre-apply hook failed: %v", err)
			return result, fmt.Errorf("pre-apply hook failed: %w", err)
		}
	}

	// Apply packages (unless dotfiles-only)
	if !o.dotfilesOnly {
		packageResult, err := packages.Apply(ctx, o.configDir, o.config, o.dryRun)
		result.Packages = &packageResult
		if err != nil {
			result.AddPackageError(fmt.Errorf("package apply failed: %w", err))
		}
	}

	// Apply dotfiles (unless packages-only)
	if !o.packagesOnly {
		dotfileResult, err := dotfiles.Apply(ctx, o.configDir, o.homeDir, o.config, o.dryRun)
		result.Dotfiles = &dotfileResult
		if err != nil {
			result.AddDotfileError(fmt.Errorf("dotfile apply failed: %w", err))
		}
	}

	// Run post-apply hooks only if we had some success
	if o.config != nil && len(o.config.Hooks.PostApply) > 0 {
		if err := o.hookRunner.RunPostApply(ctx, o.config.Hooks.PostApply); err != nil {
			// Post-apply hook failure is not fatal, just add to errors
			result.Error = fmt.Sprintf("post-apply hook failed: %v", err)
		}
	}

	// Determine overall success
	// Success if we had no critical errors and at least some operations succeeded
	if result.Packages != nil {
		if !o.dryRun && result.Packages.TotalInstalled > 0 {
			result.Success = true
		} else if o.dryRun && result.Packages.TotalWouldInstall > 0 {
			result.Success = true
		}
	}
	if result.Dotfiles != nil {
		if !o.dryRun && result.Dotfiles.Summary.Added > 0 {
			result.Success = true
		} else if o.dryRun && result.Dotfiles.Summary.Added > 0 {
			result.Success = true
		}
	}

	// If we had any failures, return an error even if some succeeded
	if result.HasErrors() {
		return result, result.GetCombinedError()
	}

	return result, nil
}
