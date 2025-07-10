package managers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCargoManager_IsAvailable(t *testing.T) {
	manager := NewCargoManager()
	ctx := context.Background()

	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Errorf("IsAvailable should not return an error: %v", err)
	}

	// Log availability for debugging
	t.Logf("Cargo available: %v", available)

	// Test behavior when cargo is not available
	if !available {
		t.Run("operations_when_unavailable", func(t *testing.T) {
			_, err := manager.ListInstalled(ctx)
			if err == nil {
				t.Error("Expected error when cargo is not available")
			}
			t.Logf("ListInstalled error (expected): %v", err)

			err = manager.Install(ctx, "test-package")
			if err == nil {
				t.Error("Expected error when cargo is not available")
			}
			t.Logf("Install error (expected): %v", err)

			err = manager.Uninstall(ctx, "test-package")
			if err == nil {
				t.Error("Expected error when cargo is not available")
			}
			t.Logf("Uninstall error (expected): %v", err)

			_, err = manager.IsInstalled(ctx, "test-package")
			if err == nil {
				t.Error("Expected error when cargo is not available")
			}
			t.Logf("IsInstalled error (expected): %v", err)
		})
	}
}

func TestCargoManager_ContextCancellation(t *testing.T) {
	manager := NewCargoManager()

	// First check if cargo is available - if not, skip context tests
	ctx := context.Background()
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if cargo is available: %v", err)
	}
	if !available {
		t.Skip("Cargo not available, skipping context cancellation tests")
	}

	t.Run("ListInstalled_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.ListInstalled(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("ListInstalled_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		_, err := manager.ListInstalled(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("Install_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := manager.Install(ctx, "nonexistent-package")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("Install_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		err := manager.Install(ctx, "nonexistent-package")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("Uninstall_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := manager.Uninstall(ctx, "nonexistent-package")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("Uninstall_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		err := manager.Uninstall(ctx, "nonexistent-package")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("IsInstalled_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.IsInstalled(ctx, "nonexistent-package")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("IsInstalled_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		_, err := manager.IsInstalled(ctx, "nonexistent-package")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("IsAvailable_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.IsAvailable(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("IsAvailable_ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		_, err := manager.IsAvailable(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}
