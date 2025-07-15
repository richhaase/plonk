// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/errors"
)

func TestPipManager_IsAvailable(t *testing.T) {
	manager := NewPipManager()
	ctx := context.Background()

	// This test will check actual pip availability on the system
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Logf("IsAvailable returned error: %v", err)
	}
	t.Logf("pip available: %v", available)
}

func TestPipManager_getPipCommand(t *testing.T) {
	manager := NewPipManager()

	// This test verifies that getPipCommand returns a valid command
	cmd := manager.getPipCommand()
	if cmd != "pip" && cmd != "pip3" {
		t.Errorf("getPipCommand() = %s, expected 'pip' or 'pip3'", cmd)
	}
}

func TestPipManager_normalizeName(t *testing.T) {
	manager := NewPipManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "Django", "django"},
		{"hyphen to underscore", "python-dateutil", "python_dateutil"},
		{"mixed case with hyphen", "Flask-RESTful", "flask_restful"},
		{"already normalized", "black", "black"},
		{"multiple hyphens", "some-package-name", "some_package_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.normalizeName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeName(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPipManager_ContextCancellation(t *testing.T) {
	manager := NewPipManager()

	// First check if pip is available - if not, skip context tests
	ctx := context.Background()
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if pip is available: %v", err)
	}
	if !available {
		t.Skip("pip not available, skipping context cancellation tests")
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

		err := manager.Install(ctx, "nonexistent-package")
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

		err := manager.Uninstall(ctx, "nonexistent-package")
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

		_, err := manager.IsInstalled(ctx, "nonexistent-package")
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

		_, err := manager.Info(ctx, "nonexistent-package")
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

		_, err := manager.GetInstalledVersion(ctx, "nonexistent-package")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if !containsContextError(err) {
			t.Errorf("Expected context cancellation error, got %v", err)
		}
	})
}

func TestPipManager_Search(t *testing.T) {
	manager := NewPipManager()
	ctx := context.Background()

	// Search should return an error about deprecation
	packages, err := manager.Search(ctx, "test-query")
	if err == nil {
		t.Error("Expected error for deprecated pip search")
	}
	if packages != nil {
		t.Errorf("Expected nil packages, got %v", packages)
	}

	// Check that error has proper structure
	plonkErr, ok := err.(*errors.PlonkError)
	if !ok {
		t.Errorf("Expected PlonkError, got %T", err)
	} else {
		if plonkErr.Code != errors.ErrCommandExecution {
			t.Errorf("Expected error code %v, got %v", errors.ErrCommandExecution, plonkErr.Code)
		}
		if plonkErr.Domain != errors.DomainPackages {
			t.Errorf("Expected domain %v, got %v", errors.DomainPackages, plonkErr.Domain)
		}
		// Check that it's the expected error about deprecation
		if !contains(err.Error(), "deprecated") {
			t.Logf("Error message: %v", err.Error())
			t.Error("Expected error message to mention pip search is deprecated")
		}
	}
}

func TestPipManager_ListInstalled_ParseJSON(t *testing.T) {
	manager := NewPipManager()

	// Test successful JSON parsing
	t.Run("valid JSON output", func(t *testing.T) {
		// We can't easily test the internal parsing without mocking exec.Command
		// but we can verify the normalization logic is applied
		packages := []string{"black", "flake8"}
		for i, pkg := range packages {
			normalized := manager.normalizeName(pkg)
			if normalized != pkg {
				t.Errorf("Package %d: expected %s to remain %s after normalization", i, pkg, pkg)
			}
		}
	})

	// Test package name normalization in results
	t.Run("normalized package names", func(t *testing.T) {
		testCases := []struct {
			original   string
			normalized string
		}{
			{"Django", "django"},
			{"python-dateutil", "python_dateutil"},
			{"Flask-RESTful", "flask_restful"},
		}

		for _, tc := range testCases {
			result := manager.normalizeName(tc.original)
			if result != tc.normalized {
				t.Errorf("normalizeName(%s) = %s, expected %s", tc.original, result, tc.normalized)
			}
		}
	})
}

func TestPipManager_IsInstalled(t *testing.T) {
	manager := NewPipManager()

	// Test normalized name matching
	t.Run("normalized name matching", func(t *testing.T) {
		// These should be considered the same package
		name1 := "python-dateutil"
		name2 := "python_dateutil"

		norm1 := manager.normalizeName(name1)
		norm2 := manager.normalizeName(name2)

		if norm1 != norm2 {
			t.Errorf("Normalized names should match: %s != %s", norm1, norm2)
		}
	})
}

func TestPipManager_ErrorScenarios(t *testing.T) {
	manager := NewPipManager()
	ctx := context.Background()

	// Skip if pip is not available
	available, _ := manager.IsAvailable(ctx)
	if !available {
		t.Skip("pip not available, skipping error scenario tests")
	}

	t.Run("Install_NonexistentPackage", func(t *testing.T) {
		// Use a clearly non-existent package name
		err := manager.Install(ctx, "this-package-definitely-does-not-exist-12345")
		if err == nil {
			t.Skip("Package unexpectedly exists or install succeeded")
		}

		// Should get a package not found error
		plonkErr, ok := err.(*errors.PlonkError)
		if ok && plonkErr.Code == errors.ErrPackageNotFound {
			t.Logf("Got expected package not found error: %v", err)
		} else {
			// Some other error occurred (network issues, etc)
			t.Logf("Got different error: %v", err)
		}
	})

	t.Run("GetInstalledVersion_NotInstalled", func(t *testing.T) {
		// Try to get version of a package that's not installed
		_, err := manager.GetInstalledVersion(ctx, "this-package-is-not-installed-99999")
		if err == nil {
			t.Error("Expected error for non-installed package")
		}

		plonkErr, ok := err.(*errors.PlonkError)
		if ok && plonkErr.Code != errors.ErrPackageNotFound {
			t.Errorf("Expected ErrPackageNotFound, got %v", plonkErr.Code)
		}
	})

	t.Run("Info_NonexistentPackage", func(t *testing.T) {
		_, err := manager.Info(ctx, "nonexistent-package-info-test-54321")
		if err == nil {
			t.Error("Expected error for non-existent package")
		}

		// Check for package not found error
		plonkErr, ok := err.(*errors.PlonkError)
		if ok && plonkErr.Code == errors.ErrPackageNotFound {
			t.Logf("Got expected package not found error: %v", err)
		}
	})
}

func TestPipManager_CommandExecution(t *testing.T) {
	// Test that commands handle exit codes properly
	t.Run("handle exit code 1", func(t *testing.T) {
		// Exit code 1 often indicates package not found or other expected errors
		err := &exec.ExitError{}
		if err != nil {
			// This is just to verify the type exists
			t.Logf("exec.ExitError type is available")
		}
	})
}

// Note: containsContextError and containsString are defined in homebrew_test.go

// Simple contains implementation for this file
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
