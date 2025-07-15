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

func TestCargoManager_IsAvailable(t *testing.T) {
	// Uses BaseManager.IsAvailable, so just test that it's properly initialized
	manager := NewCargoManager()
	if manager.BaseManager == nil {
		t.Error("BaseManager not initialized")
	}
	if manager.Config.BinaryName != "cargo" {
		t.Errorf("BinaryName = %v, want cargo", manager.Config.BinaryName)
	}
}

func TestCargoManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(m *mocks.MockCommandExecutor)
		want      []string
		wantErr   bool
	}{
		{
			name: "successful list with packages",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(`bat v0.18.3:
    bat
fd-find v8.2.1:
    fd
ripgrep v13.0.0:
    rg`), nil)
			},
			want:    []string{"bat", "fd-find", "ripgrep"},
			wantErr: false,
		},
		{
			name: "empty package list",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(""), nil)
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "list with complex output",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(`cargo-edit v0.8.0:
    cargo-add
    cargo-rm
    cargo-upgrade
exa v0.10.1:
    exa`), nil)
			},
			want:    []string{"cargo-edit", "exa"},
			wantErr: false,
		},
		{
			name: "list command fails",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return(nil, fmt.Errorf("command failed"))
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

			manager := NewCargoManagerWithExecutor(mockExecutor)
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

func TestCargoManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        bool
		wantErr     bool
	}{
		{
			name:        "package is installed",
			packageName: "ripgrep",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(`ripgrep v13.0.0:
    rg
bat v0.18.3:
    bat`), nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:        "package not installed",
			packageName: "tokei",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(`ripgrep v13.0.0:
    rg`), nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:        "list command error",
			packageName: "ripgrep",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
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

			manager := NewCargoManagerWithExecutor(mockExecutor)
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

func TestCargoManager_Search(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		mockSetup func(m *mocks.MockCommandExecutor)
		want      []string
		wantErr   bool
	}{
		{
			name:  "successful search",
			query: "serde",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "search", "serde").
					Return([]byte(`serde = "1.0.136"    # A generic serialization/deserialization framework
serde_json = "1.0.79" # A JSON serialization file format
serde_yaml = "0.8.23" # YAML support for Serde`), nil)
			},
			want:    []string{"serde", "serde_json", "serde_yaml"},
			wantErr: false,
		},
		{
			name:  "no results",
			query: "nonexistent-crate-12345",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "search", "nonexistent-crate-12345").
					Return([]byte("no crates found"), &mockExitError{code: 101})
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name:  "search error",
			query: "test",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "cargo", "search", "test").
					Return(nil, fmt.Errorf("network error"))
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

			manager := NewCargoManagerWithExecutor(mockExecutor)
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

func TestCargoManager_Info(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        *PackageInfo
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "package info found and installed",
			packageName: "serde",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Search for info
				m.EXPECT().Execute(gomock.Any(), "cargo", "search", "serde", "--limit", "1").
					Return([]byte(`serde = "1.0.136"    # A generic serialization/deserialization framework`), nil)
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(`serde v1.0.136:
    serde`), nil)
			},
			want: &PackageInfo{
				Name:        "serde",
				Version:     "1.0.136",
				Description: "A generic serialization/deserialization framework",
				Manager:     "cargo",
				Installed:   true,
			},
			wantErr: false,
		},
		{
			name:        "package info found but not installed",
			packageName: "tokei",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Search for info
				m.EXPECT().Execute(gomock.Any(), "cargo", "search", "tokei", "--limit", "1").
					Return([]byte(`tokei = "12.1.2"    # Count your code, quickly.`), nil)
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(""), nil)
			},
			want: &PackageInfo{
				Name:        "tokei",
				Version:     "12.1.2",
				Description: "Count your code, quickly.",
				Manager:     "cargo",
				Installed:   false,
			},
			wantErr: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Search returns empty
				m.EXPECT().Execute(gomock.Any(), "cargo", "search", "nonexistent", "--limit", "1").
					Return([]byte(""), nil)
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

			manager := NewCargoManagerWithExecutor(mockExecutor)
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
			if !tt.wantErr && got != nil && tt.want != nil {
				if got.Name != tt.want.Name || got.Version != tt.want.Version {
					t.Errorf("Info() basic fields = {%v, %v}, want {%v, %v}",
						got.Name, got.Version, tt.want.Name, tt.want.Version)
				}
				if got.Description != tt.want.Description {
					t.Errorf("Info() description = %v, want %v", got.Description, tt.want.Description)
				}
				if got.Manager != tt.want.Manager || got.Installed != tt.want.Installed {
					t.Errorf("Info() meta fields = {%v, %v}, want {%v, %v}",
						got.Manager, got.Installed, tt.want.Manager, tt.want.Installed)
				}
			}
		})
	}
}

func TestCargoManager_GetInstalledVersion(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        string
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "get version successfully",
			packageName: "ripgrep",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(`ripgrep v13.0.0:
    rg
bat v0.18.3:
    bat`), nil).Times(2)
			},
			want:    "13.0.0",
			wantErr: false,
		},
		{
			name:        "package not installed",
			packageName: "tokei",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(`ripgrep v13.0.0:
    rg`), nil)
			},
			wantErr:     true,
			wantErrCode: errors.ErrPackageNotFound,
		},
		{
			name:        "version with v prefix",
			packageName: "cargo-edit",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "cargo", "install", "--list").
					Return([]byte(`cargo-edit v0.8.0:
    cargo-add
    cargo-rm
    cargo-upgrade`), nil).Times(2)
			},
			want:    "0.8.0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewCargoManagerWithExecutor(mockExecutor)
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

func TestCargoManager_Install(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "successful install",
			packageName: "ripgrep",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "cargo", "install", "ripgrep").
					Return([]byte("Installed package ripgrep"), nil)
			},
			wantErr: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent-crate",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "cargo", "install", "nonexistent-crate").
					Return([]byte("error: could not find `nonexistent-crate` in registry"), &mockExitError{code: 101})
			},
			wantErr:     true,
			wantErrCode: errors.ErrPackageNotFound,
		},
		{
			name:        "already installed",
			packageName: "ripgrep",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "cargo", "install", "ripgrep").
					Return([]byte("error: binary `rg` already exists in destination"), &mockExitError{code: 101})
			},
			wantErr: false, // Already installed is treated as success
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewCargoManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			err := manager.Install(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if plonkErr, ok := err.(*errors.PlonkError); ok {
					if plonkErr.Code != tt.wantErrCode {
						t.Errorf("Install() error code = %v, want %v", plonkErr.Code, tt.wantErrCode)
					}
				}
			}
		})
	}
}

func TestCargoManager_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		wantErr     bool
	}{
		{
			name:        "successful uninstall",
			packageName: "ripgrep",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "cargo", "uninstall", "ripgrep").
					Return([]byte("Removing ripgrep"), nil)
			},
			wantErr: false,
		},
		{
			name:        "package not installed",
			packageName: "tokei",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "cargo", "uninstall", "tokei").
					Return([]byte("error: package `tokei` is not installed"), &mockExitError{code: 101})
			},
			wantErr: false, // Not installed is treated as success for uninstall
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockCommandExecutor(ctrl)
			tt.mockSetup(mockExecutor)

			manager := NewCargoManagerWithExecutor(mockExecutor)
			ctx := context.Background()

			err := manager.Uninstall(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Uninstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
