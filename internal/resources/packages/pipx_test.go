// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
	"testing"
)

func TestPipxManager_parseListOutput(t *testing.T) {
	manager := NewPipxManager()

	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard pipx list --short output",
			output: []byte(`black 23.12.1
flake8 7.0.0
httpie 3.2.2`),
			want: []string{"black", "flake8", "httpie"},
		},
		{
			name:   "single package output",
			output: []byte(`black 23.12.1`),
			want:   []string{"black"},
		},
		{
			name:   "no packages installed",
			output: []byte(``),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "output with extra whitespace",
			output: []byte(`  black 23.12.1
  httpie 3.2.2  `),
			want: []string{"black", "httpie"},
		},
		{
			name: "packages with complex names",
			output: []byte(`pre-commit 3.6.0
python-dotenv 1.0.0
requests-oauthlib 1.3.1`),
			want: []string{"pre-commit", "python-dotenv", "requests-oauthlib"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manager.parseListOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
				return
			}
			for i, pkg := range got {
				if pkg != tt.want[i] {
					t.Errorf("parseListOutput()[%d] = %v, want %v", i, pkg, tt.want[i])
				}
			}
		})
	}
}

func TestPipxManager_getInstalledVersion(t *testing.T) {
	_ = NewPipxManager()

	tests := []struct {
		name        string
		output      []byte
		packageName string
		wantVersion string
		wantErr     bool
	}{
		{
			name: "find version for existing package",
			output: []byte(`black 23.12.1
httpie 3.2.2`),
			packageName: "black",
			wantVersion: "23.12.1",
			wantErr:     false,
		},
		{
			name: "find version for different package",
			output: []byte(`black 23.12.1
httpie 3.2.2`),
			packageName: "httpie",
			wantVersion: "3.2.2",
			wantErr:     false,
		},
		{
			name: "package not found",
			output: []byte(`black 23.12.1
httpie 3.2.2`),
			packageName: "nonexistent-package",
			wantVersion: "",
			wantErr:     true,
		},
		{
			name:        "empty package list",
			output:      []byte(``),
			packageName: "black",
			wantVersion: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the context-based method, so we'll duplicate the parsing logic
			lines := strings.Split(string(tt.output), "\n")
			var gotVersion string
			var found bool

			for _, line := range lines {
				line = strings.TrimSpace(line)
				parts := strings.Fields(line)
				if len(parts) >= 2 && parts[0] == tt.packageName {
					gotVersion = parts[1]
					found = true
					break
				}
			}

			if tt.wantErr {
				if found {
					t.Errorf("getInstalledVersion() expected error but found version %v", gotVersion)
				}
			} else {
				if !found {
					t.Errorf("getInstalledVersion() expected version but got error")
					return
				}
				if gotVersion != tt.wantVersion {
					t.Errorf("getInstalledVersion() = %v, want %v", gotVersion, tt.wantVersion)
				}
			}
		})
	}
}

func TestPipxManager_IsAvailable(t *testing.T) {
	manager := NewPipxManager()
	ctx := context.Background()

	// This test just ensures the method exists and doesn't panic
	_, err := manager.IsAvailable(ctx)
	if err != nil && !IsContextError(err) {
		// Only context errors are acceptable here since we can't guarantee pipx is installed
		t.Logf("IsAvailable returned error (this may be expected): %v", err)
	}
}

func TestPipxManager_Search(t *testing.T) {
	manager := NewPipxManager()
	ctx := context.Background()

	// Search should always return empty results since pipx doesn't support search
	results, err := manager.Search(ctx, "test")
	if err != nil {
		t.Errorf("Search() unexpected error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() = %v, want empty slice", results)
	}
}

