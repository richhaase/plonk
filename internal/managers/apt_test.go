// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/errors"
)

func TestAptManager_IsAvailable(t *testing.T) {
	manager := NewAptManager()
	ctx := context.Background()

	// This test will check actual apt availability on the system
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Logf("IsAvailable returned error: %v", err)
	}

	// APT should only be available on Linux
	if runtime.GOOS != "linux" {
		if available {
			t.Error("APT should not be available on non-Linux systems")
		}
	}

	t.Logf("apt available: %v (OS: %s)", available, runtime.GOOS)
}

func TestAptManager_ContextCancellation(t *testing.T) {
	manager := NewAptManager()

	// Skip these tests on non-Linux systems
	if runtime.GOOS != "linux" {
		t.Skip("Skipping APT tests on non-Linux system")
	}

	// First check if apt is available - if not, skip context tests
	ctx := context.Background()
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if apt is available: %v", err)
	}
	if !available {
		t.Skip("apt not available, skipping context cancellation tests")
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

		err := manager.Install(ctx, "curl")
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

		err := manager.Uninstall(ctx, "curl")
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

		_, err := manager.IsInstalled(ctx, "curl")
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

		_, err := manager.Search(ctx, "curl")
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

		_, err := manager.Info(ctx, "curl")
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

		_, err := manager.GetInstalledVersion(ctx, "curl")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})
}

func TestAptManager_PermissionErrors(t *testing.T) {
	manager := NewAptManager()
	ctx := context.Background()

	// Skip these tests on non-Linux systems
	if runtime.GOOS != "linux" {
		t.Skip("Skipping APT tests on non-Linux system")
	}

	// Check if apt is available
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if apt is available: %v", err)
	}
	if !available {
		t.Skip("apt not available, skipping permission tests")
	}

	// Check if we're running as root - if so, skip permission tests
	if isRunningAsRoot() {
		t.Skip("Running as root, skipping permission tests")
	}

	t.Run("Install_PermissionDenied", func(t *testing.T) {
		// Try to install a package without sudo - should fail with permission error
		err := manager.Install(ctx, "this-package-does-not-exist-12345")
		if err == nil {
			t.Error("Expected permission error when installing without sudo")
			return
		}

		plonkErr, ok := err.(*errors.PlonkError)
		if !ok {
			t.Errorf("Expected PlonkError, got %T", err)
			return
		}
		if plonkErr.Code != errors.ErrFilePermission {
			t.Errorf("Expected ErrFilePermission, got %v", plonkErr.Code)
		}
		if !strings.Contains(err.Error(), "sudo") {
			t.Errorf("Expected 'sudo' in error message, got %v", err)
		}
	})

	t.Run("Uninstall_PermissionDenied", func(t *testing.T) {
		// Try to uninstall a package without sudo - should fail with permission error
		err := manager.Uninstall(ctx, "this-package-does-not-exist-12345")
		if err == nil {
			t.Error("Expected permission error when uninstalling without sudo")
			return
		}

		plonkErr, ok := err.(*errors.PlonkError)
		if !ok {
			t.Errorf("Expected PlonkError, got %T", err)
			return
		}
		if plonkErr.Code != errors.ErrFilePermission {
			t.Errorf("Expected ErrFilePermission, got %v", plonkErr.Code)
		}
		if !strings.Contains(err.Error(), "sudo") {
			t.Errorf("Expected 'sudo' in error message, got %v", err)
		}
	})
}

func TestAptManager_ListInstalled(t *testing.T) {
	manager := NewAptManager()
	ctx := context.Background()

	// Skip these tests on non-Linux systems
	if runtime.GOOS != "linux" {
		t.Skip("Skipping APT tests on non-Linux system")
	}

	// Check if apt is available
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if apt is available: %v", err)
	}
	if !available {
		t.Skip("apt not available, skipping list tests")
	}

	packages, err := manager.ListInstalled(ctx)
	if err != nil {
		t.Errorf("ListInstalled() error = %v", err)
		return
	}

	t.Logf("Found %d installed packages", len(packages))

	// On a Debian-based system, there should be at least some packages
	if len(packages) == 0 {
		t.Log("Warning: No packages found - this might be unexpected on a real system")
	}

	// Show a sample of packages
	for i, pkg := range packages {
		if i < 5 {
			t.Logf("  - %s", pkg)
		}
	}
}

