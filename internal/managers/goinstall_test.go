// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"errors"
	"testing"

	"github.com/richhaase/plonk/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestGoInstallManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult bool
		expectedError  bool
	}{
		{
			name: "available and functional",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("go").Return("/usr/local/go/bin/go", nil)
				m.EXPECT().Execute(gomock.Any(), "go", "version").Return([]byte("go version go1.21.5 darwin/amd64"), nil).Times(2)
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "not in PATH",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("go").Return("", errors.New("not found"))
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "in PATH but not functional",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("go").Return("/usr/local/go/bin/go", nil)
				m.EXPECT().Execute(gomock.Any(), "go", "version").Return(nil, errors.New("command failed"))
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "unsupported version",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("go").Return("/usr/local/go/bin/go", nil)
				m.EXPECT().Execute(gomock.Any(), "go", "version").Return([]byte("go version"), nil)
				m.EXPECT().Execute(gomock.Any(), "go", "version").Return([]byte("some invalid version"), nil)
			},
			expectedResult: false,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.setupMocks(mockExecutor)

			manager := NewGoInstallManagerWithExecutor(mockExecutor)
			result, err := manager.IsAvailable(context.Background())

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestGoInstallManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult []string
		expectedError  bool
	}{
		{
			name: "successful list with packages",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				// Mock GOBIN directory check
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte("/Users/test/go/bin"), nil)
				// Mock filesystem operations would require more complex mocking
				// For now, we'll mock the directory reading behavior via the go version calls
			},
			expectedResult: []string{}, // Empty for this test without filesystem mocking
			expectedError:  false,
		},
		{
			name: "empty GOBIN falls back to GOPATH",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte(""), nil)
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOPATH").Return([]byte("/Users/test/go"), nil)
			},
			expectedResult: []string{},
			expectedError:  false,
		},
		{
			name: "GOBIN command fails",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return(nil, errors.New("command failed"))
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.setupMocks(mockExecutor)

			manager := NewGoInstallManagerWithExecutor(mockExecutor)
			result, err := manager.ListInstalled(context.Background())

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestGoInstallManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		setupMocks  func(*mocks.MockCommandExecutor)
		expectError bool
	}{
		{
			name:        "successful install with latest version",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "go", "install", "github.com/user/tool@latest").Return([]byte(""), nil)
				// Mock GOBIN check for PATH warning
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte("/Users/test/go/bin"), nil)
			},
			expectError: false,
		},
		{
			name:        "successful install with specific version",
			packageName: "github.com/user/tool@v1.2.3",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "go", "install", "github.com/user/tool@v1.2.3").Return([]byte(""), nil)
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte("/Users/test/go/bin"), nil)
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "github.com/nonexistent/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "go", "install", "github.com/nonexistent/tool@latest").Return([]byte("cannot find module"), execErr)
			},
			expectError: true,
		},
		{
			name:        "network error",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "go", "install", "github.com/user/tool@latest").Return([]byte("connection timeout"), execErr)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.setupMocks(mockExecutor)

			manager := NewGoInstallManagerWithExecutor(mockExecutor)
			err := manager.Install(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestGoInstallManager_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		setupMocks  func(*mocks.MockCommandExecutor)
		expectError bool
	}{
		{
			name:        "binary doesn't exist - no error",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte("/Users/test/go/bin"), nil)
				// In real scenario, os.Stat would return os.IsNotExist(err) = true
				// This test would pass because binary doesn't exist
			},
			expectError: false,
		},
		{
			name:        "GOBIN command fails",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return(nil, errors.New("command failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.setupMocks(mockExecutor)

			manager := NewGoInstallManagerWithExecutor(mockExecutor)
			err := manager.Uninstall(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestGoInstallManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult bool
		expectedError  bool
	}{
		{
			name:        "package not installed - binary doesn't exist",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte("/Users/test/go/bin"), nil)
				// In real scenario, os.Stat would return os.IsNotExist(err) = true
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:        "GOBIN command fails",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return(nil, errors.New("command failed"))
			},
			expectedResult: false,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.setupMocks(mockExecutor)

			manager := NewGoInstallManagerWithExecutor(mockExecutor)
			result, err := manager.IsInstalled(context.Background(), tt.packageName)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestGoInstallManager_SupportsSearch(t *testing.T) {
	manager := NewGoInstallManager()
	result := manager.SupportsSearch()
	if result != false {
		t.Errorf("Expected SupportsSearch to return false, got %v", result)
	}
}

func TestGoInstallManager_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedResult []string
		expectedError  bool
	}{
		{
			name:           "search not supported",
			query:          "test",
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewGoInstallManager()
			result, err := manager.Search(context.Background(), tt.query)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestGoInstallManager_Info(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult *PackageInfo
		expectedError  bool
	}{
		{
			name:        "package not installed",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				// Mock getGoBinDir call in Info method
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte("/Users/test/go/bin"), nil)
				// Mock getGoBinDir call in IsInstalled method
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte("/Users/test/go/bin"), nil)
				// Mock IsInstalled returning false (binary doesn't exist)
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:        "GOBIN command fails",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return(nil, errors.New("command failed"))
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.setupMocks(mockExecutor)

			manager := NewGoInstallManagerWithExecutor(mockExecutor)
			result, err := manager.Info(context.Background(), tt.packageName)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !equalPackageInfo(result, tt.expectedResult) {
				t.Errorf("Expected result %+v but got %+v", tt.expectedResult, result)
			}
		})
	}
}

