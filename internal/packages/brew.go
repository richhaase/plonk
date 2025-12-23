// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"strings"
)

// BrewManager implements PackageManager for Homebrew.
type BrewManager struct {
	exec CommandExecutor
}

// NewBrewManager creates a new Homebrew manager.
func NewBrewManager(exec CommandExecutor) *BrewManager {
	if exec == nil {
		exec = defaultExecutor
	}
	return &BrewManager{exec: exec}
}

// IsAvailable checks if brew is available on the system.
func (b *BrewManager) IsAvailable(ctx context.Context) (bool, error) {
	if _, err := b.exec.LookPath("brew"); err != nil {
		return false, nil
	}

	_, err := b.exec.Execute(ctx, "brew", "--version")
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ListInstalled lists all packages installed by Homebrew.
func (b *BrewManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := b.exec.Execute(ctx, "brew", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return parseLines(output), nil
}

// Install installs a package via Homebrew (idempotent).
func (b *BrewManager) Install(ctx context.Context, name string) error {
	output, err := b.exec.CombinedOutput(ctx, "brew", "install", name)
	if err != nil {
		if isIdempotent(string(output), "already installed") {
			return nil
		}
		return fmt.Errorf("failed to install %s: %w", name, err)
	}
	return nil
}

// Uninstall removes a package via Homebrew (idempotent).
func (b *BrewManager) Uninstall(ctx context.Context, name string) error {
	output, err := b.exec.CombinedOutput(ctx, "brew", "uninstall", name)
	if err != nil {
		if isIdempotent(string(output), "no such keg") {
			return nil
		}
		return fmt.Errorf("failed to uninstall %s: %w", name, err)
	}
	return nil
}

// Upgrade upgrades packages to their latest versions.
// If packages is empty, upgrades all installed packages.
func (b *BrewManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return b.upgradeAll(ctx)
	}

	for _, pkg := range packages {
		output, err := b.exec.CombinedOutput(ctx, "brew", "upgrade", pkg)
		if err != nil {
			if isIdempotent(string(output), "already up-to-date", "already installed") {
				continue
			}
			return fmt.Errorf("failed to upgrade %s: %w", pkg, err)
		}
	}
	return nil
}

func (b *BrewManager) upgradeAll(ctx context.Context) error {
	output, err := b.exec.CombinedOutput(ctx, "brew", "upgrade")
	if err != nil {
		if isIdempotent(string(output), "already up-to-date") {
			return nil
		}
		return fmt.Errorf("failed to upgrade all packages: %w", err)
	}
	return nil
}

// SelfInstall installs Homebrew using the official installation script.
func (b *BrewManager) SelfInstall(ctx context.Context) error {
	// Check if already installed
	available, _ := b.IsAvailable(ctx)
	if available {
		return nil
	}

	// Use the official Homebrew installation script
	script := `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`
	_, err := b.exec.CombinedOutput(ctx, "sh", "-c", script)
	if err != nil {
		return fmt.Errorf("failed to install Homebrew: %w", err)
	}

	return nil
}

// parseLines splits output by newlines and returns non-empty lines.
func parseLines(data []byte) []string {
	result := strings.TrimSpace(string(data))
	if result == "" {
		return []string{}
	}

	lines := strings.Split(result, "\n")
	var packages []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Take the first whitespace-delimited token
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}
	return packages
}

// isIdempotent checks if output contains any of the given patterns (case-insensitive).
func isIdempotent(output string, patterns ...string) bool {
	outputLower := strings.ToLower(output)
	for _, pattern := range patterns {
		if strings.Contains(outputLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}
