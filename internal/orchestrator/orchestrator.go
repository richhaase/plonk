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
	ctx        context.Context
	config     *config.Config
	lock       lock.LockWriter
	configDir  string
	homeDir    string
	hookRunner *HookRunner
	dryRun     bool
}

// Option is a functional option for configuring the orchestrator
type Option func(*Orchestrator)

// WithConfig sets the configuration
func WithConfig(cfg *config.Config) Option {
	return func(o *Orchestrator) {
		o.config = cfg
	}
}

// WithConfigDir sets the config directory
func WithConfigDir(dir string) Option {
	return func(o *Orchestrator) {
		o.configDir = dir
	}
}

// WithHomeDir sets the home directory
func WithHomeDir(dir string) Option {
	return func(o *Orchestrator) {
		o.homeDir = dir
	}
}

// WithDryRun enables dry-run mode
func WithDryRun(dryRun bool) Option {
	return func(o *Orchestrator) {
		o.dryRun = dryRun
	}
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	DryRun   bool        `json:"dry_run" yaml:"dry_run"`
	Success  bool        `json:"success" yaml:"success"`
	Packages interface{} `json:"packages,omitempty" yaml:"packages,omitempty"`
	Dotfiles interface{} `json:"dotfiles,omitempty" yaml:"dotfiles,omitempty"`
	Error    string      `json:"error,omitempty" yaml:"error,omitempty"`
}

// New creates a new orchestrator instance with options
func New(opts ...Option) *Orchestrator {
	o := &Orchestrator{
		hookRunner: NewHookRunner(),
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.configDir != "" {
		o.lock = lock.NewYAMLLockService(o.configDir)
	}

	return o
}

// NewOrchestrator creates a new orchestrator instance (legacy constructor)
func NewOrchestrator(ctx context.Context, cfg *config.Config, configDir, homeDir string) *Orchestrator {
	lockService := lock.NewYAMLLockService(configDir)

	return &Orchestrator{
		ctx:        ctx,
		config:     cfg,
		lock:       lockService,
		configDir:  configDir,
		homeDir:    homeDir,
		hookRunner: NewHookRunner(),
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
func (o *Orchestrator) Sync(ctx context.Context) (SyncResult, error) {
	result := SyncResult{
		DryRun:  o.dryRun,
		Success: false,
	}

	// Store context
	o.ctx = ctx

	// Run pre-sync hooks
	if o.config != nil && len(o.config.Hooks.PreSync) > 0 {
		if err := o.hookRunner.RunPreSync(ctx, o.config.Hooks.PreSync); err != nil {
			result.Error = fmt.Sprintf("pre-sync hook failed: %v", err)
			return result, fmt.Errorf("pre-sync hook failed: %w", err)
		}
	}

	// Sync packages using existing function
	packageResult, err := SyncPackages(ctx, o.configDir, o.config, o.dryRun)
	if err != nil {
		result.Error = fmt.Sprintf("package sync failed: %v", err)
		return result, fmt.Errorf("package sync failed: %w", err)
	}
	result.Packages = packageResult

	// Sync dotfiles using existing function
	dotfileResult, err := SyncDotfiles(ctx, o.configDir, o.homeDir, o.config, o.dryRun, false)
	if err != nil {
		result.Error = fmt.Sprintf("dotfile sync failed: %v", err)
		return result, fmt.Errorf("dotfile sync failed: %w", err)
	}
	result.Dotfiles = dotfileResult

	// Run post-sync hooks
	if o.config != nil && len(o.config.Hooks.PostSync) > 0 {
		if err := o.hookRunner.RunPostSync(ctx, o.config.Hooks.PostSync); err != nil {
			result.Error = fmt.Sprintf("post-sync hook failed: %v", err)
			return result, fmt.Errorf("post-sync hook failed: %w", err)
		}
	}

	result.Success = true
	return result, nil
}

// SyncLegacy orchestrates the synchronization using the original interface
func (o *Orchestrator) SyncLegacy() error {
	// Run pre-sync hooks
	if err := o.hookRunner.RunPreSync(o.ctx, o.config.Hooks.PreSync); err != nil {
		return fmt.Errorf("pre-sync hook failed: %w", err)
	}

	resourceList := o.GetResources()

	// Reconcile all resources
	results, err := resources.ReconcileResources(o.ctx, resourceList)
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
	if err := o.writeLock(resourceList); err != nil {
		return err
	}

	// Run post-sync hooks
	if err := o.hookRunner.RunPostSync(o.ctx, o.config.Hooks.PostSync); err != nil {
		return fmt.Errorf("post-sync hook failed: %w", err)
	}

	return nil
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
	return orchestrator.SyncLegacy()
}

// ReconcileAll reconciles all domains - coordination logic belongs in orchestrator
func ReconcileAll(ctx context.Context, homeDir, configDir string) (map[string]resources.Result, error) {
	results := make(map[string]resources.Result)

	// Reconcile dotfiles using domain package
	dotfileResult, err := dotfiles.Reconcile(ctx, homeDir, configDir)
	if err != nil {
		return nil, err
	}
	results["dotfile"] = dotfileResult

	// Reconcile packages using domain package
	packageResult, err := packages.Reconcile(ctx, configDir)
	if err != nil {
		return nil, err
	}
	results["package"] = packageResult

	return results, nil
}
