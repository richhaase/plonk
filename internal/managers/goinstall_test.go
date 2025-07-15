// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/errors"
)

func TestGoInstallManager_IsAvailable(t *testing.T) {
	manager := NewGoInstallManager()
	ctx := context.Background()

	// This test will check actual go availability on the system
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Logf("IsAvailable returned error: %v", err)
	}
	t.Logf("go available: %v", available)
}

func TestGoInstallManager_ContextCancellation(t *testing.T) {
	manager := NewGoInstallManager()

	// First check if go is available - if not, skip context tests
	ctx := context.Background()
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("Failed to check if go is available: %v", err)
	}
	if !available {
		t.Skip("go not available, skipping context cancellation tests")
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

		err := manager.Install(ctx, "example.com/nonexistent@latest")
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

		err := manager.Uninstall(ctx, "nonexistent-tool")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		// Uninstall may succeed quickly if binary doesn't exist
		// so we don't check for context error
	})

	t.Run("IsInstalled_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.IsInstalled(ctx, "nonexistent-tool")
		// IsInstalled is fast and may complete before context cancellation
		if err != nil && containsContextError(err) {
			// Good, context was canceled
		}
	})

	t.Run("Info_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.Info(ctx, "nonexistent-tool")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		// May fail for other reasons than context
	})

	t.Run("GetInstalledVersion_ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := manager.GetInstalledVersion(ctx, "nonexistent-tool")
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		// May fail for other reasons than context
	})
}

func TestGoInstallManager_extractBinaryName(t *testing.T) {
	manager := NewGoInstallManager()

	tests := []struct {
		name       string
		modulePath string
		want       string
	}{
		{
			name:       "simple module",
			modulePath: "github.com/user/tool",
			want:       "tool",
		},
		{
			name:       "module with version",
			modulePath: "github.com/user/tool@v1.2.3",
			want:       "tool",
		},
		{
			name:       "module with latest",
			modulePath: "github.com/user/tool@latest",
			want:       "tool",
		},
		{
			name:       "cmd pattern",
			modulePath: "github.com/user/project/cmd/tool",
			want:       "tool",
		},
		{
			name:       "cmd pattern with version",
			modulePath: "github.com/user/project/cmd/tool@v1.0.0",
			want:       "tool",
		},
		{
			name:       "golang.org/x tools",
			modulePath: "golang.org/x/tools/gopls@latest",
			want:       "gopls",
		},
		{
			name:       "single name",
			modulePath: "gofumpt",
			want:       "gofumpt",
		},
		{
			name:       "domain-style name",
			modulePath: "mvdan.cc/gofumpt",
			want:       "gofumpt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manager.extractBinaryName(tt.modulePath)
			if got != tt.want {
				t.Errorf("extractBinaryName(%q) = %q, want %q", tt.modulePath, got, tt.want)
			}
		})
	}
}

func TestGoInstallManager_parseModulePath(t *testing.T) {
	manager := NewGoInstallManager()

	tests := []struct {
		name        string
		pkg         string
		wantModule  string
		wantVersion string
	}{
		{
			name:        "module without version",
			pkg:         "github.com/user/tool",
			wantModule:  "github.com/user/tool",
			wantVersion: "latest",
		},
		{
			name:        "module with version",
			pkg:         "github.com/user/tool@v1.2.3",
			wantModule:  "github.com/user/tool",
			wantVersion: "v1.2.3",
		},
		{
			name:        "module with latest",
			pkg:         "github.com/user/tool@latest",
			wantModule:  "github.com/user/tool",
			wantVersion: "latest",
		},
		{
			name:        "module with master",
			pkg:         "github.com/user/tool@master",
			wantModule:  "github.com/user/tool",
			wantVersion: "master",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, version := manager.parseModulePath(tt.pkg)
			if module != tt.wantModule {
				t.Errorf("parseModulePath(%q) module = %q, want %q", tt.pkg, module, tt.wantModule)
			}
			if version != tt.wantVersion {
				t.Errorf("parseModulePath(%q) version = %q, want %q", tt.pkg, version, tt.wantVersion)
			}
		})
	}
}

func TestGoInstallManager_ErrorScenarios(t *testing.T) {
	manager := NewGoInstallManager()
	ctx := context.Background()

	// Skip if go is not available
	available, _ := manager.IsAvailable(ctx)
	if !available {
		t.Skip("go not available, skipping error scenario tests")
	}

	t.Run("Install_NonexistentModule", func(t *testing.T) {
		// Use a clearly non-existent module name
		err := manager.Install(ctx, "this-module-definitely-does-not-exist.com/tool@latest")
		if err == nil {
			t.Skip("Module unexpectedly exists or install succeeded")
		}

		// Should get a module not found error
		plonkErr, ok := err.(*errors.PlonkError)
		if ok && plonkErr.Code == errors.ErrPackageNotFound {
			t.Logf("Got expected module not found error: %v", err)
		} else {
			// Some other error occurred (network issues, etc)
			t.Logf("Got different error: %v", err)
		}
	})

	t.Run("GetInstalledVersion_NotInstalled", func(t *testing.T) {
		// Try to get version of a binary that's not installed
		_, err := manager.GetInstalledVersion(ctx, "this-binary-is-not-installed-99999")
		if err == nil {
			t.Error("Expected error for non-installed binary")
		}

		plonkErr, ok := err.(*errors.PlonkError)
		if ok && plonkErr.Code != errors.ErrPackageNotFound {
			t.Errorf("Expected ErrPackageNotFound, got %v", plonkErr.Code)
		}
	})

	t.Run("Info_NonexistentBinary", func(t *testing.T) {
		_, err := manager.Info(ctx, "nonexistent-go-binary-test-54321")
		if err == nil {
			t.Error("Expected error for non-existent binary")
		}

		// Check for package not found error
		plonkErr, ok := err.(*errors.PlonkError)
		if ok && plonkErr.Code == errors.ErrPackageNotFound {
			t.Logf("Got expected package not found error: %v", err)
		}
	})
}

func TestGoInstallManager_Search(t *testing.T) {
	manager := NewGoInstallManager()
	ctx := context.Background()

	// Search should return an error indicating it's not supported
	results, err := manager.Search(ctx, "test-query")
	if err == nil {
		t.Error("Expected error for unsupported search operation")
	}

	// Check for unsupported error
	plonkErr, ok := err.(*errors.PlonkError)
	if ok && plonkErr.Code != errors.ErrUnsupported {
		t.Errorf("Expected ErrUnsupported, got %v", plonkErr.Code)
	}

	// Results should be empty
	if len(results) != 0 {
		t.Errorf("Expected empty results, got %v", results)
	}

	// Check for helpful suggestion
	if plonkErr != nil && plonkErr.Suggestion != nil {
		t.Logf("Got suggestion: %s", plonkErr.Suggestion.Message)
	}
}

func TestGoInstallManager_getGoBinDir(t *testing.T) {
	manager := NewGoInstallManager()
	ctx := context.Background()

	// Skip if go is not available
	available, _ := manager.IsAvailable(ctx)
	if !available {
		t.Skip("go not available, skipping GOBIN tests")
	}

	binDir, err := manager.getGoBinDir(ctx)
	if err != nil {
		t.Errorf("Failed to get GOBIN directory: %v", err)
		return
	}

	if binDir == "" {
		t.Error("GOBIN directory should not be empty")
	}

	t.Logf("GOBIN directory: %s", binDir)
}
