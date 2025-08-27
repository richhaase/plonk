// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
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
			result := SplitLines(tt.output)
			if len(result) != len(tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
			for i, expected := range tt.expectedResult {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
					break
				}
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
			if len(result) != len(tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			} else {
				for i, expected := range tt.expectedResult {
					if result[i] != expected {
						t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
						break
					}
				}
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

func TestHomebrewManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name: "brew is available",
			mockResponses: map[string]CommandResponse{
				"brew --version": {
					Output: []byte("Homebrew 3.6.0"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:          "brew not found",
			mockResponses: map[string]CommandResponse{
				// Empty responses means LookPath will fail
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "brew exists but not functional",
			mockResponses: map[string]CommandResponse{
				"brew": {}, // Makes LookPath succeed
				"brew --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewHomebrewManager()
			result, err := manager.IsAvailable(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "vim",
			mockResponses: map[string]CommandResponse{
				"brew install vim": {
					Output: []byte("Installing vim..."),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"brew install nonexistent": {
					Output: []byte("Error: No available formula with the name \"nonexistent\""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "already installed",
			packageName: "git",
			mockResponses: map[string]CommandResponse{
				"brew install git": {
					Output: []byte("Warning: git 2.37.1 is already installed"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "permission denied",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"brew install test": {
					Output: []byte("Error: Permission denied"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewHomebrewManager()
			err := manager.Install(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && tt.errorContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}

			// Verify the command was called
			if len(mock.Commands) != 1 {
				t.Errorf("Expected 1 command call, got %d", len(mock.Commands))
			} else {
				cmd := mock.Commands[0]
				if cmd.Name != "brew" || len(cmd.Args) != 2 || cmd.Args[0] != "install" || cmd.Args[1] != tt.packageName {
					t.Errorf("Unexpected command: %s %v", cmd.Name, cmd.Args)
				}
			}
		})
	}
}

func TestHomebrewManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "vim",
			mockResponses: map[string]CommandResponse{
				"brew uninstall vim": {
					Output: []byte("Uninstalling vim..."),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"brew uninstall nonexistent": {
					Output: []byte("Error: No such keg: /usr/local/Cellar/nonexistent"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not installed is success for uninstall
		},
		{
			name:        "dependency conflict",
			packageName: "python",
			mockResponses: map[string]CommandResponse{
				"brew uninstall python": {
					Output: []byte("Error: Refusing to uninstall because it is required by vim"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "dependency conflicts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewHomebrewManager()
			err := manager.Uninstall(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && tt.errorContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestHomebrewManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expectedCount int
		expectError   bool
	}{
		{
			name: "simple list",
			mockResponses: map[string]CommandResponse{
				"brew list": {
					Output: []byte("git\nvim\nnode"),
					Error:  nil,
				},
			},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"brew list": {
					Output: []byte(""),
					Error:  nil,
				},
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "list with JSON enrichment",
			mockResponses: map[string]CommandResponse{
				"brew list": {
					Output: []byte("git\nvim"),
					Error:  nil,
				},
				"brew info --installed --json=v2": {
					Output: []byte(`{"formulae":[{"name":"git","aliases":["gitscm"]},{"name":"vim","aliases":["vi"]}],"casks":[]}`),
					Error:  nil,
				},
			},
			expectedCount: 4, // git, gitscm, vim, vi
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewHomebrewManager()
			result, err := manager.ListInstalled(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d packages but got %d: %v", tt.expectedCount, len(result), result)
			}
		})
	}
}

func TestHomebrewManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name:        "package is installed",
			packageName: "git",
			mockResponses: map[string]CommandResponse{
				"brew info --installed --json=v2": {
					Output: []byte(`{"formulae":[{"name":"git","aliases":[],"installed":[{"version":"2.37.1"}]}],"casks":[]}`),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:        "package not installed",
			packageName: "vim",
			mockResponses: map[string]CommandResponse{
				"brew info --installed --json=v2": {
					Output: []byte(`{"formulae":[{"name":"git","aliases":[]}],"casks":[]}`),
					Error:  nil,
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:        "check by alias",
			packageName: "vi",
			mockResponses: map[string]CommandResponse{
				"brew info --installed --json=v2": {
					Output: []byte(`{"formulae":[{"name":"vim","aliases":["vi"]}],"casks":[]}`),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:        "fallback to brew list",
			packageName: "git",
			mockResponses: map[string]CommandResponse{
				"brew info --installed --json=v2": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
				"brew list git": {
					Output: []byte("/usr/local/bin/git"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewHomebrewManager()
			result, err := manager.IsInstalled(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManager_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockResponses  map[string]CommandResponse
		expectedResult []string
		expectError    bool
	}{
		{
			name:  "successful search",
			query: "git",
			mockResponses: map[string]CommandResponse{
				"brew search git": {
					Output: []byte("git\ngit-flow\ngithub-cli"),
					Error:  nil,
				},
			},
			expectedResult: []string{"git", "git-flow", "github-cli"},
			expectError:    false,
		},
		{
			name:  "no results",
			query: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"brew search nonexistent": {
					Output: []byte("No formula found"),
					Error:  nil,
				},
			},
			expectedResult: []string{},
			expectError:    false,
		},
		{
			name:  "search error",
			query: "test",
			mockResponses: map[string]CommandResponse{
				"brew search test": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewHomebrewManager()
			result, err := manager.Search(context.Background(), tt.query)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult string
		expectError    bool
	}{
		{
			name:        "get version from JSON",
			packageName: "git",
			mockResponses: map[string]CommandResponse{
				"brew info --installed --json=v2": {
					Output: []byte(`{"formulae":[{"name":"git","installed":[{"version":"2.37.1"}],"versions":{"stable":"2.37.1"}}],"casks":[]}`),
					Error:  nil,
				},
			},
			expectedResult: "2.37.1",
			expectError:    false,
		},
		{
			name:        "fallback to brew list --versions",
			packageName: "vim",
			mockResponses: map[string]CommandResponse{
				"brew info --installed --json=v2": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
				"brew list --versions vim": {
					Output: []byte("vim 9.0.0"),
					Error:  nil,
				},
			},
			expectedResult: "9.0.0",
			expectError:    false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"brew info --installed --json=v2": {
					Output: []byte(`{"formulae":[],"casks":[]}`),
					Error:  nil,
				},
			},
			expectedResult: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewHomebrewManager()
			result, err := manager.InstalledVersion(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManager_Info(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult *PackageInfo
		expectError    bool
	}{
		{
			name:        "successful info",
			packageName: "git",
			mockResponses: map[string]CommandResponse{
				"brew info git": {
					Output: []byte("git: stable 2.37.1\nDistributed revision control system\nFrom: https://github.com/git/git"),
					Error:  nil,
				},
				"brew info --installed --json=v2": {
					Output: []byte(`{"formulae":[{"name":"git","installed":[{"version":"2.37.1"}]}],"casks":[]}`),
					Error:  nil,
				},
			},
			expectedResult: &PackageInfo{
				Name:        "git",
				Version:     "2.37.1",
				Description: "Distributed revision control system",
				Homepage:    "https://github.com/git/git",
				Manager:     "brew",
				Installed:   true,
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"brew info nonexistent": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewHomebrewManager()
			result, err := manager.Info(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && !equalPackageInfo(result, tt.expectedResult) {
				t.Errorf("Expected result %+v but got %+v", tt.expectedResult, result)
			}
		})
	}
}
