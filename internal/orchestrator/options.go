// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"github.com/richhaase/plonk/internal/config"
)

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

// WithPackagesOnly applies packages only
func WithPackagesOnly(packagesOnly bool) Option {
	return func(o *Orchestrator) {
		o.packagesOnly = packagesOnly
	}
}

// WithDotfilesOnly applies dotfiles only
func WithDotfilesOnly(dotfilesOnly bool) Option {
	return func(o *Orchestrator) {
		o.dotfilesOnly = dotfilesOnly
	}
}
