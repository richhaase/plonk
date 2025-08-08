// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
)

func TestPixiManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name:   "standard pixi global list output",
			output: []byte("Global environments as specified in '/Users/user/.pixi/manifests/pixi-global.toml'\n└── hello: 2.12.2 \n    └─ exposes: hello"),
			want:   []string{"hello"},
		},
		{
			name:   "multiple environments",
			output: []byte("Global environments as specified in '/Users/user/.pixi/manifests/pixi-global.toml'\n└── hello: 2.12.2 \n    └─ exposes: hello\n└── ripgrep: 14.1.1\n    └─ exposes: rg"),
			want:   []string{"hello", "ripgrep"},
		},
		{
			name:   "no environments installed",
			output: []byte("No global environments found."),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "single environment with complex name",
			output: []byte("└── python-jupyter: 3.11.0\n    └─ exposes: python, jupyter"),
			want:   []string{"python-jupyter"},
		},
		{
			name:   "environment with version containing spaces",
			output: []byte("└── test-package: 1.2.3 \n    └─ exposes: test"),
			want:   []string{"test-package"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPixiManager()
			got := manager.parseListOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
			} else {
				for i, expected := range tt.want {
					if got[i] != expected {
						t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
						break
					}
				}
			}
		})
	}
}

func TestPixiManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard pixi search output",
			output: []byte(`ripgrep-14.1.1-h0ef69ab_1 (+ 1 build)
-------------------------------------

Name                ripgrep
Version             14.1.1
Build               h0ef69ab_1
Size                1373159`),
			want: []string{"ripgrep"},
		},
		{
			name: "search with complex package name",
			output: []byte(`python-jupyter-3.11.0-h1234567_0
-------------------------------------

Name                python-jupyter
Version             3.11.0`),
			want: []string{"python-jupyter"},
		},
		{
			name:   "empty search results",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "no results found",
			output: []byte("No packages found matching query."),
			want:   []string{},
		},
		{
			name: "package with underscores",
			output: []byte(`test_package-1.0.0-py311_0
-------------------------------------

Name                test_package`),
			want: []string{"test_package"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPixiManager()
			got := manager.parseSearchOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
			} else {
				for i, expected := range tt.want {
					if i >= len(got) || got[i] != expected {
						t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
						break
					}
				}
			}
		})
	}
}

