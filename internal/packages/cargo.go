// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"strings"
)

// CargoManager implements PackageManager for Rust's Cargo.
type CargoManager struct {
	BaseManager
}

// NewCargoManager creates a new Cargo manager.
func NewCargoManager(exec CommandExecutor) *CargoManager {
	return &CargoManager{
		BaseManager: NewBaseManager(exec, "cargo", "--version"),
	}
}

// ListInstalled lists all packages installed by Cargo.
// Output format: "package_name v1.2.3:\n    binary1\n    binary2\n"
func (c *CargoManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := c.Exec().Execute(ctx, "cargo", "install", "--list")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return parseCargoList(output), nil
}

// Install installs a package via Cargo (idempotent).
func (c *CargoManager) Install(ctx context.Context, name string) error {
	return c.RunIdempotent(ctx,
		[]string{"already exists", "already installed"},
		fmt.Sprintf("failed to install %s", name),
		"cargo", "install", name,
	)
}

// Uninstall removes a package via Cargo (idempotent).
func (c *CargoManager) Uninstall(ctx context.Context, name string) error {
	return c.RunIdempotent(ctx,
		[]string{"is not installed", "not installed"},
		fmt.Sprintf("failed to uninstall %s", name),
		"cargo", "uninstall", name,
	)
}

// Upgrade upgrades packages to their latest versions.
// Cargo doesn't have a native upgrade, so we use --force to reinstall.
// Empty packages slice is not supported (no upgrade-all for cargo).
func (c *CargoManager) Upgrade(ctx context.Context, packages []string) error {
	return c.UpgradeEach(ctx, packages, false,
		func(pkg string) []string { return []string{"cargo", "install", "--force", pkg} },
		[]string{"already up-to-date", "up to date"},
	)
}

// SelfInstall installs Rust and Cargo using rustup.
func (c *CargoManager) SelfInstall(ctx context.Context) error {
	return c.SelfInstallWithBrewFallback(ctx, c.IsAvailable, "rust",
		`curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y`,
		"failed to install Rust via rustup",
	)
}

// parseCargoList parses cargo install --list output.
// Format: "package_name v1.2.3:\n    binary1\n    binary2\n"
// We only want the package names (lines ending with ":").
func parseCargoList(data []byte) []string {
	result := strings.TrimSpace(string(data))
	if result == "" {
		return []string{}
	}

	lines := strings.Split(result, "\n")
	var packages []string
	for _, line := range lines {
		// Check for indentation BEFORE trimming
		// Package lines start at column 0, binary lines are indented
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Take just the package name (first token before version)
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}
	return packages
}
