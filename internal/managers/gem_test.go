// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/errors"
)

func TestGemManager_IsAvailable(t *testing.T) {
	manager := NewGemManager()
	ctx := context.Background()

	// This test will check actual gem availability on the system
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Logf("IsAvailable returned error: %v", err)
	}
	t.Logf("gem available: %v", available)
}

func TestGemManager_ContextCancellation(t *testing.T) {
	manager := NewGemManager()

	// First check if gem is available - if not, skip context tests
	ctx := context.Background()
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if gem is available: %v", err)
	}
	if !available {
		t.Skip("gem not available, skipping context cancellation tests")
	}

	t.Run("ListInstalled_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.ListInstalled(ctx)
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("ListInstalled_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		_, err := manager.ListInstalled(ctx)
		if err == nil {
			t.Error("Expected error when context times out")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context timeout error, got %v", err)
		}
	})

	t.Run("Install_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := manager.Install(ctx, "nonexistent-gem")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("Uninstall_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := manager.Uninstall(ctx, "nonexistent-gem")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("IsInstalled_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.IsInstalled(ctx, "nonexistent-gem")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("Search_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.Search(ctx, "test-query")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("Info_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.Info(ctx, "nonexistent-gem")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("GetInstalledVersion_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.GetInstalledVersion(ctx, "nonexistent-gem")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})
}

func TestGemManager_hasExecutables(t *testing.T) {
	manager := NewGemManager()
	ctx := context.Background()

	// Skip if gem is not available
	available, _ := manager.IsAvailable(ctx)
	if !available {
		t.Skip("gem not available, skipping executable detection tests")
	}

	t.Run("gem with executables", func(t *testing.T) {
		// bundler should have executables
		if manager.hasExecutables(ctx, "bundler") {
			t.Log("bundler correctly detected as having executables")
		} else {
			t.Log("bundler not installed or detection failed")
		}
	})
}

func TestGemManager_ErrorScenarios(t *testing.T) {
	manager := NewGemManager()
	ctx := context.Background()

	// Skip if gem is not available
	available, _ := manager.IsAvailable(ctx)
	if !available {
		t.Skip("gem not available, skipping error scenario tests")
	}

	t.Run("Install_NonexistentGem", func(t *testing.T) {
		// Use a clearly non-existent gem name
		err := manager.Install(ctx, "this-gem-definitely-does-not-exist-12345")
		if err == nil {
			t.Skip("Gem unexpectedly exists or install succeeded")
		}

		// Should get a gem not found error
		plonkErr, ok := err.(*errors.PlonkError)
		if ok && plonkErr.Code == errors.ErrPackageNotFound {
			t.Logf("Got expected gem not found error: %v", err)
		} else {
			// Some other error occurred (network issues, etc)
			t.Logf("Got different error: %v", err)
		}
	})

	t.Run("GetInstalledVersion_NotInstalled", func(t *testing.T) {
		// Try to get version of a gem that's not installed
		_, err := manager.GetInstalledVersion(ctx, "this-gem-is-not-installed-99999")
		if err == nil {
			t.Error("Expected error for non-installed gem")
		}

		plonkErr, ok := err.(*errors.PlonkError)
		if ok && plonkErr.Code != errors.ErrPackageNotFound {
			t.Errorf("Expected ErrPackageNotFound, got %v", plonkErr.Code)
		}
	})

	t.Run("Info_NonexistentGem", func(t *testing.T) {
		_, err := manager.Info(ctx, "nonexistent-gem-info-test-54321")
		if err == nil {
			t.Error("Expected error for non-existent gem")
		}

		// Check for gem not found error
		plonkErr, ok := err.(*errors.PlonkError)
		if ok && plonkErr.Code == errors.ErrPackageNotFound {
			t.Logf("Got expected gem not found error: %v", err)
		}
	})
}

func TestGemManager_Search(t *testing.T) {
	manager := NewGemManager()
	ctx := context.Background()

	// Skip if gem is not available
	available, _ := manager.IsAvailable(ctx)
	if !available {
		t.Skip("gem not available, skipping search tests")
	}

	t.Run("search for common gem", func(t *testing.T) {
		gems, err := manager.Search(ctx, "bundler")
		if err != nil {
			t.Errorf("Search failed: %v", err)
			return
		}

		if len(gems) > 0 {
			t.Logf("Found %d gems matching 'bundler'", len(gems))
			// Check if bundler is in the results
			found := false
			for _, gem := range gems {
				if gem == "bundler" {
					found = true
					break
				}
			}
			if found {
				t.Log("bundler found in search results")
			} else {
				t.Log("bundler not found in search results")
			}
		} else {
			t.Log("No gems found matching 'bundler'")
		}
	})

	t.Run("search for non-existent gem", func(t *testing.T) {
		gems, err := manager.Search(ctx, "xyznonexistentgem123456")
		if err != nil {
			t.Errorf("Search failed: %v", err)
			return
		}

		if len(gems) == 0 {
			t.Log("Correctly returned empty results for non-existent gem")
		} else {
			t.Errorf("Unexpected results found: %v", gems)
		}
	})
}

func TestGemManager_VersionParsing(t *testing.T) {
	manager := NewGemManager()
	ctx := context.Background()

	// Skip if gem is not available
	available, _ := manager.IsAvailable(ctx)
	if !available {
		t.Skip("gem not available, skipping version parsing tests")
	}

	// Test with a gem that's likely to be installed
	installed, err := manager.IsInstalled(ctx, "bundler")
	if err != nil {
		t.Logf("Failed to check if bundler is installed: %v", err)
		return
	}

	if installed {
		version, err := manager.GetInstalledVersion(ctx, "bundler")
		if err != nil {
			t.Errorf("Failed to get bundler version: %v", err)
		} else {
			t.Logf("bundler version: %s", version)
			// Version should not be empty
			if version == "" {
				t.Error("Version should not be empty for installed gem")
			}
		}
	} else {
		t.Log("bundler not installed, skipping version test")
	}
}
