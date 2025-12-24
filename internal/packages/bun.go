// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
)

// BunManager implements PackageManager for Bun (JavaScript runtime and package manager).
type BunManager struct {
	BaseManager
}

// NewBunManager creates a new Bun manager.
func NewBunManager(exec CommandExecutor) *BunManager {
	return &BunManager{
		BaseManager: NewBaseManager(exec, "bun", "--version"),
	}
}

// ListInstalled lists all globally installed bun packages.
func (b *BunManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := b.Exec().Execute(ctx, "bun", "pm", "ls", "-g")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	// bun outputs simple text lines, skip paths and "dependencies:" header
	return parseOutput(output, ParseConfig{TakeFirstToken: true, SkipPrefixes: []string{"/", "dependencies:"}}), nil
}

// Install installs a package globally via bun (idempotent).
func (b *BunManager) Install(ctx context.Context, name string) error {
	return b.RunIdempotent(ctx,
		[]string{"already installed"},
		fmt.Sprintf("failed to install %s", name),
		"bun", "add", "-g", name,
	)
}

// Uninstall removes a package globally via bun (idempotent).
func (b *BunManager) Uninstall(ctx context.Context, name string) error {
	return b.RunIdempotent(ctx,
		[]string{"not installed", "not found"},
		fmt.Sprintf("failed to uninstall %s", name),
		"bun", "remove", "-g", name,
	)
}

// Upgrade upgrades packages to their latest versions.
// Bun does not support upgrading all packages at once.
func (b *BunManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return fmt.Errorf("bun does not support upgrading all packages at once")
	}

	// bun doesn't have update, reinstall with @latest
	return b.UpgradeEach(ctx, packages, false,
		func(pkg string) []string { return []string{"bun", "add", "-g", pkg + "@latest"} },
		[]string{"already up-to-date", "up to date"},
	)
}

// SelfInstall installs bun using brew or the official installer.
func (b *BunManager) SelfInstall(ctx context.Context) error {
	return b.SelfInstallWithBrewFallback(ctx, b.IsAvailable, "bun",
		`curl -fsSL https://bun.sh/install | bash`,
		"failed to install bun",
	)
}
