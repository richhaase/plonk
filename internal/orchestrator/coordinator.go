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

	// Derive per-domain timeouts
	t := config.GetTimeouts(o.config)

	// Apply packages (unless dotfiles-only)
	if !o.dotfilesOnly {
		pctx, pcancel := context.WithTimeout(ctx, t.Package)
		packageResult, err := packages.Apply(pctx, o.configDir, o.config, o.dryRun)
		pcancel()
		result.Packages = &packageResult
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
