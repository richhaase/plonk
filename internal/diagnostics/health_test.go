// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package diagnostics

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHomebrewPath(t *testing.T) {
	// Save original values
	originalOS := runtime.GOOS
	originalArch := runtime.GOARCH

	// Note: We can't actually change runtime.GOOS/GOARCH in tests,
	// so we'll test the current platform behavior
	result := getHomebrewPath()

	switch originalOS {
	case "darwin":
		if originalArch == "arm64" {
			assert.Equal(t, "/opt/homebrew/bin", result)
		} else {
			assert.Equal(t, "/usr/local/bin", result)
		}
	case "linux":
		assert.Equal(t, "/home/linuxbrew/.linuxbrew/bin", result)
	default:
		assert.Equal(t, "/opt/homebrew/bin", result)
	}
}

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name      string
		shellPath string
		expected  shellInfo
	}{
		{
			name:      "zsh shell",
			shellPath: "/bin/zsh",
			expected: shellInfo{
				name:       "zsh",
				configFile: "~/.zshrc",
				reload:     "source ~/.zshrc",
			},
		},
		{
			name:      "bash shell",
			shellPath: "/bin/bash",
			expected: shellInfo{
				name:       "bash",
				configFile: "~/.bashrc",
				reload:     "source ~/.bashrc",
			},
		},
		{
			name:      "fish shell",
			shellPath: "/usr/local/bin/fish",
			expected: shellInfo{
				name:       "fish",
				configFile: "~/.config/fish/config.fish",
				reload:     "source ~/.config/fish/config.fish",
			},
		},
		{
			name:      "sh shell defaults to bash",
			shellPath: "/bin/sh",
			expected: shellInfo{
				name:       "bash",
				configFile: "~/.bashrc",
				reload:     "source ~/.bashrc",
			},
		},
		{
			name:      "empty shell path",
			shellPath: "",
			expected: shellInfo{
				name:       "bash",
				configFile: "~/.bashrc",
				reload:     "source ~/.bashrc",
			},
		},
		{
			name:      "unknown shell",
			shellPath: "/bin/exotic-shell",
			expected: shellInfo{
				name:       "bash",
				configFile: "~/.bashrc",
				reload:     "source ~/.bashrc",
			},
		},
		{
			name:      "zsh with full path",
			shellPath: "/usr/local/bin/zsh",
			expected: shellInfo{
				name:       "zsh",
				configFile: "~/.zshrc",
				reload:     "source ~/.zshrc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectShell(tt.shellPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneratePathExport(t *testing.T) {
	tests := []struct {
		name         string
		missingPaths []string
		expected     string
	}{
		{
			name:         "single path",
			missingPaths: []string{"/usr/local/bin"},
			expected:     `export PATH="/usr/local/bin:$PATH"`,
		},
		{
			name:         "multiple paths",
			missingPaths: []string{"/usr/local/bin", "/opt/bin"},
			expected:     `export PATH="/usr/local/bin:/opt/bin:$PATH"`,
		},
		{
			name:         "empty paths",
			missingPaths: []string{},
			expected:     "",
		},
		{
			name:         "paths with spaces",
			missingPaths: []string{"/path with spaces/bin"},
			expected:     `export PATH="/path with spaces/bin:$PATH"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generatePathExport(tt.missingPaths)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateShellCommands(t *testing.T) {
	tests := []struct {
		name       string
		shell      shellInfo
		pathExport string
		expected   []string
	}{
		{
			name: "zsh shell",
			shell: shellInfo{
				name:       "zsh",
				configFile: "~/.zshrc",
				reload:     "source ~/.zshrc",
			},
			pathExport: `export PATH="/usr/local/bin:$PATH"`,
			expected: []string{
				`echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc`,
				"source ~/.zshrc",
			},
		},
		{
			name: "fish shell",
			shell: shellInfo{
				name:       "fish",
				configFile: "~/.config/fish/config.fish",
				reload:     "source ~/.config/fish/config.fish",
			},
			pathExport: `export PATH="/usr/local/bin:$PATH"`,
			expected: []string{
				`fish_add_path /usr/local/bin`,
			},
		},
		{
			name: "bash shell",
			shell: shellInfo{
				name:       "bash",
				configFile: "~/.bashrc",
				reload:     "source ~/.bashrc",
			},
			pathExport: `export PATH="/opt/homebrew/bin:$PATH"`,
			expected: []string{
				`echo 'export PATH="/opt/homebrew/bin:$PATH"' >> ~/.bashrc`,
				"source ~/.bashrc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateShellCommands(tt.shell, tt.pathExport)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateOverallHealth(t *testing.T) {
	tests := []struct {
		name     string
		checks   []HealthCheck
		expected HealthStatus
	}{
		{
			name: "all pass",
			checks: []HealthCheck{
				{Status: "pass"},
				{Status: "pass"},
				{Status: "pass"},
			},
			expected: HealthStatus{Status: "healthy", Message: "All systems operational"},
		},
		{
			name: "one warning",
			checks: []HealthCheck{
				{Status: "pass"},
				{Status: "warn"},
				{Status: "pass"},
			},
			expected: HealthStatus{Status: "warning", Message: "Some issues detected"},
		},
		{
			name: "one error",
			checks: []HealthCheck{
				{Status: "pass"},
				{Status: "warn"},
				{Status: "fail"},
			},
			expected: HealthStatus{Status: "unhealthy", Message: "Critical issues detected"},
		},
		{
			name: "error takes precedence over warning",
			checks: []HealthCheck{
				{Status: "fail"},
				{Status: "warn"},
				{Status: "warn"},
			},
			expected: HealthStatus{Status: "unhealthy", Message: "Critical issues detected"},
		},
		{
			name:     "empty checks",
			checks:   []HealthCheck{},
			expected: HealthStatus{Status: "healthy", Message: "All systems operational"},
		},
		{
			name: "unknown status treated as pass",
			checks: []HealthCheck{
				{Status: "unknown"},
				{Status: "pass"},
			},
			expected: HealthStatus{Status: "healthy", Message: "All systems operational"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateOverallHealth(tt.checks)
			assert.Equal(t, tt.expected, result)
		})
	}
}
