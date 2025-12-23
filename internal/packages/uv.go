// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
)

// UVManager implements PackageManager for uv (Python package manager).
type UVManager struct {
	BaseManager
}

// NewUVManager creates a new uv manager.
func NewUVManager(exec CommandExecutor) *UVManager {
	return &UVManager{
		BaseManager: NewBaseManager(exec, "uv", "--version"),
	}
}

// ListInstalled lists all tools installed by uv.
func (u *UVManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := u.Exec().Execute(ctx, "uv", "tool", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return parseOutput(output, ParseConfig{TakeFirstToken: true}), nil
}

// Install installs a tool via uv (idempotent).
func (u *UVManager) Install(ctx context.Context, name string) error {
	return u.RunIdempotent(ctx,
		[]string{"already installed"},
		fmt.Sprintf("failed to install %s", name),
		"uv", "tool", "install", name,
	)
}

// Uninstall removes a tool via uv (idempotent).
func (u *UVManager) Uninstall(ctx context.Context, name string) error {
	return u.RunIdempotent(ctx,
		[]string{"not installed"},
		fmt.Sprintf("failed to uninstall %s", name),
		"uv", "tool", "uninstall", name,
	)
}

// Upgrade upgrades tools to their latest versions.
// If packages is empty, upgrades all installed tools.
func (u *UVManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return u.UpgradeAll(ctx, []string{"already up-to-date", "up to date"}, "uv", "tool", "upgrade", "--all")
	}

	return u.UpgradeEach(ctx, packages, true,
		func(pkg string) []string { return []string{"uv", "tool", "upgrade", pkg} },
		[]string{"already up-to-date", "up to date"},
	)
}

// SelfInstall installs uv using brew or the official installer.
func (u *UVManager) SelfInstall(ctx context.Context) error {
	return u.SelfInstallWithBrewFallback(ctx, u.IsAvailable, "uv",
		`curl -LsSf https://astral.sh/uv/install.sh | sh`,
		"failed to install uv",
	)
}

