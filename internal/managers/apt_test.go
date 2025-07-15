// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/richhaase/plonk/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestAptManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult bool
		expectedError  bool
	}{
		{
			name: "available and functional on Linux",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				if runtime.GOOS == "linux" {
					m.EXPECT().LookPath("apt").Return("/usr/bin/apt", nil)
					m.EXPECT().Execute(gomock.Any(), "apt", "--version").Return([]byte("apt 2.4.8"), nil)
				}
			},
			expectedResult: runtime.GOOS == "linux",
			expectedError:  false,
		},
		{
			name: "not available on non-Linux",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				// No mocks needed for non-Linux
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "not in PATH on Linux",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				if runtime.GOOS == "linux" {
					m.EXPECT().LookPath("apt").Return("", errors.New("not found"))
				}
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

			manager := NewAptManagerWithExecutor(mockExecutor)
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

func TestAptManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult []string
		expectedError  bool
	}{
		{
			name: "successful list with packages",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				output := `Listing... Done
curl/stable,now 7.68.0-1ubuntu2.18 amd64 [installed]
git/stable,now 1:2.25.1-1ubuntu3.10 amd64 [installed]
vim/stable,now 2:8.1.2269-1ubuntu5.15 amd64 [installed]`
				m.EXPECT().Execute(gomock.Any(), "apt", "list", "--installed").Return([]byte(output), nil)
			},
			expectedResult: []string{"curl", "git", "vim"},
			expectedError:  false,
		},
		{
			name: "empty list",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "apt", "list", "--installed").Return([]byte("Listing... Done"), nil)
			},
			expectedResult: []string{},
			expectedError:  false,
		},
		{
			name: "command fails",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "apt", "list", "--installed").Return(nil, errors.New("command failed"))
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

			manager := NewAptManagerWithExecutor(mockExecutor)
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

func TestAptManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		setupMocks  func(*mocks.MockCommandExecutor)
		expectError bool
	}{
		{
			name:        "successful install",
			packageName: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "install", "-y", "curl").Return([]byte("Setting up curl"), nil)
			},
			expectError: false,
		},
		{
			name:        "package already installed",
			packageName: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 0}
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "install", "-y", "curl").Return([]byte("curl is already the newest version"), execErr)
			},
			expectError: false, // already installed should not be an error
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 100}
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "install", "-y", "nonexistent").Return([]byte("Unable to locate package nonexistent"), execErr)
			},
			expectError: true,
		},
		{
			name:        "permission denied",
			packageName: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "install", "-y", "curl").Return([]byte("Permission denied"), execErr)
			},
			expectError: true,
		},
		{
			name:        "database locked",
			packageName: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "install", "-y", "curl").Return([]byte("Could not get lock /var/lib/dpkg/lock"), execErr)
			},
			expectError: true,
		},
		{
			name:        "broken dependencies",
			packageName: "broken-package",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "install", "-y", "broken-package").Return([]byte("Depends: libfoo but it is not installable"), execErr)
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

			manager := NewAptManagerWithExecutor(mockExecutor)
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

func TestAptManager_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		setupMocks  func(*mocks.MockCommandExecutor)
		expectError bool
	}{
		{
			name:        "successful uninstall",
			packageName: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "remove", "-y", "curl").Return([]byte("Removing curl"), nil)
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 0}
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "remove", "-y", "nonexistent").Return([]byte("nonexistent is not installed"), execErr)
			},
			expectError: false, // not installed should not be an error for uninstall
		},
		{
			name:        "dependency conflict",
			packageName: "libssl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 1}
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "remove", "-y", "libssl").Return([]byte("Broken packages"), execErr)
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

			manager := NewAptManagerWithExecutor(mockExecutor)
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

func TestAptManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult bool
		expectedError  bool
	}{
		{
			name:        "package is installed",
			packageName: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				output := `curl/stable,now 7.68.0-1ubuntu2.18 amd64 [installed]
git/stable,now 1:2.25.1-1ubuntu3.10 amd64 [installed]`
				m.EXPECT().Execute(gomock.Any(), "apt", "list", "--installed").Return([]byte(output), nil)
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:        "package is not installed",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				output := `curl/stable,now 7.68.0-1ubuntu2.18 amd64 [installed]
git/stable,now 1:2.25.1-1ubuntu3.10 amd64 [installed]`
				m.EXPECT().Execute(gomock.Any(), "apt", "list", "--installed").Return([]byte(output), nil)
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:        "list command fails",
			packageName: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "apt", "list", "--installed").Return(nil, errors.New("command failed"))
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

			manager := NewAptManagerWithExecutor(mockExecutor)
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

func TestAptManager_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult []string
		expectedError  bool
	}{
		{
			name:  "successful search with results",
			query: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				output := `curl/stable 7.68.0-1ubuntu2.18 amd64
  command line tool for transferring data with URL syntax

curlftpfs/stable 0.9.2-9build1 amd64
  filesystem to access FTP hosts based on FUSE and cURL`
				m.EXPECT().Execute(gomock.Any(), "apt", "search", "curl").Return([]byte(output), nil)
			},
			expectedResult: []string{"curl", "curlftpfs"},
			expectedError:  false,
		},
		{
			name:  "search with no results",
			query: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "apt", "search", "nonexistent").Return([]byte(""), nil)
			},
			expectedResult: []string{},
			expectedError:  false,
		},
		{
			name:  "search command fails",
			query: "test",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "apt", "search", "test").Return(nil, errors.New("command failed"))
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

			manager := NewAptManagerWithExecutor(mockExecutor)
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

