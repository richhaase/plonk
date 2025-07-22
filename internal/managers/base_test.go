// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"testing"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestBaseManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name       string
		config     ManagerConfig
		mockSetup  func(m *mocks.MockCommandExecutor)
		wantAvail  bool
		wantErr    bool
		wantBinary string
	}{
		{
			name: "primary binary available and functional",
			config: ManagerConfig{
				BinaryName:  "npm",
				VersionArgs: []string{"--version"},
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("npm").Return("/usr/bin/npm", nil)
				m.EXPECT().Execute(gomock.Any(), "npm", "--version").
					Return([]byte("8.19.2"), nil)
			},
			wantAvail:  true,
			wantErr:    false,
			wantBinary: "npm",
		},
		{
			name: "primary binary not found, fallback works",
			config: ManagerConfig{
				BinaryName:       "pip",
				FallbackBinaries: []string{"pip3"},
				VersionArgs:      []string{"--version"},
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("", fmt.Errorf("not found"))
				m.EXPECT().LookPath("pip3").Return("/usr/bin/pip3", nil)
				m.EXPECT().Execute(gomock.Any(), "pip3", "--version").
					Return([]byte("pip 21.0"), nil)
			},
			wantAvail:  true,
			wantErr:    false,
			wantBinary: "pip3",
		},
		{
			name: "binary exists but not functional",
			config: ManagerConfig{
				BinaryName:  "npm",
				VersionArgs: []string{"--version"},
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("npm").Return("/usr/bin/npm", nil)
				m.EXPECT().Execute(gomock.Any(), "npm", "--version").
					Return(nil, fmt.Errorf("command failed"))
			},
			wantAvail:  false,
			wantErr:    false,
			wantBinary: "",
		},
		{
			name: "no binaries found",
			config: ManagerConfig{
				BinaryName:       "pip",
				FallbackBinaries: []string{"pip3", "python3-pip"},
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("", fmt.Errorf("not found"))
				m.EXPECT().LookPath("pip3").Return("", fmt.Errorf("not found"))
				m.EXPECT().LookPath("python3-pip").Return("", fmt.Errorf("not found"))
			},
			wantAvail:  false,
			wantErr:    false,
			wantBinary: "",
		},
		{
			name: "context canceled during verification",
			config: ManagerConfig{
				BinaryName:  "npm",
				VersionArgs: []string{"--version"},
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("npm").Return("/usr/bin/npm", nil)
				m.EXPECT().Execute(gomock.Any(), "npm", "--version").
					Return(nil, context.Canceled)
			},
			wantAvail: false,
			wantErr:   true,
		},
		{
			name: "custom version args",
			config: ManagerConfig{
				BinaryName:  "go",
				VersionArgs: []string{"version"},
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("go").Return("/usr/local/go/bin/go", nil)
				m.EXPECT().Execute(gomock.Any(), "go", "version").
					Return([]byte("go version go1.19.1 darwin/amd64"), nil)
			},
			wantAvail:  true,
			wantErr:    false,
			wantBinary: "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			base := NewBaseManagerWithExecutor(tt.config, mockExecutor)
			ctx := context.Background()

			avail, err := base.IsAvailable(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsAvailable() error = %v, wantErr %v", err, tt.wantErr)
			}
			if avail != tt.wantAvail {
				t.Errorf("IsAvailable() = %v, want %v", avail, tt.wantAvail)
			}
			if tt.wantBinary != "" && base.GetBinary() != tt.wantBinary {
				t.Errorf("GetBinary() = %v, want %v", base.GetBinary(), tt.wantBinary)
			}
		})
	}
}

