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

func TestGemManagerV2_IsAvailable(t *testing.T) {
	// Uses BaseManager.IsAvailable, so just test that it's properly initialized
	manager := NewGemManagerV2()
	if manager.BaseManager == nil {
		t.Error("BaseManager not initialized")
	}
	if manager.Config.BinaryName != "gem" {
		t.Errorf("BinaryName = %v, want gem", manager.Config.BinaryName)
	}
}

func TestGemManagerV2_ListInstalled(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(m *mocks.MockCommandExecutor)
		want      []string
		wantErr   bool
	}{
		{
			name: "successful list with gems",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "--no-versions").
					Return([]byte(`bundler
rails
rake
rspec`), nil)
			},
			want:    []string{"bundler", "rails", "rake", "rspec"},
			wantErr: false,
		},
		{
			name: "empty gem list",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "--no-versions").
					Return([]byte(""), nil)
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "list with header lines",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "--no-versions").
					Return([]byte(`*** LOCAL GEMS ***

bundler
rails`), nil)
			},
			want:    []string{"bundler", "rails"},
			wantErr: false,
		},
		{
			name: "list command fails",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "--no-versions").
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

			manager := NewGemManagerV2WithExecutor(mockExecutor)
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

func TestGemManagerV2_IsInstalled(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        bool
		wantErr     bool
	}{
		{
			name:        "gem is installed",
			packageName: "rails",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "rails").
					Return([]byte("rails (7.0.4, 6.1.7)"), nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:        "gem not installed",
			packageName: "sinatra",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "sinatra").
					Return([]byte(""), nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:        "exact match only",
			packageName: "rake",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "rake").
					Return([]byte("rake (13.0.6)"), nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:        "list command error",
			packageName: "rails",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "rails").
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

			manager := NewGemManagerV2WithExecutor(mockExecutor)
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

func TestGemManagerV2_Search(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		mockSetup func(m *mocks.MockCommandExecutor)
		want      []string
		wantErr   bool
	}{
		{
			name:  "successful search",
			query: "rails",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "search", "rails").
					Return([]byte(`rails (7.0.4)
rails-dom-testing (2.0.3)
rails-html-sanitizer (1.4.4)`), nil)
			},
			want:    []string{"rails", "rails-dom-testing", "rails-html-sanitizer"},
			wantErr: false,
		},
		{
			name:  "no results",
			query: "nonexistent-gem-12345",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "search", "nonexistent-gem-12345").
					Return([]byte(""), &mockExitError{code: 1})
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name:  "search error",
			query: "test",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().Execute(gomock.Any(), "gem", "search", "test").
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

			manager := NewGemManagerV2WithExecutor(mockExecutor)
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

func TestGemManagerV2_Info(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		want        *PackageInfo
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "gem info found and installed",
			packageName: "rails",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "rails").
					Return([]byte("rails (7.0.4)"), nil)
				// Get specification
				m.EXPECT().Execute(gomock.Any(), "gem", "specification", "rails").
					Return([]byte(`--- !ruby/object:Gem::Specification
name: rails
version: !ruby/object:Gem::Version
  version: 7.0.4
summary: Full-stack web application framework.
homepage: https://rubyonrails.org`), nil)
				// Get dependencies
				m.EXPECT().Execute(gomock.Any(), "gem", "dependency", "rails").
					Return([]byte(`Gem rails-7.0.4
  actioncable (= 7.0.4)
  actionmailbox (= 7.0.4)
  actionmailer (= 7.0.4)`), nil)
			},
			want: &PackageInfo{
				Name:         "rails",
				Version:      "7.0.4",
				Description:  "Full-stack web application framework.",
				Homepage:     "https://rubyonrails.org",
				Manager:      "gem",
				Installed:    true,
				Dependencies: []string{"actioncable", "actionmailbox", "actionmailer"},
			},
			wantErr: false,
		},
		{
			name:        "gem not found",
			packageName: "nonexistent",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "nonexistent").
					Return([]byte(""), nil)
				// Get specification fails
				m.EXPECT().Execute(gomock.Any(), "gem", "specification", "nonexistent").
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

			manager := NewGemManagerV2WithExecutor(mockExecutor)
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
				if got.Description != tt.want.Description || got.Homepage != tt.want.Homepage {
					t.Errorf("Info() description/homepage = {%v, %v}, want {%v, %v}",
						got.Description, got.Homepage, tt.want.Description, tt.want.Homepage)
				}
				if got.Manager != tt.want.Manager || got.Installed != tt.want.Installed {
					t.Errorf("Info() meta fields = {%v, %v}, want {%v, %v}",
						got.Manager, got.Installed, tt.want.Manager, tt.want.Installed)
				}
				if len(got.Dependencies) != len(tt.want.Dependencies) {
					t.Errorf("Info() dependencies count = %v, want %v",
						len(got.Dependencies), len(tt.want.Dependencies))
				}
			}
		})
	}
}

