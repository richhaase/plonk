// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
)

func TestPnpmManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name: "list with packages",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "5.3.3"},
							"prettier": {"version": "3.1.0"}
						}
					}`),
					Error: nil,
				},
			},
			expected:    []string{"prettier", "typescript"}, // sorted
			expectError: false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{"dependencies": {}}`),
					Error:  nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name: "command error",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
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

			manager := NewPnpmManager()
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

func TestPnpmManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"pnpm add -g typescript": {
					Output: []byte("+ typescript@5.3.3\nPackage installed successfully"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pnpm add -g nonexistent": {
					Output: []byte("ERR_PNPM_PACKAGE_NOT_FOUND Package 'nonexistent' not found"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "already installed",
			packageName: "prettier",
			mockResponses: map[string]CommandResponse{
				"pnpm add -g prettier": {
					Output: []byte("Package already exists"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Should return nil for already installed
		},
		{
			name:        "permission denied",
			packageName: "eslint",
			mockResponses: map[string]CommandResponse{
				"pnpm add -g eslint": {
					Output: []byte("error: Permission denied"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:        "network error",
			packageName: "lodash",
			mockResponses: map[string]CommandResponse{
				"pnpm add -g lodash": {
					Output: []byte("Network error: connection refused"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "network error",
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

			manager := NewPnpmManager()
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

func TestPnpmManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"pnpm remove -g typescript": {
					Output: []byte("- typescript@5.3.3\nPackage removed successfully"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pnpm remove -g nonexistent": {
					Output: []byte("Package 'nonexistent' not installed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Should return nil for not installed
		},
		{
			name:        "permission denied",
			packageName: "prettier",
			mockResponses: map[string]CommandResponse{
				"pnpm remove -g prettier": {
					Output: []byte("error: Permission denied"),
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

			manager := NewPnpmManager()
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

func TestPnpmManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      bool
		expectError   bool
	}{
		{
			name:        "package is installed",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "5.3.3"},
							"prettier": {"version": "3.1.0"}
						}
					}`),
					Error: nil,
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "5.3.3"}
						}
					}`),
					Error: nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "empty list",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{"dependencies": {}}`),
					Error:  nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "command error",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
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

			manager := NewPnpmManager()
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

func TestPnpmManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      string
		expectError   bool
	}{
		{
			name:        "get version of installed package",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "5.3.3"},
							"prettier": {"version": "3.1.0"}
						}
					}`),
					Error: nil,
				},
			},
			expected:    "5.3.3",
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "5.3.3"}
						}
					}`),
					Error: nil,
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name:        "command error",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
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

			manager := NewPnpmManager()
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

func TestPnpmManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      bool
	}{
		{
			name: "pnpm is available",
			mockResponses: map[string]CommandResponse{
				"pnpm --version": {
					Output: []byte("8.15.1"),
					Error:  nil,
				},
			},
			expected: true,
		},
		{
			name: "pnpm not found",
			mockResponses: map[string]CommandResponse{
				"pnpm --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 127},
				},
			},
			expected: false,
		},
		{
			name: "pnpm exists but not functional",
			mockResponses: map[string]CommandResponse{
				"pnpm --version": {
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

			manager := NewPnpmManager()
			result, _ := manager.IsAvailable(context.Background())

			if result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestPnpmManager_Search(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "search query",
			query:       "typescript",
			expectError: false, // pnpm Search returns empty results, no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPnpmManager()
			result, err := manager.Search(context.Background(), tt.query)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && len(result) != 0 {
				t.Errorf("Expected empty result but got %v", result)
			}
		})
	}
}

func TestPnpmManager_Info(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		expectInfo    bool
	}{
		{
			name:        "info for installed package",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "5.3.3"}
						}
					}`),
					Error: nil,
				},
				"pnpm view typescript --json": {
					Output: []byte(`{
						"name": "typescript",
						"version": "5.3.3",
						"description": "TypeScript is a language for application scale JavaScript development",
						"homepage": "https://www.typescriptlang.org/"
					}`),
					Error: nil,
				},
			},
			expectError: false,
			expectInfo:  true,
		},
		{
			name:        "info for non-installed package",
			packageName: "lodash",
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{"dependencies": {}}`),
					Error:  nil,
				},
				"pnpm view lodash --json": {
					Output: []byte(`{
						"name": "lodash",
						"version": "4.17.21",
						"description": "Lodash modular utilities",
						"homepage": "https://lodash.com/"
					}`),
					Error: nil,
				},
			},
			expectError: false,
			expectInfo:  true,
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

			manager := NewPnpmManager()
			result, err := manager.Info(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectInfo && result == nil {
				t.Errorf("Expected PackageInfo but got nil")
			}
			if tt.expectInfo && result != nil {
				if result.Name != tt.packageName {
					t.Errorf("Expected name %s but got %s", tt.packageName, result.Name)
				}
				if result.Manager != "pnpm" {
					t.Errorf("Expected manager 'pnpm' but got '%s'", result.Manager)
				}
			}
		})
	}
}

func TestPnpmManager_Upgrade(t *testing.T) {
	tests := []struct {
		name          string
		packages      []string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:     "upgrade specific packages",
			packages: []string{"typescript", "prettier"},
			mockResponses: map[string]CommandResponse{
				"pnpm update -g typescript": {
					Output: []byte("+ typescript@5.4.0\nPackage upgraded successfully"),
					Error:  nil,
				},
				"pnpm update -g prettier": {
					Output: []byte("+ prettier@3.2.0\nPackage upgraded successfully"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade all packages",
			packages: []string{}, // empty means all packages
			mockResponses: map[string]CommandResponse{
				"pnpm list -g --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "5.3.3"},
							"prettier": {"version": "3.1.0"}
						}
					}`),
					Error: nil,
				},
				"pnpm update -g typescript": {
					Output: []byte("+ typescript@5.4.0\nPackage upgraded successfully"),
					Error:  nil,
				},
				"pnpm update -g prettier": {
					Output: []byte("+ prettier@3.2.0\nPackage upgraded successfully"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade with package not found error",
			packages: []string{"nonexistent"},
			mockResponses: map[string]CommandResponse{
				"pnpm update -g nonexistent": {
					Output: []byte("ERR_PNPM_PACKAGE_NOT_FOUND Package 'nonexistent' not found"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:     "upgrade with permission error",
			packages: []string{"typescript"},
			mockResponses: map[string]CommandResponse{
				"pnpm update -g typescript": {
					Output: []byte("error: Permission denied"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:     "upgrade already up-to-date",
			packages: []string{"typescript"},
			mockResponses: map[string]CommandResponse{
				"pnpm update -g typescript": {
					Output: []byte("Package already up-to-date"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Should return nil for already up-to-date
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

			manager := NewPnpmManager()
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

func TestPnpmManager_handleInstallError(t *testing.T) {
	manager := NewPnpmManager()

	tests := []struct {
		name         string
		output       []byte
		packageName  string
		exitCode     int
		wantContains string
		wantNil      bool
	}{
		{
			name:         "package not found",
			output:       []byte("ERR_PNPM_PACKAGE_NOT_FOUND Package 'nonexistent' not found"),
			packageName:  "nonexistent",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "permission denied",
			output:       []byte("Permission denied accessing /usr/local"),
			packageName:  "testpkg",
			exitCode:     1,
			wantContains: "permission denied",
		},
		{
			name:        "already installed",
			output:      []byte("Package already exists"),
			packageName: "testpkg",
			exitCode:    1,
			wantNil:     true, // Should return nil
		},
		{
			name:         "network error",
			output:       []byte("Network error: connection refused"),
			packageName:  "testpkg",
			exitCode:     1,
			wantContains: "network error",
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

			if tt.wantNil {
				if err != nil {
					t.Errorf("handleInstallError() = %v, want nil", err)
				}
				return
			}

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

func TestPnpmManager_handleUninstallError(t *testing.T) {
	manager := NewPnpmManager()

	tests := []struct {
		name        string
		output      []byte
		packageName string
		exitCode    int
		wantErr     bool
	}{
		{
			name:        "package not installed",
			output:      []byte("Package 'notinstalled' not installed"),
			packageName: "notinstalled",
			exitCode:    1,
			wantErr:     false, // Should return nil - not an error for uninstall
		},
		{
			name:        "package not found",
			output:      []byte("not found"),
			packageName: "missing",
			exitCode:    1,
			wantErr:     false, // Should return nil - not an error for uninstall
		},
		{
			name:        "permission denied",
			output:      []byte("Permission denied accessing package directory"),
			packageName: "testpkg",
			exitCode:    1,
			wantErr:     true,
		},
		{
			name:        "generic error",
			output:      []byte("Some uninstallation error"),
			packageName: "testpkg",
			exitCode:    2,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleUninstallError(mockErr, tt.output, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleUninstallError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Note: Uses MockExitError from executor.go and stringContains/stringSlicesEqual from test_helpers.go
