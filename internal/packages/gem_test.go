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

func TestGemManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]CommandResponse
		expected  bool
	}{
		{
			name: "gem available",
			responses: map[string]CommandResponse{
				"gem --version": {Output: []byte("3.4.22")},
			},
			expected: true,
		},
		{
			name: "gem command fails",
			responses: map[string]CommandResponse{
				"gem --version": {Error: fmt.Errorf("command failed")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewGemManager(mock)

			result, err := mgr.IsAvailable(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGemManager_IsAvailable_NotInPath(t *testing.T) {
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{},
	}
	mgr := NewGemManager(mock)

	result, err := mgr.IsAvailable(context.Background())

	assert.NoError(t, err)
	assert.False(t, result)
}

func TestGemManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []string
	}{
		{
			name: "multiple packages",
			output: `colorize
rake
bundler
rspec
`,
			expected: []string{"colorize", "rake", "bundler", "rspec"},
		},
		{
			name:     "empty list",
			output:   "",
			expected: []string{},
		},
		{
			name:   "single package",
			output: "colorize\n",
			expected: []string{"colorize"},
		},
		{
			name: "packages with extra whitespace",
			output: `  colorize
  rake
bundler
`,
			expected: []string{"colorize", "rake", "bundler"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"gem list --no-versions": {Output: []byte(tt.output)},
				},
			}
			mgr := NewGemManager(mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGemManager_ListInstalled_Error(t *testing.T) {
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"gem list --no-versions": {Error: fmt.Errorf("gem not available")},
		},
	}
	mgr := NewGemManager(mock)

	_, err := mgr.ListInstalled(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list packages")
}

func TestGemManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "successful install",
			pkg:         "colorize",
			output:      "Successfully installed colorize-0.8.1\n",
			err:         nil,
			expectError: false,
		},
		{
			name:        "already installed (idempotent)",
			pkg:         "colorize",
			output:      "Successfully installed colorize-0.8.1\n1 gem already installed",
			err:         nil,
			expectError: false,
		},
		{
			name:        "install failure",
			pkg:         "nonexistent-gem",
			output:      "ERROR: Could not find a valid gem 'nonexistent-gem'",
			err:         fmt.Errorf("exit status 1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"gem install --user-install " + tt.pkg: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewGemManager(mock)

			err := mgr.Install(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGemManager_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		output      string
		err         error
		expectError bool
	}{
		{
			name:        "successful uninstall",
			pkg:         "colorize",
			output:      "Successfully uninstalled colorize-0.8.1\n",
			err:         nil,
			expectError: false,
		},
		{
			name:        "not installed (idempotent)",
			pkg:         "colorize",
			output:      "ERROR: gem 'colorize' is not installed",
			err:         fmt.Errorf("exit status 1"),
			expectError: false,
		},
		{
			name:        "uninstall failure",
			pkg:         "protected",
			output:      "ERROR: cannot uninstall protected gem",
			err:         fmt.Errorf("exit status 1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"gem uninstall " + tt.pkg + " -x -a": {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewGemManager(mock)

			err := mgr.Uninstall(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGemManager_Upgrade(t *testing.T) {
	tests := []struct {
		name        string
		packages    []string
		responses   map[string]CommandResponse
		expectError bool
	}{
		{
			name:     "upgrade single package",
			packages: []string{"colorize"},
			responses: map[string]CommandResponse{
				"gem update colorize": {Output: []byte("Updating colorize...")},
			},
			expectError: false,
		},
		{
			name:     "upgrade multiple packages",
			packages: []string{"colorize", "rake"},
			responses: map[string]CommandResponse{
				"gem update colorize": {Output: []byte("Updating colorize...")},
				"gem update rake":     {Output: []byte("Updating rake...")},
			},
			expectError: false,
		},
		{
			name:     "upgrade all packages",
			packages: []string{},
			responses: map[string]CommandResponse{
				"gem update": {Output: []byte("Updating installed gems...")},
			},
			expectError: false,
		},
		{
			name:     "already up-to-date (idempotent)",
			packages: []string{"colorize"},
			responses: map[string]CommandResponse{
				"gem update colorize": {
					Output: []byte("Nothing to update"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade failure",
			packages: []string{"nonexistent"},
			responses: map[string]CommandResponse{
				"gem update nonexistent": {
					Output: []byte("ERROR: While executing gem ..."),
					Error:  fmt.Errorf("exit status 1"),
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewGemManager(mock)

			err := mgr.Upgrade(context.Background(), tt.packages)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGemManager_SelfInstall(t *testing.T) {
	t.Run("already installed", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"gem --version": {Output: []byte("3.4.22")},
			},
		}
		mgr := NewGemManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		// Should not have tried to install since already available
		assert.Len(t, mock.Commands, 1)
		assert.Equal(t, "gem", mock.Commands[0].Name)
	})

	t.Run("installs via brew when available", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				// gem --version fails (not installed)
				"gem --version": {Error: fmt.Errorf("command not found")},
				// brew is available
				"brew --version":      {Output: []byte("Homebrew 4.2.0")},
				"brew install ruby":   {Output: []byte("Installing ruby...")},
			},
		}
		mgr := NewGemManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		// Should have called gem --version, brew --version, brew install ruby
		require.GreaterOrEqual(t, len(mock.Commands), 3)
	})

	t.Run("returns error when no install method available", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				// gem --version fails (not installed)
				"gem --version": {Error: fmt.Errorf("command not found")},
				// brew not available
				"brew --version": {Error: fmt.Errorf("command not found")},
			},
		}
		mgr := NewGemManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not installed")
	})
}

func TestGemManager_ImplementsInterfaces(t *testing.T) {
	mgr := NewGemManager(nil)

	// Verify PackageManager interface
	var _ PackageManager = mgr

	// Verify PackageUpgrader interface
	var _ PackageUpgrader = mgr
}

func TestParseGemList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "standard output",
			input: `colorize
rake
bundler`,
			expected: []string{"colorize", "rake", "bundler"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single package",
			input:    "colorize",
			expected: []string{"colorize"},
		},
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			expected: []string{},
		},
		{
			name: "trailing newline",
			input: `colorize
rake
`,
			expected: []string{"colorize", "rake"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOutput([]byte(tt.input), ParseConfig{})
			assert.Equal(t, tt.expected, result)
		})
	}
}
