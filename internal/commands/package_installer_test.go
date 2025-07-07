// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"plonk/pkg/config"
)

func TestPackageInstaller_ExtractInstalledPackages(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]string
		expected []string
	}{
		{
			name: "extracts packages with configs from all managers",
			input: map[string][]string{
				"homebrew": {"git", "vim"},
				"asdf":     {"nodejs"},
				"npm":      {"typescript"},
			},
			expected: []string{"git", "vim", "nodejs", "typescript"},
		},
		{
			name:     "handles empty input",
			input:    map[string][]string{},
			expected: []string{},
		},
		{
			name: "handles single manager",
			input: map[string][]string{
				"homebrew": {"git"},
			},
			expected: []string{"git"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractInstalledPackages(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d packages, got %d", len(tt.expected), len(result))
			}

			// Check all expected packages are present
			resultMap := make(map[string]bool)
			for _, pkg := range result {
				resultMap[pkg] = true
			}

			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("Expected package %q not found in result", expected)
				}
			}
		})
	}
}

func TestPackageInstaller_ShouldInstallPackage(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		isInstalled bool
		expected    bool
	}{
		{
			name:        "should install when not installed",
			packageName: "git",
			isInstalled: false,
			expected:    true,
		},
		{
			name:        "should not install when already installed",
			packageName: "git",
			isInstalled: true,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldInstallPackage(tt.packageName, tt.isInstalled)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPackageInstaller_GetPackageDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		pkg      interface{}
		expected string
	}{
		{
			name:     "homebrew package",
			pkg:      config.HomebrewPackage{Name: "git"},
			expected: "git",
		},
		{
			name:     "asdf tool",
			pkg:      config.ASDFTool{Name: "nodejs", Version: "20.0.0"},
			expected: "nodejs@20.0.0",
		},
		{
			name:     "npm package with custom package name",
			pkg:      config.NPMPackage{Name: "tool", Package: "@scope/actual-package"},
			expected: "@scope/actual-package",
		},
		{
			name:     "npm package without custom package name",
			pkg:      config.NPMPackage{Name: "typescript"},
			expected: "typescript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPackageDisplayName(tt.pkg)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPackageInstaller_GetPackageConfig(t *testing.T) {
	tests := []struct {
		name     string
		pkg      interface{}
		expected string
	}{
		{
			name:     "homebrew package with config",
			pkg:      config.HomebrewPackage{Name: "git", Config: "config/git/"},
			expected: "config/git/",
		},
		{
			name:     "homebrew package without config",
			pkg:      config.HomebrewPackage{Name: "vim"},
			expected: "",
		},
		{
			name:     "asdf tool with config",
			pkg:      config.ASDFTool{Name: "nodejs", Version: "20.0.0", Config: "config/node/"},
			expected: "config/node/",
		},
		{
			name:     "npm package with config",
			pkg:      config.NPMPackage{Name: "typescript", Config: "config/ts/"},
			expected: "config/ts/",
		},
		{
			name:     "unknown package type",
			pkg:      "invalid",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPackageConfig(tt.pkg)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPackageInstaller_GetPackageName(t *testing.T) {
	tests := []struct {
		name     string
		pkg      interface{}
		expected string
	}{
		{
			name:     "homebrew package",
			pkg:      config.HomebrewPackage{Name: "git"},
			expected: "git",
		},
		{
			name:     "asdf tool",
			pkg:      config.ASDFTool{Name: "nodejs", Version: "20.0.0"},
			expected: "nodejs",
		},
		{
			name:     "npm package with custom package name",
			pkg:      config.NPMPackage{Name: "tool", Package: "@scope/actual-package"},
			expected: "tool", // Should return Name, not Package
		},
		{
			name:     "unknown package type",
			pkg:      struct{}{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPackageName(tt.pkg)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
