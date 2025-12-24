// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/packages"
)

// ReconcileAllResult contains domain-specific reconciliation results
type ReconcileAllResult struct {
	Dotfiles dotfiles.Result
	Packages packages.ReconcileResult
}

// ReconcileAllWithConfig reconciles all domains using injected config
func ReconcileAllWithConfig(ctx context.Context, homeDir, configDir string, cfg *config.Config) (ReconcileAllResult, error) {
	var result ReconcileAllResult

	// Dotfiles with injected config
	dotfileResult, err := dotfiles.ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return ReconcileAllResult{}, err
	}
	result.Dotfiles = dotfileResult

	// Packages
	packageResult, err := packages.ReconcileWithConfig(ctx, configDir, cfg)
	if err != nil {
		return ReconcileAllResult{}, err
	}
	result.Packages = packageResult

	return result, nil
}
