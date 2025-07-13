// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package testing provides common test utilities and helpers for plonk tests.
// This package helps reduce test boilerplate and ensures consistent test patterns.
package testing

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/mocks"
	"go.uber.org/mock/gomock"
)

// TestContext provides a common setup for tests with temporary directories
type TestContext struct {
	T         *testing.T
	TempDir   string
	ConfigDir string
	HomeDir   string
	CleanupFn func()
}

// NewTestContext creates a new test context with temporary directories
func NewTestContext(t *testing.T) *TestContext {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	homeDir := filepath.Join(tempDir, "home")

	// Create directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}

	return &TestContext{
		T:         t,
		TempDir:   tempDir,
		ConfigDir: configDir,
		HomeDir:   homeDir,
		CleanupFn: func() {}, // t.TempDir() handles cleanup automatically
	}
}

// WithConfig creates a config file in the test context
func (tc *TestContext) WithConfig(cfg *config.Config) *TestContext {
	configManager := config.NewConfigManager(tc.ConfigDir)
	if err := configManager.Save(cfg); err != nil {
		tc.T.Fatalf("Failed to save test config: %v", err)
	}

	return tc
}

// WithFile creates a file with given content in the test context
func (tc *TestContext) WithFile(path, content string) *TestContext {
	fullPath := filepath.Join(tc.TempDir, path)

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		tc.T.Fatalf("Failed to create parent directory for %s: %v", path, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		tc.T.Fatalf("Failed to create test file %s: %v", path, err)
	}

	return tc
}

// MockManagerSetup provides common mock manager setup
type MockManagerSetup struct {
	Ctrl    *gomock.Controller
	Manager *mocks.MockPackageManager
	Config  *mocks.MockPackageConfigLoader
}

// NewMockManagerSetup creates a new mock manager setup
func NewMockManagerSetup(t *testing.T) *MockManagerSetup {
	ctrl := gomock.NewController(t)
	return &MockManagerSetup{
		Ctrl:    ctrl,
		Manager: mocks.NewMockPackageManager(ctrl),
		Config:  mocks.NewMockPackageConfigLoader(ctrl),
	}
}

// WithAvailability sets up the manager availability expectation
func (ms *MockManagerSetup) WithAvailability(available bool) *MockManagerSetup {
	ms.Manager.EXPECT().
		IsAvailable(gomock.Any()).
		Return(available, nil).
		AnyTimes()
	return ms
}

// WithPackages sets up the manager to return specific installed packages
func (ms *MockManagerSetup) WithPackages(packages []string) *MockManagerSetup {
	ms.Manager.EXPECT().
		ListInstalled(gomock.Any()).
		Return(packages, nil).
		AnyTimes()
	return ms
}

// TestTimeout creates a context with a reasonable test timeout
func TestTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// AssertNoError is a helper that fails the test if err is not nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected no error but got: %v. %v", err, msgAndArgs[0])
		} else {
			t.Fatalf("Expected no error but got: %v", err)
		}
	}
}

// AssertError is a helper that fails the test if err is nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected an error but got nil. %v", msgAndArgs[0])
		} else {
			t.Fatalf("Expected an error but got nil")
		}
	}
}

// AssertStringContains checks if a string contains a substring
func AssertStringContains(t *testing.T, str, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	if !contains(str, substr) {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected string to contain %q but got %q. %v", substr, str, msgAndArgs[0])
		} else {
			t.Fatalf("Expected string to contain %q but got %q", substr, str)
		}
	}
}

// contains is a simple string contains check
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (len(substr) == 0 || findIndex(str, substr) >= 0)
}

// findIndex finds the index of substr in str, returns -1 if not found
func findIndex(str, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
