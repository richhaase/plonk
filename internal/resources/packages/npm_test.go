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
		provider  NPMProvider
		responses map[string]CommandResponse
		expected  bool
	}{
		{
			name:     "npm available",
			provider: ProviderNPM,
			responses: map[string]CommandResponse{
				"npm --version": {Output: []byte("10.2.0")},
			},
			expected: true,
		},
		{
			name:     "pnpm available",
			provider: ProviderPNPM,
			responses: map[string]CommandResponse{
				"pnpm --version": {Output: []byte("8.10.0")},
			},
			expected: true,
		},
		{
			name:     "bun available",
			provider: ProviderBun,
			responses: map[string]CommandResponse{
				"bun --version": {Output: []byte("1.0.0")},
			},
			expected: true,
		},
		{
			name:     "npm command fails",
			provider: ProviderNPM,
			responses: map[string]CommandResponse{
				"npm --version": {Error: fmt.Errorf("command failed")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewNPMManager(tt.provider, mock)

			result, err := mgr.IsAvailable(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNPMManager_ListInstalled_NPM(t *testing.T) {
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
			mgr := NewNPMManager(ProviderNPM, mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestNPMManager_ListInstalled_PNPM(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"pnpm list -g --depth=0 --json": {Output: []byte(tt.output)},
				},
			}
			mgr := NewNPMManager(ProviderPNPM, mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestNPMManager_ListInstalled_Bun(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"bun pm ls -g": {Output: []byte(tt.output)},
				},
			}
			mgr := NewNPMManager(ProviderBun, mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestNPMManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		provider    NPMProvider
		pkg         string
		cmdKey      string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "npm install success",
			provider:    ProviderNPM,
			pkg:         "typescript",
			cmdKey:      "npm install -g typescript",
			output:      "added 1 package",
			expectError: false,
		},
		{
			name:        "pnpm install success",
			provider:    ProviderPNPM,
			pkg:         "typescript",
			cmdKey:      "pnpm add -g typescript",
			output:      "added 1 package",
			expectError: false,
		},
		{
			name:        "bun install success",
			provider:    ProviderBun,
			pkg:         "typescript",
			cmdKey:      "bun add -g typescript",
			output:      "installed typescript",
			expectError: false,
		},
		{
			name:        "already installed (idempotent)",
			provider:    ProviderNPM,
			pkg:         "typescript",
			cmdKey:      "npm install -g typescript",
			output:      "package already installed",
			err:         fmt.Errorf("exit status 1"),
			expectError: false,
		},
		{
			name:        "install failure",
			provider:    ProviderNPM,
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
			mgr := NewNPMManager(tt.provider, mock)

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
		provider    NPMProvider
		pkg         string
		cmdKey      string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "npm uninstall success",
			provider:    ProviderNPM,
			pkg:         "typescript",
			cmdKey:      "npm uninstall -g typescript",
			output:      "removed 1 package",
			expectError: false,
		},
		{
			name:        "pnpm uninstall success",
			provider:    ProviderPNPM,
			pkg:         "typescript",
			cmdKey:      "pnpm remove -g typescript",
			output:      "removed 1 package",
			expectError: false,
		},
		{
			name:        "bun uninstall success",
			provider:    ProviderBun,
			pkg:         "typescript",
			cmdKey:      "bun remove -g typescript",
			output:      "removed typescript",
			expectError: false,
		},
		{
			name:        "not installed (idempotent)",
			provider:    ProviderNPM,
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
			mgr := NewNPMManager(tt.provider, mock)

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
		provider    NPMProvider
		packages    []string
		responses   map[string]CommandResponse
		expectError bool
		errContains string
	}{
		{
			name:     "npm upgrade single",
			provider: ProviderNPM,
			packages: []string{"typescript"},
			responses: map[string]CommandResponse{
				"npm update -g typescript": {Output: []byte("updated")},
			},
			expectError: false,
		},
		{
			name:     "pnpm upgrade single",
			provider: ProviderPNPM,
			packages: []string{"typescript"},
			responses: map[string]CommandResponse{
				"pnpm update -g typescript": {Output: []byte("updated")},
			},
			expectError: false,
		},
		{
			name:     "bun upgrade uses add @latest",
			provider: ProviderBun,
			packages: []string{"typescript"},
			responses: map[string]CommandResponse{
				"bun add -g typescript@latest": {Output: []byte("updated")},
			},
			expectError: false,
		},
		{
			name:     "npm upgrade all",
			provider: ProviderNPM,
			packages: []string{},
			responses: map[string]CommandResponse{
				"npm update -g": {Output: []byte("updated all")},
			},
			expectError: false,
		},
		{
			name:        "bun upgrade all not supported",
			provider:    ProviderBun,
			packages:    []string{},
			responses:   map[string]CommandResponse{},
			expectError: true,
			errContains: "does not support upgrading all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewNPMManager(tt.provider, mock)

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

func TestNPMManager_SelfInstall(t *testing.T) {
	t.Run("npm already installed", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"npm --version": {Output: []byte("10.0.0")},
			},
		}
		mgr := NewNPMManager(ProviderNPM, mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		assert.Len(t, mock.Commands, 1)
	})

	t.Run("pnpm installs via brew", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"pnpm --version":      {Error: fmt.Errorf("not found")},
				"brew install pnpm": {Output: []byte("installed")},
			},
		}
		mgr := NewNPMManager(ProviderPNPM, mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
	})

	t.Run("bun installs via brew", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"bun --version":      {Error: fmt.Errorf("not found")},
				"brew install bun": {Output: []byte("installed")},
			},
		}
		mgr := NewNPMManager(ProviderBun, mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
	})
}

func TestNPMManager_ImplementsInterfaces(t *testing.T) {
	mgr := NewNPMManager(ProviderNPM, nil)

	// Verify PackageManager interface
	var _ PackageManager = mgr

	// Verify PackageUpgrader interface
	var _ PackageUpgrader = mgr
}

func TestNPMManager_Provider(t *testing.T) {
	tests := []struct {
		provider NPMProvider
	}{
		{ProviderNPM},
		{ProviderPNPM},
		{ProviderBun},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			mgr := NewNPMManager(tt.provider, nil)
			assert.Equal(t, tt.provider, mgr.Provider())
		})
	}
}

