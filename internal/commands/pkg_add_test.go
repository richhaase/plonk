// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/operations"
	"go.uber.org/mock/gomock"
)

func TestPkgAddCommand_Creation(t *testing.T) {
	// Test that the pkg add command is created correctly
	if pkgAddCmd == nil {
		t.Fatal("pkgAddCmd is nil")
	}

	if pkgAddCmd.Use != "add [package1] [package2] ..." {
		t.Errorf("Expected Use to be 'add [package1] [package2] ...', got '%s'", pkgAddCmd.Use)
	}

	if pkgAddCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if pkgAddCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if pkgAddCmd.RunE == nil {
		t.Error("RunE should not be nil")
	}
}

func TestPkgAddCommand_Flags(t *testing.T) {
	// Test that flags are set up correctly
	flag := pkgAddCmd.Flags().Lookup("manager")
	if flag == nil {
		t.Error("manager flag not found")
	}

	flag = pkgAddCmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Error("dry-run flag not found")
	}
}

func TestAddSinglePackage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create minimal config
	cfg := &config.Config{}

	// Create mock package manager
	mockMgr := managers.NewMockPackageManager(ctrl)

	tests := []struct {
		name        string
		packageName string
		dryRun      bool
		setupMock   func(*managers.MockPackageManager)
		setupLock   func(*lock.YAMLLockService)
		expected    operations.OperationResult
	}{
		{
			name:        "successful install",
			packageName: "git",
			dryRun:      false,
			setupMock: func(m *managers.MockPackageManager) {
				m.EXPECT().IsInstalled(gomock.Any(), "git").Return(false, nil)
				m.EXPECT().Install(gomock.Any(), "git").Return(nil)
				m.EXPECT().GetInstalledVersion(gomock.Any(), "git").Return("2.43.0", nil)
			},
			setupLock: func(ls *lock.YAMLLockService) {
				// Package not in lock file initially
			},
			expected: operations.OperationResult{
				Name:    "git",
				Manager: "homebrew",
				Version: "2.43.0",
				Status:  "added",
			},
		},
		{
			name:        "already managed",
			packageName: "git",
			dryRun:      false,
			setupMock: func(m *managers.MockPackageManager) {
				// No mock calls expected since package is already managed
			},
			setupLock: func(ls *lock.YAMLLockService) {
				// Add package to lock file first
				ls.AddPackage("homebrew", "git", "2.43.0")
			},
			expected: operations.OperationResult{
				Name:           "git",
				Manager:        "homebrew",
				Status:         "skipped",
				AlreadyManaged: true,
			},
		},
		{
			name:        "dry run",
			packageName: "neovim",
			dryRun:      true,
			setupMock: func(m *managers.MockPackageManager) {
				// No mock calls expected for dry run
			},
			setupLock: func(ls *lock.YAMLLockService) {
				// Package not in lock file
			},
			expected: operations.OperationResult{
				Name:    "neovim",
				Manager: "homebrew",
				Status:  "would-add",
			},
		},
		{
			name:        "already installed but not managed",
			packageName: "ripgrep",
			dryRun:      false,
			setupMock: func(m *managers.MockPackageManager) {
				m.EXPECT().IsInstalled(gomock.Any(), "ripgrep").Return(true, nil)
				m.EXPECT().GetInstalledVersion(gomock.Any(), "ripgrep").Return("13.0.0", nil)
			},
			setupLock: func(ls *lock.YAMLLockService) {
				// Package not in lock file initially
			},
			expected: operations.OperationResult{
				Name:    "ripgrep",
				Manager: "homebrew",
				Version: "13.0.0",
				Status:  "added",
			},
		},
		{
			name:        "install failure",
			packageName: "nonexistent",
			dryRun:      false,
			setupMock: func(m *managers.MockPackageManager) {
				m.EXPECT().IsInstalled(gomock.Any(), "nonexistent").Return(false, nil)
				m.EXPECT().Install(gomock.Any(), "nonexistent").Return(fmt.Errorf("package not found"))
			},
			setupLock: func(ls *lock.YAMLLockService) {
				// Package not in lock file
			},
			expected: operations.OperationResult{
				Name:    "nonexistent",
				Manager: "homebrew",
				Status:  "failed",
				Error:   fmt.Errorf("package not found"), // Error will be wrapped, so we'll check for presence
			},
		},
		{
			name:        "version lookup failure but install succeeds",
			packageName: "test-pkg",
			dryRun:      false,
			setupMock: func(m *managers.MockPackageManager) {
				m.EXPECT().IsInstalled(gomock.Any(), "test-pkg").Return(false, nil)
				m.EXPECT().Install(gomock.Any(), "test-pkg").Return(nil)
				m.EXPECT().GetInstalledVersion(gomock.Any(), "test-pkg").Return("", fmt.Errorf("version not found"))
			},
			setupLock: func(ls *lock.YAMLLockService) {
				// Package not in lock file
			},
			expected: operations.OperationResult{
				Name:    "test-pkg",
				Manager: "homebrew",
				Version: "unknown",
				Status:  "added",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup fresh lock service for each test
			testTempDir := t.TempDir()
			testLockService := lock.NewYAMLLockService(testTempDir)
			tt.setupLock(testLockService)

			// Setup mock
			tt.setupMock(mockMgr)

			// Execute function
			ctx := context.Background()
			result := addSinglePackage(ctx, cfg, testLockService, mockMgr, "homebrew", tt.packageName, tt.dryRun)

			// Verify results
			if result.Name != tt.expected.Name {
				t.Errorf("Name: got %s, want %s", result.Name, tt.expected.Name)
			}
			if result.Manager != tt.expected.Manager {
				t.Errorf("Manager: got %s, want %s", result.Manager, tt.expected.Manager)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("Status: got %s, want %s", result.Status, tt.expected.Status)
			}
			if result.Version != tt.expected.Version {
				t.Errorf("Version: got %s, want %s", result.Version, tt.expected.Version)
			}
			if result.AlreadyManaged != tt.expected.AlreadyManaged {
				t.Errorf("AlreadyManaged: got %v, want %v", result.AlreadyManaged, tt.expected.AlreadyManaged)
			}

			// For error cases, just check that an error is present
			if tt.expected.Error != nil && result.Error == nil {
				t.Error("Expected an error but got none")
			}
			if tt.expected.Error == nil && result.Error != nil {
				t.Errorf("Expected no error but got: %v", result.Error)
			}
		})
	}
}

