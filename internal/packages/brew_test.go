// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrewManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]CommandResponse
		expected  bool
	}{
		{
			name: "brew available",
			responses: map[string]CommandResponse{
				"brew --version": {Output: []byte("Homebrew 4.2.0")},
			},
			expected: true,
		},
		{
			name: "brew command fails",
			responses: map[string]CommandResponse{
				"brew --version": {Error: fmt.Errorf("command failed")},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewBrewManager(mock)

			result, err := mgr.IsAvailable(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBrewManager_IsAvailable_NotInPath(t *testing.T) {
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{},
	}
	mgr := NewBrewManager(mock)

	result, err := mgr.IsAvailable(context.Background())

	assert.NoError(t, err)
	assert.False(t, result)
}

func TestBrewManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []string
	}{
		{
			name:     "multiple packages",
			output:   "git\nwget\ncurl\njq\n",
			expected: []string{"git", "wget", "curl", "jq"},
		},
		{
			name:     "empty list",
			output:   "",
			expected: []string{},
		},
		{
			name:     "packages with extra whitespace",
			output:   "  git  \n  wget\n\ncurl\n",
			expected: []string{"git", "wget", "curl"},
		},
		{
			name:     "single package",
			output:   "ripgrep\n",
			expected: []string{"ripgrep"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"brew list": {Output: []byte(tt.output)},
				},
			}
			mgr := NewBrewManager(mock)

			result, err := mgr.ListInstalled(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBrewManager_ListInstalled_Error(t *testing.T) {
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"brew list": {Error: fmt.Errorf("brew not available")},
		},
	}
	mgr := NewBrewManager(mock)

	_, err := mgr.ListInstalled(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list packages")
}

func TestBrewManager_Install(t *testing.T) {
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
			output:      "==> Installing ripgrep\n",
			err:         nil,
			expectError: false,
		},
		{
			name:        "already installed (idempotent)",
			pkg:         "ripgrep",
			output:      "Warning: ripgrep 14.0.0 is already installed",
			err:         &exec.ExitError{},
			expectError: false,
		},
		{
			name:        "install failure",
			pkg:         "nonexistent",
			output:      "Error: No formula or cask found for nonexistent",
			err:         fmt.Errorf("exit status 1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"brew install " + tt.pkg: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewBrewManager(mock)

			err := mgr.Install(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBrewManager_Uninstall(t *testing.T) {
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
			output:      "Uninstalling /opt/homebrew/Cellar/ripgrep/14.0.0...",
			err:         nil,
			expectError: false,
		},
		{
			name:        "not installed (idempotent)",
			pkg:         "ripgrep",
			output:      "Error: No such keg: /opt/homebrew/Cellar/ripgrep",
			err:         fmt.Errorf("exit status 1"),
			expectError: false,
		},
		{
			name:        "uninstall failure",
			pkg:         "protected",
			output:      "Error: Refusing to uninstall protected",
			err:         fmt.Errorf("exit status 1"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				Responses: map[string]CommandResponse{
					"brew uninstall " + tt.pkg: {Output: []byte(tt.output), Error: tt.err},
				},
			}
			mgr := NewBrewManager(mock)

			err := mgr.Uninstall(context.Background(), tt.pkg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBrewManager_Upgrade(t *testing.T) {
	tests := []struct {
		name        string
		packages    []string
		responses   map[string]CommandResponse
		expectError bool
	}{
		{
			name:     "upgrade single package",
			packages: []string{"ripgrep"},
			responses: map[string]CommandResponse{
				"brew upgrade ripgrep": {Output: []byte("Upgrading ripgrep...")},
			},
			expectError: false,
		},
		{
			name:     "upgrade multiple packages",
			packages: []string{"ripgrep", "fd"},
			responses: map[string]CommandResponse{
				"brew upgrade ripgrep": {Output: []byte("Upgrading ripgrep...")},
				"brew upgrade fd":      {Output: []byte("Upgrading fd...")},
			},
			expectError: false,
		},
		{
			name:     "upgrade all packages",
			packages: []string{},
			responses: map[string]CommandResponse{
				"brew upgrade": {Output: []byte("Upgrading all packages...")},
			},
			expectError: false,
		},
		{
			name:     "already up-to-date (idempotent)",
			packages: []string{"ripgrep"},
			responses: map[string]CommandResponse{
				"brew upgrade ripgrep": {
					Output: []byte("Warning: ripgrep 14.0.0 already installed"),
					Error:  fmt.Errorf("exit status 1"),
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade failure",
			packages: []string{"nonexistent"},
			responses: map[string]CommandResponse{
				"brew upgrade nonexistent": {
					Output: []byte("Error: nonexistent not installed"),
					Error:  fmt.Errorf("exit status 1"),
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{Responses: tt.responses}
			mgr := NewBrewManager(mock)

			err := mgr.Upgrade(context.Background(), tt.packages)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBrewManager_SelfInstall(t *testing.T) {
	t.Run("already installed", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				"brew --version": {Output: []byte("Homebrew 4.2.0")},
			},
		}
		mgr := NewBrewManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		// Should not have tried to install since already available
		assert.Len(t, mock.Commands, 1)
		assert.Equal(t, "brew", mock.Commands[0].Name)
	})

	t.Run("installs when not available", func(t *testing.T) {
		mock := &MockCommandExecutor{
			Responses: map[string]CommandResponse{
				// brew --version fails (not installed)
				"brew --version": {Error: fmt.Errorf("command not found")},
				// install script succeeds
			},
			DefaultResponse: CommandResponse{Output: []byte("Installation successful")},
		}
		mgr := NewBrewManager(mock)

		err := mgr.SelfInstall(context.Background())

		assert.NoError(t, err)
		// Should have called brew --version, then sh -c with install script
		require.GreaterOrEqual(t, len(mock.Commands), 2)
		assert.Equal(t, "sh", mock.Commands[1].Name)
	})
}

func TestBrewManager_ImplementsInterfaces(t *testing.T) {
	mgr := NewBrewManager(nil)

	// Verify PackageManager interface
	var _ PackageManager = mgr

	// Verify PackageUpgrader interface
	var _ PackageUpgrader = mgr
}

func TestParseLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple lines",
			input:    "one\ntwo\nthree",
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			expected: []string{},
		},
		{
			name:     "lines with versions",
			input:    "ripgrep v14.0.0\nfd v8.7.0",
			expected: []string{"ripgrep", "fd"},
		},
		{
			name:     "trailing newline",
			input:    "one\ntwo\n",
			expected: []string{"one", "two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOutput([]byte(tt.input), ParseConfig{TakeFirstToken: true})
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsIdempotent(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		patterns []string
		expected bool
	}{
		{
			name:     "matches single pattern",
			output:   "Warning: ripgrep is already installed",
			patterns: []string{"already installed"},
			expected: true,
		},
		{
			name:     "matches case insensitive",
			output:   "ALREADY INSTALLED",
			patterns: []string{"already installed"},
			expected: true,
		},
		{
			name:     "matches one of multiple patterns",
			output:   "No such keg found",
			patterns: []string{"already installed", "no such keg"},
			expected: true,
		},
		{
			name:     "no match",
			output:   "Some other error",
			patterns: []string{"already installed"},
			expected: false,
		},
		{
			name:     "empty patterns",
			output:   "already installed",
			patterns: []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIdempotent(tt.output, tt.patterns...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
