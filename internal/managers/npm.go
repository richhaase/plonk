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

// IsAvailable checks if NPM is installed and accessible.
func (n *NpmManager) IsAvailable(ctx context.Context) (bool, error) {
	_, err := exec.LookPath("npm")
	if err != nil {
		// Binary not found in PATH - this is not an error condition
		return false, nil
	}
	
	// Verify npm is actually functional by running a simple command
	cmd := exec.CommandContext(ctx, "npm", "--version")
	err = cmd.Run()
	if err != nil {
		// npm exists but is not functional - this is an error
		return false, fmt.Errorf("npm binary found but not functional: %w", err)
	}
	
	return true, nil
}

// ListInstalled lists all globally installed NPM packages.
func (n *NpmManager) ListInstalled(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", "--depth=0", "--parseable")
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if exitError, ok := err.(*exec.ExitError); ok {
			// npm list can return non-zero exit codes even when working correctly
			// (e.g., when there are peer dependency warnings)
			// Only treat it as an error if the exit code indicates a real failure
			if exitError.ExitCode() > 1 {
				return nil, fmt.Errorf("failed to list npm packages: %w", err)
			}
			// Exit code 1 might just be warnings - continue with parsing
		} else {
			// Non-exit errors (e.g., command not found, context cancellation)
			return nil, fmt.Errorf("failed to execute npm list: %w", err)
		}
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		// No packages installed - this is normal, not an error
		return []string{}, nil
	}

	// Parse output to extract package names
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
func (n *NpmManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", name)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Package not found - this is not an error condition
			return false, nil
		}
		// Real error (npm not found, permission issues, etc.)
		return false, fmt.Errorf("failed to check package %s: %w", name, err)
	}
	return true, nil
}

