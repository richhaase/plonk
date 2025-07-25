// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"
)

func TestHomebrewManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         []byte
		expectedResult []string
	}{
		{
			name:           "normal output",
			output:         []byte("git\nnode\npython@3.9"),
			expectedResult: []string{"git", "node", "python@3.9"},
		},
		{
			name:           "empty output",
			output:         []byte(""),
			expectedResult: []string{},
		},
		{
			name:           "output with extra whitespace",
			output:         []byte("  git  \n  node  \n  python@3.9  "),
			expectedResult: []string{"git", "node", "python@3.9"},
		},
		{
			name:           "single package with newline",
			output:         []byte("git\n"),
			expectedResult: []string{"git"},
		},
		{
			name:           "packages with spaces in names",
			output:         []byte("docker-compose\npython@3.11\nnode@18"),
			expectedResult: []string{"docker-compose", "python@3.11", "node@18"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewHomebrewManager()
			result := manager.parseListOutput(tt.output)

			if !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         []byte
		expectedResult []string
	}{
		{
			name:           "normal output",
			output:         []byte("git\ngit-flow\ngithub-cli"),
			expectedResult: []string{"git", "git-flow", "github-cli"},
		},
		{
			name:           "no results",
			output:         []byte("No formula found"),
			expectedResult: []string{},
		},
		{
			name:           "empty output",
			output:         []byte(""),
			expectedResult: []string{},
		},
		{
			name:           "output with headers",
			output:         []byte("==> Formulae\ngit\ngit-flow\n==> Casks\ngithub-desktop"),
			expectedResult: []string{"git", "git-flow", "github-desktop"},
		},
		{
			name:           "output with info messages",
			output:         []byte("git\ngit-flow\nIf you meant \"git\" specifically:\nbrew install git"),
			expectedResult: []string{"git", "git-flow", "brew install git"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewHomebrewManager()
			result := manager.parseSearchOutput(tt.output)

			if !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         []byte
		packageName    string
		expectedResult *PackageInfo
	}{
		{
			name:        "normal output",
			output:      []byte("git: stable 2.37.1\nDistributed revision control system\nFrom: https://github.com/git/git"),
			packageName: "git",
			expectedResult: &PackageInfo{
				Name:        "git",
				Version:     "2.37.1",
				Description: "Distributed revision control system",
				Homepage:    "https://github.com/git/git",
			},
		},
		{
			name:        "minimal output",
			output:      []byte("git: stable 2.37.1"),
			packageName: "git",
			expectedResult: &PackageInfo{
				Name:        "git",
				Version:     "2.37.1",
				Description: "",
				Homepage:    "",
			},
		},
		{
			name:        "output with URL prefix",
			output:      []byte("node: stable 18.12.1\nPlatform built on V8 to build network applications\nURL: https://nodejs.org/"),
			packageName: "node",
			expectedResult: &PackageInfo{
				Name:        "node",
				Version:     "18.12.1",
				Description: "Platform built on V8 to build network applications",
				Homepage:    "https://nodejs.org/",
			},
		},
		{
			name:        "complex output with description on same line",
			output:      []byte("python@3.11: stable 3.11.6 High-level, general-purpose programming language\nFrom: https://www.python.org/"),
			packageName: "python@3.11",
			expectedResult: &PackageInfo{
				Name:        "python@3.11",
				Version:     "3.11.6",
				Description: "High-level, general-purpose programming language",
				Homepage:    "https://www.python.org/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewHomebrewManager()
			result := manager.parseInfoOutput(tt.output, tt.packageName)

			if !equalPackageInfo(result, tt.expectedResult) {
				t.Errorf("Expected result %+v but got %+v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManager_extractVersion(t *testing.T) {
	tests := []struct {
		name           string
		output         []byte
		packageName    string
		expectedResult string
	}{
		{
			name:           "normal output",
			output:         []byte("git 2.37.1 2.36.0"),
			packageName:    "git",
			expectedResult: "2.37.1",
		},
		{
			name:           "single version",
			output:         []byte("git 2.37.1"),
			packageName:    "git",
			expectedResult: "2.37.1",
		},
		{
			name:           "no version",
			output:         []byte("git"),
			packageName:    "git",
			expectedResult: "",
		},
		{
			name:           "empty output",
			output:         []byte(""),
			packageName:    "git",
			expectedResult: "",
		},
		{
			name:           "wrong package name",
			output:         []byte("node 18.12.1"),
			packageName:    "git",
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewHomebrewManager()
			result := manager.extractVersion(tt.output, tt.packageName)

			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

// Helper functions for testing
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalPackageInfo(a, b *PackageInfo) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Name == b.Name &&
		a.Version == b.Version &&
		a.Description == b.Description &&
		a.Homepage == b.Homepage &&
		a.Manager == b.Manager &&
		a.Installed == b.Installed
}