func TestAptManager_Search(t *testing.T) {
	manager := NewAptManager()
	ctx := context.Background()

	// Skip these tests on non-Linux systems
	if runtime.GOOS != "linux" {
		t.Skip("Skipping APT tests on non-Linux system")
	}

	// Check if apt is available
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if apt is available: %v", err)
	}
	if !available {
		t.Skip("apt not available, skipping search tests")
	}

	// Search for a common package
	results, err := manager.Search(ctx, "curl")
	if err != nil {
		t.Errorf("Search() error = %v", err)
		return
	}

	t.Logf("Found %d packages matching 'curl'", len(results))

	// Should find at least the curl package itself
	foundCurl := false
	for _, pkg := range results {
		if pkg == "curl" {
			foundCurl = true
			break
		}
		// Show first few results
		if len(results) < 10 {
			t.Logf("  - %s", pkg)
		}
	}

	if !foundCurl && len(results) > 0 {
		t.Log("Warning: 'curl' package not found in search results")
	}
}

func TestAptManager_IsInstalled(t *testing.T) {
	manager := NewAptManager()
	ctx := context.Background()

	// Skip these tests on non-Linux systems
	if runtime.GOOS != "linux" {
		t.Skip("Skipping APT tests on non-Linux system")
	}

	// Check if apt is available
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if apt is available: %v", err)
	}
	if !available {
		t.Skip("apt not available, skipping IsInstalled tests")
	}

	// Check a package that's likely to be installed on most systems
	installed, err := manager.IsInstalled(ctx, "apt")
	if err != nil {
		t.Errorf("IsInstalled() error = %v", err)
		return
	}

	if !installed {
		t.Log("Warning: 'apt' package not installed - this is unusual for a Debian-based system")
	}

	// Check a package that definitely doesn't exist
	installed, err = manager.IsInstalled(ctx, "this-package-definitely-does-not-exist-12345")
	if err != nil {
		t.Errorf("IsInstalled() error = %v", err)
		return
	}

	if installed {
		t.Error("Non-existent package should not be installed")
	}
}

func TestAptManager_Info(t *testing.T) {
	manager := NewAptManager()
	ctx := context.Background()

	// Skip these tests on non-Linux systems
	if runtime.GOOS != "linux" {
		t.Skip("Skipping APT tests on non-Linux system")
	}

	// Check if apt is available
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if apt is available: %v", err)
	}
	if !available {
		t.Skip("apt not available, skipping info tests")
	}

	// Get info for a common package
	info, err := manager.Info(ctx, "curl")
	if err != nil {
		// It's okay if curl info is not available
		t.Logf("Info() error = %v (this is okay if curl is not in the package cache)", err)
		return
	}

	if info.Name != "curl" {
		t.Errorf("Info().Name = %v, want curl", info.Name)
	}
	if info.Manager != "apt" {
		t.Errorf("Info().Manager = %v, want apt", info.Manager)
	}

	t.Logf("Package info: %+v", info)
}

func TestAptManager_PackageNotFound(t *testing.T) {
	manager := NewAptManager()
	ctx := context.Background()

	// Skip these tests on non-Linux systems
	if runtime.GOOS != "linux" {
		t.Skip("Skipping APT tests on non-Linux system")
	}

	// Check if apt is available
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if apt is available: %v", err)
	}
	if !available {
		t.Skip("apt not available, skipping package not found tests")
	}

	// Try to get info for non-existent package
	_, err = manager.Info(ctx, "this-package-definitely-does-not-exist-12345")
	if err == nil {
		t.Error("Expected error when getting info for non-existent package")
		return
	}

	plonkErr, ok := err.(*errors.PlonkError)
	if !ok {
		t.Errorf("Expected PlonkError, got %T", err)
		return
	}
	if plonkErr.Code != errors.ErrPackageNotFound {
		t.Errorf("Expected ErrPackageNotFound, got %v", plonkErr.Code)
	}
}

func TestAptManager_GetInstalledVersionNotInstalled(t *testing.T) {
	manager := NewAptManager()
	ctx := context.Background()

	// Skip these tests on non-Linux systems
	if runtime.GOOS != "linux" {
		t.Skip("Skipping APT tests on non-Linux system")
	}

	// Check if apt is available
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if apt is available: %v", err)
	}
	if !available {
		t.Skip("apt not available, skipping version tests")
	}

	// Try to get version for non-installed package
	_, err = manager.GetInstalledVersion(ctx, "this-package-definitely-does-not-exist-12345")
	if err == nil {
		t.Error("Expected error when getting version for non-installed package")
		return
	}

	plonkErr, ok := err.(*errors.PlonkError)
	if !ok {
		t.Errorf("Expected PlonkError, got %T", err)
		return
	}
	if plonkErr.Code != errors.ErrPackageNotFound {
		t.Errorf("Expected ErrPackageNotFound, got %v", plonkErr.Code)
	}
}

// Helper function to check if running as root
func isRunningAsRoot() bool {
	// This is a simple check - in production you might use os.Geteuid() == 0
	// but that's not available on all platforms
	return false // For testing, assume not root unless we know otherwise
}
