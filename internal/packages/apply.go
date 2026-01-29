// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/lock"
)

// SimpleApplyResult holds the result of applying packages
type SimpleApplyResult struct {
	Installed []string
	Skipped   []string
	Failed    []string
	Errors    []error
}

// SimpleApply installs all tracked packages that are missing
func SimpleApply(ctx context.Context, configDir string, dryRun bool) (*SimpleApplyResult, error) {
	lockSvc := lock.NewLockV3Service(configDir)
	lockFile, err := lockSvc.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	result := &SimpleApplyResult{}

	// Process each manager
	for manager, pkgs := range lockFile.Packages {
		mgr, err := GetManager(manager)
		if err != nil {
			result.Failed = append(result.Failed, manager+":*")
			result.Errors = append(result.Errors, fmt.Errorf("manager %s: %w", manager, err))
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

			// Install
			if dryRun {
				fmt.Printf("Would install %s\n", spec)
				result.Installed = append(result.Installed, spec)
				continue
			}

			fmt.Printf("Installing %s...\n", spec)
			if err := mgr.Install(ctx, pkg); err != nil {
				result.Failed = append(result.Failed, spec)
				result.Errors = append(result.Errors, fmt.Errorf("%s: %w", spec, err))
				continue
			}

			result.Installed = append(result.Installed, spec)
		}
	}

	return result, nil
}
