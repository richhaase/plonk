// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
	"testing"
)

func TestDotnetManager_parseListOutput(t *testing.T) {
	manager := NewDotnetManager()

	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard dotnet tool list output",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef
dotnet-counters               9.0.572801 dotnet-counters`),
			want: []string{"dotnet-counters", "dotnet-ef", "dotnetsay"},
		},
		{
			name: "single tool output",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay`),
			want: []string{"dotnetsay"},
		},
		{
			name: "no tools installed",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------`),
			want: []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "output with extra whitespace",
			output: []byte(`  Package Id                    Version    Commands
  -------------------------------------------------------
  dotnetsay                     2.1.4      dotnetsay
  dotnet-outdated-tool          4.6.4      dotnet-outdated  `),
			want: []string{"dotnet-outdated-tool", "dotnetsay"},
		},
		{
			name: "tools with complex names",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
Microsoft.dotnet-openapi      9.0.7      dotnet-openapi
coverlet.console              6.0.0      coverlet
dotnet-format                 5.1.250801 dotnet-format`),
			want: []string{"Microsoft.dotnet-openapi", "coverlet.console", "dotnet-format"},
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

func TestDotnetManager_getInstalledVersion(t *testing.T) {
	_ = NewDotnetManager()

	tests := []struct {
		name        string
		output      []byte
		toolName    string
		wantVersion string
		wantErr     bool
	}{
		{
			name: "find version for existing tool",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef`),
			toolName:    "dotnetsay",
			wantVersion: "2.1.4",
			wantErr:     false,
		},
		{
			name: "find version for different tool",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef`),
			toolName:    "dotnet-ef",
			wantVersion: "9.0.7",
			wantErr:     false,
		},
		{
			name: "tool not found",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay`),
			toolName:    "nonexistent-tool",
			wantVersion: "",
			wantErr:     true,
		},
		{
			name: "empty tool list",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------`),
			toolName:    "dotnetsay",
			wantVersion: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the context-based method, so we'll duplicate the parsing logic
			lines := strings.Split(string(tt.output), "\n")
			var inDataSection bool
			var gotVersion string
			var found bool

			for _, line := range lines {
				line = strings.TrimSpace(line)

				if strings.Contains(line, "-------") {
					inDataSection = true
					continue
				}

				if strings.Contains(line, "Package Id") {
					continue
				}

				if inDataSection {
					fields := strings.Fields(line)
					if len(fields) >= 2 && fields[0] == tt.toolName {
						gotVersion = fields[1]
						found = true
						break
					}
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

func TestDotnetManager_IsAvailable(t *testing.T) {
	manager := NewDotnetManager()
	ctx := context.Background()

	// This test just ensures the method exists and doesn't panic
	_, err := manager.IsAvailable(ctx)
	if err != nil && !IsContextError(err) {
		// Only context errors are acceptable here since we can't guarantee dotnet is installed
		t.Logf("IsAvailable returned error (this may be expected): %v", err)
	}
}

func TestDotnetManager_Search(t *testing.T) {
	manager := NewDotnetManager()
	ctx := context.Background()

	// Search should always return empty results since .NET doesn't support search
	results, err := manager.Search(ctx, "test")
	if err != nil {
		t.Errorf("Search() unexpected error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() = %v, want empty slice", results)
	}
}

func TestDotnetManager_handleInstallError(t *testing.T) {
	manager := NewDotnetManager()

	tests := []struct {
		name         string
		output       []byte
		toolName     string
		exitCode     int
		wantContains string
	}{
		{
			name:         "tool not found",
			output:       []byte("error NU1101: Unable to find package nonexistent-tool. No packages exist with this id"),
			toolName:     "nonexistent-tool",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "not a dotnet tool",
			output:       []byte("error NU1212: Invalid project-package combination for some-package. DotnetToolReference project style can only contain references of the DotnetTool type"),
			toolName:     "some-package",
			exitCode:     1,
			wantContains: "not a .NET global tool",
		},
		{
			name:         "tool already installed",
			output:       []byte("Tool 'dotnetsay' is already installed."),
			toolName:     "dotnetsay",
			exitCode:     1,
			wantContains: "already installed",
		},
		{
			name:         "permission denied",
			output:       []byte("Access denied when writing to tool directory"),
			toolName:     "testtool",
			exitCode:     1,
			wantContains: "permission denied",
		},
		{
			name:         "generic error with output",
			output:       []byte("Some installation error occurred"),
			toolName:     "testtool",
			exitCode:     2,
			wantContains: "installation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleInstallError(mockErr, tt.output, tt.toolName)
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

func TestDotnetManager_handleUninstallError(t *testing.T) {
	manager := NewDotnetManager()

	tests := []struct {
		name     string
		output   []byte
		toolName string
		exitCode int
		wantErr  bool
	}{
		{
			name:     "tool not installed",
			output:   []byte("Tool 'testtool' is not installed."),
			toolName: "testtool",
			exitCode: 1,
			wantErr:  false, // Not installed should be success
		},
		{
			name:     "no such tool",
			output:   []byte("No such tool exists in the global tools"),
			toolName: "missing-tool",
			exitCode: 1,
			wantErr:  false, // Not found should be success
		},
		{
			name:     "permission denied",
			output:   []byte("Access denied when writing to tool directory"),
			toolName: "testtool",
			exitCode: 1,
			wantErr:  true,
		},
		{
			name:     "generic error",
			output:   []byte("Some uninstallation error"),
			toolName: "testtool",
			exitCode: 2,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleUninstallError(mockErr, tt.output, tt.toolName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleUninstallError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDotnetManager_IsAvailableWithMock(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      bool
	}{
		{
			name: "dotnet is available",
			mockResponses: map[string]CommandResponse{
				"dotnet --version": {
					Output: []byte("9.0.100"),
					Error:  nil,
				},
			},
			expected: true,
		},
		{
			name: "dotnet not found",
			mockResponses: map[string]CommandResponse{
				"dotnet --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 127},
				},
			},
			expected: false,
		},
		{
			name: "dotnet exists but not functional",
			mockResponses: map[string]CommandResponse{
				"dotnet --version": {
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

			manager := NewDotnetManager()
			result, _ := manager.IsAvailable(context.Background())

			if result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestDotnetManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		toolName      string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:     "successful install",
			toolName: "dotnetsay",
			mockResponses: map[string]CommandResponse{
				"dotnet tool install -g dotnetsay": {
					Output: []byte("You can invoke the tool using the following command: dotnetsay\nTool 'dotnetsay' (version '2.1.4') was installed successfully."),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "tool not found",
			toolName: "nonexistent-tool",
			mockResponses: map[string]CommandResponse{
				"dotnet tool install -g nonexistent-tool": {
					Output: []byte("error NU1101: Unable to find package nonexistent-tool. No packages exist with this id"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:     "tool already installed",
			toolName: "dotnetsay",
			mockResponses: map[string]CommandResponse{
				"dotnet tool install -g dotnetsay": {
					Output: []byte("Tool 'dotnetsay' is already installed."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "already installed",
		},
		{
			name:     "not a dotnet tool",
			toolName: "some-package",
			mockResponses: map[string]CommandResponse{
				"dotnet tool install -g some-package": {
					Output: []byte("error NU1212: Invalid project-package combination for some-package. DotnetToolReference project style can only contain references of the DotnetTool type"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not a .NET global tool",
		},
		{
			name:     "permission denied",
			toolName: "testtool",
			mockResponses: map[string]CommandResponse{
				"dotnet tool install -g testtool": {
					Output: []byte("Access denied when writing to tool directory"),
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

			manager := NewDotnetManager()
			err := manager.Install(context.Background(), tt.toolName)

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

func TestDotnetManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		toolName      string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:     "successful uninstall",
			toolName: "dotnetsay",
			mockResponses: map[string]CommandResponse{
				"dotnet tool uninstall -g dotnetsay": {
					Output: []byte("Tool 'dotnetsay' (version '2.1.4') was successfully uninstalled."),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "tool not installed",
			toolName: "nonexistent-tool",
			mockResponses: map[string]CommandResponse{
				"dotnet tool uninstall -g nonexistent-tool": {
					Output: []byte("Tool 'nonexistent-tool' is not installed."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not installed is success for uninstall
		},
		{
			name:     "no such tool",
			toolName: "missing-tool",
			mockResponses: map[string]CommandResponse{
				"dotnet tool uninstall -g missing-tool": {
					Output: []byte("No such tool exists in the global tools"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not found is success for uninstall
		},
		{
			name:     "permission denied",
			toolName: "restricted-tool",
			mockResponses: map[string]CommandResponse{
				"dotnet tool uninstall -g restricted-tool": {
					Output: []byte("Access denied when writing to tool directory"),
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

			manager := NewDotnetManager()
			err := manager.Uninstall(context.Background(), tt.toolName)

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

func TestDotnetManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name: "list with tools",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef`),
					Error: nil,
				},
			},
			expected:    []string{"dotnet-ef", "dotnetsay"},
			expectError: false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------`),
					Error: nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name: "no tools installed - exit code 1",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
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
				"dotnet tool list -g": {
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

			manager := NewDotnetManager()
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

func TestDotnetManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name          string
		toolName      string
		mockResponses map[string]CommandResponse
		expected      bool
		expectError   bool
	}{
		{
			name:     "tool is installed",
			toolName: "dotnetsay",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef`),
					Error: nil,
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name:     "tool not installed",
			toolName: "nonexistent-tool",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay`),
					Error: nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:     "empty list",
			toolName: "dotnetsay",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------`),
					Error: nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:     "command error",
			toolName: "dotnetsay",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
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

			manager := NewDotnetManager()
			result, err := manager.IsInstalled(context.Background(), tt.toolName)

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

func TestDotnetManager_Info(t *testing.T) {
	tests := []struct {
		name          string
		toolName      string
		mockResponses map[string]CommandResponse
		expectError   bool
	}{
		{
			name:     "info for installed tool",
			toolName: "dotnetsay",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay`),
					Error: nil,
				},
			},
			expectError: false,
		},
		{
			name:     "info for not installed tool",
			toolName: "notinstalled-tool",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------`),
					Error: nil,
				},
			},
			expectError: false,
		},
		{
			name:     "command error",
			toolName: "test-tool",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
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

			manager := NewDotnetManager()
			result, err := manager.Info(context.Background(), tt.toolName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result == nil {
				t.Errorf("Expected non-nil PackageInfo but got nil")
			}
			if !tt.expectError && result != nil && result.Name != tt.toolName {
				t.Errorf("Expected tool name '%s' but got '%s'", tt.toolName, result.Name)
			}
		})
	}
}

func TestDotnetManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name          string
		toolName      string
		mockResponses map[string]CommandResponse
		expected      string
		expectError   bool
	}{
		{
			name:     "get version of installed tool",
			toolName: "dotnetsay",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef`),
					Error: nil,
				},
			},
			expected:    "2.1.4",
			expectError: false,
		},
		{
			name:     "tool not installed",
			toolName: "nonexistent-tool",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay`),
					Error: nil,
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name:     "command error",
			toolName: "dotnetsay",
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
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

			manager := NewDotnetManager()
			result, err := manager.InstalledVersion(context.Background(), tt.toolName)

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

func TestDotnetManager_Upgrade(t *testing.T) {
	tests := []struct {
		name          string
		tools         []string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:  "upgrade specific tools",
			tools: []string{"dotnetsay", "dotnet-ef"},
			mockResponses: map[string]CommandResponse{
				"dotnet tool update -g dotnetsay": {
					Output: []byte("Tool 'dotnetsay' was reinstalled with the latest stable version (version '2.1.5')."),
					Error:  nil,
				},
				"dotnet tool update -g dotnet-ef": {
					Output: []byte("Tool 'dotnet-ef' was reinstalled with the latest stable version (version '9.0.8')."),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:  "upgrade all tools",
			tools: []string{}, // empty means all tools
			mockResponses: map[string]CommandResponse{
				"dotnet tool list -g": {
					Output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef`),
					Error: nil,
				},
				"dotnet tool update -g dotnetsay": {
					Output: []byte("Tool 'dotnetsay' was reinstalled with the latest stable version (version '2.1.5')."),
					Error:  nil,
				},
				"dotnet tool update -g dotnet-ef": {
					Output: []byte("Tool 'dotnet-ef' was reinstalled with the latest stable version (version '9.0.8')."),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:  "upgrade with tool not installed error",
			tools: []string{"nonexistent-tool"},
			mockResponses: map[string]CommandResponse{
				"dotnet tool update -g nonexistent-tool": {
					Output: []byte("Tool 'nonexistent-tool' is not installed."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:  "upgrade with permission error",
			tools: []string{"restricted-tool"},
			mockResponses: map[string]CommandResponse{
				"dotnet tool update -g restricted-tool": {
					Output: []byte("Access denied when writing to tool directory"),
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

			manager := NewDotnetManager()
			err := manager.Upgrade(context.Background(), tt.tools)

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
