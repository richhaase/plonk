// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/packages"
)

// ReconcileAllWithConfig reconciles all domains using injected config
func ReconcileAllWithConfig(ctx context.Context, homeDir, configDir string, cfg *config.Config) (map[string]resources.Result, error) {
	results := make(map[string]resources.Result)

	// Dotfiles with injected config - convert domain-specific result to resources.Result
	dotfileResult, err := dotfiles.ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return nil, err
	}
	results["dotfile"] = convertDotfileResultToResources(dotfileResult)

	// Packages - convert domain-specific result to resources.Result
	packageResult, err := packages.ReconcileWithConfig(ctx, configDir, cfg)
	if err != nil {
		return nil, err
	}
	results["package"] = convertPackageResultToResources(packageResult)

	return results, nil
}

// convertDotfileResultToResources converts dotfiles.Result to resources.Result
// This delegates to the dotfiles package's own conversion function
func convertDotfileResultToResources(r dotfiles.Result) resources.Result {
	return resources.Result{
		Domain:    r.Domain,
		Managed:   convertDotfileItemsToResources(r.Managed),
		Missing:   convertDotfileItemsToResources(r.Missing),
		Untracked: convertDotfileItemsToResources(r.Untracked),
	}
}

// convertDotfileItemsToResources converts DotfileItem slice to resources.Item slice
func convertDotfileItemsToResources(items []dotfiles.DotfileItem) []resources.Item {
	result := make([]resources.Item, len(items))
	for i, item := range items {
		result[i] = convertDotfileItemToResource(item)
	}
	return result
}

// convertDotfileItemToResource converts a DotfileItem to resources.Item
func convertDotfileItemToResource(item dotfiles.DotfileItem) resources.Item {
	// Convert state
	var state resources.ItemState
	switch item.State {
	case dotfiles.StateManaged:
		state = resources.StateManaged
	case dotfiles.StateMissing:
		state = resources.StateMissing
	case dotfiles.StateUntracked:
		state = resources.StateUntracked
	case dotfiles.StateDegraded:
		state = resources.StateDegraded
	}

	// Build metadata
	metadata := make(map[string]interface{})
	if item.Metadata != nil {
		for k, v := range item.Metadata {
			metadata[k] = v
		}
	}
	metadata["source"] = item.Source
	metadata["destination"] = item.Destination
	metadata["isTemplate"] = item.IsTemplate
	metadata["isDirectory"] = item.IsDirectory
	if item.CompareFunc != nil {
		metadata["compare_fn"] = item.CompareFunc
	}

	return resources.Item{
		Name:     item.Name,
		State:    state,
		Domain:   "dotfile",
		Path:     item.Destination,
		Error:    item.Error,
		Metadata: metadata,
	}
}

// convertPackageResultToResources converts packages.ReconcileResult to resources.Result
func convertPackageResultToResources(r packages.ReconcileResult) resources.Result {
	managed := make([]resources.Item, 0, len(r.Managed))
	for _, pkg := range r.Managed {
		managed = append(managed, resources.Item{
			Name:    pkg.Name,
			Manager: pkg.Manager,
			Domain:  "package",
			State:   resources.StateManaged,
		})
	}

	missing := make([]resources.Item, 0, len(r.Missing))
	for _, pkg := range r.Missing {
		missing = append(missing, resources.Item{
			Name:    pkg.Name,
			Manager: pkg.Manager,
			Domain:  "package",
			State:   resources.StateMissing,
		})
	}

	untracked := make([]resources.Item, 0, len(r.Untracked))
	for _, pkg := range r.Untracked {
		untracked = append(untracked, resources.Item{
			Name:    pkg.Name,
			Manager: pkg.Manager,
			Domain:  "package",
			State:   resources.StateUntracked,
		})
	}

	return resources.Result{
		Domain:    "package",
		Managed:   managed,
		Missing:   missing,
		Untracked: untracked,
	}
}
