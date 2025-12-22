// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"strings"
)

// GemManager implements PackageManager for Ruby gems.
type GemManager struct {
	exec CommandExecutor
}

// NewGemManager creates a new Ruby gem manager.
func NewGemManager(exec CommandExecutor) *GemManager {
	if exec == nil {
		exec = defaultExecutor
	}
	return &GemManager{exec: exec}
}

// IsAvailable checks if gem is available on the system.
func (g *GemManager) IsAvailable(ctx context.Context) (bool, error) {
	if _, err := g.exec.LookPath("gem"); err != nil {
		return false, nil
	}

	_, err := g.exec.Execute(ctx, "gem", "--version")
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ListInstalled lists all packages installed by gem.
// Output format: "package_name (version, version2)"
func (g *GemManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := g.exec.Execute(ctx, "gem", "list", "--no-versions")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return parseGemList(output), nil
}

// Install installs a package via gem (idempotent).
// Uses --user-install to install to user's home directory (avoids permission issues on Linux).
func (g *GemManager) Install(ctx context.Context, name string) error {
	output, err := g.exec.CombinedOutput(ctx, "gem", "install", "--user-install", name)
	if err != nil {
		if isIdempotent(string(output), "already installed") {
			return nil
		}
		return fmt.Errorf("failed to install %s: %w", name, err)
	}
	return nil
}

// Uninstall removes a package via gem (idempotent).
// Uses -x to remove executables and -a to remove all versions.
func (g *GemManager) Uninstall(ctx context.Context, name string) error {
	output, err := g.exec.CombinedOutput(ctx, "gem", "uninstall", name, "-x", "-a")
	if err != nil {
		if isIdempotent(string(output), "is not installed", "not installed") {
			return nil
		}
		return fmt.Errorf("failed to uninstall %s: %w", name, err)
	}
	return nil
}

// Upgrade upgrades packages to their latest versions.
// If packages is empty, upgrades all installed packages.
func (g *GemManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return g.upgradeAll(ctx)
	}

	for _, pkg := range packages {
		output, err := g.exec.CombinedOutput(ctx, "gem", "update", pkg)
		if err != nil {
			if isIdempotent(string(output), "already up-to-date", "nothing to update") {
				continue
			}
			return fmt.Errorf("failed to upgrade %s: %w", pkg, err)
		}
	}
	return nil
}

func (g *GemManager) upgradeAll(ctx context.Context) error {
	output, err := g.exec.CombinedOutput(ctx, "gem", "update")
	if err != nil {
		if isIdempotent(string(output), "nothing to update") {
			return nil
		}
		return fmt.Errorf("failed to upgrade all packages: %w", err)
	}
	return nil
}

// SelfInstall installs Ruby via Homebrew (macOS) or provides guidance.
func (g *GemManager) SelfInstall(ctx context.Context) error {
	// Check if already installed
	available, _ := g.IsAvailable(ctx)
	if available {
		return nil
	}

	// Try to install via Homebrew first
	brew := NewBrewManager(g.exec)
	if brewAvailable, _ := brew.IsAvailable(ctx); brewAvailable {
		if err := brew.Install(ctx, "ruby"); err != nil {
			return fmt.Errorf("failed to install Ruby via Homebrew: %w", err)
		}
		return nil
	}

	return fmt.Errorf("Ruby is not installed and no supported installation method is available")
}

// parseGemList parses gem list output.
// With --no-versions, format is just "package_name" per line.
func parseGemList(data []byte) []string {
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
		// With --no-versions, each line is just the package name
		packages = append(packages, line)
	}
	return packages
}
