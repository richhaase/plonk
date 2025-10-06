// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCapabilityDetection verifies that capability checking functions work correctly
func TestCapabilityDetection(t *testing.T) {
	// Create instances of each manager
	managers := map[string]PackageManager{
		"brew":  NewHomebrewManager(),
		"npm":   NewNpmManager(),
		"pnpm":  NewPnpmManager(),
		"cargo": NewCargoManager(),
		"pipx":  NewPipxManager(),
		"conda": NewCondaManager(),
		"gem":   NewGemManager(),
		"go":    NewGoInstallManager(),
		"uv":    NewUvManager(),
		"pixi":  NewPixiManager(),
	}

	// Test each manager's capabilities
	for name, mgr := range managers {
		t.Run(name, func(t *testing.T) {
			// All managers should support basic operations (this is enforced by interface)
			assert.NotNil(t, mgr, "manager should not be nil")

			// Check capabilities - most managers support all features
			// We're just verifying the capability detection works, not enforcing specific capabilities
			hasSearch := SupportsSearch(mgr)
			hasInfo := SupportsInfo(mgr)
			hasUpgrade := SupportsUpgrade(mgr)
			hasHealth := SupportsHealthCheck(mgr)

			t.Logf("%s capabilities: search=%v info=%v upgrade=%v health=%v",
				name, hasSearch, hasInfo, hasUpgrade, hasHealth)

			// Verify that capability detection is consistent
			// If we detect a capability, we should be able to type assert to it
			if hasSearch {
				_, ok := mgr.(PackageSearcher)
				assert.True(t, ok, "SupportsSearch returned true but type assertion failed")
			}

			if hasInfo {
				_, ok := mgr.(PackageInfoProvider)
				assert.True(t, ok, "SupportsInfo returned true but type assertion failed")
			}

			if hasUpgrade {
				_, ok := mgr.(PackageUpgrader)
				assert.True(t, ok, "SupportsUpgrade returned true but type assertion failed")
			}

			if hasHealth {
				_, ok := mgr.(PackageHealthChecker)
				assert.True(t, ok, "SupportsHealthCheck returned true but type assertion failed")
			}
		})
	}
}

// TestAllManagersSupportCore verifies all managers implement core PackageManager interface
func TestAllManagersSupportCore(t *testing.T) {
	managers := []PackageManager{
		NewHomebrewManager(),
		NewNpmManager(),
		NewPnpmManager(),
		NewCargoManager(),
		NewPipxManager(),
		NewCondaManager(),
		NewGemManager(),
		NewGoInstallManager(),
		NewUvManager(),
		NewPixiManager(),
	}

	for _, mgr := range managers {
		// If this compiles and runs, the manager implements PackageManager
		assert.NotNil(t, mgr)
	}
}

// TestCapabilityFunctionsReturnBool verifies capability functions return proper booleans
func TestCapabilityFunctionsReturnBool(t *testing.T) {
	// Test with a manager that has all capabilities (brew)
	brew := NewHomebrewManager()

	// These should all return boolean values (not panic)
	assert.IsType(t, true, SupportsSearch(brew))
	assert.IsType(t, true, SupportsInfo(brew))
	assert.IsType(t, true, SupportsUpgrade(brew))
	assert.IsType(t, true, SupportsHealthCheck(brew))
}