func TestGemManagerV2_GetInstalledVersion(t *testing.T) {
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
			packageName: "rails",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "rails").
					Return([]byte("rails (7.0.4)"), nil).Times(2)
			},
			want:    "7.0.4",
			wantErr: false,
		},
		{
			name:        "multiple versions installed",
			packageName: "rake",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "rake").
					Return([]byte("rake (13.0.6, 13.0.3, 12.3.3)"), nil).Times(2)
			},
			want:    "13.0.6",
			wantErr: false,
		},
		{
			name:        "gem not installed",
			packageName: "sinatra",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// Check if installed
				m.EXPECT().Execute(gomock.Any(), "gem", "list", "--local", "sinatra").
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

			manager := NewGemManagerV2WithExecutor(mockExecutor)
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

func TestGemManagerV2_Install(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name:        "successful install",
			packageName: "sinatra",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "gem", "install", "sinatra", "--user-install").
					Return([]byte("Successfully installed sinatra-3.0.5"), nil)
			},
			wantErr: false,
		},
		{
			name:        "gem not found",
			packageName: "nonexistent-gem",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "gem", "install", "nonexistent-gem", "--user-install").
					Return([]byte("ERROR: Could not find a valid gem 'nonexistent-gem' (>= 0) in any repository"),
						&mockExitError{code: 2})
			},
			wantErr:     true,
			wantErrCode: errors.ErrPackageNotFound,
		},
		{
			name:        "already installed",
			packageName: "rails",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "gem", "install", "rails", "--user-install").
					Return([]byte("rails-7.0.4 already installed"), &mockExitError{code: 0})
			},
			wantErr: false,
		},
		{
			name:        "user-install fails, retry without",
			packageName: "bundler",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				// First try with --user-install fails
				m.EXPECT().ExecuteCombined(gomock.Any(), "gem", "install", "bundler", "--user-install").
					Return([]byte("ERROR: While executing gem ... (Gem::FilePermissionError)\n    You don't have write permissions for the /usr/local/bin directory.\n    Use --user-install"),
						&mockExitError{code: 1})
				// Retry without --user-install succeeds
				m.EXPECT().ExecuteCombined(gomock.Any(), "gem", "install", "bundler").
					Return([]byte("Successfully installed bundler-2.4.6"), nil)
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

			manager := NewGemManagerV2WithExecutor(mockExecutor)
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

func TestGemManagerV2_Uninstall(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		mockSetup   func(m *mocks.MockCommandExecutor)
		wantErr     bool
	}{
		{
			name:        "successful uninstall",
			packageName: "sinatra",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "gem", "uninstall", "sinatra", "-x", "-a", "-I").
					Return([]byte("Successfully uninstalled sinatra-3.0.5"), nil)
			},
			wantErr: false,
		},
		{
			name:        "gem not installed",
			packageName: "flask",
			mockSetup: func(m *mocks.MockCommandExecutor) {
				m.EXPECT().ExecuteCombined(gomock.Any(), "gem", "uninstall", "flask", "-x", "-a", "-I").
					Return([]byte("ERROR: While executing gem ... (Gem::InstallError)\n    flask is not installed"),
						&mockExitError{code: 1})
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

			manager := NewGemManagerV2WithExecutor(mockExecutor)
			ctx := context.Background()

			err := manager.Uninstall(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Uninstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
