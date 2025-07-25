// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// Orchestrator manages resources and their lock file state
type Orchestrator struct {
	ctx       context.Context
	config    *config.Config
	lock      lock.LockWriter
	configDir string
	homeDir   string
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(ctx context.Context, cfg *config.Config, configDir, homeDir string) *Orchestrator {
	lockService := lock.NewYAMLLockService(configDir)

	return &Orchestrator{
		ctx:       ctx,
		config:    cfg,
		lock:      lockService,
		configDir: configDir,
		homeDir:   homeDir,
	}
}

// GetResources returns all configured resources
func (o *Orchestrator) GetResources() []resources.Resource {
	var resourceList []resources.Resource

	// Add package resource
	packageResource := packages.NewMultiPackageResource()

	// Get configured packages from existing lock file for now
	// TODO: In future, get from config directly
	lockService := lock.NewYAMLLockService(o.configDir)
	lockFile, err := lockService.Load()
	if err == nil {
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
	}
	resourceList = append(resourceList, packageResource)

	// Add dotfile resource
	manager := dotfiles.NewManager(o.homeDir, o.configDir)
	dotfileResource := dotfiles.NewDotfileResource(manager)

	// Get configured dotfiles
	configured, err := manager.GetConfiguredDotfiles()
	if err == nil {
		dotfileResource.SetDesired(configured)
	}
	resourceList = append(resourceList, dotfileResource)

	return resourceList
}

// Sync orchestrates the synchronization of all resources
func (o *Orchestrator) Sync() error {
	resourceList := o.GetResources()

	// Reconcile all resources
	results, err := ReconcileResources(o.ctx, resourceList)
	if err != nil {
		return fmt.Errorf("reconciling resources: %w", err)
	}

	// Apply changes for all resources
	for _, resource := range resourceList {
		reconciled := results[resource.ID()]

		// Apply missing items
		for _, item := range reconciled {
			if item.State == resources.StateMissing {
				if err := resource.Apply(o.ctx, item); err != nil {
					return fmt.Errorf("applying %s resource item %s: %w", resource.ID(), item.Name, err)
				}
			}
		}
	}

	// Write updated lock file
	return o.writeLock(resourceList)
}

// writeLock writes the current resource state to the lock file
func (o *Orchestrator) writeLock(resourceList []resources.Resource) error {
	lockData := &lock.LockData{
		Version:   lock.CurrentVersion,
		Packages:  make(map[string][]lock.Package),
		Resources: []lock.ResourceEntry{},
	}

	for _, res := range resourceList {
		items := res.Actual(o.ctx)
		for _, item := range items {
			if item.State == resources.StateManaged {
				entry := lock.ResourceEntry{
					Type:        res.ID(),
					ID:          item.Name,
					State:       "managed",
					InstalledAt: time.Now().Format(time.RFC3339),
					Metadata:    make(map[string]interface{}),
				}

				// Add resource-specific metadata
				switch res.ID() {
				case "package":
					if item.Manager != "" {
						entry.ID = fmt.Sprintf("%s:%s", item.Manager, item.Name)
						entry.Metadata["manager"] = item.Manager
						entry.Metadata["name"] = item.Name
						if version, ok := item.Metadata["version"].(string); ok {
							entry.Metadata["version"] = version
						}

						// Also maintain backward compatibility in packages section
						if lockData.Packages[item.Manager] == nil {
							lockData.Packages[item.Manager] = []lock.Package{}
						}
						lockData.Packages[item.Manager] = append(lockData.Packages[item.Manager], lock.Package{
							Name:        item.Name,
							Version:     item.Metadata["version"].(string),
							InstalledAt: entry.InstalledAt,
						})
					}
				case "dotfile":
					if item.Path != "" {
						entry.Metadata["source"] = item.Path
						entry.Metadata["target"] = item.Name
					}
				}

				lockData.Resources = append(lockData.Resources, entry)
			}
		}
	}

	return o.lock.Write(lockData)
}

// SyncWithLockUpdate performs sync and updates lock file to v2 format
// This method bridges the gap between existing commands and new v2 format
func SyncWithLockUpdate(ctx context.Context, configDir, homeDir string, cfg *config.Config) error {
	orchestrator := NewOrchestrator(ctx, cfg, configDir, homeDir)
	return orchestrator.Sync()
}
