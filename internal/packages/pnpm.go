// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
)

// PNPMManager implements PackageManager for pnpm (Node.js package manager).
type PNPMManager struct {
	BaseManager
}

// NewPNPMManager creates a new pnpm manager.
func NewPNPMManager(exec CommandExecutor) *PNPMManager {
	return &PNPMManager{
		BaseManager: NewBaseManager(exec, "pnpm", "--version"),
	}
}

// ListInstalled lists all globally installed pnpm packages.
func (p *PNPMManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := p.Exec().Execute(ctx, "pnpm", "list", "-g", "--depth=0", "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	// pnpm outputs JSON array format: [{"dependencies": {...}}]
	return parseJSONDependencies(output, true)
}

// Install installs a package globally via pnpm (idempotent).
func (p *PNPMManager) Install(ctx context.Context, name string) error {
	return p.RunIdempotent(ctx,
		[]string{"already installed"},
		fmt.Sprintf("failed to install %s", name),
		"pnpm", "add", "-g", name,
	)
}

// Uninstall removes a package globally via pnpm (idempotent).
func (p *PNPMManager) Uninstall(ctx context.Context, name string) error {
	return p.RunIdempotent(ctx,
		[]string{"not installed", "not found"},
		fmt.Sprintf("failed to uninstall %s", name),
		"pnpm", "remove", "-g", name,
	)
}

// Upgrade upgrades packages to their latest versions.
// If packages is empty, upgrades all installed packages.
func (p *PNPMManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return p.UpgradeAll(ctx, []string{"already up-to-date", "up to date"}, "pnpm", "update", "-g")
	}

	return p.UpgradeEach(ctx, packages, true,
		func(pkg string) []string { return []string{"pnpm", "update", "-g", pkg} },
		[]string{"already up-to-date", "up to date"},
	)
}

// SelfInstall installs pnpm via Homebrew or the official installer.
func (p *PNPMManager) SelfInstall(ctx context.Context) error {
	return p.SelfInstallWithBrewFallback(ctx, p.IsAvailable, "pnpm",
		`curl -fsSL https://get.pnpm.io/install.sh | sh -`,
		"failed to install pnpm",
	)
}
