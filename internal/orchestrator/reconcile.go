// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"

	"os"

	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// ReconcileItems performs generic reconciliation logic for any domain
func ReconcileItems(configured, actual []resources.Item, domain string) resources.Result {
	// Build lookup map for actual items by name
	actualMap := make(map[string]*resources.Item)
	for i := range actual {
		actualMap[actual[i].Name] = &actual[i]
	}

	// Build lookup map for configured items by name
	configuredMap := make(map[string]*resources.Item)
	for i := range configured {
		configuredMap[configured[i].Name] = &configured[i]
	}

	result := resources.Result{
		Domain:    domain,
		Managed:   make([]resources.Item, 0),
		Missing:   make([]resources.Item, 0),
		Untracked: make([]resources.Item, 0),
	}

	// Check each configured item against actual
	for _, configItem := range configured {
		item := configItem // Copy the item
		if actualItem, exists := actualMap[configItem.Name]; exists {
			// Item is managed (in config AND present)
			item.State = resources.StateManaged
			// Merge metadata from actual if needed
			if item.Metadata == nil {
				item.Metadata = actualItem.Metadata
			} else if actualItem.Metadata != nil {
				// Merge actual metadata into configured
				for k, v := range actualItem.Metadata {
					if _, exists := item.Metadata[k]; !exists {
						item.Metadata[k] = v
					}
				}
			}
			// Use actual path if available (for dotfiles)
			if item.Path == "" && actualItem.Path != "" {
				item.Path = actualItem.Path
			}
			result.Managed = append(result.Managed, item)
		} else {
			// Item is missing (in config BUT not present)
			item.State = resources.StateMissing
			result.Missing = append(result.Missing, item)
		}
	}

	// Check each actual item against configured
	for _, actualItem := range actual {
		if _, exists := configuredMap[actualItem.Name]; !exists {
			// Item is untracked (present BUT not in config)
			item := actualItem // Copy the item
			item.State = resources.StateUntracked
			result.Untracked = append(result.Untracked, item)
		}
	}

	return result
}

// ReconcileDotfiles performs dotfile reconciliation
func ReconcileDotfiles(ctx context.Context, homeDir, configDir string) (resources.Result, error) {
	// Get configured dotfiles
	configured, err := dotfiles.GetConfiguredDotfiles(homeDir, configDir)
	if err != nil {
		return resources.Result{}, fmt.Errorf("getting configured dotfiles: %w", err)
	}

	// Get actual dotfiles
	actual, err := dotfiles.GetActualDotfiles(ctx, homeDir, configDir)
	if err != nil {
		return resources.Result{}, fmt.Errorf("getting actual dotfiles: %w", err)
	}

	// Reconcile
	return ReconcileItems(configured, actual, "dotfile"), nil
}

// ReconcilePackages performs package reconciliation
func ReconcilePackages(ctx context.Context, configDir string) (resources.Result, error) {
	// Get configured packages from lock file
	lockService := lock.NewYAMLLockService(configDir)
	lockFile, err := lockService.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			return resources.Result{}, fmt.Errorf("loading lock file: %w", err)
		}
		// No lock file means no configured packages
		lockFile = &lock.LockFile{
			Version:  1,
			Packages: make(map[string][]lock.PackageEntry),
		}
	}

	configured := make([]resources.Item, 0)
	for manager, packages := range lockFile.Packages {
		for _, pkg := range packages {
			configured = append(configured, resources.Item{
				Name:    pkg.Name,
				State:   resources.StateMissing, // Will be updated during reconciliation
				Domain:  "package",
				Manager: manager,
				Metadata: map[string]interface{}{
					"version": pkg.Version,
				},
			})
		}
	}

	// Get actual packages
	actual, err := packages.GetActualPackages(ctx)
	if err != nil {
		return resources.Result{}, fmt.Errorf("getting actual packages: %w", err)
	}

	// Create maps by manager+name for proper reconciliation
	configuredByKey := make(map[string]*resources.Item)
	for i := range configured {
		key := configured[i].Manager + ":" + configured[i].Name
		configuredByKey[key] = &configured[i]
	}

	actualByKey := make(map[string]*resources.Item)
	for i := range actual {
		key := actual[i].Manager + ":" + actual[i].Name
		actualByKey[key] = &actual[i]
	}

	result := resources.Result{
		Domain:    "package",
		Managed:   make([]resources.Item, 0),
		Missing:   make([]resources.Item, 0),
		Untracked: make([]resources.Item, 0),
	}

	// Check configured against actual
	for key, configItem := range configuredByKey {
		item := *configItem // Copy
		if actualItem, exists := actualByKey[key]; exists {
			// Managed
			item.State = resources.StateManaged
			if actualItem.Metadata != nil {
				if item.Metadata == nil {
					item.Metadata = actualItem.Metadata
				} else {
					// Merge actual version if available
					if v, ok := actualItem.Metadata["version"]; ok {
						item.Metadata["actualVersion"] = v
					}
				}
			}
			result.Managed = append(result.Managed, item)
		} else {
			// Missing
			item.State = resources.StateMissing
			result.Missing = append(result.Missing, item)
		}
	}

	// Check actual against configured
	for key, actualItem := range actualByKey {
		if _, exists := configuredByKey[key]; !exists {
			// Untracked
			item := *actualItem
			item.State = resources.StateUntracked
			result.Untracked = append(result.Untracked, item)
		}
	}

	return result, nil
}

// ReconcileAll reconciles all domains
func ReconcileAll(ctx context.Context, homeDir, configDir string) (map[string]resources.Result, error) {
	results := make(map[string]resources.Result)

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