func TestNPMManager_DefaultProvider(t *testing.T) {
	mgr := NewNPMManager("", nil)
	assert.Equal(t, ProviderNPM, mgr.Provider())
}

func TestParseNPMJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []string
		expectError bool
	}{
		{
			name:     "valid json",
			input:    `{"dependencies": {"pkg1": {}, "pkg2": {}}}`,
			expected: []string{"pkg1", "pkg2"},
		},
		{
			name:     "empty dependencies",
			input:    `{"dependencies": {}}`,
			expected: []string{},
		},
		{
			name:        "invalid json",
			input:       `not json`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseNPMJSON([]byte(tt.input))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestParsePNPMJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []string
		expectError bool
	}{
		{
			name:     "valid json array",
			input:    `[{"dependencies": {"pkg1": {}, "pkg2": {}}}]`,
			expected: []string{"pkg1", "pkg2"},
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: []string{},
		},
		{
			name:        "invalid json",
			input:       `not json`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePNPMJSON([]byte(tt.input))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestParseBunList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple list",
			input:    "pkg1 v1.0.0\npkg2 v2.0.0",
			expected: []string{"pkg1", "pkg2"},
		},
		{
			name:     "empty",
			input:    "",
			expected: []string{},
		},
		{
			name:     "skips paths and headers",
			input:    "/path/to/global\ndependencies:\npkg1 v1.0.0",
			expected: []string{"pkg1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBunList([]byte(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}