func TestAptManager_Info(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult *PackageInfo
		expectedError  bool
	}{
		{
			name:        "successful info for installed package",
			packageName: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				infoOutput := `Package: curl
Version: 7.68.0-1ubuntu2.18
Description: command line tool for transferring data with URL syntax
Homepage: https://curl.haxx.se/
Depends: libc6 (>= 2.17), libcurl4 (= 7.68.0-1ubuntu2.18)
Installed-Size: 411`
				m.EXPECT().Execute(gomock.Any(), "apt", "show", "curl").Return([]byte(infoOutput), nil)
				listOutput := `curl/stable,now 7.68.0-1ubuntu2.18 amd64 [installed]`
				m.EXPECT().Execute(gomock.Any(), "apt", "list", "--installed").Return([]byte(listOutput), nil)
			},
			expectedResult: &PackageInfo{
				Name:          "curl",
				Version:       "7.68.0-1ubuntu2.18",
				Description:   "command line tool for transferring data with URL syntax",
				Homepage:      "https://curl.haxx.se/",
				Dependencies:  []string{"libc6", "libcurl4"},
				InstalledSize: "411",
				Manager:       "apt",
				Installed:     true,
			},
			expectedError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				execErr := &mockExitError{code: 100}
				m.EXPECT().Execute(gomock.Any(), "apt", "show", "nonexistent").Return(nil, execErr)
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

			manager := NewAptManagerWithExecutor(mockExecutor)
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

func TestAptManager_GetInstalledVersion(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		setupMocks     func(*mocks.MockCommandExecutor)
		expectedResult string
		expectedError  bool
	}{
		{
			name:        "successful version retrieval",
			packageName: "curl",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				listOutput := `curl/stable,now 7.68.0-1ubuntu2.18 amd64 [installed]`
				m.EXPECT().Execute(gomock.Any(), "apt", "list", "--installed").Return([]byte(listOutput), nil)
				infoOutput := `Package: curl
Version: 7.68.0-1ubuntu2.18
Description: command line tool for transferring data with URL syntax`
				m.EXPECT().Execute(gomock.Any(), "apt", "show", "curl").Return([]byte(infoOutput), nil)
			},
			expectedResult: "7.68.0-1ubuntu2.18",
			expectedError:  false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			setupMocks: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "apt", "list", "--installed").Return([]byte(""), nil)
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

			manager := NewAptManagerWithExecutor(mockExecutor)
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

func TestAptManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         []byte
		expectedResult []string
	}{
		{
			name: "normal output",
			output: []byte(`Listing... Done
curl/stable,now 7.68.0-1ubuntu2.18 amd64 [installed]
git/stable,now 1:2.25.1-1ubuntu3.10 amd64 [installed]
vim/stable,now 2:8.1.2269-1ubuntu5.15 amd64 [installed]`),
			expectedResult: []string{"curl", "git", "vim"},
		},
		{
			name:           "empty output",
			output:         []byte("Listing... Done"),
			expectedResult: []string{},
		},
		{
			name: "output with warnings",
			output: []byte(`WARNING: apt does not have a stable CLI interface. Use with caution in scripts.
Listing... Done
curl/stable,now 7.68.0-1ubuntu2.18 amd64 [installed]`),
			expectedResult: []string{"curl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewAptManager()
			result := manager.parseListOutput(tt.output)

			if !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestAptManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         []byte
		expectedResult []string
	}{
		{
			name: "normal output",
			output: []byte(`curl/stable 7.68.0-1ubuntu2.18 amd64
  command line tool for transferring data with URL syntax

curlftpfs/stable 0.9.2-9build1 amd64
  filesystem to access FTP hosts based on FUSE and cURL`),
			expectedResult: []string{"curl", "curlftpfs"},
		},
		{
			name:           "no results",
			output:         []byte("No packages found"),
			expectedResult: []string{},
		},
		{
			name:           "empty output",
			output:         []byte(""),
			expectedResult: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewAptManager()
			result := manager.parseSearchOutput(tt.output)

			if !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestAptManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         []byte
		packageName    string
		expectedResult *PackageInfo
	}{
		{
			name: "normal output",
			output: []byte(`Package: curl
Version: 7.68.0-1ubuntu2.18
Description: command line tool for transferring data with URL syntax
Homepage: https://curl.haxx.se/
Depends: libc6 (>= 2.17), libcurl4 (= 7.68.0-1ubuntu2.18)
Installed-Size: 411`),
			packageName: "curl",
			expectedResult: &PackageInfo{
				Name:          "curl",
				Version:       "7.68.0-1ubuntu2.18",
				Description:   "command line tool for transferring data with URL syntax",
				Homepage:      "https://curl.haxx.se/",
				Dependencies:  []string{"libc6", "libcurl4"},
				InstalledSize: "411",
			},
		},
		{
			name: "minimal output",
			output: []byte(`Package: curl
Version: 7.68.0-1ubuntu2.18`),
			packageName: "curl",
			expectedResult: &PackageInfo{
				Name:    "curl",
				Version: "7.68.0-1ubuntu2.18",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewAptManager()
			result := manager.parseInfoOutput(tt.output, tt.packageName)

			if !equalPackageInfo(result, tt.expectedResult) {
				t.Errorf("Expected result %+v but got %+v", tt.expectedResult, result)
			}
		})
	}
}

func TestAptManager_extractVersion(t *testing.T) {
	tests := []struct {
		name           string
		output         []byte
		expectedResult string
	}{
		{
			name: "normal output",
			output: []byte(`Package: curl
Version: 7.68.0-1ubuntu2.18
Description: command line tool for transferring data with URL syntax`),
			expectedResult: "7.68.0-1ubuntu2.18",
		},
		{
			name: "no version",
			output: []byte(`Package: curl
Description: command line tool for transferring data with URL syntax`),
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewAptManager()
			result := manager.extractVersion(tt.output)

			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

// Helper functions (defined in base_test.go and homebrew_refactored_test.go)
