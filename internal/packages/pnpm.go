// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PNPMSimple implements Manager for pnpm
type PNPMSimple struct{}

// NewPNPMSimple creates a new pnpm manager
func NewPNPMSimple() *PNPMSimple {
	return &PNPMSimple{}
}

// IsInstalled checks if a package is globally installed via pnpm
func (p *PNPMSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "pnpm", "list", "-g", "--depth=0", "--json")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list pnpm packages: %w", err)
	}

	// pnpm outputs JSON array: [{"dependencies": {...}}]
	var result []struct {
		Dependencies map[string]any `json:"dependencies"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return false, fmt.Errorf("failed to parse pnpm output: %w", err)
	}

	for _, item := range result {
		if _, ok := item.Dependencies[name]; ok {
			return true, nil
		}
	}
	return false, nil
}

// Install installs a package globally via pnpm
func (p *PNPMSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "pnpm", "add", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed
		if strings.Contains(strings.ToLower(string(output)), "already installed") {
			return nil
		}
		return err
	}
	return nil
}
