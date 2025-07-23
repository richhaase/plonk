// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package core contains the core business logic for plonk.
// This package should never import from internal/commands or internal/cli.
package core

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/runtime"
	"github.com/richhaase/plonk/internal/state"
)

// LoadOrCreateConfig loads existing config or creates a new one
func LoadOrCreateConfig(configDir string) (*config.Config, error) {
	manager := config.NewConfigManager(configDir)
	return manager.LoadOrCreate()
}

// CreatePackageProvider creates a multi-manager package provider using lock file
func CreatePackageProvider(ctx context.Context, configDir string) (*state.MultiManagerPackageProvider, error) {
	// Use SharedContext to create provider
	sharedCtx := runtime.GetSharedContext()
	return sharedCtx.CreatePackageProvider(ctx)
}

// CreateDotfileProvider creates a dotfile provider
func CreateDotfileProvider(homeDir string, configDir string, cfg *config.Config) *state.DotfileProvider {
	// Use SharedContext to create provider
	sharedCtx := runtime.GetSharedContext()
	provider, _ := sharedCtx.CreateDotfileProvider()
	return provider
}
