// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"
)

// WithTemporaryRegistry replaces the global manager registry with a fresh instance
// for the duration of a test, then restores the original via t.Cleanup.
//
// Usage:
//
//	packages.WithTemporaryRegistry(t, func(r *packages.ManagerRegistry) {
//	    r.Register("fake", func() packages.PackageManager { return newFakeManager() })
//	})
//
// This avoids interacting with real package managers during tests.
func WithTemporaryRegistry(t *testing.T, register func(*ManagerRegistry)) {
	t.Helper()

	original := defaultRegistry

	// Create a clean registry for tests
	temp := &ManagerRegistry{managers: make(map[string]*managerEntry)}
	defaultRegistry = temp

	// Allow caller to register desired fake managers
	if register != nil {
		register(temp)
	}

	// Restore original after the test completes
	t.Cleanup(func() {
		defaultRegistry = original
	})
}
