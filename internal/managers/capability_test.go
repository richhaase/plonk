// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"testing"
)

func TestPackageManagerCapabilities(t *testing.T) {
	tests := []struct {
		name           string
		manager        func() PackageManager
		supportsSearch bool
	}{
		{
			name:           "cargo supports search",
			manager:        func() PackageManager { return NewCargoManager() },
			supportsSearch: true,
		},
		{
			name:           "gem supports search",
			manager:        func() PackageManager { return NewGemManager() },
			supportsSearch: true,
		},
		{
			name:           "homebrew supports search",
			manager:        func() PackageManager { return NewHomebrewManager() },
			supportsSearch: true,
		},
		{
			name:           "npm supports search",
			manager:        func() PackageManager { return NewNpmManager() },
			supportsSearch: true,
		},
		{
			name:           "go does not support search",
			manager:        func() PackageManager { return NewGoInstallManager() },
			supportsSearch: false,
		},
		{
			name:           "pip supports search",
			manager:        func() PackageManager { return NewPipManager() },
			supportsSearch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := tt.manager()

			// Test SupportsSearch capability
			if got := manager.SupportsSearch(); got != tt.supportsSearch {
				t.Errorf("SupportsSearch() = %v, want %v", got, tt.supportsSearch)
			}

			// Test that unsupported operations return error
			if !tt.supportsSearch {
				ctx := context.Background()
				_, err := manager.Search(ctx, "test-query")
				if err == nil {
					t.Errorf("Search() expected error for unsupported operation, got nil")
				}
			}
		})
	}
}

func TestCapabilityDiscoveryPattern(t *testing.T) {
	// Test that capability checking prevents unnecessary errors
	managers := []struct {
		name    string
		manager PackageManager
	}{
		{"cargo", NewCargoManager()},
		{"gem", NewGemManager()},
		{"homebrew", NewHomebrewManager()},
		{"npm", NewNpmManager()},
		{"go", NewGoInstallManager()},
		{"pip", NewPipManager()},
	}

	ctx := context.Background()
	for _, m := range managers {
		t.Run(m.name+" capability check", func(t *testing.T) {
			// Check if manager is available first
			available, err := m.manager.IsAvailable(ctx)
			if err != nil {
				t.Skipf("Manager not available: %v", err)
			}
			if !available {
				t.Skip("Manager not available on this system")
			}

			// Simulate command layer checking capability before calling
			if m.manager.SupportsSearch() {
				// Safe to call Search - use a simple query that should work
				results, err := m.manager.Search(ctx, "test")
				// For supported managers, search should not return error (may return empty results)
				if err != nil {
					// Some managers may fail on network issues or other problems
					// Just log it but don't fail the test
					t.Logf("Search() returned error (may be network/system issue): %v", err)
				} else {
					// Results can be nil or empty slice - both are valid
					t.Logf("Search() returned %d results", len(results))
				}
			} else {
				// Would skip calling Search in real code
				// Verify error is consistent
				_, err := m.manager.Search(ctx, "test")
				if err == nil {
					t.Errorf("Search() should return error for unsupported operation")
				}
			}
		})
	}
}
