// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCargoManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]CommandResponse
		expected  bool
	}{
		{
			name: "cargo available",
			responses: map[string]CommandResponse{
				"cargo --version": {Output: []byte("cargo 1.75.0 (1d8b05cdd 2023-11-20)")},
			},
			expected: true,
		},
		{
			name: "cargo command fails",
			responses: map[string]CommandResponse{
				"cargo --version": {Error: fmt.Errorf("command failed")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewCargoManager(mock)

			result, err := mgr.IsAvailable(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCargoManager_IsAvailable_NotInPath(t *testing.T) {
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{},
	}
	mgr := NewCargoManager(mock)

	result, err := mgr.IsAvailable(context.Background())

	assert.NoError(t, err)
	assert.False(t, result)
}

func TestCargoManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []string
	}{
		{
			name: "multiple packages with binaries",
			output: `ripgrep v14.1.0:
    rg
fd-find v9.0.0:
    fd
bat v0.24.0:
    bat
`,
			expected: []string{"ripgrep", "fd-find", "bat"},
		},
		{
			name:     "empty list",
			output:   "",
			expected: []string{},
		},
		{
			name: "single package",
			output: `ripgrep v14.1.0:
    rg
`,
			expected: []string{"ripgrep"},
		},
		{
			name: "package with multiple binaries",
			output: `cargo-edit v0.12.2:
    cargo-add
    cargo-rm
    cargo-upgrade
`,
			expected: []string{"cargo-edit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"cargo install --list": {Output: []byte(tt.output)},
				},
			}
			mgr := NewCargoManager(mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCargoManager_ListInstalled_Error(t *testing.T) {
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"cargo install --list": {Error: fmt.Errorf("cargo not available")},
		},
	}
	mgr := NewCargoManager(mock)

	_, err := mgr.ListInstalled(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list packages")
}

func TestCargoManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "successful install",
			pkg:         "ripgrep",
			output:      "Installing ripgrep v14.1.0\n",
			err:         nil,
			expectError: false,
		},
		{
			name:        "already installed - already exists",
			pkg:         "ripgrep",
			output:      "error: binary `rg` already exists in destination",
			err:         fmt.Errorf("exit status 101"),
			expectError: false,
		},
		{
			name:        "already installed - already installed message",
			pkg:         "ripgrep",
			output:      "warning: package `ripgrep` is already installed",
			err:         fmt.Errorf("exit status 0"),
			expectError: false,
		},
		{
			name:        "install failure",
			pkg:         "nonexistent-crate",
			output:      "error: could not find `nonexistent-crate` in registry",
			err:         fmt.Errorf("exit status 101"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"cargo install " + tt.pkg: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewCargoManager(mock)

			err := mgr.Install(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCargoManager_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "successful uninstall",
			pkg:         "ripgrep",
			output:      "Removing ripgrep v14.1.0\n",
			err:         nil,
			expectError: false,
		},
		{
			name:        "not installed (idempotent)",
			pkg:         "ripgrep",
			output:      "error: package `ripgrep` is not installed",
			err:         fmt.Errorf("exit status 101"),
			expectError: false,
		},
		{
			name:        "uninstall failure",
			pkg:         "protected",
			output:      "error: cannot uninstall protected package",
			err:         fmt.Errorf("exit status 101"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"cargo uninstall " + tt.pkg: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewCargoManager(mock)

			err := mgr.Uninstall(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCargoManager_Upgrade(t *testing.T) {
	tests := []struct {
		name        string
		packages    []string
		responses   map[string]CommandResponse
		expectError bool
		errContains string
	}{
		{
			name:     "upgrade single package",
			packages: []string{"ripgrep"},
			responses: map[string]CommandResponse{
				"cargo install --force ripgrep": {Output: []byte("Compiling ripgrep...")},
			},
			expectError: false,
		},
		{
			name:     "upgrade multiple packages",
			packages: []string{"ripgrep", "fd-find"},
			responses: map[string]CommandResponse{
				"cargo install --force ripgrep":  {Output: []byte("Compiling ripgrep...")},
				"cargo install --force fd-find": {Output: []byte("Compiling fd-find...")},
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
			name:     "already up-to-date (idempotent)",
			packages: []string{"ripgrep"},
			responses: map[string]CommandResponse{
				"cargo install --force ripgrep": {
					Output: []byte("warning: package already up-to-date"),
					Error:  fmt.Errorf("exit status 0"),
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade failure",
			packages: []string{"nonexistent"},
			responses: map[string]CommandResponse{
				"cargo install --force nonexistent": {
					Output: []byte("error: could not find crate"),
					Error:  fmt.Errorf("exit status 101"),
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewCargoManager(mock)

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

func TestCargoManager_SelfInstall(t *testing.T) {
	t.Run("already installed", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"cargo --version": {Output: []byte("cargo 1.75.0")},
			},
		}
		mgr := NewCargoManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		// Should not have tried to install since already available
		assert.Len(t, mock.Commands, 1)
		assert.Equal(t, "cargo", mock.Commands[0].Name)
	})

	t.Run("installs when not available", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				// cargo --version fails (not installed)
				"cargo --version": {Error: fmt.Errorf("command not found")},
			},
			DefaultResponse: CommandResponse{Output: []byte("Installation successful")},
		}
		mgr := NewCargoManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		// Should have called cargo --version, then sh -c with install script
		require.GreaterOrEqual(t, len(mock.Commands), 2)
		assert.Equal(t, "sh", mock.Commands[1].Name)
	})
}

func TestCargoManager_ImplementsInterfaces(t *testing.T) {
	mgr := NewCargoManager(nil)

	// Verify PackageManager interface
	var _ PackageManager = mgr

	// Verify PackageUpgrader interface
	var _ PackageUpgrader = mgr
}

func TestParseCargoList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "standard output",
			input: `ripgrep v14.1.0:
    rg
fd-find v9.0.0:
    fd`,
			expected: []string{"ripgrep", "fd-find"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name: "package with multiple binaries",
			input: `cargo-edit v0.12.2:
    cargo-add
    cargo-rm
    cargo-upgrade`,
			expected: []string{"cargo-edit"},
		},
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOutput([]byte(tt.input), ParseConfig{SkipIndented: true, TakeFirstToken: true})
			assert.Equal(t, tt.expected, result)
		})
	}
}
