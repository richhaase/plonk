// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"strings"
)

// UVManager implements PackageManager for uv (Python package manager).
type UVManager struct {
	exec CommandExecutor
}

// NewUVManager creates a new uv manager.
func NewUVManager(exec CommandExecutor) *UVManager {
	if exec == nil {
		exec = defaultExecutor
	}
	return &UVManager{exec: exec}
}

// IsAvailable checks if uv is available on the system.
func (u *UVManager) IsAvailable(ctx context.Context) (bool, error) {
	if _, err := u.exec.LookPath("uv"); err != nil {
		return false, nil
	}

	_, err := u.exec.Execute(ctx, "uv", "--version")
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ListInstalled lists all tools installed by uv.
func (u *UVManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := u.exec.Execute(ctx, "uv", "tool", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return parseUVList(output), nil
}

// Install installs a tool via uv (idempotent).
func (u *UVManager) Install(ctx context.Context, name string) error {
	output, err := u.exec.CombinedOutput(ctx, "uv", "tool", "install", name)
	if err != nil {
		if isIdempotent(string(output), "already installed") {
			return nil
		}
		return fmt.Errorf("failed to install %s: %w", name, err)
	}
	return nil
}

// Uninstall removes a tool via uv (idempotent).
func (u *UVManager) Uninstall(ctx context.Context, name string) error {
	output, err := u.exec.CombinedOutput(ctx, "uv", "tool", "uninstall", name)
	if err != nil {
		if isIdempotent(string(output), "not installed") {
			return nil
		}
		return fmt.Errorf("failed to uninstall %s: %w", name, err)
	}
	return nil
}

// Upgrade upgrades tools to their latest versions.
// If packages is empty, upgrades all installed tools.
func (u *UVManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return u.upgradeAll(ctx)
	}

	for _, pkg := range packages {
		output, err := u.exec.CombinedOutput(ctx, "uv", "tool", "upgrade", pkg)
		if err != nil {
			if isIdempotent(string(output), "already up-to-date", "up to date") {
				continue
			}
			return fmt.Errorf("failed to upgrade %s: %w", pkg, err)
		}
	}
	return nil
}

func (u *UVManager) upgradeAll(ctx context.Context) error {
	output, err := u.exec.CombinedOutput(ctx, "uv", "tool", "upgrade", "--all")
	if err != nil {
		if isIdempotent(string(output), "already up-to-date", "up to date") {
			return nil
		}
		return fmt.Errorf("failed to upgrade all tools: %w", err)
	}
	return nil
}

// SelfInstall installs uv using brew or the official installer.
func (u *UVManager) SelfInstall(ctx context.Context) error {
	// Check if already installed
	available, _ := u.IsAvailable(ctx)
	if available {
		return nil
	}

	// Try brew first
	if _, err := u.exec.LookPath("brew"); err == nil {
		_, err := u.exec.CombinedOutput(ctx, "brew", "install", "uv")
		if err == nil {
			return nil
		}
	}

	// Fall back to official installer script
	script := `curl -LsSf https://astral.sh/uv/install.sh | sh`
	_, err := u.exec.CombinedOutput(ctx, "sh", "-c", script)
	if err != nil {
		return fmt.Errorf("failed to install uv: %w", err)
	}

	return nil
}

// parseUVList parses uv tool list output.
// Format: "tool_name v1.2.3\n" or "tool_name v1.2.3 (extras)\n"
func parseUVList(data []byte) []string {
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
		// Take just the package name (first token)
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}
	return packages
}
