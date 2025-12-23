// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"strings"
)

// BaseManager provides shared functionality for package managers.
// Embed this struct in concrete managers to reduce code duplication.
type BaseManager struct {
	exec       CommandExecutor
	binary     string
	versionArg string // "--version" or "version"
}

// NewBaseManager creates a new BaseManager with the given configuration.
func NewBaseManager(exec CommandExecutor, binary, versionArg string) BaseManager {
	if exec == nil {
		exec = defaultExecutor
	}
	return BaseManager{
		exec:       exec,
		binary:     binary,
		versionArg: versionArg,
	}
}

// Exec returns the command executor.
func (b *BaseManager) Exec() CommandExecutor {
	return b.exec
}

// Binary returns the binary name.
func (b *BaseManager) Binary() string {
	return b.binary
}

// IsAvailable checks if the binary is available on the system.
func (b *BaseManager) IsAvailable(ctx context.Context) (bool, error) {
	if _, err := b.exec.LookPath(b.binary); err != nil {
		return false, nil
	}

	_, err := b.exec.Execute(ctx, b.binary, b.versionArg)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// RunIdempotent executes a command and returns nil if the output matches any idempotent pattern.
func (b *BaseManager) RunIdempotent(ctx context.Context, patterns []string, errMsg string, args ...string) error {
	output, err := b.exec.CombinedOutput(ctx, args[0], args[1:]...)
	if err != nil {
		if isIdempotent(string(output), patterns...) {
			return nil
		}
		return fmt.Errorf("%s: %w", errMsg, err)
	}
	return nil
}

// UpgradeEach upgrades packages one at a time using the provided command builder.
// Returns an error for managers that don't support upgrade-all when packages is empty.
func (b *BaseManager) UpgradeEach(ctx context.Context, packages []string, supportsAll bool, buildArgs func(pkg string) []string, patterns []string) error {
	if len(packages) == 0 && !supportsAll {
		return fmt.Errorf("%s does not support upgrading all packages at once", b.binary)
	}

	for _, pkg := range packages {
		args := buildArgs(pkg)
		output, err := b.exec.CombinedOutput(ctx, args[0], args[1:]...)
		if err != nil {
			if isIdempotent(string(output), patterns...) {
				continue
			}
			return fmt.Errorf("failed to upgrade %s: %w", pkg, err)
		}
	}
	return nil
}

// UpgradeAll runs an upgrade-all command with idempotency checking.
func (b *BaseManager) UpgradeAll(ctx context.Context, patterns []string, args ...string) error {
	output, err := b.exec.CombinedOutput(ctx, args[0], args[1:]...)
	if err != nil {
		if isIdempotent(string(output), patterns...) {
			return nil
		}
		return fmt.Errorf("failed to upgrade all packages: %w", err)
	}
	return nil
}

// SelfInstallWithBrewFallback attempts to install using brew, then falls back to the provided script.
func (b *BaseManager) SelfInstallWithBrewFallback(ctx context.Context, isAvailable func(context.Context) (bool, error), brewPkg, fallbackScript, errMsg string) error {
	// Check if already installed
	available, _ := isAvailable(ctx)
	if available {
		return nil
	}

	// Try brew first
	if _, err := b.exec.LookPath("brew"); err == nil {
		_, err := b.exec.CombinedOutput(ctx, "brew", "install", brewPkg)
		if err == nil {
			return nil
		}
	}

	// Fall back to script if provided
	if fallbackScript != "" {
		_, err := b.exec.CombinedOutput(ctx, "sh", "-c", fallbackScript)
		if err != nil {
			return fmt.Errorf("%s: %w", errMsg, err)
		}
		return nil
	}

	return fmt.Errorf("%s", errMsg)
}

// parseLines splits output by newlines and returns non-empty lines.
func parseLines(data []byte) []string {
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
		// Take the first whitespace-delimited token
		parts := strings.Fields(line)
		if len(parts) > 0 {
			packages = append(packages, parts[0])
		}
	}
	return packages
}

// isIdempotent checks if output contains any of the given patterns (case-insensitive).
func isIdempotent(output string, patterns ...string) bool {
	outputLower := strings.ToLower(output)
	for _, pattern := range patterns {
		if strings.Contains(outputLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}
