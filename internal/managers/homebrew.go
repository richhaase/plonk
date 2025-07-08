// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// HomebrewManager manages Homebrew packages.
type HomebrewManager struct{}

// NewHomebrewManager creates a new Homebrew manager.
func NewHomebrewManager() *HomebrewManager {
	return &HomebrewManager{}
}

// IsAvailable checks if Homebrew is installed and accessible.
func (h *HomebrewManager) IsAvailable(ctx context.Context) (bool, error) {
	_, err := exec.LookPath("brew")
	if err != nil {
		// Binary not found in PATH - this is not an error condition
		return false, nil
	}
	
	// Verify brew is actually functional by running a simple command
	cmd := exec.CommandContext(ctx, "brew", "--version")
	err = cmd.Run()
	if err != nil {
		// brew exists but is not functional - this is an error
		return false, fmt.Errorf("brew binary found but not functional: %w", err)
	}
	
	return true, nil
}

// ListInstalled lists all installed Homebrew packages.
func (h *HomebrewManager) ListInstalled(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "brew", "list")
	output, err := cmd.Output()
	if err != nil {
		// Check if this is a real error vs expected conditions
		if _, ok := err.(*exec.ExitError); ok {
			// For brew list, any non-zero exit usually indicates a real problem
			// (brew list returns exit 0 even with no packages installed)
			return nil, fmt.Errorf("failed to list homebrew packages: %w", err)
		}
		// Non-exit errors (e.g., command not found, context cancellation)
		return nil, fmt.Errorf("failed to execute brew list: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		// No packages installed - this is normal, not an error
		return []string{}, nil
	}

	// Parse output into package list
	packages := strings.Split(result, "\n")
	// Filter out any empty strings that might result from parsing
	var filteredPackages []string
	for _, pkg := range packages {
		if trimmed := strings.TrimSpace(pkg); trimmed != "" {
			filteredPackages = append(filteredPackages, trimmed)
		}
	}

	return filteredPackages, nil
}

// Install installs a Homebrew package.
func (h *HomebrewManager) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "brew", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// Uninstall removes a Homebrew package.
func (h *HomebrewManager) Uninstall(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "brew", "uninstall", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// IsInstalled checks if a specific package is installed.
func (h *HomebrewManager) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "brew", "list", name)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Package not found - this is not an error condition
			return false, nil
		}
		// Real error (brew not found, permission issues, etc.)
		return false, fmt.Errorf("failed to check package %s: %w", name, err)
	}
	return true, nil
}

