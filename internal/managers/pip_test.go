// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"testing"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestPipManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(m *mocks.MockCommandExecutor)
		want      bool
		wantErr   bool
	}{
		{
			name: "pip available and functional",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil)
				m.EXPECT().Execute(gomock.Any(), "pip", "--version").Return([]byte("pip 21.0.1"), nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "pip3 available as fallback",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("", &ExitError{Code: 1})
				m.EXPECT().LookPath("pip3").Return("/usr/bin/pip3", nil)
				m.EXPECT().Execute(gomock.Any(), "pip3", "--version").Return([]byte("pip 21.0.1"), nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "neither pip nor pip3 available",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("", &ExitError{Code: 1})
				m.EXPECT().LookPath("pip3").Return("", &ExitError{Code: 1})
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "pip exists but not functional, pip3 not found",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil)
				m.EXPECT().Execute(gomock.Any(), "pip", "--version").Return(nil, &ExitError{Code: 1})
				m.EXPECT().LookPath("pip3").Return("", &ExitError{Code: 1})
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "pip not functional but pip3 works",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil)
				m.EXPECT().Execute(gomock.Any(), "pip", "--version").Return(nil, &ExitError{Code: 1})
				m.EXPECT().LookPath("pip3").Return("/usr/bin/pip3", nil)
				m.EXPECT().Execute(gomock.Any(), "pip3", "--version").Return([]byte("pip 21.0.1"), nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "context canceled during pip check",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil)
				m.EXPECT().Execute(gomock.Any(), "pip", "--version").Return(nil, context.Canceled)
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewPipManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.IsAvailable(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAvailable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(m *mocks.MockCommandExecutor)
		want      []string
		wantErr   bool
	}{
		{
			name: "successful JSON list",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return([]byte(`[{"name": "requests", "version": "2.28.0"}, {"name": "Django", "version": "4.0"}]`), nil)
			},
			want:    []string{"requests", "django"},
			wantErr: false,
		},
		{
			name: "empty package list",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return([]byte("[]"), nil)
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "fallback to plain text when --user not supported",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return([]byte("unknown option --user"), &ExitError{Code: 1})
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user").
					Return([]byte("unknown option --user"), &ExitError{Code: 1})
				m.EXPECT().Execute(gomock.Any(), "pip", "list").
					Return([]byte("Package    Version\n----------  -------\nrequests   2.28.0\ndjango     4.0"), nil)
			},
			want:    []string{"requests", "django"},
			wantErr: false,
		},
		{
			name: "command execution error",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return(nil, &ExitError{Code: 1})
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewPipManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.ListInstalled(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListInstalled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !equalStringSlices(got, tt.want) {
				t.Errorf("ListInstalled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "successful install",
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "install", "--user", "requests").
					Return([]byte("Successfully installed requests-2.28.0"), nil)
			},
			wantErr: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent-package",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "install", "--user", "nonexistent-package").
					Return([]byte("ERROR: Could not find a version that satisfies the requirement"), &ExitError{Code: 1})
			},
			wantErr:     true,
			wantErrCode: errors.ErrPackageNotFound,
		},
		{
			name:        "already installed",
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "install", "--user", "requests").
					Return([]byte("Requirement already satisfied: requests in /usr/local/lib"), nil)
			},
			wantErr: false, // Not an error - package is already installed
		},
		{
			name:        "permission denied",
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "install", "--user", "requests").
					Return([]byte("Permission denied: /usr/local/lib"), &ExitError{Code: 1})
			},
			wantErr:     true,
			wantErrCode: errors.ErrFilePermission,
		},
		{
			name:        "fallback to system install when --user fails",
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "install", "--user", "requests").
					Return([]byte("error: --user not supported"), &ExitError{Code: 1})
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "install", "requests").
					Return([]byte("Successfully installed requests-2.28.0"), nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewPipManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			err := manager.Install(ctx, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrCode != "" {
				plonkErr, ok := err.(*errors.PlonkError)
				if !ok {
					t.Errorf("Install() error type = %T, want *PlonkError", err)
					return
				}
				if plonkErr.Code != tt.wantErrCode {
					t.Errorf("Install() error code = %v, want %v", plonkErr.Code, tt.wantErrCode)
				}
			}
		})
	}
}

