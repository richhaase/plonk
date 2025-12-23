// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	exec     CommandExecutor
	provider NPMProvider
}

// NewNPMManager creates a new npm-family manager with the specified provider.
func NewNPMManager(provider NPMProvider, exec CommandExecutor) *NPMManager {
	if exec == nil {
		exec = defaultExecutor
	}
	if provider == "" {
		provider = ProviderNPM
	}
	return &NPMManager{exec: exec, provider: provider}
}

// Provider returns the current provider.
func (n *NPMManager) Provider() NPMProvider {
	return n.provider
}

// IsAvailable checks if the provider is available on the system.
func (n *NPMManager) IsAvailable(ctx context.Context) (bool, error) {
	binary := string(n.provider)
	if _, err := n.exec.LookPath(binary); err != nil {
		return false, nil
	}

	_, err := n.exec.Execute(ctx, binary, "--version")
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ListInstalled lists all globally installed packages.
func (n *NPMManager) ListInstalled(ctx context.Context) ([]string, error) {
	binary := string(n.provider)

	switch n.provider {
	case ProviderNPM:
		output, err := n.exec.Execute(ctx, binary, "list", "-g", "--depth=0", "--json")
		if err != nil {
			// npm list returns exit code 1 when there are peer dep warnings
			// but still outputs valid JSON, so we try to parse anyway
			if len(output) == 0 {
				return nil, fmt.Errorf("failed to list packages: %w", err)
			}
		}
		return parseNPMJSON(output)

	case ProviderPNPM:
		output, err := n.exec.Execute(ctx, binary, "list", "-g", "--depth=0", "--json")
		if err != nil {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
		return parsePNPMJSON(output)

	case ProviderBun:
		// bun pm ls -g outputs simple lines
		output, err := n.exec.Execute(ctx, binary, "pm", "ls", "-g")
		if err != nil {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
		return parseBunList(output), nil

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
		args = []string{"install", "-g", name}
	case ProviderPNPM:
		args = []string{"add", "-g", name}
	case ProviderBun:
		args = []string{"add", "-g", name}
	default:
		return fmt.Errorf("unknown provider: %s", n.provider)
	}

	output, err := n.exec.CombinedOutput(ctx, binary, args...)
	if err != nil {
		if isIdempotent(string(output), "already installed") {
			return nil
		}
		return fmt.Errorf("failed to install %s: %w", name, err)
	}
	return nil
}

// Uninstall removes a package globally (idempotent).
func (n *NPMManager) Uninstall(ctx context.Context, name string) error {
	binary := string(n.provider)
	var args []string

	switch n.provider {
	case ProviderNPM:
		args = []string{"uninstall", "-g", name}
	case ProviderPNPM:
		args = []string{"remove", "-g", name}
	case ProviderBun:
		args = []string{"remove", "-g", name}
	default:
		return fmt.Errorf("unknown provider: %s", n.provider)
	}

	output, err := n.exec.CombinedOutput(ctx, binary, args...)
	if err != nil {
		if isIdempotent(string(output), "not installed", "not found") {
			return nil
		}
		return fmt.Errorf("failed to uninstall %s: %w", name, err)
	}
	return nil
}

// Upgrade upgrades packages to their latest versions.
// If packages is empty, upgrades all installed packages.
func (n *NPMManager) Upgrade(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return n.upgradeAll(ctx)
	}

	binary := string(n.provider)

	for _, pkg := range packages {
		var args []string

		switch n.provider {
		case ProviderNPM:
			args = []string{"update", "-g", pkg}
		case ProviderPNPM:
			args = []string{"update", "-g", pkg}
		case ProviderBun:
			// bun doesn't have update, reinstall with latest
			args = []string{"add", "-g", pkg + "@latest"}
		default:
			return fmt.Errorf("unknown provider: %s", n.provider)
		}

		output, err := n.exec.CombinedOutput(ctx, binary, args...)
		if err != nil {
			if isIdempotent(string(output), "already up-to-date", "up to date") {
				continue
			}
			return fmt.Errorf("failed to upgrade %s: %w", pkg, err)
		}
	}
	return nil
}

func (n *NPMManager) upgradeAll(ctx context.Context) error {
	binary := string(n.provider)
	var args []string

	switch n.provider {
	case ProviderNPM:
		args = []string{"update", "-g"}
	case ProviderPNPM:
		args = []string{"update", "-g"}
	case ProviderBun:
		// bun doesn't have upgrade-all, would need to list and upgrade each
		return fmt.Errorf("bun does not support upgrading all packages at once")
	default:
		return fmt.Errorf("unknown provider: %s", n.provider)
	}

	output, err := n.exec.CombinedOutput(ctx, binary, args...)
	if err != nil {
		if isIdempotent(string(output), "already up-to-date", "up to date") {
			return nil
		}
		return fmt.Errorf("failed to upgrade all packages: %w", err)
	}
	return nil
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
	if _, err := n.exec.LookPath("brew"); err == nil {
		_, err := n.exec.CombinedOutput(ctx, "brew", "install", "node")
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("install Node.js from https://nodejs.org/ to get npm")
}

func (n *NPMManager) installPNPM(ctx context.Context) error {
	// Try brew first
	if _, err := n.exec.LookPath("brew"); err == nil {
		_, err := n.exec.CombinedOutput(ctx, "brew", "install", "pnpm")
		if err == nil {
			return nil
		}
	}

	// Try npm if available
	if _, err := n.exec.LookPath("npm"); err == nil {
		_, err := n.exec.CombinedOutput(ctx, "npm", "install", "-g", "pnpm")
		if err == nil {
			return nil
		}
	}

	// Fall back to official installer
	script := `curl -fsSL https://get.pnpm.io/install.sh | sh -`
	_, err := n.exec.CombinedOutput(ctx, "sh", "-c", script)
	if err != nil {
		return fmt.Errorf("failed to install pnpm: %w", err)
	}
	return nil
}

func (n *NPMManager) installBun(ctx context.Context) error {
	// Try brew first
	if _, err := n.exec.LookPath("brew"); err == nil {
		_, err := n.exec.CombinedOutput(ctx, "brew", "install", "bun")
		if err == nil {
			return nil
		}
	}

	// Fall back to official installer
	script := `curl -fsSL https://bun.sh/install | bash`
	_, err := n.exec.CombinedOutput(ctx, "sh", "-c", script)
	if err != nil {
		return fmt.Errorf("failed to install bun: %w", err)
	}
	return nil
}

// parseNPMJSON parses npm list -g --json output.
// Format: {"dependencies": {"pkg1": {...}, "pkg2": {...}}}
func parseNPMJSON(data []byte) ([]string, error) {
	var result struct {
		Dependencies map[string]interface{} `json:"dependencies"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse npm JSON: %w", err)
	}

	packages := make([]string, 0, len(result.Dependencies))
	for name := range result.Dependencies {
		packages = append(packages, name)
	}
	return packages, nil
}

// parsePNPMJSON parses pnpm list -g --json output.
// Format: [{"dependencies": {"pkg1": {...}, "pkg2": {...}}}]
func parsePNPMJSON(data []byte) ([]string, error) {
	var result []struct {
		Dependencies map[string]interface{} `json:"dependencies"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pnpm JSON: %w", err)
	}

	var packages []string
	for _, item := range result {
		for name := range item.Dependencies {
			packages = append(packages, name)
		}
	}
	return packages, nil
}

// parseBunList parses bun pm ls -g output (simple lines).
func parseBunList(data []byte) []string {
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
		// Take the first token (package name)
		parts := strings.Fields(line)
		if len(parts) > 0 {
			// Skip header lines or non-package lines
			name := parts[0]
			if strings.HasPrefix(name, "/") || name == "dependencies:" {
				continue
			}
			packages = append(packages, name)
		}
	}
	return packages
}
