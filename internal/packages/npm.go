// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
)

// NPMProvider represents the Node.js package manager provider.
type NPMProvider string

const (
	ProviderNPM  NPMProvider = "npm"
	ProviderPNPM NPMProvider = "pnpm"
	ProviderBun  NPMProvider = "bun"
)

// NPMManager implements PackageManager for Node.js package managers.
// It supports npm, pnpm, and bun as providers.
type NPMManager struct {
	BaseManager
	provider NPMProvider
}

// NewNPMManager creates a new npm-family manager with the specified provider.
func NewNPMManager(provider NPMProvider, exec CommandExecutor) *NPMManager {
	if provider == "" {
		provider = ProviderNPM
	}
	return &NPMManager{
		BaseManager: NewBaseManager(exec, string(provider), "--version"),
		provider:    provider,
	}
}

// Provider returns the current provider.
func (n *NPMManager) Provider() NPMProvider {
	return n.provider
}

// ListInstalled lists all globally installed packages.
func (n *NPMManager) ListInstalled(ctx context.Context) ([]string, error) {
	binary := string(n.provider)

	switch n.provider {
	case ProviderNPM:
		output, err := n.Exec().Execute(ctx, binary, "list", "-g", "--depth=0", "--json")
		if err != nil {
			// npm list returns exit code 1 when there are peer dep warnings
			// but still outputs valid JSON, so we try to parse anyway
			if len(output) == 0 {
				return nil, fmt.Errorf("failed to list packages: %w", err)
			}
		}
		return parseJSONDependencies(output, false)

	case ProviderPNPM:
		output, err := n.Exec().Execute(ctx, binary, "list", "-g", "--depth=0", "--json")
		if err != nil {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
		return parseJSONDependencies(output, true)

	case ProviderBun:
		// bun pm ls -g outputs simple lines
		output, err := n.Exec().Execute(ctx, binary, "pm", "ls", "-g")
		if err != nil {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
		return parseOutput(output, ParseConfig{TakeFirstToken: true, SkipPrefixes: []string{"/", "dependencies:"}}), nil

	default:
		return nil, fmt.Errorf("unknown provider: %s", n.provider)
	}
}

// Install installs a package globally (idempotent).
func (n *NPMManager) Install(ctx context.Context, name string) error {
	binary := string(n.provider)
	var args []string

	switch n.provider {
	case ProviderNPM:
		args = []string{binary, "install", "-g", name}
	case ProviderPNPM:
		args = []string{binary, "add", "-g", name}
	case ProviderBun:
		args = []string{binary, "add", "-g", name}
	default:
		return fmt.Errorf("unknown provider: %s", n.provider)
	}

	return n.RunIdempotent(ctx,
		[]string{"already installed"},
		fmt.Sprintf("failed to install %s", name),
		args...,
	)
}

// Uninstall removes a package globally (idempotent).
func (n *NPMManager) Uninstall(ctx context.Context, name string) error {
	binary := string(n.provider)
	var args []string

	switch n.provider {
	case ProviderNPM:
		args = []string{binary, "uninstall", "-g", name}
	case ProviderPNPM:
		args = []string{binary, "remove", "-g", name}
	case ProviderBun:
		args = []string{binary, "remove", "-g", name}
	default:
		return fmt.Errorf("unknown provider: %s", n.provider)
	}

	return n.RunIdempotent(ctx,
		[]string{"not installed", "not found"},
		fmt.Sprintf("failed to uninstall %s", name),
		args...,
	)
}

// Upgrade upgrades packages to their latest versions.
// If packages is empty, upgrades all installed packages.
func (n *NPMManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return n.upgradeAll(ctx)
	}

	binary := string(n.provider)
	patterns := []string{"already up-to-date", "up to date"}

	return n.UpgradeEach(ctx, packages, true,
		func(pkg string) []string {
			switch n.provider {
			case ProviderNPM:
				return []string{binary, "update", "-g", pkg}
			case ProviderPNPM:
				return []string{binary, "update", "-g", pkg}
			case ProviderBun:
				// bun doesn't have update, reinstall with latest
				return []string{binary, "add", "-g", pkg + "@latest"}
			default:
				return []string{binary, "update", "-g", pkg}
			}
		},
		patterns,
	)
}

func (n *NPMManager) upgradeAll(ctx context.Context) error {
	binary := string(n.provider)

	switch n.provider {
	case ProviderNPM:
		return n.UpgradeAll(ctx, []string{"already up-to-date", "up to date"}, binary, "update", "-g")
	case ProviderPNPM:
		return n.UpgradeAll(ctx, []string{"already up-to-date", "up to date"}, binary, "update", "-g")
	case ProviderBun:
		// bun doesn't have upgrade-all, would need to list and upgrade each
		return fmt.Errorf("bun does not support upgrading all packages at once")
	default:
		return fmt.Errorf("unknown provider: %s", n.provider)
	}
}

// SelfInstall installs the provider.
func (n *NPMManager) SelfInstall(ctx context.Context) error {
	// Check if already installed
	available, _ := n.IsAvailable(ctx)
	if available {
		return nil
	}

	switch n.provider {
	case ProviderNPM:
		return n.installNPM(ctx)
	case ProviderPNPM:
		return n.installPNPM(ctx)
	case ProviderBun:
		return n.installBun(ctx)
	default:
		return fmt.Errorf("unknown provider: %s", n.provider)
	}
}

func (n *NPMManager) installNPM(ctx context.Context) error {
	// npm comes with Node.js, try brew first
	if _, err := n.Exec().LookPath("brew"); err == nil {
		_, err := n.Exec().CombinedOutput(ctx, "brew", "install", "node")
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("install Node.js from https://nodejs.org/ to get npm")
}

func (n *NPMManager) installPNPM(ctx context.Context) error {
	// Try brew first
	if _, err := n.Exec().LookPath("brew"); err == nil {
		_, err := n.Exec().CombinedOutput(ctx, "brew", "install", "pnpm")
		if err == nil {
			return nil
		}
	}

	// Try npm if available
	if _, err := n.Exec().LookPath("npm"); err == nil {
		_, err := n.Exec().CombinedOutput(ctx, "npm", "install", "-g", "pnpm")
		if err == nil {
			return nil
		}
	}

	// Fall back to official installer
	script := `curl -fsSL https://get.pnpm.io/install.sh | sh -`
	_, err := n.Exec().CombinedOutput(ctx, "sh", "-c", script)
	if err != nil {
		return fmt.Errorf("failed to install pnpm: %w", err)
	}
	return nil
}

func (n *NPMManager) installBun(ctx context.Context) error {
	// Try brew first
	if _, err := n.Exec().LookPath("brew"); err == nil {
		_, err := n.Exec().CombinedOutput(ctx, "brew", "install", "bun")
		if err == nil {
			return nil
		}
	}

	// Fall back to official installer
	script := `curl -fsSL https://bun.sh/install | bash`
	_, err := n.Exec().CombinedOutput(ctx, "sh", "-c", script)
	if err != nil {
		return fmt.Errorf("failed to install bun: %w", err)
	}
	return nil
}