func TestPipManager_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "successful uninstall",
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "uninstall", "-y", "requests").
					Return([]byte("Successfully uninstalled requests-2.28.0"), nil)
			},
			wantErr: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent-package",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "uninstall", "-y", "nonexistent-package").
					Return([]byte("WARNING: Package 'nonexistent-package' is not installed"), &ExitError{Code: 1})
			},
			wantErr: false, // Not an error - package is not installed
		},
		{
			name:        "permission denied",
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().ExecuteCombined(gomock.Any(), "pip", "uninstall", "-y", "requests").
					Return([]byte("Permission denied: /usr/local/lib"), &ExitError{Code: 1})
			},
			wantErr:     true,
			wantErrCode: errors.ErrFilePermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewPipManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			err := manager.Uninstall(ctx, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uninstall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrCode != "" {
				plonkErr, ok := err.(*errors.PlonkError)
				if !ok {
					t.Errorf("Uninstall() error type = %T, want *PlonkError", err)
					return
				}
				if plonkErr.Code != tt.wantErrCode {
					t.Errorf("Uninstall() error code = %v, want %v", plonkErr.Code, tt.wantErrCode)
				}
			}
		})
	}
}

func TestPipManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        bool
		wantErr     bool
	}{
		{
			name:        "package is installed",
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return([]byte(`[{"name": "requests", "version": "2.28.0"}, {"name": "django", "version": "4.0"}]`), nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:        "package not installed",
			packageName: "flask",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return([]byte(`[{"name": "requests", "version": "2.28.0"}, {"name": "django", "version": "4.0"}]`), nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:        "normalized package name matching",
			packageName: "python-dateutil",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return([]byte(`[{"name": "python_dateutil", "version": "2.8.0"}]`), nil)
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewPipManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.IsInstalled(ctx, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsInstalled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsInstalled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipManager_GetInstalledVersion(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        string
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "get version of installed package",
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				// First check if installed
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return([]byte(`[{"name": "requests", "version": "2.28.0"}]`), nil)
				// Then get version
				m.EXPECT().Execute(gomock.Any(), "pip", "show", "requests").
					Return([]byte("Name: requests\nVersion: 2.28.0\nSummary: HTTP library"), nil)
			},
			want:    "2.28.0",
			wantErr: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return([]byte(`[]`), nil)
			},
			want:        "",
			wantErr:     true,
			wantErrCode: errors.ErrPackageNotFound,
		},
		{
			name:        "version not found in output",
			packageName: "requests",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().LookPath("pip").Return("/usr/bin/pip", nil).AnyTimes()
				m.EXPECT().Execute(gomock.Any(), "pip", "list", "--user", "--format=json").
					Return([]byte(`[{"name": "requests", "version": "2.28.0"}]`), nil)
				m.EXPECT().Execute(gomock.Any(), "pip", "show", "requests").
					Return([]byte("Name: requests\nSummary: HTTP library"), nil) // No version line
			},
			want:        "",
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

			manager := NewPipManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.GetInstalledVersion(ctx, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInstalledVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetInstalledVersion() = %v, want %v", got, tt.want)
			}
			if err != nil && tt.wantErrCode != "" {
				plonkErr, ok := err.(*errors.PlonkError)
				if !ok {
					t.Errorf("GetInstalledVersion() error type = %T, want *PlonkError", err)
					return
				}
				if plonkErr.Code != tt.wantErrCode {
					t.Errorf("GetInstalledVersion() error code = %v, want %v", plonkErr.Code, tt.wantErrCode)
				}
			}
		})
	}
}

func TestPipManager_normalizeName(t *testing.T) {
	manager := NewPipManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "Django", "django"},
		{"hyphen to underscore", "python-dateutil", "python_dateutil"},
		{"mixed case with hyphen", "Flask-RESTful", "flask_restful"},
		{"already normalized", "black", "black"},
		{"multiple hyphens", "some-package-name", "some_package_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.normalizeName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeName(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
