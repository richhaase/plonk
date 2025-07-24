// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/state"
)

// ReconcileDotfiles performs dotfile reconciliation without SharedContext
func ReconcileDotfiles(ctx context.Context, homeDir, configDir string) (state.Result, error) {
	// Create fresh provider (no caching)
	cfg := config.LoadConfigWithDefaults(configDir)
	dotfileConfigLoader := state.NewConfigBasedDotfileLoader(cfg.IgnorePatterns, cfg.ExpandDirectories)
	provider := state.NewDotfileProvider(homeDir, configDir, dotfileConfigLoader)

	// Get items
	configured, err := provider.GetConfiguredItems()
	if err != nil {
		return state.Result{}, err
	}

	actual, err := provider.GetActualItems(ctx)
	if err != nil {
		return state.Result{}, err
	}

	// Reconcile (copy exact logic from runtime/context_simple.go)
	return reconcileDotfileItems(provider, configured, actual), nil
}

// ReconcilePackages performs package reconciliation without SharedContext
func ReconcilePackages(ctx context.Context, configDir string) (state.Result, error) {
	// Create fresh providers (no caching)
	lockService := lock.NewYAMLLockService(configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)
	registry := managers.NewManagerRegistry()
	provider, err := registry.CreateMultiProvider(ctx, lockAdapter)
	if err != nil {
		return state.Result{}, err
	}

	// Get items
	configured, err := provider.GetConfiguredItems()
	if err != nil {
		return state.Result{}, err
	}

	actual, err := provider.GetActualItems(ctx)
	if err != nil {
		return state.Result{}, err
	}

	// Reconcile (copy exact logic from runtime/context_simple.go)
	return reconcilePackageItems(provider, configured, actual), nil
}

// ReconcileAll reconciles all domains
func ReconcileAll(ctx context.Context, homeDir, configDir string) (map[string]state.Result, error) {
	results := make(map[string]state.Result)

	// Reconcile dotfiles
	dotfileResult, err := ReconcileDotfiles(ctx, homeDir, configDir)
	if err != nil {
		return nil, err
	}
	results["dotfile"] = dotfileResult

	// Reconcile packages
	packageResult, err := ReconcilePackages(ctx, configDir)
	if err != nil {
		return nil, err
	}
	results["package"] = packageResult

	return results, nil
}

// reconcileDotfileItems performs the core reconciliation logic for dotfiles
func reconcileDotfileItems(provider *state.DotfileProvider, configured []state.ConfigItem, actual []state.ActualItem) state.Result {
	// Build lookup sets
	actualSet := make(map[string]*state.ActualItem)
	for i := range actual {
		actualSet[actual[i].Name] = &actual[i]
	}

	configuredSet := make(map[string]*state.ConfigItem)
	for i := range configured {
		configuredSet[configured[i].Name] = &configured[i]
	}

	result := state.Result{
		Domain:    "dotfile",
		Managed:   make([]state.Item, 0),
		Missing:   make([]state.Item, 0),
		Untracked: make([]state.Item, 0),
	}

	// Check each configured item against actual
	for _, configItem := range configured {
		if actualItem, exists := actualSet[configItem.Name]; exists {
			// Item is managed (in config AND present)
			item := provider.CreateItem(configItem.Name, state.StateManaged, &configItem, actualItem)
			result.Managed = append(result.Managed, item)
		} else {
			// Item is missing (in config BUT not present)
			item := provider.CreateItem(configItem.Name, state.StateMissing, &configItem, nil)
			result.Missing = append(result.Missing, item)
		}
	}

	// Check each actual item against configured
	for _, actualItem := range actual {
		if _, exists := configuredSet[actualItem.Name]; !exists {
			// Item is untracked (present BUT not in config)
			item := provider.CreateItem(actualItem.Name, state.StateUntracked, nil, &actualItem)
			result.Untracked = append(result.Untracked, item)
		}
	}

	return result
}

// reconcilePackageItems performs the core reconciliation logic for packages
func reconcilePackageItems(provider *managers.MultiManagerPackageProvider, configured []state.ConfigItem, actual []state.ActualItem) state.Result {
	// Build lookup sets
	actualSet := make(map[string]*state.ActualItem)
	for i := range actual {
		actualSet[actual[i].Name] = &actual[i]
	}

	configuredSet := make(map[string]*state.ConfigItem)
	for i := range configured {
		configuredSet[configured[i].Name] = &configured[i]
	}

	result := state.Result{
		Domain:    "package",
		Managed:   make([]state.Item, 0),
		Missing:   make([]state.Item, 0),
		Untracked: make([]state.Item, 0),
	}

	// Check each configured item against actual
	for _, configItem := range configured {
		if actualItem, exists := actualSet[configItem.Name]; exists {
			// Item is managed (in config AND present)
			item := provider.CreateItem(configItem.Name, state.StateManaged, &configItem, actualItem)
			result.Managed = append(result.Managed, item)
		} else {
			// Item is missing (in config BUT not present)
			item := provider.CreateItem(configItem.Name, state.StateMissing, &configItem, nil)
			result.Missing = append(result.Missing, item)
		}
	}

	// Check each actual item against configured
	for _, actualItem := range actual {
		if _, exists := configuredSet[actualItem.Name]; !exists {
			// Item is untracked (present BUT not in config)
			item := provider.CreateItem(actualItem.Name, state.StateUntracked, nil, &actualItem)
			result.Untracked = append(result.Untracked, item)
		}
	}

	return result
}
