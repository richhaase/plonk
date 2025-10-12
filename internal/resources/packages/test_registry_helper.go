// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"

	"github.com/richhaase/plonk/internal/config"
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
	// Legacy helper no longer registers code managers; kept for compatibility where tests
	// only need an isolated v2 registry. Callers should use v2 config + MockCommandExecutor.
	t.Helper()
	original := defaultRegistry
	temp := &ManagerRegistry{v2Managers: make(map[string]config.ManagerConfig), enableV2: true}
	defaultRegistry = temp
	if register != nil {
		register(temp)
	}
	t.Cleanup(func() { defaultRegistry = original })
}
