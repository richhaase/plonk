// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"sort"

	"github.com/richhaase/plonk/internal/lock"
)

// SimpleApplyResult holds the result of applying packages
type SimpleApplyResult struct {
	Installed    []string // Packages that were actually installed
	WouldInstall []string // Packages that would be installed (dry-run only)
	Skipped      []string // Packages already installed
	Failed       []string // Packages that failed to install
	Errors       []error  // Errors for failed packages
}

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

	// Process each manager in sorted order
	for _, manager := range managers {
		pkgs := lockFile.Packages[manager]
		mgr, err := GetManager(manager)
		if err != nil {
			// Record failure for each package individually
			for _, pkg := range pkgs {
				spec := manager + ":" + pkg
				result.Failed = append(result.Failed, spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: manager not available: %w", spec, err))
			}
			continue
		}

		for _, pkg := range pkgs {
			spec := manager + ":" + pkg

			// Check if installed
			installed, err := mgr.IsInstalled(ctx, pkg)
			if err != nil {
				result.Failed = append(result.Failed, spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: %w", spec, err))
				continue
			}

			if installed {
				result.Skipped = append(result.Skipped, spec)
				continue
			}

			// Install (or mark as would-install for dry-run)
			if dryRun {
				result.WouldInstall = append(result.WouldInstall, spec)
				continue
			}

			if err := mgr.Install(ctx, pkg); err != nil {
				result.Failed = append(result.Failed, spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: %w", spec, err))
				continue
			}

			result.Installed = append(result.Installed, spec)
		}
	}

	// Return error if any packages failed
	if len(result.Failed) > 0 {
		return result, fmt.Errorf("%d package(s) failed to install", len(result.Failed))
	}

	return result, nil
}
