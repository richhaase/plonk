// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNPMManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]CommandResponse
		expected  bool
	}{
		{
			name: "npm available",
			responses: map[string]CommandResponse{
				"npm --version": {Output: []byte("10.2.0")},
			},
			expected: true,
		},
		{
			name: "npm command fails",
			responses: map[string]CommandResponse{
				"npm --version": {Error: fmt.Errorf("command failed")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewNPMManager(mock)

			result, err := mgr.IsAvailable(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNPMManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []string
	}{
		{
			name: "multiple packages",
			output: `{
				"dependencies": {
					"typescript": {"version": "5.0.0"},
					"prettier": {"version": "3.0.0"}
				}
			}`,
			expected: []string{"typescript", "prettier"},
		},
		{
			name:     "empty dependencies",
			output:   `{"dependencies": {}}`,
			expected: []string{},
		},
		{
			name: "scoped packages",
			output: `{
				"dependencies": {
					"@types/node": {"version": "20.0.0"},
					"typescript": {"version": "5.0.0"}
				}
			}`,
			expected: []string{"@types/node", "typescript"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"npm list -g --depth=0 --json": {Output: []byte(tt.output)},
				},
			}
			mgr := NewNPMManager(mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestNPMManager_Install(t *testing.T) {
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
			cmdKey:      "npm install -g typescript",
			output:      "added 1 package",
			expectError: false,
		},
		{
			name:        "already installed (idempotent)",
			pkg:         "typescript",
			cmdKey:      "npm install -g typescript",
			output:      "package already installed",
			err:         fmt.Errorf("exit status 1"),
			expectError: false,
		},
		{
			name:        "install failure",
			pkg:         "nonexistent",
			cmdKey:      "npm install -g nonexistent",
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
			mgr := NewNPMManager(mock)

			err := mgr.Install(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNPMManager_Uninstall(t *testing.T) {
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
			cmdKey:      "npm uninstall -g typescript",
			output:      "removed 1 package",
			expectError: false,
		},
		{
			name:        "not installed (idempotent)",
			pkg:         "typescript",
			cmdKey:      "npm uninstall -g typescript",
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
			mgr := NewNPMManager(mock)

			err := mgr.Uninstall(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNPMManager_Upgrade(t *testing.T) {
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
				"npm update -g typescript": {Output: []byte("updated")},
			},
			expectError: false,
		},
		{
			name:     "upgrade all packages",
			packages: []string{},
			responses: map[string]CommandResponse{
				"npm update -g": {Output: []byte("updated all")},
			},
			expectError: false,
		},
		{
			name:     "upgrade multiple packages",
			packages: []string{"typescript", "prettier"},
			responses: map[string]CommandResponse{
				"npm update -g typescript": {Output: []byte("updated")},
				"npm update -g prettier":   {Output: []byte("updated")},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewNPMManager(mock)

			err := mgr.Upgrade(context.Background(), tt.packages)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNPMManager_SelfInstall(t *testing.T) {
	t.Run("already installed", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"npm --version": {Output: []byte("10.0.0")},
			},
		}
		mgr := NewNPMManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		assert.Len(t, mock.Commands, 1)
	})

	t.Run("installs via brew", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"npm --version":       {Error: fmt.Errorf("not found")},
				"brew install node": {Output: []byte("installed")},
			},
		}
		mgr := NewNPMManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
	})

	t.Run("returns error when brew not available", func(t *testing.T) {
		// Don't include any brew responses, so LookPath will fail
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"npm --version": {Error: fmt.Errorf("not found")},
			},
		}
		mgr := NewNPMManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nodejs.org")
	})
}

func TestNPMManager_ImplementsInterfaces(t *testing.T) {
	mgr := NewNPMManager(nil)

	// Verify PackageManager interface
	var _ PackageManager = mgr

	// Verify PackageUpgrader interface
	var _ PackageUpgrader = mgr
}