func TestPixiManager_extractPackageName(t *testing.T) {
	tests := []struct {
		name        string
		packageInfo string
		want        string
	}{
		{
			name:        "standard package with version",
			packageInfo: "ripgrep-14.1.1-h0ef69ab_1",
			want:        "ripgrep",
		},
		{
			name:        "package with underscores",
			packageInfo: "python_package-3.11.0-py311_0",
			want:        "python_package",
		},
		{
			name:        "complex package name",
			packageInfo: "jupyter-notebook-6.5.2-pyh6c4a22f_0",
			want:        "jupyter-notebook",
		},
		{
			name:        "simple package name",
			packageInfo: "hello-2.12.2-h0e07e94_0",
			want:        "hello",
		},
		{
			name:        "package without version pattern",
			packageInfo: "simple-package",
			want:        "simple",
		},
		{
			name:        "single word package",
			packageInfo: "git",
			want:        "git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPixiManager()
			got := manager.extractPackageName(tt.packageInfo)
			if got != tt.want {
				t.Errorf("extractPackageName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPixiManager_handleInstallError(t *testing.T) {
	manager := NewPixiManager()

	tests := []struct {
		name         string
		output       []byte
		packageName  string
		exitCode     int
		wantContains string
	}{
		{
			name:         "package not found",
			output:       []byte("No candidates were found for nonexistent"),
			packageName:  "nonexistent",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "dependency resolution failure",
			output:       []byte("failed to solve the environment"),
			packageName:  "testpkg",
			exitCode:     1,
			wantContains: "resolve dependencies",
		},
		{
			name:         "cannot solve request",
			output:       []byte("Cannot solve the request because of: No candidates"),
			packageName:  "missing",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "permission denied",
			output:       []byte("Permission denied accessing directory"),
			packageName:  "testpkg",
			exitCode:     1,
			wantContains: "permission denied",
		},
		{
			name:         "generic error with output",
			output:       []byte("Some installation error occurred"),
			packageName:  "testpkg",
			exitCode:     2,
			wantContains: "installation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleInstallError(mockErr, tt.output, tt.packageName)
			if err == nil {
				t.Errorf("handleInstallError() = nil, want error")
				return
			}

			if tt.wantContains != "" && !stringContains(err.Error(), tt.wantContains) {
				t.Errorf("handleInstallError() error = %v, want to contain %s", err, tt.wantContains)
			}
		})
	}
}

func TestPixiManager_handleUninstallError(t *testing.T) {
	manager := NewPixiManager()

	tests := []struct {
		name            string
		output          []byte
		environmentName string
		exitCode        int
		wantErr         bool
	}{
		{
			name:            "environment not found",
			output:          []byte("No environment named 'notfound' exists"),
			environmentName: "notfound",
			exitCode:        1,
			wantErr:         false, // Should return nil - not an error for uninstall
		},
		{
			name:            "environment does not exist",
			output:          []byte("does not exist"),
			environmentName: "missing",
			exitCode:        1,
			wantErr:         false, // Should return nil - not an error for uninstall
		},
		{
			name:            "permission denied",
			output:          []byte("Permission denied accessing environment directory"),
			environmentName: "testenv",
			exitCode:        1,
			wantErr:         true,
		},
		{
			name:            "generic error",
			output:          []byte("Some uninstallation error"),
			environmentName: "testenv",
			exitCode:        2,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleUninstallError(mockErr, tt.output, tt.environmentName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleUninstallError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPixiManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      bool
	}{
		{
			name: "pixi is available",
			mockResponses: map[string]CommandResponse{
				"pixi --version": {
					Output: []byte("pixi 0.13.0"),
					Error:  nil,
				},
			},
			expected: true,
		},
		{
			name: "pixi not found",
			mockResponses: map[string]CommandResponse{
				"pixi --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 127},
				},
			},
			expected: false,
		},
		{
			name: "pixi exists but not functional",
			mockResponses: map[string]CommandResponse{
				"pixi --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPixiManager()
			result, _ := manager.IsAvailable(context.Background())

			if result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestPixiManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"pixi global install ripgrep": {
					Output: []byte("✔ Added ripgrep@14.1.1 to global manifest"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pixi global install nonexistent": {
					Output: []byte("No candidates were found for nonexistent"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "dependency resolution failure",
			packageName: "problematic",
			mockResponses: map[string]CommandResponse{
				"pixi global install problematic": {
					Output: []byte("failed to solve the environment"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "resolve dependencies",
		},
		{
			name:        "permission denied",
			packageName: "restricted",
			mockResponses: map[string]CommandResponse{
				"pixi global install restricted": {
					Output: []byte("Permission denied accessing directory"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:        "network error",
			packageName: "test-pkg",
			mockResponses: map[string]CommandResponse{
				"pixi global install test-pkg": {
					Output: []byte("Failed to download package from remote"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "installation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPixiManager()
			err := manager.Install(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !stringContains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestPixiManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"pixi global uninstall ripgrep": {
					Output: []byte("✔ Removed ripgrep from global manifest"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "environment not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pixi global uninstall nonexistent": {
					Output: []byte("No environment named 'nonexistent' exists"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not an error for uninstall
		},
		{
			name:        "environment does not exist",
			packageName: "missing",
			mockResponses: map[string]CommandResponse{
				"pixi global uninstall missing": {
					Output: []byte("does not exist"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not an error for uninstall
		},
		{
			name:        "permission denied",
			packageName: "restricted",
			mockResponses: map[string]CommandResponse{
				"pixi global uninstall restricted": {
					Output: []byte("Permission denied accessing environment directory"),
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
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPixiManager()
			err := manager.Uninstall(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !stringContains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestPixiManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name: "list with environments",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("Global environments as specified in '/Users/user/.pixi/manifests/pixi-global.toml'\n└── hello: 2.12.2 \n    └─ exposes: hello\n└── ripgrep: 14.1.1\n    └─ exposes: rg"),
					Error:  nil,
				},
			},
			expected:    []string{"hello", "ripgrep"},
			expectError: false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("No global environments found."),
					Error:  nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name: "command error",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPixiManager()
			result, err := manager.ListInstalled(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && !stringSlicesEqual(result, tt.expected) {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestPixiManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      bool
		expectError   bool
	}{
		{
			name:        "environment is installed",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("└── ripgrep: 14.1.1\n    └─ exposes: rg\n└── hello: 2.12.2\n    └─ exposes: hello"),
					Error:  nil,
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name:        "environment not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("└── ripgrep: 14.1.1\n    └─ exposes: rg"),
					Error:  nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "empty list",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("No global environments found."),
					Error:  nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "command error",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPixiManager()
			result, err := manager.IsInstalled(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestPixiManager_Search(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name:  "successful search",
			query: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"pixi search ripgrep": {
					Output: []byte("ripgrep-14.1.1-h0ef69ab_1 (+ 1 build)\n-------------------------------------\n\nName                ripgrep\nVersion             14.1.1"),
					Error:  nil,
				},
			},
			expected:    []string{"ripgrep"},
			expectError: false,
		},
		{
			name:  "no results found",
			query: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pixi search nonexistent": {
					Output: []byte("No packages found matching query."),
					Error:  nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name:  "command error",
			query: "test",
			mockResponses: map[string]CommandResponse{
				"pixi search test": {
					Output: []byte("error: search failed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPixiManager()
			result, err := manager.Search(context.Background(), tt.query)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && !stringSlicesEqual(result, tt.expected) {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestPixiManager_Info(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
	}{
		{
			name:        "info for installed package",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("└── ripgrep: 14.1.1\n    └─ exposes: rg"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "info for not installed package",
			packageName: "notinstalled",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("No global environments found."),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "command error",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPixiManager()
			result, err := manager.Info(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result == nil {
				t.Errorf("Expected non-nil PackageInfo but got nil")
			}
			if !tt.expectError && result != nil && result.Name != tt.packageName {
				t.Errorf("Expected package name '%s' but got '%s'", tt.packageName, result.Name)
			}
		})
	}
}

func TestPixiManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      string
		expectError   bool
	}{
		{
			name:        "get version of installed environment",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("└── ripgrep: 14.1.1\n    └─ exposes: rg\n└── hello: 2.12.2\n    └─ exposes: hello"),
					Error:  nil,
				},
			},
			expected:    "14.1.1",
			expectError: false,
		},
		{
			name:        "environment not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("└── ripgrep: 14.1.1\n    └─ exposes: rg"),
					Error:  nil,
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name:        "command error",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPixiManager()
			result, err := manager.InstalledVersion(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected '%s' but got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPixiManager_Upgrade(t *testing.T) {
	tests := []struct {
		name          string
		packages      []string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:     "upgrade specific environments",
			packages: []string{"ripgrep", "hello"},
			mockResponses: map[string]CommandResponse{
				"pixi global update ripgrep hello": {
					Output: []byte("✔ Updated ripgrep from 14.1.0 to 14.1.1\n✔ Updated hello from 2.12.1 to 2.12.2"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade all environments",
			packages: []string{}, // empty means all environments
			mockResponses: map[string]CommandResponse{
				"pixi global list": {
					Output: []byte("└── ripgrep: 14.1.1\n    └─ exposes: rg\n└── hello: 2.12.2\n    └─ exposes: hello"),
					Error:  nil,
				},
				"pixi global update ripgrep": {
					Output: []byte("✔ Updated ripgrep from 14.1.1 to 14.1.2"),
					Error:  nil,
				},
				"pixi global update hello": {
					Output: []byte("✔ Updated hello from 2.12.2 to 2.12.3"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade with environment not found error",
			packages: []string{"nonexistent"},
			mockResponses: map[string]CommandResponse{
				"pixi global update nonexistent": {
					Output: []byte("No environment named 'nonexistent' exists"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:     "upgrade with permission error",
			packages: []string{"restricted"},
			mockResponses: map[string]CommandResponse{
				"pixi global update restricted": {
					Output: []byte("Permission denied accessing environment"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:     "upgrade multiple packages returns error on failure",
			packages: []string{"ripgrep", "badenv"},
			mockResponses: map[string]CommandResponse{
				"pixi global update ripgrep badenv": {
					Output: []byte("No environment named 'badenv' exists"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPixiManager()
			err := manager.Upgrade(context.Background(), tt.packages)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !stringContains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

// Note: Uses MockExitError from executor.go and stringContains, stringSlicesEqual from test_helpers.go