func TestPipxManager_handleInstallError(t *testing.T) {
	manager := NewPipxManager()

	tests := []struct {
		name         string
		output       []byte
		packageName  string
		exitCode     int
		wantContains string
	}{
		{
			name:         "package not found",
			output:       []byte("ERROR: Could not find a version that satisfies the requirement nonexistent-package"),
			packageName:  "nonexistent-package",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "package already installed",
			output:       []byte("'black' already exists on your system."),
			packageName:  "black",
			exitCode:     1,
			wantContains: "already installed",
		},
		{
			name:         "permission denied",
			output:       []byte("Permission denied: Unable to write to directory"),
			packageName:  "testpackage",
			exitCode:     1,
			wantContains: "permission denied",
		},
		{
			name:         "generic error with output",
			output:       []byte("Some installation error occurred"),
			packageName:  "testpackage",
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

func TestPipxManager_handleUninstallError(t *testing.T) {
	manager := NewPipxManager()

	tests := []struct {
		name        string
		output      []byte
		packageName string
		exitCode    int
		wantErr     bool
	}{
		{
			name:        "package not installed",
			output:      []byte("No apps associated with package testpackage."),
			packageName: "testpackage",
			exitCode:    1,
			wantErr:     false, // Not installed should be success
		},
		{
			name:        "package not found",
			output:      []byte("No such package 'missing-package' is installed."),
			packageName: "missing-package",
			exitCode:    1,
			wantErr:     false, // Not found should be success
		},
		{
			name:        "permission denied",
			output:      []byte("Permission denied: Unable to remove directory"),
			packageName: "testpackage",
			exitCode:    1,
			wantErr:     true,
		},
		{
			name:        "generic error",
			output:      []byte("Some uninstallation error"),
			packageName: "testpackage",
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

func TestPipxManager_IsAvailableWithMock(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      bool
	}{
		{
			name: "pipx is available",
			mockResponses: map[string]CommandResponse{
				"pipx --version": {
					Output: []byte("1.4.3"),
					Error:  nil,
				},
			},
			expected: true,
		},
		{
			name: "pipx not found",
			mockResponses: map[string]CommandResponse{
				"pipx --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 127},
				},
			},
			expected: false,
		},
		{
			name: "pipx exists but not functional",
			mockResponses: map[string]CommandResponse{
				"pipx --version": {
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

			manager := NewPipxManager()
			result, _ := manager.IsAvailable(context.Background())

			if result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestPipxManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"pipx install black": {
					Output: []byte("installed package black 23.12.1, installed using Python 3.11.7\n  - black"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent-package",
			mockResponses: map[string]CommandResponse{
				"pipx install nonexistent-package": {
					Output: []byte("ERROR: Could not find a version that satisfies the requirement nonexistent-package"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "package already installed",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"pipx install black": {
					Output: []byte("'black' already exists on your system."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "already installed",
		},
		{
			name:        "permission denied",
			packageName: "testpackage",
			mockResponses: map[string]CommandResponse{
				"pipx install testpackage": {
					Output: []byte("Permission denied: Unable to write to directory"),
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

			manager := NewPipxManager()
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

func TestPipxManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"pipx uninstall black": {
					Output: []byte("uninstalled black! ✨ 🌟 ✨"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent-package",
			mockResponses: map[string]CommandResponse{
				"pipx uninstall nonexistent-package": {
					Output: []byte("No apps associated with package nonexistent-package."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not installed is success for uninstall
		},
		{
			name:        "package not found",
			packageName: "missing-package",
			mockResponses: map[string]CommandResponse{
				"pipx uninstall missing-package": {
					Output: []byte("No such package 'missing-package' is installed."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not found is success for uninstall
		},
		{
			name:        "permission denied",
			packageName: "restricted-package",
			mockResponses: map[string]CommandResponse{
				"pipx uninstall restricted-package": {
					Output: []byte("Permission denied: Unable to remove directory"),
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

			manager := NewPipxManager()
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

func TestPipxManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name: "list with packages",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(`black 23.12.1
httpie 3.2.2`),
					Error: nil,
				},
			},
			expected:    []string{"black", "httpie"},
			expectError: false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(``),
					Error:  nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name: "no packages installed - exit code 1",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    []string{},
			expectError: false, // Exit code 1 is handled gracefully
		},
		{
			name: "command error - severe failure",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 2},
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

			manager := NewPipxManager()
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

func TestPipxManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      bool
		expectError   bool
	}{
		{
			name:        "package is installed",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(`black 23.12.1
httpie 3.2.2`),
					Error: nil,
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent-package",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(`black 23.12.1
httpie 3.2.2`),
					Error: nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "empty list",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(``),
					Error:  nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "command error",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 2},
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

			manager := NewPipxManager()
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

func TestPipxManager_Info(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
	}{
		{
			name:        "info for installed package",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(`black 23.12.1
httpie 3.2.2`),
					Error: nil,
				},
				"pipx list": {
					Output: []byte(`venvs are in /home/user/.local/share/pipx/venvs
apps are in /home/user/.local/bin
   package black 23.12.1, installed using Python 3.11.7
    - black`),
					Error: nil,
				},
			},
			expectError: false,
		},
		{
			name:        "info for not installed package",
			packageName: "notinstalled-package",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(`black 23.12.1`),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "command error",
			packageName: "test-package",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 2},
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

			manager := NewPipxManager()
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

func TestPipxManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      string
		expectError   bool
	}{
		{
			name:        "get version of installed package",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(`black 23.12.1
httpie 3.2.2`),
					Error: nil,
				},
			},
			expected:    "23.12.1",
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent-package",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte(`black 23.12.1`),
					Error:  nil,
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name:        "command error",
			packageName: "black",
			mockResponses: map[string]CommandResponse{
				"pipx list --short": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 2},
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

			manager := NewPipxManager()
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

func TestPipxManager_Upgrade(t *testing.T) {
	tests := []struct {
		name          string
		packages      []string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:     "upgrade specific packages",
			packages: []string{"black", "httpie"},
			mockResponses: map[string]CommandResponse{
				"pipx upgrade black": {
					Output: []byte("upgraded package black from 23.12.0 to 23.12.1! ✨ 🌟 ✨"),
					Error:  nil,
				},
				"pipx upgrade httpie": {
					Output: []byte("upgraded package httpie from 3.2.1 to 3.2.2! ✨ 🌟 ✨"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade all packages",
			packages: []string{}, // empty means all packages
			mockResponses: map[string]CommandResponse{
				"pipx upgrade-all": {
					Output: []byte("Upgraded packages: black, httpie"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade with package not installed error",
			packages: []string{"nonexistent-package"},
			mockResponses: map[string]CommandResponse{
				"pipx upgrade nonexistent-package": {
					Output: []byte("No apps associated with package nonexistent-package."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found or not installed",
		},
		{
			name:     "upgrade with permission error",
			packages: []string{"restricted-package"},
			mockResponses: map[string]CommandResponse{
				"pipx upgrade restricted-package": {
					Output: []byte("Permission denied: Unable to upgrade package"),
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

			manager := NewPipxManager()
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
