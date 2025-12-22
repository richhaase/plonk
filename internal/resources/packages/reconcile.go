// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources"
)

// ReconcileWithConfig performs package reconciliation with injected config
func ReconcileWithConfig(ctx context.Context, configDir string, cfg *config.Config) (resources.Result, error) {
	// Get configured packages from lock file
	lockService := lock.NewYAMLLockService(configDir)
	lockData, err := lockService.Read()
	if err != nil {
		return resources.Result{}, err
	}

	// Create multi-package resource
	cfgUsed := cfg
	if cfgUsed == nil {
		cfgUsed = config.LoadWithDefaults(configDir)
	}
	packageResource := NewMultiPackageResource(cfgUsed)

	// Convert lock file entries to desired items
	desired := make([]resources.Item, 0)
	for _, resource := range lockData.Resources {
		if resource.Type == "package" {
			// Extract manager and name from metadata
			manager, _ := resource.Metadata["manager"].(string)
			name, _ := resource.Metadata["name"].(string)
			version, _ := resource.Metadata["version"].(string)

			desired = append(desired, resources.Item{
				Name:    name,
				Domain:  "package",
				Manager: manager,
				Metadata: map[string]interface{}{
					"version": version,
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