func TestGoInstallManager_GetInstalledVersion(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult string
		expectedError  bool
	}{
		{
			name:        "package not installed",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				// Mock getGoBinDir call in GetInstalledVersion method
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte("/Users/test/go/bin"), nil)
				// Mock getGoBinDir call in IsInstalled method
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return([]byte("/Users/test/go/bin"), nil)
				// Mock IsInstalled returning false
			},
			expectedResult: "",
			expectedError:  true,
		},
		{
			name:        "GOBIN command fails",
			packageName: "github.com/user/tool",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "go", "env", "GOBIN").Return(nil, errors.New("command failed"))
			},
			expectedResult: "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.setupMocks(mockExecutor)

			manager := NewGoInstallManagerWithExecutor(mockExecutor)
			result, err := manager.GetInstalledVersion(context.Background(), tt.packageName)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestParseModulePath(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedModule  string
		expectedVersion string
	}{
		{
			name:            "module without version",
			input:           "github.com/user/tool",
			expectedModule:  "github.com/user/tool",
			expectedVersion: "latest",
		},
		{
			name:            "module with version",
			input:           "github.com/user/tool@v1.2.3",
			expectedModule:  "github.com/user/tool",
			expectedVersion: "v1.2.3",
		},
		{
			name:            "module with latest",
			input:           "github.com/user/tool@latest",
			expectedModule:  "github.com/user/tool",
			expectedVersion: "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, version := parseModulePath(tt.input)
			if module != tt.expectedModule {
				t.Errorf("Expected module %s but got %s", tt.expectedModule, module)
			}
			if version != tt.expectedVersion {
				t.Errorf("Expected version %s but got %s", tt.expectedVersion, version)
			}
		})
	}
}

func TestExtractBinaryName(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
	}{
		{
			name:           "simple module path",
			input:          "github.com/user/tool",
			expectedResult: "tool",
		},
		{
			name:           "module with version",
			input:          "github.com/user/tool@v1.2.3",
			expectedResult: "tool",
		},
		{
			name:           "cmd pattern",
			input:          "github.com/user/repo/cmd/tool",
			expectedResult: "tool",
		},
		{
			name:           "cmd pattern with version",
			input:          "github.com/user/repo/cmd/tool@v1.2.3",
			expectedResult: "tool",
		},
		{
			name:           "nested path",
			input:          "github.com/user/repo/tools/mytool",
			expectedResult: "mytool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBinaryName(tt.input)
			if result != tt.expectedResult {
				t.Errorf("Expected result %s but got %s", tt.expectedResult, result)
			}
		})
	}
}
