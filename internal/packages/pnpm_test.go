// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPNPMManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]CommandResponse
		expected  bool
	}{
		{
			name: "pnpm available",
			responses: map[string]CommandResponse{
				"pnpm --version": {Output: []byte("8.10.0")},
			},
			expected: true,
		},
		{
			name: "pnpm command fails",
			responses: map[string]CommandResponse{
				"pnpm --version": {Error: fmt.Errorf("command failed")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewPNPMManager(mock)

			result, err := mgr.IsAvailable(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPNPMManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []string
	}{
		{
			name: "multiple packages",
			output: `[{
				"dependencies": {
					"typescript": {"version": "5.0.0"},
					"prettier": {"version": "3.0.0"}
				}
			}]`,
			expected: []string{"typescript", "prettier"},
		},
		{
			name:     "empty array",
			output:   `[]`,
			expected: []string{},
		},
		{
			name: "scoped packages",
			output: `[{
				"dependencies": {
					"@types/node": {"version": "20.0.0"},
					"typescript": {"version": "5.0.0"}
				}
			}]`,
			expected: []string{"@types/node", "typescript"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"pnpm list -g --depth=0 --json": {Output: []byte(tt.output)},
				},
			}
			mgr := NewPNPMManager(mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestPNPMManager_Install(t *testing.T) {
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
			cmdKey:      "pnpm add -g typescript",
			output:      "added 1 package",
			expectError: false,
		},
		{
			name:        "already installed (idempotent)",
			pkg:         "typescript",
			cmdKey:      "pnpm add -g typescript",
			output:      "package already installed",
			err:         fmt.Errorf("exit status 1"),
			expectError: false,
		},
		{
			name:        "install failure",
			pkg:         "nonexistent",
			cmdKey:      "pnpm add -g nonexistent",
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
			mgr := NewPNPMManager(mock)

			err := mgr.Install(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPNPMManager_Uninstall(t *testing.T) {
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
			cmdKey:      "pnpm remove -g typescript",
			output:      "removed 1 package",
			expectError: false,
		},
		{
			name:        "not installed (idempotent)",
			pkg:         "typescript",
			cmdKey:      "pnpm remove -g typescript",
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
			mgr := NewPNPMManager(mock)

			err := mgr.Uninstall(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPNPMManager_Upgrade(t *testing.T) {
	tests := []struct {
		name        string
		packages    []string
		responses   map[string]CommandResponse
		expectError bool
	}{
		{
			name:     "upgrade single package",
			packages: []string{"typescript"},
			responses: map[string]CommandResponse{
				"pnpm update -g typescript": {Output: []byte("updated")},
			},
			expectError: false,
		},
		{
			name:     "upgrade all packages",
			packages: []string{},
			responses: map[string]CommandResponse{
				"pnpm update -g": {Output: []byte("updated all")},
			},
			expectError: false,
		},
		{
			name:     "upgrade multiple packages",
			packages: []string{"typescript", "prettier"},
			responses: map[string]CommandResponse{
				"pnpm update -g typescript": {Output: []byte("updated")},
				"pnpm update -g prettier":   {Output: []byte("updated")},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewPNPMManager(mock)

			err := mgr.Upgrade(context.Background(), tt.packages)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPNPMManager_SelfInstall(t *testing.T) {
	t.Run("already installed", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"pnpm --version": {Output: []byte("8.0.0")},
			},
		}
		mgr := NewPNPMManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		assert.Len(t, mock.Commands, 1)
		assert.Equal(t, "pnpm", mock.Commands[0].Name)
	})

	t.Run("installs via brew when available", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				// pnpm --version fails (not installed)
				"pnpm --version":     {Error: fmt.Errorf("command not found")},
				// brew install pnpm succeeds (LookPath checks for "brew" prefix)
				"brew install pnpm": {Output: []byte("Installing pnpm...")},
			},
		}
		mgr := NewPNPMManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		// Should have called pnpm --version, brew install pnpm
		assert.Len(t, mock.Commands, 2)
	})

	t.Run("installs via curl when brew not available", func(t *testing.T) {
		// Don't include any brew responses so LookPath fails, fall back to curl
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				// pnpm --version fails (not installed)
				"pnpm --version": {Error: fmt.Errorf("command not found")},
				// curl fallback succeeds
				"sh -c curl -fsSL https://get.pnpm.io/install.sh | sh -": {Output: []byte("Installing pnpm...")},
			},
		}
		mgr := NewPNPMManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
	})
}

func TestPNPMManager_ImplementsInterfaces(t *testing.T) {
	mgr := NewPNPMManager(nil)

	// Verify PackageManager interface
	var _ PackageManager = mgr

	// Verify PackageUpgrader interface
	var _ PackageUpgrader = mgr
}
