// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
	"testing"
)

func TestParseModulePath(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		wantModule  string
		wantVersion string
	}{
		{
			name:        "simple module path",
			pkg:         "github.com/user/repo",
			wantModule:  "github.com/user/repo",
			wantVersion: "latest",
		},
		{
			name:        "module path with version",
			pkg:         "github.com/user/repo@v1.2.3",
			wantModule:  "github.com/user/repo",
			wantVersion: "v1.2.3",
		},
		{
			name:        "module path with latest",
			pkg:         "github.com/user/repo@latest",
			wantModule:  "github.com/user/repo",
			wantVersion: "latest",
		},
		{
			name:        "module path with commit hash",
			pkg:         "github.com/user/repo@abc123",
			wantModule:  "github.com/user/repo",
			wantVersion: "abc123",
		},
		{
			name:        "module path with subpath",
			pkg:         "github.com/user/repo/cmd/tool@v1.0.0",
			wantModule:  "github.com/user/repo/cmd/tool",
			wantVersion: "v1.0.0",
		},
		{
			name:        "complex module path",
			pkg:         "go.uber.org/zap@v1.24.0",
			wantModule:  "go.uber.org/zap",
			wantVersion: "v1.24.0",
		},
		{
			name:        "module without version",
			pkg:         "golang.org/x/tools/cmd/goimports",
			wantModule:  "golang.org/x/tools/cmd/goimports",
			wantVersion: "latest",
		},
		{
			name:        "empty package",
			pkg:         "",
			wantModule:  "",
			wantVersion: "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModule, gotVersion := parseModulePath(tt.pkg)
			if gotModule != tt.wantModule {
				t.Errorf("parseModulePath() module = %v, want %v", gotModule, tt.wantModule)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("parseModulePath() version = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestExtractBinaryName(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		want       string
	}{
		{
			name:       "simple module",
			modulePath: "github.com/user/tool",
			want:       "tool",
		},
		{
			name:       "module with cmd pattern",
			modulePath: "github.com/user/project/cmd/tool",
			want:       "tool",
		},
		{
			name:       "module with version",
			modulePath: "github.com/user/tool@v1.2.3",
			want:       "tool",
		},
		{
			name:       "golang.org/x tools",
			modulePath: "golang.org/x/tools/cmd/goimports",
			want:       "goimports",
		},
		{
			name:       "single path component",
			modulePath: "tool",
			want:       "tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBinaryName(tt.modulePath)
			if got != tt.want {
				t.Errorf("extractBinaryName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractBinaryNameFromPath(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		want       string
	}{
		{
			name:       "simple module",
			modulePath: "github.com/user/tool",
			want:       "tool",
		},
		{
			name:       "module with cmd pattern",
			modulePath: "github.com/user/project/cmd/tool",
			want:       "tool",
		},
		{
			name:       "module with version",
			modulePath: "github.com/user/tool@v1.2.3",
			want:       "tool",
		},
		{
			name:       "golang.org/x tools",
			modulePath: "golang.org/x/tools/cmd/goimports",
			want:       "goimports",
		},
		{
			name:       "single path component",
			modulePath: "tool",
			want:       "tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractBinaryNameFromPath(tt.modulePath)
			if got != tt.want {
				t.Errorf("ExtractBinaryNameFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoInstallManager_Configuration(t *testing.T) {
	manager := NewGoInstallManager()

	if manager.binary != "go" {
		t.Errorf("binary = %v, want go", manager.binary)
	}
}

func TestGoInstallManager_SupportsSearch(t *testing.T) {
	manager := NewGoInstallManager()
	if manager.SupportsSearch() {
		t.Error("GoInstallManager should not support search")
	}
}

func TestGoInstallManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name: "go is available",
			mockResponses: map[string]CommandResponse{
				"go version": {
					Output: []byte("go version go1.21.5 darwin/arm64"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:          "go not found",
			mockResponses: map[string]CommandResponse{
				// Empty responses means LookPath will fail
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "go exists but not functional",
			mockResponses: map[string]CommandResponse{
				"go": {}, // Makes LookPath succeed
				"go version": {
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

			manager := NewGoInstallManager()
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

func TestGoInstallManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "github.com/golangci/golangci-lint/cmd/golangci-lint",
			mockResponses: map[string]CommandResponse{
				"go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest": {
					Output: []byte(""),
					Error:  nil,
				},
				"go env GOBIN": {
					Output: []byte("/home/user/go/bin"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "successful install with version",
			packageName: "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.0",
			mockResponses: map[string]CommandResponse{
				"go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.0": {
					Output: []byte(""),
					Error:  nil,
				},
				"go env GOBIN": {
					Output: []byte("/home/user/go/bin"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "github.com/nonexistent/package",
			mockResponses: map[string]CommandResponse{
				"go install github.com/nonexistent/package@latest": {
					Output: []byte("go: github.com/nonexistent/package@latest: module github.com/nonexistent/package: cannot find module providing package github.com/nonexistent/package"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "build error",
			packageName: "github.com/broken/package",
			mockResponses: map[string]CommandResponse{
				"go install github.com/broken/package@latest": {
					Output: []byte("# github.com/broken/package\n./main.go:10:5: undefined: SomeFunction"),
					Error:  &MockExitError{Code: 2},
				},
			},
			expectError:   true,
			errorContains: "package installation failed",
		},
		{
			name:        "network error",
			packageName: "github.com/test/package",
			mockResponses: map[string]CommandResponse{
				"go install github.com/test/package@latest": {
					Output: []byte("go: downloading github.com/test/package v1.0.0\ngo: github.com/test/package@v1.0.0: Get \"https://proxy.golang.org/github.com/test/package/@v/v1.0.0.mod\": dial tcp: lookup proxy.golang.org: no such host"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "package installation failed",
		},
		{
			name:        "incompatible go version",
			packageName: "github.com/modern/package",
			mockResponses: map[string]CommandResponse{
				"go install github.com/modern/package@latest": {
					Output: []byte("go: github.com/modern/package@latest requires go >= 1.21 (running go 1.19.5)"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "dependency conflict",
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

			manager := NewGoInstallManager()
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
		})
	}
}

func TestGoInstallManager_Search(t *testing.T) {
	manager := NewGoInstallManager()
	_, err := manager.Search(context.Background(), "test")

	if err == nil {
		t.Error("Expected error for Search method")
	}

	expectedError := "go does not have a built-in search command"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s' but got: %v", expectedError, err)
	}
}

func TestGoInstallManager_getGoBinDir(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		want          string
		expectError   bool
	}{
		{
			name: "GOBIN is set",
			mockResponses: map[string]CommandResponse{
				"go env GOBIN": {
					Output: []byte("/custom/go/bin\n"),
					Error:  nil,
				},
			},
			want:        "/custom/go/bin",
			expectError: false,
		},
		{
			name: "GOBIN not set, use GOPATH",
			mockResponses: map[string]CommandResponse{
				"go env GOBIN": {
					Output: []byte("\n"),
					Error:  nil,
				},
				"go env GOPATH": {
					Output: []byte("/home/user/go\n"),
					Error:  nil,
				},
			},
			want:        "/home/user/go/bin",
			expectError: false,
		},
		{
			name: "GOBIN command fails",
			mockResponses: map[string]CommandResponse{
				"go env GOBIN": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			want:        "",
			expectError: true,
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

			manager := NewGoInstallManager()
			got, err := manager.getGoBinDir(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && got != tt.want {
				t.Errorf("getGoBinDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoInstallManager_isGoBinary(t *testing.T) {
	tests := []struct {
		name          string
		binaryPath    string
		mockResponses map[string]CommandResponse
		want          bool
	}{
		{
			name:       "valid go binary",
			binaryPath: "/home/user/go/bin/golangci-lint",
			mockResponses: map[string]CommandResponse{
				"go version -m /home/user/go/bin/golangci-lint": {
					Output: []byte("/home/user/go/bin/golangci-lint: go1.21.5\n\tmod\tgithub.com/golangci/golangci-lint\tv1.50.0"),
					Error:  nil,
				},
			},
			want: true,
		},
		{
			name:       "not a go binary",
			binaryPath: "/usr/bin/ls",
			mockResponses: map[string]CommandResponse{
				"go version -m /usr/bin/ls": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			want: false,
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

			manager := NewGoInstallManager()
			got := manager.isGoBinary(context.Background(), tt.binaryPath)

			if got != tt.want {
				t.Errorf("isGoBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGoInstallManager_InstalledVersion is omitted because it depends on os.Stat
// to check if binary files exist on disk, which cannot be mocked with our
// Command Executor pattern. This limitation is documented in COMMAND_EXECUTOR_INTERFACE_PLAN.md

// TestGoInstallManager_Info is omitted because it depends on os.Stat
// to check if binary files exist on disk, which cannot be mocked with our
// Command Executor pattern. This limitation is documented in COMMAND_EXECUTOR_INTERFACE_PLAN.md
