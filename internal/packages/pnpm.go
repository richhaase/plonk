// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// PNPMSimple implements Manager for pnpm
type PNPMSimple struct {
	mu        sync.Mutex
	installed map[string]bool
}

// NewPNPMSimple creates a new pnpm manager
func NewPNPMSimple() *PNPMSimple {
	return &PNPMSimple{}
}

// IsInstalled checks if a package is globally installed via pnpm
func (p *PNPMSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Load installed list on first call
	if p.installed == nil {
		if err := p.loadInstalled(ctx); err != nil {
			return false, err
		}
	}

	return p.installed[name], nil
}

// loadInstalled fetches all globally installed pnpm packages
func (p *PNPMSimple) loadInstalled(ctx context.Context) error {
	installed := make(map[string]bool)

	cmd := exec.CommandContext(ctx, "pnpm", "list", "-g", "--depth=0", "--json")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list pnpm packages: %w", err)
	}

	// pnpm outputs JSON array: [{"dependencies": {...}}]
	var result []struct {
		Dependencies map[string]any `json:"dependencies"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("failed to parse pnpm output: %w", err)
	}

	for _, item := range result {
		for name := range item.Dependencies {
			installed[name] = true
		}
	}

	// Only set the cache after successful loading
	p.installed = installed
	return nil
}

// Install installs a package globally via pnpm
func (p *PNPMSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "pnpm", "add", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed
		if strings.Contains(strings.ToLower(string(output)), "already installed") {
			p.markInstalled(name)
			return nil
		}
		return fmt.Errorf("pnpm add -g %s: %s: %w", name, strings.TrimSpace(string(output)), err)
	}

	// Update cache after successful install
	p.markInstalled(name)
	return nil
}

// markInstalled updates the cache to mark a package as installed
func (p *PNPMSimple) markInstalled(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.installed != nil {
		p.installed[name] = true
	}
}
