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

// ReconcileResource performs reconciliation for a Resource interface
func ReconcileResource(ctx context.Context, resource resources.Resource) ([]resources.Item, error) {
	desired := resource.Desired()
	actual := resource.Actual(ctx)

	// Use the reconciliation helper from resources package
	reconciled := resources.ReconcileItems(desired, actual)
	return reconciled, nil
}

// ReconcileResources performs reconciliation for multiple resources
func ReconcileResources(ctx context.Context, resourceList []resources.Resource) (map[string][]resources.Item, error) {
	results := make(map[string][]resources.Item)

	for _, resource := range resourceList {
		reconciled, err := ReconcileResource(ctx, resource)
		if err != nil {
			return nil, fmt.Errorf("reconciling resource %s: %w", resource.ID(), err)
		}
		results[resource.ID()] = reconciled
	}

	return results, nil
}

// ReconcileDotfiles performs dotfile reconciliation (backward compatibility)
func ReconcileDotfiles(ctx context.Context, homeDir, configDir string) (resources.Result, error) {
	// Create dotfile resource
	manager := dotfiles.NewManager(homeDir, configDir)
	dotfileResource := dotfiles.NewDotfileResource(manager)

	// Get configured dotfiles and set as desired
	configured, err := manager.GetConfiguredDotfiles()
	if err != nil {
		return resources.Result{}, err
	}
	dotfileResource.SetDesired(configured)

	// Reconcile using the resource interface
	reconciled, err := ReconcileResource(ctx, dotfileResource)
	if err != nil {
		return resources.Result{}, err
	}

	// Convert to Result format for backward compatibility
	managed, missing, untracked := resources.GroupItemsByState(reconciled)
	return resources.Result{
		Domain:    "dotfile",
		Managed:   managed,
		Missing:   missing,
		Untracked: untracked,
	}, nil
}

// ReconcilePackages performs package reconciliation (backward compatibility)
func ReconcilePackages(ctx context.Context, configDir string) (resources.Result, error) {
	// Get configured packages from lock file
	lockService := lock.NewYAMLLockService(configDir)
	lockFile, err := lockService.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			return resources.Result{}, err
		}
		// No lock file means no configured packages
		lockFile = &lock.LockFile{
			Version:  1,
			Packages: make(map[string][]lock.PackageEntry),
		}
	}

	// Create multi-package resource
	packageResource := packages.NewMultiPackageResource()

	// Convert lock file entries to desired items
	desired := make([]resources.Item, 0)
	for manager, pkgs := range lockFile.Packages {
		for _, pkg := range pkgs {
			desired = append(desired, resources.Item{
				Name:    pkg.Name,
				Domain:  "package",
				Manager: manager,
				Metadata: map[string]interface{}{
					"version": pkg.Version,
				},
			})
		}
	}
	packageResource.SetDesired(desired)

	// Use custom key function for package reconciliation
	keyFunc := func(item resources.Item) string {
		return item.Manager + ":" + item.Name
	}
	reconciled := resources.ReconcileItemsWithKey(packageResource.Desired(), packageResource.Actual(ctx), keyFunc)

	// Convert to Result format for backward compatibility
	managed, missing, untracked := resources.GroupItemsByState(reconciled)
	return resources.Result{
		Domain:    "package",
		Managed:   managed,
		Missing:   missing,
		Untracked: untracked,
	}, nil
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