func TestBaseManager_ExecuteList(t *testing.T) {
	tests := []struct {
		name        string
		config      ManagerConfig
		mockSetup   func(m *mocks.MockCommandExecutor)
		wantOutput  []byte
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name: "successful list",
			config: ManagerConfig{
				BinaryName: "npm",
				ListArgs:   func() []string { return []string{"list", "-g", "--depth=0"} },
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "--depth=0").
					Return([]byte("npm packages list"), nil)
			},
			wantOutput: []byte("npm packages list"),
			wantErr:    false,
		},
		{
			name: "list with JSON flag",
			config: ManagerConfig{
				BinaryName: "npm",
				ListArgs:   func() []string { return []string{"list"} },
				PreferJSON: true,
				JSONFlag:   "--json",
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "--json").
					Return([]byte(`{"packages": []}`), nil)
			},
			wantOutput: []byte(`{"packages": []}`),
			wantErr:    false,
		},
		{
			name: "list command not configured",
			config: ManagerConfig{
				BinaryName: "npm",
				ListArgs:   nil,
			},
			mockSetup:   func(m *mocks.MockCommandExecutor) {},
			wantErr:     true,
			wantErrCode: errors.ErrCommandExecution,
		},
		{
			name: "list command fails",
			config: ManagerConfig{
				BinaryName: "npm",
				ListArgs:   func() []string { return []string{"list"} },
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list").
					Return(nil, fmt.Errorf("command failed"))
			},
			wantErr:     true,
			wantErrCode: errors.ErrCommandExecution,
		},
		{
			name: "uses cached binary",
			config: ManagerConfig{
				BinaryName:       "pip",
				FallbackBinaries: []string{"pip3"},
				ListArgs:         func() []string { return []string{"list"} },
			},
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// This simulates that pip3 was cached during IsAvailable
				m.EXPECT().Execute(gomock.Any(), "pip3", "list").
					Return([]byte("package list"), nil)
			},
			wantOutput: []byte("package list"),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			base := NewBaseManagerWithExecutor(tt.config, mockExecutor)
			// Simulate cached binary for the last test
			if tt.name == "uses cached binary" {
				base.binaryCache = "pip3"
			}

			ctx := context.Background()
			output, err := base.ExecuteList(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if plonkErr, ok := err.(*errors.PlonkError); ok {
					if plonkErr.Code != tt.wantErrCode {
						t.Errorf("ExecuteList() error code = %v, want %v", plonkErr.Code, tt.wantErrCode)
					}
				}
			}
			if !tt.wantErr && string(output) != string(tt.wantOutput) {
				t.Errorf("ExecuteList() output = %v, want %v", string(output), string(tt.wantOutput))
			}
		})
	}
}

