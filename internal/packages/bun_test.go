// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBunManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]CommandResponse
		expected  bool
	}{
		{
			name: "bun available",
			responses: map[string]CommandResponse{
				"bun --version": {Output: []byte("1.0.0")},
			},
			expected: true,
		},
		{
			name: "bun command fails",
			responses: map[string]CommandResponse{
				"bun --version": {Error: fmt.Errorf("command failed")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewBunManager(mock)

			result, err := mgr.IsAvailable(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBunManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []string
	}{
		{
			name:     "multiple packages",
			output:   "typescript v5.0.0\nprettier v3.0.0\n",
			expected: []string{"typescript", "prettier"},
		},
		{
			name:     "empty list",
			output:   "",
			expected: []string{},
		},
		{
			name:     "skips paths and headers",
			output:   "/path/to/global\ndependencies:\ntypescript v5.0.0",
			expected: []string{"typescript"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"bun pm ls -g": {Output: []byte(tt.output)},
				},
			}
			mgr := NewBunManager(mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestBunManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		cmdKey      string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "install success",
			pkg:         "typescript",
			cmdKey:      "bun add -g typescript",
			output:      "installed typescript",
			expectError: false,
		},
		{
			name:        "already installed (idempotent)",
			pkg:         "typescript",
			cmdKey:      "bun add -g typescript",
			output:      "package already installed",
			err:         fmt.Errorf("exit status 1"),
			expectError: false,
		},
		{
			name:        "install failure",
			pkg:         "nonexistent",
			cmdKey:      "bun add -g nonexistent",
			output:      "404 Not Found",
			err:         fmt.Errorf("exit status 1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					tt.cmdKey: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewBunManager(mock)

			err := mgr.Install(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBunManager_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		cmdKey      string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "uninstall success",
			pkg:         "typescript",
			cmdKey:      "bun remove -g typescript",
			output:      "removed typescript",
			expectError: false,
		},
		{
			name:        "not installed (idempotent)",
			pkg:         "typescript",
			cmdKey:      "bun remove -g typescript",
			output:      "package not installed",
			err:         fmt.Errorf("exit status 1"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					tt.cmdKey: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewBunManager(mock)

			err := mgr.Uninstall(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBunManager_Upgrade(t *testing.T) {
	tests := []struct {
		name        string
		packages    []string
		responses   map[string]CommandResponse
		expectError bool
		errContains string
	}{
		{
			name:     "upgrade single package uses add @latest",
			packages: []string{"typescript"},
			responses: map[string]CommandResponse{
				"bun add -g typescript@latest": {Output: []byte("updated")},
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
			name:     "upgrade multiple packages",
			packages: []string{"typescript", "prettier"},
			responses: map[string]CommandResponse{
				"bun add -g typescript@latest": {Output: []byte("updated")},
				"bun add -g prettier@latest":   {Output: []byte("updated")},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewBunManager(mock)

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

func TestBunManager_SelfInstall(t *testing.T) {
	t.Run("already installed", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"bun --version": {Output: []byte("1.0.0")},
			},
		}
		mgr := NewBunManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		assert.Len(t, mock.Commands, 1)
	})

	t.Run("installs via brew", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"bun --version":    {Error: fmt.Errorf("not found")},
				"brew install bun": {Output: []byte("installed")},
			},
		}
		mgr := NewBunManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
	})

	t.Run("installs via curl when brew not available", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"bun --version": {Error: fmt.Errorf("not found")},
				"sh -c curl -fsSL https://bun.sh/install | bash": {Output: []byte("installed")},
			},
		}
		mgr := NewBunManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
	})
}

func TestBunManager_ImplementsInterfaces(t *testing.T) {
	mgr := NewBunManager(nil)

	// Verify PackageManager interface
	var _ PackageManager = mgr

	// Verify PackageUpgrader interface
	var _ PackageUpgrader = mgr
}
