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

func TestNpmManager_IsAvailable(t *testing.T) {
	// Uses BaseManager.IsAvailable, so just test that it's properly initialized
	manager := NewNpmManager()
	if manager.BaseManager == nil {
		t.Error("BaseManager not initialized")
	}
	if manager.Config.BinaryName != "npm" {
		t.Errorf("BinaryName = %v, want npm", manager.Config.BinaryName)
	}
}

func TestNpmManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(m *mocks.MockCommandExecutor)
		want      []string
		wantErr   bool
	}{
		{
			name: "successful list with packages",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "--depth=0", "--json").
					Return([]byte(`{
  "dependencies": {
    "typescript": {
      "version": "5.0.0"
    },
    "eslint": {
      "version": "8.0.0"
    },
    "@angular/cli": {
      "version": "16.0.0"
    }
  }
}`), nil)
			},
			want:    []string{"@angular/cli", "eslint", "typescript"},
			wantErr: false,
		},
		{
			name: "empty package list",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "--depth=0", "--json").
					Return([]byte(`{"dependencies":{}}`), nil)
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "list with warnings (exit code 1)",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// npm list returns exit code 1 with output for warnings
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "--depth=0", "--json").
					Return([]byte(`{
  "dependencies": {
    "typescript": {
      "version": "5.0.0"
    }
  }
}`), &mockExitError{code: 1})
			},
			want:    []string{"typescript"},
			wantErr: false,
		},
		{
			name: "list with severe error (exit code 2)",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "--depth=0", "--json").
					Return(nil, &mockExitError{code: 2})
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid JSON output",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "--depth=0", "--json").
					Return([]byte(`invalid json`), nil)
			},
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewNpmManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.ListInstalled(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListInstalled() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !equalSlices(got, tt.want) {
				t.Errorf("ListInstalled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNpmManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        bool
		wantErr     bool
	}{
		{
			name:        "package is installed",
			packageName: "typescript",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "typescript").
					Return([]byte("typescript@4.5.0"), nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "nonexistent").
					Return(nil, &mockExitError{code: 1})
			},
			want:    false,
			wantErr: false,
		},
		{
			name:        "command error",
			packageName: "typescript",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "typescript").
					Return(nil, fmt.Errorf("command failed"))
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

			manager := NewNpmManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.IsInstalled(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsInstalled() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("IsInstalled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNpmManager_Search(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		mockSetup func(m *mocks.MockCommandExecutor)
		want      []string
		wantErr   bool
	}{
		{
			name:  "successful search with JSON results",
			query: "typescript",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "search", "typescript", "--json").
					Return([]byte(`[
						{"name": "typescript", "version": "4.5.0"},
						{"name": "typescript-eslint-parser", "version": "1.0.0"}
					]`), nil)
			},
			want:    []string{"typescript", "typescript-eslint-parser"},
			wantErr: false,
		},
		{
			name:  "no search results",
			query: "nonexistent-package-12345",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "search", "nonexistent-package-12345", "--json").
					Return(nil, &mockExitError{code: 1})
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name:  "search with malformed JSON fallback",
			query: "react",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "npm", "search", "react", "--json").
					Return([]byte(`
						"name": "react",
						"name": "react-dom",
						"name": "react-router"
					`), nil)
			},
			want:    []string{"react", "react-dom", "react-router"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewNpmManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.Search(ctx, tt.query)

			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !equalSlices(got, tt.want) {
				t.Errorf("Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNpmManager_Info(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        *PackageInfo
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "package info with JSON",
			packageName: "typescript",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "typescript").
					Return([]byte("typescript@4.5.0"), nil)
				// Get info
				m.EXPECT().Execute(gomock.Any(), "npm", "view", "typescript", "--json").
					Return([]byte(`{
						"name": "typescript",
						"version": "4.5.0",
						"description": "TypeScript is a language for application scale JavaScript development",
						"homepage": "https://www.typescriptlang.org/",
						"dependencies": {
							"source-map": "^0.7.3",
							"tslib": "^2.3.0"
						}
					}`), nil)
			},
			want: &PackageInfo{
				Name:         "typescript",
				Version:      "4.5.0",
				Description:  "TypeScript is a language for application scale JavaScript development",
				Homepage:     "https://www.typescriptlang.org/",
				Manager:      "npm",
				Installed:    true,
				Dependencies: []string{"source-map", "tslib"},
			},
			wantErr: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "nonexistent").
					Return(nil, &mockExitError{code: 1})
				// Try to get info
				m.EXPECT().Execute(gomock.Any(), "npm", "view", "nonexistent", "--json").
					Return(nil, &mockExitError{code: 1})
			},
			wantErr:     true,
			wantErrCode: errors.ErrPackageNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewNpmManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.Info(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Info() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if plonkErr, ok := err.(*errors.PlonkError); ok {
					if plonkErr.Code != tt.wantErrCode {
						t.Errorf("Info() error code = %v, want %v", plonkErr.Code, tt.wantErrCode)
					}
				}
			}
			if !tt.wantErr && got != nil {
				if got.Name != tt.want.Name || got.Version != tt.want.Version {
					t.Errorf("Info() basic fields = {%v, %v}, want {%v, %v}",
						got.Name, got.Version, tt.want.Name, tt.want.Version)
				}
				if got.Manager != tt.want.Manager || got.Installed != tt.want.Installed {
					t.Errorf("Info() meta fields = {%v, %v}, want {%v, %v}",
						got.Manager, got.Installed, tt.want.Manager, tt.want.Installed)
				}
			}
		})
	}
}

func TestNpmManager_GetInstalledVersion(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        string
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "get version with JSON",
			packageName: "typescript",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "typescript").
					Return([]byte("typescript@4.5.0"), nil)
				// Get version
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "typescript", "--depth=0", "--json").
					Return([]byte(`{
						"dependencies": {
							"typescript": {
								"version": "4.5.0"
							}
						}
					}`), nil)
			},
			want:    "4.5.0",
			wantErr: false,
		},
		{
			name:        "get version with ls fallback",
			packageName: "eslint",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "eslint").
					Return([]byte("eslint@8.0.0"), nil)
				// JSON fails
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "eslint", "--depth=0", "--json").
					Return(nil, fmt.Errorf("json parse error"))
				// Fallback to ls
				m.EXPECT().Execute(gomock.Any(), "npm", "ls", "-g", "eslint", "--depth=0").
					Return([]byte(`/usr/local/lib
└── eslint@8.0.0`), nil)
			},
			want:    "8.0.0",
			wantErr: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "npm", "list", "-g", "nonexistent").
					Return(nil, &mockExitError{code: 1})
			},
			wantErr:     true,
			wantErrCode: errors.ErrPackageNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewNpmManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			got, err := manager.GetInstalledVersion(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInstalledVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if plonkErr, ok := err.(*errors.PlonkError); ok {
					if plonkErr.Code != tt.wantErrCode {
						t.Errorf("GetInstalledVersion() error code = %v, want %v", plonkErr.Code, tt.wantErrCode)
					}
				}
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetInstalledVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare string slices
func equalSlices(a, b []string) bool {
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
