// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"slices"
)

// Manager defines the simplified package manager interface.
// Only two operations: check if installed, install if missing.
type Manager interface {
	// IsInstalled checks if a package is installed
	IsInstalled(ctx context.Context, name string) (bool, error)

	// Install installs a package (should be idempotent)
	Install(ctx context.Context, name string) error
}

// SupportedManagers lists all available package managers
var SupportedManagers = []string{"brew", "cargo", "go", "pnpm", "uv"}

// IsSupportedManager checks if a manager name is valid
func IsSupportedManager(name string) bool {
	return slices.Contains(SupportedManagers, name)
}
