// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
)

// BrewManager implements PackageManager for Homebrew.
type BrewManager struct {
	BaseManager
}

// NewBrewManager creates a new Homebrew manager.
func NewBrewManager(exec CommandExecutor) *BrewManager {
	return &BrewManager{
		BaseManager: NewBaseManager(exec, "brew", "--version"),
	}
}

// ListInstalled lists all packages installed by Homebrew.
func (b *BrewManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := b.Exec().Execute(ctx, "brew", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return parseLines(output), nil
}

// Install installs a package via Homebrew (idempotent).
func (b *BrewManager) Install(ctx context.Context, name string) error {
	return b.RunIdempotent(ctx,
		[]string{"already installed"},
		fmt.Sprintf("failed to install %s", name),
		"brew", "install", name,
	)
}

// Uninstall removes a package via Homebrew (idempotent).
func (b *BrewManager) Uninstall(ctx context.Context, name string) error {
	return b.RunIdempotent(ctx,
		[]string{"no such keg"},
		fmt.Sprintf("failed to uninstall %s", name),
		"brew", "uninstall", name,
	)
}

// Upgrade upgrades packages to their latest versions.
// If packages is empty, upgrades all installed packages.
func (b *BrewManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return b.UpgradeAll(ctx, []string{"already up-to-date"}, "brew", "upgrade")
	}

	return b.UpgradeEach(ctx, packages, true,
		func(pkg string) []string { return []string{"brew", "upgrade", pkg} },
		[]string{"already up-to-date", "already installed"},
	)
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
	_, err := b.Exec().CombinedOutput(ctx, "sh", "-c", script)
	if err != nil {
		return fmt.Errorf("failed to install Homebrew: %w", err)
	}

	return nil
}
