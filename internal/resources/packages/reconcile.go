// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os"

	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources"
)

// Reconcile performs package reconciliation (backward compatibility)
func Reconcile(ctx context.Context, configDir string) (resources.Result, error) {
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
	packageResource := NewMultiPackageResource()

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
