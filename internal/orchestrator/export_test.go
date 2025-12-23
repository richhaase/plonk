// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/richhaase/plonk/internal/resources"
)

// ReconcileAll reconciles all domains.
// This is a test helper - production code should use ReconcileAllWithConfig.
func ReconcileAll(ctx context.Context, homeDir, configDir string) (map[string]resources.Result, error) {
	cfg := config.LoadWithDefaults(configDir)
	results := make(map[string]resources.Result)

	// Reconcile dotfiles using domain package and convert at boundary
	dotfileResult, err := dotfiles.ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return nil, err
	}
	results["dotfile"] = convertDotfileResultToResources(dotfileResult)

	// Reconcile packages using domain package and convert at boundary
	packageResult, err := packages.ReconcileWithConfig(ctx, configDir, cfg)
	if err != nil {
		return nil, err
	}
	results["package"] = convertPackageResultToResources(packageResult)

	return results, nil
}
