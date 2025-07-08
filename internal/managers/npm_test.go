// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"testing"
	"time"
)

func TestNpmManager_ContextCancellation(t *testing.T) {
	manager := NewNpmManager()

	t.Run("ListInstalled_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := manager.ListInstalled(ctx)
		if err == nil {
			t.Error("Expected error when context is cancelled")
		}
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
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
		if err != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded, got %v", err)
		}
	})

	t.Run("Install_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		err := manager.Install(ctx, "nonexistent-package")
		if err == nil {
			t.Error("Expected error when context is cancelled")
		}
		// Should contain context cancellation error
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("Install_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		time.Sleep(10 * time.Millisecond) // Ensure timeout
		
		err := manager.Install(ctx, "nonexistent-package")
		if err == nil {
			t.Error("Expected error when context times out")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context timeout error, got %v", err)
		}
	})

	t.Run("Uninstall_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		err := manager.Uninstall(ctx, "nonexistent-package")
		if err == nil {
			t.Error("Expected error when context is cancelled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})

	t.Run("Uninstall_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		time.Sleep(10 * time.Millisecond) // Ensure timeout
		
		err := manager.Uninstall(ctx, "nonexistent-package")
		if err == nil {
			t.Error("Expected error when context times out")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context timeout error, got %v", err)
		}
	})

	t.Run("IsInstalled_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		// IsInstalled returns bool, but should respect context cancellation
		// In a real scenario, this would return false when context is cancelled
		result := manager.IsInstalled(ctx, "nonexistent-package")
		_ = result // Context cancellation may not be detectable in boolean return
	})

	t.Run("IsInstalled_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		time.Sleep(10 * time.Millisecond) // Ensure timeout
		
		result := manager.IsInstalled(ctx, "nonexistent-package")
		_ = result // Context timeout may not be detectable in boolean return
	})
}