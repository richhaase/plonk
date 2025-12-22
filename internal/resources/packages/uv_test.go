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

func TestUVManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]CommandResponse
		expected  bool
	}{
		{
			name: "uv available",
			responses: map[string]CommandResponse{
				"uv --version": {Output: []byte("uv 0.1.0")},
			},
			expected: true,
		},
		{
			name: "uv command fails",
			responses: map[string]CommandResponse{
				"uv --version": {Error: fmt.Errorf("command failed")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewUVManager(mock)

			result, err := mgr.IsAvailable(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUVManager_IsAvailable_NotInPath(t *testing.T) {
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{},
	}
	mgr := NewUVManager(mock)

	result, err := mgr.IsAvailable(context.Background())

	assert.NoError(t, err)
	assert.False(t, result)
}

func TestUVManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []string
	}{
		{
			name:     "multiple tools",
			output:   "ruff v0.1.0\nblack v23.0.0\nmypy v1.5.0\n",
			expected: []string{"ruff", "black", "mypy"},
		},
		{
			name:     "empty list",
			output:   "",
			expected: []string{},
		},
		{
			name:     "tools with extras",
			output:   "ruff v0.1.0 (with extras)\nblack v23.0.0\n",
			expected: []string{"ruff", "black"},
		},
		{
			name:     "single tool",
			output:   "ruff v0.1.0\n",
			expected: []string{"ruff"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"uv tool list": {Output: []byte(tt.output)},
				},
			}
			mgr := NewUVManager(mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUVManager_ListInstalled_Error(t *testing.T) {
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"uv tool list": {Error: fmt.Errorf("uv not available")},
		},
	}
	mgr := NewUVManager(mock)

	_, err := mgr.ListInstalled(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list tools")
}

func TestUVManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "successful install",
			pkg:         "ruff",
			output:      "Installed ruff\n",
			err:         nil,
			expectError: false,
		},
		{
			name:        "already installed (idempotent)",
			pkg:         "ruff",
			output:      "Tool ruff is already installed",
			err:         fmt.Errorf("exit status 1"),
			expectError: false,
		},
		{
			name:        "install failure",
			pkg:         "nonexistent",
			output:      "error: Tool not found",
			err:         fmt.Errorf("exit status 1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"uv tool install " + tt.pkg: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewUVManager(mock)

			err := mgr.Install(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUVManager_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "successful uninstall",
			pkg:         "ruff",
			output:      "Uninstalled ruff\n",
			err:         nil,
			expectError: false,
		},
		{
			name:        "not installed (idempotent)",
			pkg:         "ruff",
			output:      "Tool ruff is not installed",
			err:         fmt.Errorf("exit status 1"),
			expectError: false,
		},
		{
			name:        "uninstall failure",
			pkg:         "protected",
			output:      "error: Cannot uninstall",
			err:         fmt.Errorf("exit status 1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"uv tool uninstall " + tt.pkg: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewUVManager(mock)

			err := mgr.Uninstall(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUVManager_Upgrade(t *testing.T) {
	tests := []struct {
		name        string
		packages    []string
		responses   map[string]CommandResponse
		expectError bool
	}{
		{
			name:     "upgrade single package",
			packages: []string{"ruff"},
			responses: map[string]CommandResponse{
				"uv tool upgrade ruff": {Output: []byte("Upgraded ruff")},
			},
			expectError: false,
		},
		{
			name:     "upgrade multiple packages",
			packages: []string{"ruff", "black"},
			responses: map[string]CommandResponse{
				"uv tool upgrade ruff":  {Output: []byte("Upgraded ruff")},
				"uv tool upgrade black": {Output: []byte("Upgraded black")},
			},
			expectError: false,
		},
		{
			name:     "upgrade all packages",
			packages: []string{},
			responses: map[string]CommandResponse{
				"uv tool upgrade --all": {Output: []byte("Upgraded all tools")},
			},
			expectError: false,
		},
		{
			name:     "already up-to-date (idempotent)",
			packages: []string{"ruff"},
			responses: map[string]CommandResponse{
				"uv tool upgrade ruff": {
					Output: []byte("ruff is already up-to-date"),
					Error:  fmt.Errorf("exit status 0"),
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade failure",
			packages: []string{"nonexistent"},
			responses: map[string]CommandResponse{
				"uv tool upgrade nonexistent": {
					Output: []byte("error: Tool not found"),
					Error:  fmt.Errorf("exit status 1"),
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewUVManager(mock)

			err := mgr.Upgrade(context.Background(), tt.packages)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUVManager_SelfInstall(t *testing.T) {
	t.Run("already installed", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"uv --version": {Output: []byte("uv 0.1.0")},
			},
		}
		mgr := NewUVManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		assert.Len(t, mock.Commands, 1)
	})

	t.Run("installs via brew when available", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"uv --version":      {Error: fmt.Errorf("not found")},
				"brew install uv": {Output: []byte("Installing uv...")},
			},
		}
		mgr := NewUVManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
	})

	t.Run("falls back to official script when brew unavailable", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"uv --version": {Error: fmt.Errorf("not found")},
			},
			DefaultResponse: CommandResponse{Output: []byte("Installation successful")},
		}
		mgr := NewUVManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		// Should have tried uv --version, then sh -c with install script
		require.GreaterOrEqual(t, len(mock.Commands), 2)
		assert.Equal(t, "sh", mock.Commands[1].Name)
	})
}

func TestUVManager_ImplementsInterfaces(t *testing.T) {
	mgr := NewUVManager(nil)

	// Verify PackageManager interface
	var _ PackageManager = mgr

	// Verify PackageUpgrader interface
	var _ PackageUpgrader = mgr
}

func TestParseUVList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "standard output",
			input:    "ruff v0.1.0\nblack v23.0.0",
			expected: []string{"ruff", "black"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "with extras",
			input:    "ruff v0.1.0 (lsp)\nblack v23.0.0",
			expected: []string{"ruff", "black"},
		},
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseUVList([]byte(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}
