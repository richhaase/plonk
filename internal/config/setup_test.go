// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"testing"
)

// TestMain sets up the test environment for all tests in this package.
// It initializes the valid managers list that would normally be set by
// the packages registry during init().
func TestMain(m *testing.M) {
	// Set valid managers for config validation
	// This mirrors the supported managers in the packages registry
	SetValidManagers([]string{"brew", "cargo", "gem", "go", "npm", "pnpm", "bun", "uv"})

	// Run all tests
	code := m.Run()

	os.Exit(code)
}
