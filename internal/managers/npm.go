// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// NpmManager manages NPM packages.
type NpmManager struct{}

// NewNpmManager creates a new NPM manager.
func NewNpmManager() *NpmManager {
	return &NpmManager{}
}

// IsAvailable checks if NPM is installed.
func (n *NpmManager) IsAvailable(ctx context.Context) bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

// ListInstalled lists all globally installed NPM packages.
func (n *NpmManager) ListInstalled(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", "--depth=0", "--parseable")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}

	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, "/") {
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				pkg := parts[len(parts)-1]
				if pkg != "" && pkg != "lib" {
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages, nil
}

// Install installs a global NPM package.
func (n *NpmManager) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "npm", "install", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", name, err, string(output))
	}
	
	return nil
}

// Uninstall removes a global NPM package.
func (n *NpmManager) Uninstall(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "npm", "uninstall", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall %s: %w\nOutput: %s", name, err, string(output))
	}
	
	return nil
}

// IsInstalled checks if a specific package is installed globally.
func (n *NpmManager) IsInstalled(ctx context.Context, name string) bool {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", name)
	err := cmd.Run()
	return err == nil
}

