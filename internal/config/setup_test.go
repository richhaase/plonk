// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"testing"
)

// supportedManagers mirrors the list from the packages registry for testing
var supportedManagers = []string{"brew", "cargo", "go", "pnpm", "uv"}

// TestMain sets up the test environment for all tests in this package.
// It initializes the manager checker that would normally be set by
// the packages registry during init().
func TestMain(m *testing.M) {
	// Set up manager checker for config validation
	ManagerChecker = func(name string) bool {
		for _, m := range supportedManagers {
			if name == m {
				return true
			}
		}
		return false
	}

	// Run all tests
	code := m.Run()

	os.Exit(code)
}
