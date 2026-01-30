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

// UVSimple implements Manager for uv (Python)
type UVSimple struct {
	mu        sync.Mutex
	installed map[string]bool
}

// NewUVSimple creates a new uv manager
func NewUVSimple() *UVSimple {
	return &UVSimple{}
}

// IsInstalled checks if a tool is installed via uv
func (u *UVSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Load installed list on first call
	if u.installed == nil {
		if err := u.loadInstalled(ctx); err != nil {
			return false, err
		}
	}

	return u.installed[name], nil
}

// loadInstalled fetches all installed uv tools
func (u *UVSimple) loadInstalled(ctx context.Context) error {
	u.installed = make(map[string]bool)

	cmd := exec.CommandContext(ctx, "uv", "tool", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list uv tools: %w", err)
	}

	// Parse output: tool names are first token on each line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			u.installed[fields[0]] = true
		}
	}

	return nil
}

// Install installs a tool via uv
func (u *UVSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "uv", "tool", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed
		if strings.Contains(strings.ToLower(string(output)), "already installed") {
			return nil
		}
		return fmt.Errorf("uv tool install %s: %s: %w", name, strings.TrimSpace(string(output)), err)
	}
	return nil
}
