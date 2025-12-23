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
	exec CommandExecutor
}

// NewCargoManager creates a new Cargo manager.
func NewCargoManager(exec CommandExecutor) *CargoManager {
	if exec == nil {
		exec = defaultExecutor
	}
	return &CargoManager{exec: exec}
}

// IsAvailable checks if cargo is available on the system.
func (c *CargoManager) IsAvailable(ctx context.Context) (bool, error) {
	if _, err := c.exec.LookPath("cargo"); err != nil {
		return false, nil
	}

	_, err := c.exec.Execute(ctx, "cargo", "--version")
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ListInstalled lists all packages installed by Cargo.
// Output format: "package_name v1.2.3:\n    binary1\n    binary2\n"
func (c *CargoManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := c.exec.Execute(ctx, "cargo", "install", "--list")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return parseCargoList(output), nil
}

// Install installs a package via Cargo (idempotent).
func (c *CargoManager) Install(ctx context.Context, name string) error {
	output, err := c.exec.CombinedOutput(ctx, "cargo", "install", name)
	if err != nil {
		if isIdempotent(string(output), "already exists", "already installed") {
			return nil
		}
		return fmt.Errorf("failed to install %s: %w", name, err)
	}
	return nil
}

// Uninstall removes a package via Cargo (idempotent).
func (c *CargoManager) Uninstall(ctx context.Context, name string) error {
	output, err := c.exec.CombinedOutput(ctx, "cargo", "uninstall", name)
	if err != nil {
		if isIdempotent(string(output), "is not installed", "not installed") {
			return nil
		}
		return fmt.Errorf("failed to uninstall %s: %w", name, err)
	}
	return nil
}

// Upgrade upgrades packages to their latest versions.
// Cargo doesn't have a native upgrade, so we use --force to reinstall.
// Empty packages slice is not supported (no upgrade-all for cargo).
func (c *CargoManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return fmt.Errorf("cargo does not support upgrading all packages at once")
	}

	for _, pkg := range packages {
		output, err := c.exec.CombinedOutput(ctx, "cargo", "install", "--force", pkg)
		if err != nil {
			if isIdempotent(string(output), "already up-to-date", "up to date") {
				continue
			}
			return fmt.Errorf("failed to upgrade %s: %w", pkg, err)
		}
	}
	return nil
}

// SelfInstall installs Rust and Cargo using rustup.
func (c *CargoManager) SelfInstall(ctx context.Context) error {
	// Check if already installed
	available, _ := c.IsAvailable(ctx)
	if available {
		return nil
	}

	// Use the official rustup installation script
	script := `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y`
	_, err := c.exec.CombinedOutput(ctx, "sh", "-c", script)
	if err != nil {
		return fmt.Errorf("failed to install Rust via rustup: %w", err)
	}

	return nil
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
