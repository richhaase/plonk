// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
)

// Result contains the results of dotfile reconciliation
type Result struct {
	Domain    string
	Managed   []DotfileItem
	Missing   []DotfileItem
	Untracked []DotfileItem
}

// Reconcile performs dotfile reconciliation (backward compatibility wrapper)
func Reconcile(ctx context.Context, homeDir, configDir string) (resources.Result, error) {
	cfg := config.LoadWithDefaults(configDir)
	result, err := ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return resources.Result{}, err
	}

	// Convert DotfileItem to resources.Item for backward compatibility
	return convertToResourcesResult(result), nil
}

// ReconcileWithConfig reconciles dotfiles using injected config
func ReconcileWithConfig(ctx context.Context, homeDir, configDir string, cfg *config.Config) (Result, error) {
	manager := NewManagerWithConfig(homeDir, configDir, cfg)

	// Get desired state from config
	desired, err := manager.GetConfiguredDotfiles()
	if err != nil {
		return Result{}, err
	}

	// Get actual state from filesystem
	actual, err := manager.GetActualDotfiles(ctx)
	if err != nil {
		return Result{}, err
	}

	// Reconcile desired vs actual using dotfile-specific reconciliation
	reconciled := ReconcileItems(desired, actual)

	// Group by state for the result
	managed, missing, untracked := GroupItemsByState(reconciled)

	return Result{
		Domain:    "dotfile",
		Managed:   managed,
		Missing:   missing,
		Untracked: untracked,
	}, nil
}

// convertToResourcesResult converts a dotfile Result to resources.Result for backward compatibility
func convertToResourcesResult(result Result) resources.Result {
	return resources.Result{
		Domain:    result.Domain,
		Managed:   convertItemsToResources(result.Managed),
		Missing:   convertItemsToResources(result.Missing),
		Untracked: convertItemsToResources(result.Untracked),
	}
}

// convertItemsToResources converts DotfileItem slice to resources.Item slice
func convertItemsToResources(items []DotfileItem) []resources.Item {
	result := make([]resources.Item, len(items))
	for i, item := range items {
		result[i] = convertItemToResource(item)
	}
	return result
}

// convertItemToResource converts a DotfileItem to resources.Item
func convertItemToResource(item DotfileItem) resources.Item {
	// Convert state
	var state resources.ItemState
	switch item.State {
	case StateManaged:
		state = resources.StateManaged
	case StateMissing:
		state = resources.StateMissing
	case StateUntracked:
		state = resources.StateUntracked
	case StateDegraded:
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
