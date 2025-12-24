// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]CommandResponse
		expected  bool
	}{
		{
			name: "go available",
			responses: map[string]CommandResponse{
				"go version": {Output: []byte("go version go1.21.0 darwin/arm64")},
			},
			expected: true,
		},
		{
			name: "go command fails",
			responses: map[string]CommandResponse{
				"go version": {Error: fmt.Errorf("command failed")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewGoManager(mock)

			result, err := mgr.IsAvailable(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoManager_IsAvailable_NotInPath(t *testing.T) {
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{},
	}
	mgr := NewGoManager(mock)

	result, err := mgr.IsAvailable(context.Background())

	assert.NoError(t, err)
	assert.False(t, result)
}

func TestGoManager_ListInstalled(t *testing.T) {
	// Create a temp directory to simulate GOBIN
	tmpDir := t.TempDir()
	t.Setenv("GOBIN", tmpDir)

	// Create some fake binaries
	for _, name := range []string{"gopls", "staticcheck", "goimports"} {
		f, err := os.Create(filepath.Join(tmpDir, name))
		require.NoError(t, err)
		f.Close()
	}

	// Create a hidden file that should be skipped
	f, err := os.Create(filepath.Join(tmpDir, ".hidden"))
	require.NoError(t, err)
	f.Close()

	// Create a directory that should be skipped
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755))

	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
	mgr := NewGoManager(mock)

	result, err := mgr.ListInstalled(context.Background())

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"gopls", "staticcheck", "goimports"}, result)
}

func TestGoManager_ListInstalled_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("GOBIN", tmpDir)

	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
	mgr := NewGoManager(mock)

	result, err := mgr.ListInstalled(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGoManager_ListInstalled_NonexistentDir(t *testing.T) {
	t.Setenv("GOBIN", "/nonexistent/path/that/doesnt/exist")

	mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
	mgr := NewGoManager(mock)

	result, err := mgr.ListInstalled(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGoManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		expectedPkg string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "install with @latest",
			pkg:         "github.com/golangci/golangci-lint/cmd/golangci-lint",
			expectedPkg: "github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
			output:      "",
			err:         nil,
			expectError: false,
		},
		{
			name:        "install with version",
			pkg:         "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.0",
			expectedPkg: "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.0",
			output:      "",
			err:         nil,
			expectError: false,
		},
		{
			name:        "install failure",
			pkg:         "github.com/nonexistent/package",
			expectedPkg: "github.com/nonexistent/package@latest",
			output:      "go: module not found",
			err:         fmt.Errorf("exit status 1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"go install " + tt.expectedPkg: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewGoManager(mock)

			err := mgr.Install(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGoManager_Uninstall(t *testing.T) {
	t.Run("successful uninstall", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("GOBIN", tmpDir)

		// Create the binary to uninstall
		binPath := filepath.Join(tmpDir, "gopls")
		f, err := os.Create(binPath)
		require.NoError(t, err)
		f.Close()

		mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
		mgr := NewGoManager(mock)

		err = mgr.Uninstall(context.Background(), "gopls")

		assert.NoError(t, err)
		_, err = os.Stat(binPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("uninstall package path extracts binary name", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("GOBIN", tmpDir)

		// Create the binary
		binPath := filepath.Join(tmpDir, "golangci-lint")
		f, err := os.Create(binPath)
		require.NoError(t, err)
		f.Close()

		mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
		mgr := NewGoManager(mock)

		err = mgr.Uninstall(context.Background(), "github.com/golangci/golangci-lint/cmd/golangci-lint")

		assert.NoError(t, err)
		_, err = os.Stat(binPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("uninstall strips version", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("GOBIN", tmpDir)

		binPath := filepath.Join(tmpDir, "gopls")
		f, err := os.Create(binPath)
		require.NoError(t, err)
		f.Close()

		mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
		mgr := NewGoManager(mock)

		err = mgr.Uninstall(context.Background(), "gopls@v0.14.0")

		assert.NoError(t, err)
		_, err = os.Stat(binPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("uninstall nonexistent is idempotent", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("GOBIN", tmpDir)

		mock := &MockCommandExecutor{Responses: map[string]CommandResponse{}}
		mgr := NewGoManager(mock)

		err := mgr.Uninstall(context.Background(), "nonexistent")

		assert.NoError(t, err)
	})
}

func TestGoManager_Upgrade(t *testing.T) {
	tests := []struct {
		name        string
		packages    []string
		responses   map[string]CommandResponse
		expectError bool
		errContains string
	}{
		{
			name:     "upgrade single package",
			packages: []string{"github.com/golangci/golangci-lint/cmd/golangci-lint"},
			responses: map[string]CommandResponse{
				"go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest": {Output: []byte("")},
			},
			expectError: false,
		},
		{
			name:     "upgrade strips existing version",
			packages: []string{"github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.0"},
			responses: map[string]CommandResponse{
				"go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest": {Output: []byte("")},
			},
			expectError: false,
		},
		{
			name:        "upgrade all not supported",
			packages:    []string{},
			responses:   map[string]CommandResponse{},
			expectError: true,
			errContains: "does not support upgrading all",
		},
		{
			name:     "upgrade failure",
			packages: []string{"github.com/nonexistent/package"},
			responses: map[string]CommandResponse{
				"go install github.com/nonexistent/package@latest": {
					Output: []byte("module not found"),
					Error:  fmt.Errorf("exit status 1"),
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewGoManager(mock)

			err := mgr.Upgrade(context.Background(), tt.packages)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGoManager_SelfInstall(t *testing.T) {
	t.Run("already installed", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"go version": {Output: []byte("go version go1.21.0")},
			},
		}
		mgr := NewGoManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		assert.Len(t, mock.Commands, 1)
	})

	t.Run("installs via brew when available", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"go version":      {Error: fmt.Errorf("not found")},
				"brew install go": {Output: []byte("Installing go...")},
			},
		}
		mgr := NewGoManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
	})

	t.Run("falls back to error when brew unavailable", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"go version": {Error: fmt.Errorf("not found")},
			},
		}
		mgr := NewGoManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "go.dev")
	})
}

func TestGoManager_ImplementsInterfaces(t *testing.T) {
	mgr := NewGoManager(nil)

	// Verify PackageManager interface
	var _ PackageManager = mgr

	// Verify PackageUpgrader interface
	var _ PackageUpgrader = mgr
}

func TestGoManager_goBinDir(t *testing.T) {
	t.Run("uses GOBIN when set", func(t *testing.T) {
		t.Setenv("GOBIN", "/custom/gobin")
		t.Setenv("GOPATH", "/should/not/use")

		mgr := NewGoManager(nil)
		dir := mgr.goBinDir()

		assert.Equal(t, "/custom/gobin", dir)
	})

	t.Run("uses GOPATH/bin when GOBIN not set", func(t *testing.T) {
		t.Setenv("GOBIN", "")
		t.Setenv("GOPATH", "/custom/gopath")

		mgr := NewGoManager(nil)
		dir := mgr.goBinDir()

		assert.Equal(t, "/custom/gopath/bin", dir)
	})

	t.Run("uses default ~/go/bin when neither set", func(t *testing.T) {
		t.Setenv("GOBIN", "")
		t.Setenv("GOPATH", "")

		mgr := NewGoManager(nil)
		dir := mgr.goBinDir()

		home, _ := os.UserHomeDir()
		assert.Equal(t, filepath.Join(home, "go", "bin"), dir)
	})
}
