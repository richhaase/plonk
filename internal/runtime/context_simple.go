// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// This file contains simplified state reconciliation methods that bypass
// the generic Provider interface and Reconciler abstraction.

package runtime

import (
	"context"

	"github.com/richhaase/plonk/internal/interfaces"
	"github.com/richhaase/plonk/internal/state"
)

// SimplifiedReconcileDotfiles performs direct dotfile reconciliation without Provider interface
func (sc *SharedContext) SimplifiedReconcileDotfiles(ctx context.Context) (state.Result, error) {
	// Create the dotfile provider directly
	provider, err := sc.CreateDotfileProvider()
	if err != nil {
		return state.Result{}, err
	}

	// Get configured and actual items directly
	configured, err := provider.GetConfiguredItems()
	if err != nil {
		return state.Result{}, err
	}

	actual, err := provider.GetActualItems(ctx)
	if err != nil {
		return state.Result{}, err
	}

	// Perform reconciliation inline
	return reconcileDotfileItems(provider, configured, actual), nil
}

// SimplifiedReconcilePackages performs direct package reconciliation without Provider interface
func (sc *SharedContext) SimplifiedReconcilePackages(ctx context.Context) (state.Result, error) {
	// Create the package provider directly
	provider, err := sc.CreatePackageProvider(ctx)
	if err != nil {
		return state.Result{}, err
	}

	// Get configured and actual items directly
	configured, err := provider.GetConfiguredItems()
	if err != nil {
		return state.Result{}, err
	}

	actual, err := provider.GetActualItems(ctx)
	if err != nil {
		return state.Result{}, err
	}

	// Perform reconciliation inline
	return reconcilePackageItems(provider, configured, actual), nil
}

// reconcileDotfileItems performs the core reconciliation logic for dotfiles
func reconcileDotfileItems(provider *state.DotfileProvider, configured []interfaces.ConfigItem, actual []interfaces.ActualItem) state.Result {
	// Build lookup sets
	actualSet := make(map[string]*interfaces.ActualItem)
	for i := range actual {
		actualSet[actual[i].Name] = &actual[i]
	}

	configuredSet := make(map[string]*interfaces.ConfigItem)
	for i := range configured {
		configuredSet[configured[i].Name] = &configured[i]
	}

	result := state.Result{
		Domain:    "dotfile",
		Managed:   make([]interfaces.Item, 0),
		Missing:   make([]interfaces.Item, 0),
		Untracked: make([]interfaces.Item, 0),
	}

	// Check each configured item against actual
	for _, configItem := range configured {
		if actualItem, exists := actualSet[configItem.Name]; exists {
			// Item is managed (in config AND present)
			item := provider.CreateItem(configItem.Name, interfaces.StateManaged, &configItem, actualItem)
			result.Managed = append(result.Managed, item)
		} else {
			// Item is missing (in config BUT not present)
			item := provider.CreateItem(configItem.Name, interfaces.StateMissing, &configItem, nil)
			result.Missing = append(result.Missing, item)
		}
	}

	// Check each actual item against configured
	for _, actualItem := range actual {
		if _, exists := configuredSet[actualItem.Name]; !exists {
			// Item is untracked (present BUT not in config)
			item := provider.CreateItem(actualItem.Name, interfaces.StateUntracked, nil, &actualItem)
			result.Untracked = append(result.Untracked, item)
		}
	}

	return result
}

// reconcilePackageItems performs the core reconciliation logic for packages
func reconcilePackageItems(provider *state.MultiManagerPackageProvider, configured []interfaces.ConfigItem, actual []interfaces.ActualItem) state.Result {
	// Build lookup sets
	actualSet := make(map[string]*interfaces.ActualItem)
	for i := range actual {
		actualSet[actual[i].Name] = &actual[i]
	}

	configuredSet := make(map[string]*interfaces.ConfigItem)
	for i := range configured {
		configuredSet[configured[i].Name] = &configured[i]
	}

	result := state.Result{
		Domain:    "package",
		Managed:   make([]interfaces.Item, 0),
		Missing:   make([]interfaces.Item, 0),
		Untracked: make([]interfaces.Item, 0),
	}

	// Check each configured item against actual
	for _, configItem := range configured {
		if actualItem, exists := actualSet[configItem.Name]; exists {
			// Item is managed (in config AND present)
			item := provider.CreateItem(configItem.Name, interfaces.StateManaged, &configItem, actualItem)
			result.Managed = append(result.Managed, item)
		} else {
			// Item is missing (in config BUT not present)
			item := provider.CreateItem(configItem.Name, interfaces.StateMissing, &configItem, nil)
			result.Missing = append(result.Missing, item)
		}
	}

	// Check each actual item against configured
	for _, actualItem := range actual {
		if _, exists := configuredSet[actualItem.Name]; !exists {
			// Item is untracked (present BUT not in config)
			item := provider.CreateItem(actualItem.Name, interfaces.StateUntracked, nil, &actualItem)
			result.Untracked = append(result.Untracked, item)
		}
	}

	return result
}

// SimplifiedReconcileAll reconciles all domains without using the generic Reconciler
func (sc *SharedContext) SimplifiedReconcileAll(ctx context.Context) (map[string]state.Result, error) {
	results := make(map[string]state.Result)

	// Reconcile dotfiles directly
	dotfileResult, err := sc.SimplifiedReconcileDotfiles(ctx)
	if err != nil {
		return nil, err
	}
	results["dotfile"] = dotfileResult

	// Reconcile packages directly
	packageResult, err := sc.SimplifiedReconcilePackages(ctx)
	if err != nil {
		return nil, err
	}
	results["package"] = packageResult

	return results, nil
}
