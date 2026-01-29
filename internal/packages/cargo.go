// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CargoSimple implements Manager for Rust's Cargo
type CargoSimple struct{}

// NewCargoSimple creates a new Cargo manager
func NewCargoSimple() *CargoSimple {
	return &CargoSimple{}
}

// IsInstalled checks if a package is installed via cargo
func (c *CargoSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "cargo", "install", "--list")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list cargo packages: %w", err)
	}

	// Parse output: each installed package starts at column 0
	// Format: "package_name v1.2.3:\n    binary1\n"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return true, nil
		}
	}
	return false, nil
}

// Install installs a package via cargo
func (c *CargoSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "cargo", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed (idempotent)
		outStr := strings.ToLower(string(output))
		if strings.Contains(outStr, "already exists") || strings.Contains(outStr, "already installed") {
			return nil
		}
		return err
	}
	return nil
}
