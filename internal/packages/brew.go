// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
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
	b.installed = make(map[string]bool)

	// Get formulas
	cmd := exec.CommandContext(ctx, "brew", "list", "--formula", "-1")
	output, err := cmd.Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			if line != "" {
				b.installed[line] = true
			}
		}
	}

	// Get casks
	cmd = exec.CommandContext(ctx, "brew", "list", "--cask", "-1")
	output, err = cmd.Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			if line != "" {
				b.installed[line] = true
			}
		}
	}

	return nil
}

// Install installs a package via brew
func (b *BrewSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "brew", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed (idempotent)
		if strings.Contains(strings.ToLower(string(output)), "already installed") {
			return nil
		}
		return err
	}
	return nil
}
