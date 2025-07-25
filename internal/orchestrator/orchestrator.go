// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"

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
