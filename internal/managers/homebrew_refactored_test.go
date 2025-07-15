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

func TestHomebrewManagerV2_IsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult bool
		expectedError  bool
	}{
		{
			name: "available and functional",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("brew").Return("/usr/local/bin/brew", nil)
				m.EXPECT().Execute(gomock.Any(), "brew", "--version").Return([]byte("Homebrew 3.6.0"), nil)
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "not in PATH",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("brew").Return("", errors.New("not found"))
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "in PATH but not functional",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("brew").Return("/usr/local/bin/brew", nil)
				m.EXPECT().Execute(gomock.Any(), "brew", "--version").Return(nil, errors.New("command failed"))
			},
			expectedResult: false,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.setupMocks(mockExecutor)

			manager := NewHomebrewManagerV2WithExecutor(mockExecutor)
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

func TestHomebrewManagerV2_ListInstalled(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult []string
		expectedError  bool
	}{
		{
			name: "successful list with packages",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "list").Return([]byte("git\nnode\npython@3.9"), nil)
			},
			expectedResult: []string{"git", "node", "python@3.9"},
			expectedError:  false,
		},
		{
			name: "empty list",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "list").Return([]byte(""), nil)
			},
			expectedResult: []string{},
			expectedError:  false,
		},
		{
			name: "command fails",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "list").Return(nil, errors.New("command failed"))
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

			manager := NewHomebrewManagerV2WithExecutor(mockExecutor)
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

func TestHomebrewManagerV2_Install(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		setupMocks  func(*mocks.MockCommandExecutor)
		expectError bool
	}{
		{
			name:        "successful install",
			packageName: "git",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "brew", "install", "git").Return([]byte("==> Installing git"), nil)
			},
			expectError: false,
		},
		{
			name:        "package already installed",
			packageName: "git",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "brew", "install", "git").Return([]byte("git is already installed"), execErr)
			},
			expectError: false, // already installed should not be an error
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "brew", "install", "nonexistent").Return([]byte("No available formula"), execErr)
			},
			expectError: true,
		},
		{
			name:        "dependency conflict",
			packageName: "conflicting-package",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "brew", "install", "conflicting-package").Return([]byte("because it is required by other"), execErr)
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

			manager := NewHomebrewManagerV2WithExecutor(mockExecutor)
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

func TestHomebrewManagerV2_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		setupMocks  func(*mocks.MockCommandExecutor)
		expectError bool
	}{
		{
			name:        "successful uninstall",
			packageName: "git",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "brew", "uninstall", "git").Return([]byte("==> Uninstalling git"), nil)
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "brew", "uninstall", "nonexistent").Return([]byte("No such keg"), execErr)
			},
			expectError: false, // not installed should not be an error for uninstall
		},
		{
			name:        "package has dependents",
			packageName: "python",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "brew", "uninstall", "python").Return([]byte("still has dependents"), execErr)
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

			manager := NewHomebrewManagerV2WithExecutor(mockExecutor)
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

func TestHomebrewManagerV2_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult bool
		expectedError  bool
	}{
		{
			name:        "package is installed",
			packageName: "git",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "list").Return([]byte("git\nnode\npython@3.9"), nil)
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:        "package is not installed",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "list").Return([]byte("git\nnode\npython@3.9"), nil)
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:        "list command fails",
			packageName: "git",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "list").Return(nil, errors.New("command failed"))
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

			manager := NewHomebrewManagerV2WithExecutor(mockExecutor)
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

func TestHomebrewManagerV2_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult []string
		expectedError  bool
	}{
		{
			name:  "successful search with results",
			query: "git",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "search", "git").Return([]byte("git\ngit-flow\ngithub-cli"), nil)
			},
			expectedResult: []string{"git", "git-flow", "github-cli"},
			expectedError:  false,
		},
		{
			name:  "search with no results",
			query: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "search", "nonexistent").Return([]byte("No formula found"), nil)
			},
			expectedResult: []string{},
			expectedError:  false,
		},
		{
			name:  "search command fails",
			query: "test",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "search", "test").Return(nil, errors.New("command failed"))
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

			manager := NewHomebrewManagerV2WithExecutor(mockExecutor)
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

func TestHomebrewManagerV2_Info(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult *PackageInfo
		expectedError  bool
	}{
		{
			name:        "successful info for installed package",
			packageName: "git",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "info", "git").Return([]byte("git: stable 2.37.1\nDistributed revision control system\nFrom: https://github.com/git/git"), nil)
				m.EXPECT().Execute(gomock.Any(), "brew", "list").Return([]byte("git\nnode"), nil)
			},
			expectedResult: &PackageInfo{
				Name:        "git",
				Version:     "2.37.1",
				Description: "Distributed revision control system",
				Homepage:    "https://github.com/git/git",
				Manager:     "homebrew",
				Installed:   true,
			},
			expectedError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().Execute(gomock.Any(), "brew", "info", "nonexistent").Return(nil, execErr)
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

			manager := NewHomebrewManagerV2WithExecutor(mockExecutor)
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

func TestHomebrewManagerV2_GetInstalledVersion(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult string
		expectedError  bool
	}{
		{
			name:        "successful version retrieval",
			packageName: "git",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "list").Return([]byte("git\nnode"), nil)
				m.EXPECT().Execute(gomock.Any(), "brew", "list", "--versions", "git").Return([]byte("git 2.37.1 2.36.0"), nil)
			},
			expectedResult: "2.37.1",
			expectedError:  false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "brew", "list").Return([]byte("git\nnode"), nil)
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

			manager := NewHomebrewManagerV2WithExecutor(mockExecutor)
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

func TestHomebrewManagerV2_parseListOutput(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewHomebrewManagerV2()
			result := manager.parseListOutput(tt.output)

			if !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManagerV2_parseSearchOutput(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewHomebrewManagerV2()
			result := manager.parseSearchOutput(tt.output)

			if !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManagerV2_parseInfoOutput(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewHomebrewManagerV2()
			result := manager.parseInfoOutput(tt.output, tt.packageName)

			if !equalPackageInfo(result, tt.expectedResult) {
				t.Errorf("Expected result %+v but got %+v", tt.expectedResult, result)
			}
		})
	}
}

func TestHomebrewManagerV2_extractVersion(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewHomebrewManagerV2()
			result := manager.extractVersion(tt.output, tt.packageName)

			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

// Helper functions (defined in base_test.go)
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