func TestBaseManager_ExecuteInstall(t *testing.T) {
	tests := []struct {
		name        string
		config      ManagerConfig
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name: "successful install",
			config: ManagerConfig{
				BinaryName:  "npm",
				InstallArgs: func(pkg string) []string { return []string{"install", "-g", pkg} },
			},
			packageName: "typescript",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "npm", "install", "-g", "typescript").
					Return([]byte("added 1 package"), nil)
			},
			wantErr: false,
		},
		{
			name: "package not found",
			config: ManagerConfig{
				BinaryName:  "npm",
				InstallArgs: func(pkg string) []string { return []string{"install", "-g", pkg} },
			},
			packageName: "nonexistent-package",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "npm", "install", "-g", "nonexistent-package").
					Return([]byte("npm ERR! 404 Not Found - GET https://registry.npmjs.org/nonexistent-package - Not found"),
						&mockExitError{code: 1})
			},
			wantErr:     true,
			wantErrCode: errors.ErrPackageNotFound,
		},
		{
			name: "already installed",
			config: ManagerConfig{
				BinaryName:  "apt",
				InstallArgs: func(pkg string) []string { return []string{"install", "-y", pkg} },
			},
			packageName: "curl",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "install", "-y", "curl").
					Return([]byte("curl is already the newest version (7.68.0-1ubuntu2.7)."),
						&mockExitError{code: 0})
			},
			wantErr: false,
		},
		{
			name: "permission denied",
			config: ManagerConfig{
				BinaryName:  "apt",
				InstallArgs: func(pkg string) []string { return []string{"install", "-y", pkg} },
			},
			packageName: "vim",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "install", "-y", "vim").
					Return([]byte("E: Could not open lock file /var/lib/dpkg/lock-frontend - open (13: Permission denied)"),
						&mockExitError{code: 100})
			},
			wantErr:     true,
			wantErrCode: errors.ErrFilePermission,
		},
		{
			name: "database locked",
			config: ManagerConfig{
				BinaryName:  "apt",
				InstallArgs: func(pkg string) []string { return []string{"install", "-y", pkg} },
			},
			packageName: "git",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "install", "-y", "git").
					Return([]byte("E: Could not get lock /var/lib/dpkg/lock-frontend. It is held by process 1234"),
						&mockExitError{code: 100})
			},
			wantErr:     true,
			wantErrCode: errors.ErrCommandExecution,
		},
		{
			name: "install command not configured",
			config: ManagerConfig{
				BinaryName:  "npm",
				InstallArgs: nil,
			},
			packageName: "typescript",
			mockSetup:   func(m *mocks.MockCommandExecutor) {},
			wantErr:     true,
			wantErrCode: errors.ErrCommandExecution,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			base := NewBaseManagerWithExecutor(tt.config, mockExecutor)
			ctx := context.Background()

			err := base.ExecuteInstall(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteInstall() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if plonkErr, ok := err.(*errors.PlonkError); ok {
					if plonkErr.Code != tt.wantErrCode {
						t.Errorf("ExecuteInstall() error code = %v, want %v", plonkErr.Code, tt.wantErrCode)
					}
				}
			}
		})
	}
}

func TestBaseManager_ExecuteUninstall(t *testing.T) {
	tests := []struct {
		name        string
		config      ManagerConfig
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name: "successful uninstall",
			config: ManagerConfig{
				BinaryName:    "npm",
				UninstallArgs: func(pkg string) []string { return []string{"uninstall", "-g", pkg} },
			},
			packageName: "typescript",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "npm", "uninstall", "-g", "typescript").
					Return([]byte("removed 1 package"), nil)
			},
			wantErr: false,
		},
		{
			name: "package not installed",
			config: ManagerConfig{
				BinaryName:    "pip",
				UninstallArgs: func(pkg string) []string { return []string{"uninstall", "-y", pkg} },
			},
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "uninstall", "-y", "requests").
					Return([]byte("WARNING: Skipping requests as it is not installed."),
						&mockExitError{code: 0})
			},
			wantErr: false,
		},
		{
			name: "permission denied",
			config: ManagerConfig{
				BinaryName:    "apt",
				UninstallArgs: func(pkg string) []string { return []string{"remove", "-y", pkg} },
			},
			packageName: "vim",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "apt", "remove", "-y", "vim").
					Return([]byte("E: Could not open lock file /var/lib/dpkg/lock-frontend - open (13: Permission denied)"),
						&mockExitError{code: 100})
			},
			wantErr:     true,
			wantErrCode: errors.ErrFilePermission,
		},
		{
			name: "uninstall command not configured",
			config: ManagerConfig{
				BinaryName:    "npm",
				UninstallArgs: nil,
			},
			packageName: "typescript",
			mockSetup:   func(m *mocks.MockCommandExecutor) {},
			wantErr:     true,
			wantErrCode: errors.ErrCommandExecution,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			base := NewBaseManagerWithExecutor(tt.config, mockExecutor)
			ctx := context.Background()

			err := base.ExecuteUninstall(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteUninstall() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if plonkErr, ok := err.(*errors.PlonkError); ok {
					if plonkErr.Code != tt.wantErrCode {
						t.Errorf("ExecuteUninstall() error code = %v, want %v", plonkErr.Code, tt.wantErrCode)
					}
				}
			}
		})
	}
}

// mockExitError implements the ExitCode() method for testing
type mockExitError struct {
	code int
}

func (e *mockExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func (e *mockExitError) ExitCode() int {
	return e.code
}
