// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// ReconcileAll reconciles all domains - coordination logic belongs in orchestrator
func ReconcileAll(ctx context.Context, homeDir, configDir string) (map[string]resources.Result, error) {
	cfg := config.LoadWithDefaults(configDir)
	results := make(map[string]resources.Result)

	// Reconcile dotfiles using domain package (backwards-compatible path)
	dotfileResult, err := dotfiles.Reconcile(ctx, homeDir, configDir)
	if err != nil {
		return nil, err
	}
	results["dotfile"] = dotfileResult

	// Reconcile packages using domain package
	packageResult, err := packages.ReconcileWithConfig(ctx, configDir, cfg)
	if err != nil {
		return nil, err
	}
	results["package"] = packageResult

	return results, nil
}

// ReconcileAllWithConfig reconciles all domains using injected config
func ReconcileAllWithConfig(ctx context.Context, homeDir, configDir string, cfg *config.Config) (map[string]resources.Result, error) {
	results := make(map[string]resources.Result)

	// Dotfiles with injected config
	dotfileResult, err := dotfiles.ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return nil, err
	}
	results["dotfile"] = dotfileResult

	// Packages unchanged (no config needed for Reconcile)
	packageResult, err := packages.ReconcileWithConfig(ctx, configDir, cfg)
	if err != nil {
		return nil, err
	}
	results["package"] = packageResult

	return results, nil
}