func TestCreateActionsFromResult(t *testing.T) {
	tests := []struct {
		name     string
		result   operations.OperationResult
		expected []string
	}{
		{
			name: "successful add with version",
			result: operations.OperationResult{
				Name:    "git",
				Manager: "homebrew",
				Version: "2.43.0",
				Status:  "added",
			},
			expected: []string{
				"Successfully installed git@2.43.0",
				"Added git to lock file",
			},
		},
		{
			name: "successful add without version",
			result: operations.OperationResult{
				Name:    "git",
				Manager: "homebrew",
				Version: "unknown",
				Status:  "added",
			},
			expected: []string{
				"Successfully installed git",
				"Added git to lock file",
			},
		},
		{
			name: "skipped package",
			result: operations.OperationResult{
				Name:           "git",
				Manager:        "homebrew",
				Status:         "skipped",
				AlreadyManaged: true,
			},
			expected: []string{
				"git already managed by homebrew",
			},
		},
		{
			name: "dry run",
			result: operations.OperationResult{
				Name:    "git",
				Manager: "homebrew",
				Status:  "would-add",
			},
			expected: []string{
				"Would install git",
				"Would add git to lock file",
			},
		},
		{
			name: "failed package",
			result: operations.OperationResult{
				Name:    "nonexistent",
				Manager: "homebrew",
				Status:  "failed",
				Error:   fmt.Errorf("package not found"),
			},
			expected: []string{
				"Failed to process nonexistent",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := createActionsFromResult(tt.result)

			if len(actions) != len(tt.expected) {
				t.Errorf("Expected %d actions, got %d", len(tt.expected), len(actions))
				return
			}

			for i, action := range actions {
				if action != tt.expected[i] {
					t.Errorf("Action %d: got %q, want %q", i, action, tt.expected[i])
				}
			}
		})
	}
}

func TestConvertResultsToEnhancedAdd(t *testing.T) {
	results := []operations.OperationResult{
		{
			Name:    "git",
			Manager: "homebrew",
			Version: "2.43.0",
			Status:  "added",
		},
		{
			Name:           "neovim",
			Manager:        "homebrew",
			Status:         "skipped",
			AlreadyManaged: true,
		},
		{
			Name:    "nonexistent",
			Manager: "homebrew",
			Status:  "failed",
			Error:   fmt.Errorf("package not found"),
		},
	}

	outputs := convertResultsToEnhancedAdd(results)

	if len(outputs) != len(results) {
		t.Errorf("Expected %d outputs, got %d", len(results), len(outputs))
		return
	}

	// Test first result (successful)
	if outputs[0].Package != "git" {
		t.Errorf("Package: got %s, want git", outputs[0].Package)
	}
	if !outputs[0].ConfigAdded {
		t.Error("ConfigAdded should be true for added package")
	}
	if !outputs[0].Installed {
		t.Error("Installed should be true for added package")
	}

	// Test second result (skipped)
	if outputs[1].Package != "neovim" {
		t.Errorf("Package: got %s, want neovim", outputs[1].Package)
	}
	if outputs[1].ConfigAdded {
		t.Error("ConfigAdded should be false for skipped package")
	}
	if !outputs[1].AlreadyInstalled {
		t.Error("AlreadyInstalled should be true for skipped package")
	}

	// Test third result (failed)
	if outputs[2].Package != "nonexistent" {
		t.Errorf("Package: got %s, want nonexistent", outputs[2].Package)
	}
	if outputs[2].Error == "" {
		t.Error("Error should not be empty for failed package")
	}
}
