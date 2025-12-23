// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
)

// NPMManager implements PackageManager for npm (Node.js package manager).
type NPMManager struct {
	BaseManager
}

// NewNPMManager creates a new npm manager.
func NewNPMManager(exec CommandExecutor) *NPMManager {
	return &NPMManager{
		BaseManager: NewBaseManager(exec, "npm", "--version"),
	}
}

// ListInstalled lists all globally installed npm packages.
func (n *NPMManager) ListInstalled(ctx context.Context) ([]string, error) {
	output, err := n.Exec().Execute(ctx, "npm", "list", "-g", "--depth=0", "--json")
	if err != nil {
		// npm list returns exit code 1 when there are peer dep warnings
		// but still outputs valid JSON, so we try to parse anyway
		if len(output) == 0 {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
	}
	return parseJSONDependencies(output, false)
}

// Install installs a package globally via npm (idempotent).
func (n *NPMManager) Install(ctx context.Context, name string) error {
	return n.RunIdempotent(ctx,
		[]string{"already installed"},
		fmt.Sprintf("failed to install %s", name),
		"npm", "install", "-g", name,
	)
}

// Uninstall removes a package globally via npm (idempotent).
func (n *NPMManager) Uninstall(ctx context.Context, name string) error {
	return n.RunIdempotent(ctx,
		[]string{"not installed", "not found"},
		fmt.Sprintf("failed to uninstall %s", name),
		"npm", "uninstall", "-g", name,
	)
}

// Upgrade upgrades packages to their latest versions.
// If packages is empty, upgrades all installed packages.
func (n *NPMManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return n.UpgradeAll(ctx, []string{"already up-to-date", "up to date"}, "npm", "update", "-g")
	}

	return n.UpgradeEach(ctx, packages, true,
		func(pkg string) []string { return []string{"npm", "update", "-g", pkg} },
		[]string{"already up-to-date", "up to date"},
	)
}

// SelfInstall installs npm by installing Node.js.
func (n *NPMManager) SelfInstall(ctx context.Context) error {
	// Check if already installed
	available, _ := n.IsAvailable(ctx)
	if available {
		return nil
	}

	// npm comes with Node.js, try brew first
	if _, err := n.Exec().LookPath("brew"); err == nil {
		_, err := n.Exec().CombinedOutput(ctx, "brew", "install", "node")
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("install Node.js from https://nodejs.org/ to get npm")
}
