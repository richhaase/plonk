// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
)

func TestUvManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name:   "standard uv tool list output",
			output: []byte("cowsay v6.1\n- cowsay\nruff v0.1.0\n- ruff"),
			want:   []string{"cowsay", "ruff"},
		},
		{
			name:   "single tool",
			output: []byte("black v23.1.0\n- black"),
			want:   []string{"black"},
		},
		{
			name:   "no tools installed",
			output: []byte("No tools installed"),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "tool with path info",
			output: []byte("pytest v7.2.0 (/home/user/.local/share/uv/tools/pytest)\n- pytest"),
			want:   []string{"pytest"},
		},
		{
			name:   "multiple tools with executables",
			output: []byte("cowsay v6.1\n- cowsay\n- cowthink\nruff v0.1.0\n- ruff\npytest v7.2.0\n- pytest\n- py.test"),
			want:   []string{"cowsay", "ruff", "pytest"},
		},
		{
			name:   "tools with complex names",
			output: []byte("black-formatter v1.0.0\n- black\npython-lsp-server v1.7.1\n- pylsp"),
			want:   []string{"black-formatter", "python-lsp-server"},
		},
		{
			name:   "tools with hyphenated names",
			output: []byte("pre-commit v2.20.0\n- pre-commit\nblack-formatter v1.0.0\n- black"),
			want:   []string{"pre-commit", "black-formatter"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewUvManager()
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

func TestUvManager_SupportsSearch(t *testing.T) {
	manager := NewUvManager()
	if manager.SupportsSearch() {
		t.Errorf("SupportsSearch() = true, want false - UV does not support search")
	}
}

func TestUvManager_handleInstallError(t *testing.T) {
	manager := NewUvManager()

	tests := []struct {
		name         string
		output       []byte
		packageName  string
		exitCode     int
		wantContains string
	}{
		{
			name:         "package not found",
			output:       []byte("No such package 'nonexistent'"),
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
			name:         "404 error",
			output:       []byte("404: Package not found in registry"),
			packageName:  "missing",
			exitCode:     1,
			wantContains: "not found",
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

func TestUvManager_handleUninstallError(t *testing.T) {
	manager := NewUvManager()

	tests := []struct {
		name        string
		output      []byte
		packageName string
		exitCode    int
		wantErr     bool
	}{
		{
			name:        "tool not installed",
			output:      []byte("No tool named 'notinstalled' found"),
			packageName: "notinstalled",
			exitCode:    1,
			wantErr:     false, // Should return nil - not an error for uninstall
		},
		{
			name:        "tool not found",
			output:      []byte("not found"),
			packageName: "missing",
			exitCode:    1,
			wantErr:     false, // Should return nil - not an error for uninstall
		},
		{
			name:        "permission denied",
			output:      []byte("Permission denied accessing tool directory"),
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

func TestUvManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      bool
	}{
		{
			name: "uv is available",
			mockResponses: map[string]CommandResponse{
				"uv --version": {
					Output: []byte("uv 0.1.0"),
					Error:  nil,
				},
			},
			expected: true,
		},
		{
			name: "uv not found",
			mockResponses: map[string]CommandResponse{
				"uv --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 127},
				},
			},
			expected: false,
		},
		{
			name: "uv exists but not functional",
			mockResponses: map[string]CommandResponse{
				"uv --version": {
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

			manager := NewUvManager()
			result, _ := manager.IsAvailable(context.Background())

			if result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestUvManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "cowsay",
			mockResponses: map[string]CommandResponse{
				"uv tool install cowsay": {
					Output: []byte("Resolved 1 package in 234ms\nInstalled 1 package in 567ms\n+ cowsay==6.1\nInstalled cowsay"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"uv tool install nonexistent": {
					Output: []byte("error: Package `nonexistent` not found"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "already installed",
			packageName: "ruff",
			mockResponses: map[string]CommandResponse{
				"uv tool install ruff": {
					Output: []byte("Tool `ruff` is already installed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "already installed",
		},
		{
			name:        "network error",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"uv tool install black": {
					Output: []byte("error: Failed to fetch package metadata: network unreachable"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "tool installation failed",
		},
		{
			name:        "permission denied",
			packageName: "pytest",
			mockResponses: map[string]CommandResponse{
				"uv tool install pytest": {
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

			manager := NewUvManager()
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

func TestUvManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "cowsay",
			mockResponses: map[string]CommandResponse{
				"uv tool uninstall cowsay": {
					Output: []byte("Uninstalled cowsay"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "tool not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"uv tool uninstall nonexistent": {
					Output: []byte("Tool `nonexistent` is not installed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // UV handleUninstallError returns nil for "not installed"
		},
		{
			name:        "permission denied",
			packageName: "ruff",
			mockResponses: map[string]CommandResponse{
				"uv tool uninstall ruff": {
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

			manager := NewUvManager()
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

func TestUvManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name: "list with tools",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
					Output: []byte("cowsay v6.1\n- cowsay\nruff v0.1.0\n- ruff"),
					Error:  nil,
				},
			},
			expected:    []string{"cowsay", "ruff"},
			expectError: false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
					Output: []byte("No tools installed"),
					Error:  nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name: "command error",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
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

			manager := NewUvManager()
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

func TestUvManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      bool
		expectError   bool
	}{
		{
			name:        "tool is installed",
			packageName: "cowsay",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
					Output: []byte("cowsay v6.1\n- cowsay\nruff v0.1.0\n- ruff"),
					Error:  nil,
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name:        "tool not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
					Output: []byte("cowsay v6.1\n- cowsay"),
					Error:  nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "empty list",
			packageName: "cowsay",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
					Output: []byte("No tools installed"),
					Error:  nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "command error",
			packageName: "cowsay",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
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

			manager := NewUvManager()
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

func TestUvManager_Search(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "search query",
			query:       "black",
			expectError: false, // UV Search returns empty results, no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewUvManager()
			_, err := manager.Search(context.Background(), tt.query)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestUvManager_Info(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		expectError bool
	}{
		{
			name:        "info query",
			packageName: "cowsay",
			expectError: false, // UV Info returns basic PackageInfo, no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewUvManager()
			_, err := manager.Info(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
		})
	}
}

func TestUvManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      string
		expectError   bool
	}{
		{
			name:        "get version of installed tool",
			packageName: "cowsay",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
					Output: []byte("cowsay v6.1\n- cowsay\nruff v0.1.0\n- ruff"),
					Error:  nil,
				},
			},
			expected:    "6.1",
			expectError: false,
		},
		{
			name:        "tool not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
					Output: []byte("cowsay v6.1\n- cowsay"),
					Error:  nil,
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name:        "command error",
			packageName: "cowsay",
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
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

			manager := NewUvManager()
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

func TestUvManager_Upgrade(t *testing.T) {
	tests := []struct {
		name          string
		packages      []string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:     "upgrade specific tools",
			packages: []string{"cowsay", "ruff"},
			mockResponses: map[string]CommandResponse{
				"uv tool upgrade cowsay": {
					Output: []byte("Upgraded cowsay to v6.2"),
					Error:  nil,
				},
				"uv tool upgrade ruff": {
					Output: []byte("Upgraded ruff to v0.2.0"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade all tools",
			packages: []string{}, // empty means all tools
			mockResponses: map[string]CommandResponse{
				"uv tool list": {
					Output: []byte("cowsay v6.1\n- cowsay\nruff v0.1.0\n- ruff"),
					Error:  nil,
				},
				"uv tool upgrade cowsay": {
					Output: []byte("Upgraded cowsay to v6.2"),
					Error:  nil,
				},
				"uv tool upgrade ruff": {
					Output: []byte("Upgraded ruff to v0.2.0"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade with tool not installed error",
			packages: []string{"nonexistent"},
			mockResponses: map[string]CommandResponse{
				"uv tool upgrade nonexistent": {
					Output: []byte("Tool `nonexistent` is not installed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:     "upgrade with permission error",
			packages: []string{"cowsay"},
			mockResponses: map[string]CommandResponse{
				"uv tool upgrade cowsay": {
					Output: []byte("error: Permission denied"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:     "upgrade continues on individual failures",
			packages: []string{"cowsay", "badtool"},
			mockResponses: map[string]CommandResponse{
				"uv tool upgrade cowsay": {
					Output: []byte("Upgraded cowsay to v6.2"),
					Error:  nil,
				},
				"uv tool upgrade badtool": {
					Output: []byte("Tool `badtool` is not installed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // This test expects single package upgrade, no error for this case
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

			manager := NewUvManager()
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

// Note: Uses MockExitError from executor.go and stringContains from test_helpers.go
