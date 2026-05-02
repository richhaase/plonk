// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/output"
)

// SimpleApplyResult holds the result of applying packages
type SimpleApplyResult struct {
	Installed    []string // Packages that were actually installed
	WouldInstall []string // Packages that would be installed (dry-run only)
	Skipped      []string // Packages already installed
	Failed       []string // Packages that failed to install
	Errors       []error  // Errors for failed packages
}

// PerPackageTimeout bounds a single Install or IsInstalled invocation.
// The orchestrator no longer caps the whole batch — each package gets its own budget.
const PerPackageTimeout = 10 * time.Minute

// SimpleApply installs all tracked packages that are missing
func SimpleApply(ctx context.Context, configDir string, dryRun bool) (*SimpleApplyResult, error) {
	lockSvc := lock.NewLockV3Service(configDir)
	lockFile, err := lockSvc.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	result := &SimpleApplyResult{}

	// Sort managers for deterministic order — ensures managers that provide
	// tools (e.g., brew:go) are processed before managers that depend on them
	// (e.g., go:golang.org/x/tools/gopls)
	managers := make([]string, 0, len(lockFile.Packages))
	for manager := range lockFile.Packages {
		managers = append(managers, manager)
	}
	sort.Strings(managers)

	// Phase 1: build install plan, recording skipped/would-install/failed-from-IsInstalled.
	type planEntry struct {
		spec string
		pkg  string
		mgr  Manager
	}
	var plan []planEntry

	for _, manager := range managers {
		pkgs := lockFile.Packages[manager]
		mgr, err := GetManager(manager)
		if err != nil {
			for _, pkg := range pkgs {
				spec := manager + ":" + pkg
				result.Failed = append(result.Failed, spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: manager not available: %w", spec, err))
			}
			continue
		}

		var managerBroken bool
		var managerErr error
		for _, pkg := range pkgs {
			spec := manager + ":" + pkg

			if managerBroken {
				result.Failed = append(result.Failed, spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: %w", spec, managerErr))
				continue
			}

			installed, err := callWithTimeout(ctx, func(c context.Context) (bool, error) {
				return mgr.IsInstalled(c, pkg)
			})
			if err != nil {
				managerBroken = true
				managerErr = err
				result.Failed = append(result.Failed, spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: %w", spec, err))
				continue
			}

			if installed {
				result.Skipped = append(result.Skipped, spec)
				continue
			}

			if dryRun {
				result.WouldInstall = append(result.WouldInstall, spec)
				continue
			}

			plan = append(plan, planEntry{spec: spec, pkg: pkg, mgr: mgr})
		}
	}

	// Phase 2: execute installs with live spinner feedback.
	if len(plan) > 0 {
		sm := output.NewSpinnerManager(len(plan))
		for _, p := range plan {
			spinner := sm.StartSpinner("Installing", p.spec)
			err := callWithTimeoutVoid(ctx, func(c context.Context) error {
				return p.mgr.Install(c, p.pkg)
			})
			if err != nil {
				spinner.Error(fmt.Sprintf("%s: %s", p.spec, err.Error()))
				result.Failed = append(result.Failed, p.spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: %w", p.spec, err))
				continue
			}
			spinner.Success(fmt.Sprintf("installed %s", p.spec))
			result.Installed = append(result.Installed, p.spec)
		}
	}

	// Return error if any packages failed
	if len(result.Failed) > 0 {
		return result, fmt.Errorf("%d package(s) failed to install", len(result.Failed))
	}

	return result, nil
}

// callWithTimeout runs fn with a per-call timeout derived from PerPackageTimeout,
// inheriting cancellation from the parent context.
func callWithTimeout[T any](ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	c, cancel := context.WithTimeout(ctx, PerPackageTimeout)
	defer cancel()
	return fn(c)
}

func callWithTimeoutVoid(ctx context.Context, fn func(context.Context) error) error {
	c, cancel := context.WithTimeout(ctx, PerPackageTimeout)
	defer cancel()
	return fn(c)
}
