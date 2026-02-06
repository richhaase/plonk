// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// BrewSimple implements Manager for Homebrew
type BrewSimple struct {
	mu        sync.Mutex
	installed map[string]bool
}

// NewBrewSimple creates a new Homebrew manager
func NewBrewSimple() *BrewSimple {
	return &BrewSimple{}
}

// IsInstalled checks if a package is installed via brew
func (b *BrewSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Load installed list on first call
	if b.installed == nil {
		if err := b.loadInstalled(ctx); err != nil {
			return false, err
		}
	}

	return b.installed[name], nil
}

// loadInstalled fetches all installed formulas and casks
func (b *BrewSimple) loadInstalled(ctx context.Context) error {
	installed := make(map[string]bool)

	// Get formulas
	cmd := exec.CommandContext(ctx, "brew", "list", "--formula", "-1")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list brew formulas: %w", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			installed[line] = true
		}
	}

	// Get casks — failure is non-fatal (cask support may be unavailable, e.g., on Linux)
	cmd = exec.CommandContext(ctx, "brew", "list", "--cask", "-1")
	output, err = cmd.Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			if line != "" {
				installed[line] = true
			}
		}
	}

	// Set cache with whatever we loaded (formulas always, casks if available)
	b.installed = installed
	return nil
}

// Install installs a package via brew
func (b *BrewSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "brew", "install", "--", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed (idempotent)
		if strings.Contains(strings.ToLower(string(output)), "already installed") {
			b.markInstalled(name)
			return nil
		}
		return fmt.Errorf("brew install %s: %s: %w", name, strings.TrimSpace(string(output)), err)
	}

	// Update cache after successful install
	b.markInstalled(name)
	return nil
}

// markInstalled updates the cache to mark a package as installed
func (b *BrewSimple) markInstalled(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.installed != nil {
		b.installed[name] = true
	}
}
