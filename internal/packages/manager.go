// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/richhaase/plonk/internal/config"
)

func init() {
	// Register manager checker with config validation
	config.ManagerChecker = IsSupportedManager
}

// Manager defines the simplified package manager interface.
// Check installation, install missing packages, and explicitly uninstall.
type Manager interface {
	// IsInstalled checks if a package is installed
	IsInstalled(ctx context.Context, name string) (bool, error)

	// Install installs a package (should be idempotent)
	Install(ctx context.Context, name string) error

	// Uninstall removes an installed package.
	Uninstall(ctx context.Context, name string) error
}

// SupportedManagers lists all available package managers
var SupportedManagers = []string{"brew", "cargo", "go", "pnpm", "uv"}

// IsSupportedManager checks if a manager name is valid
func IsSupportedManager(name string) bool {
	return slices.Contains(SupportedManagers, name)
}

// ParsePackageSpec parses "manager:package" format and validates the manager
func ParsePackageSpec(spec string) (manager, pkg string, err error) {
	manager, pkg, found := strings.Cut(spec, ":")
	if !found {
		return "", "", fmt.Errorf("invalid format, expected manager:package")
	}

	if !IsSupportedManager(manager) {
		return "", "", fmt.Errorf("unsupported manager: %s (supported: %v)", manager, SupportedManagers)
	}

	if pkg == "" {
		return "", "", fmt.Errorf("package name cannot be empty")
	}

	if pkg[0] == '-' {
		return "", "", fmt.Errorf("invalid package name %q: must not start with '-'", pkg)
	}

	if manager == "go" && !strings.Contains(pkg, "/") {
		return "", "", fmt.Errorf("invalid go package %q: expected full import path (e.g., golang.org/x/tools/gopls)", pkg)
	}

	return manager, pkg, nil
}
