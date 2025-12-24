// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
)

// GemManager implements PackageManager for Ruby gems.
type GemManager struct {
	BaseManager
}

// NewGemManager creates a new Ruby gem manager.
func NewGemManager(exec CommandExecutor) *GemManager {
	return &GemManager{
		BaseManager: NewBaseManager(exec, "gem", "--version"),
	}
}

// ListInstalled lists all packages installed by gem.
// Output format: "package_name (version, version2)"
func (g *GemManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := g.Exec().Execute(ctx, "gem", "list", "--no-versions")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return parseOutput(output, ParseConfig{}), nil
}

// Install installs a package via gem (idempotent).
// Uses --user-install to install to user's home directory (avoids permission issues on Linux).
func (g *GemManager) Install(ctx context.Context, name string) error {
	return g.RunIdempotent(ctx,
		[]string{"already installed"},
		fmt.Sprintf("failed to install %s", name),
		"gem", "install", "--user-install", name,
	)
}

// Uninstall removes a package via gem (idempotent).
// Uses -x to remove executables and -a to remove all versions.
func (g *GemManager) Uninstall(ctx context.Context, name string) error {
	return g.RunIdempotent(ctx,
		[]string{"is not installed", "not installed"},
		fmt.Sprintf("failed to uninstall %s", name),
		"gem", "uninstall", name, "-x", "-a",
	)
}

// Upgrade upgrades packages to their latest versions.
// If packages is empty, upgrades all installed packages.
func (g *GemManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return g.UpgradeAll(ctx, []string{"nothing to update"}, "gem", "update")
	}

	return g.UpgradeEach(ctx, packages, true,
		func(pkg string) []string { return []string{"gem", "update", pkg} },
		[]string{"already up-to-date", "nothing to update"},
	)
}

// SelfInstall installs Ruby via Homebrew (macOS) or provides guidance.
func (g *GemManager) SelfInstall(ctx context.Context) error {
	return g.SelfInstallWithBrewFallback(ctx, g.IsAvailable, "ruby", "",
		"ruby is not installed and no supported installation method is available",
	)
}

