// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources"
)

// MockLockService implements the minimal lock.LockService interface for testing
type MockLockService struct {
	packages    map[string]map[string]bool // manager -> packageName -> exists
	addCalls    []AddPackageCall
	removeCalls []RemovePackageCall
	findResults map[string][]lock.ResourceEntry // packageName -> locations
}

type AddPackageCall struct {
	Manager  string
	Name     string
	Version  string
	Metadata map[string]interface{}
}

type RemovePackageCall struct {
	Manager string
	Name    string
}

func NewMockLockService() *MockLockService {
	return &MockLockService{
		packages:    make(map[string]map[string]bool),
		findResults: make(map[string][]lock.ResourceEntry),
	}
}

func (m *MockLockService) Read() (*lock.Lock, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockLockService) Write(lock *lock.Lock) error {
	return fmt.Errorf("not implemented")
}

func (m *MockLockService) AddPackage(manager, name, version string, metadata map[string]interface{}) error {
	m.addCalls = append(m.addCalls, AddPackageCall{
		Manager:  manager,
		Name:     name,
		Version:  version,
		Metadata: metadata,
	})

	if m.packages[manager] == nil {
		m.packages[manager] = make(map[string]bool)
	}
	m.packages[manager][name] = true

	return nil
}

func (m *MockLockService) RemovePackage(manager, name string) error {
	m.removeCalls = append(m.removeCalls, RemovePackageCall{
		Manager: manager,
		Name:    name,
	})

	if m.packages[manager] != nil {
		delete(m.packages[manager], name)
	}

	return nil
}

func (m *MockLockService) GetPackages(manager string) ([]lock.ResourceEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockLockService) HasPackage(manager, name string) bool {
	if managerPackages, ok := m.packages[manager]; ok {
		return managerPackages[name]
	}
	return false
}

func (m *MockLockService) FindPackage(name string) []lock.ResourceEntry {
	if results, ok := m.findResults[name]; ok {
		return results
	}
	return []lock.ResourceEntry{}
}

// Helper function to set up package existence
func (m *MockLockService) SetPackageExists(manager, name string) {
	if m.packages[manager] == nil {
		m.packages[manager] = make(map[string]bool)
	}
	m.packages[manager][name] = true
}

// Helper function to set up find results
func (m *MockLockService) SetFindResults(packageName string, results []lock.ResourceEntry) {
	m.findResults[packageName] = results
}

// Test getManagerInstallSuggestion
func TestGetManagerInstallSuggestion(t *testing.T) {
	tests := []struct {
		name     string
		manager  string
		expected string
	}{
		{
			name:     "brew manager",
			manager:  "brew",
			expected: `install with: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
		},
		{
			name:     "npm manager",
			manager:  "npm",
			expected: "install Node.js from https://nodejs.org/",
		},
		{
			name:     "pip manager",
			manager:  "pip",
			expected: "install Python from https://python.org/",
		},
		{
			name:     "cargo manager",
			manager:  "cargo",
			expected: "install Rust from https://rustup.rs/",
		},
		{
			name:     "gem manager",
			manager:  "gem",
			expected: "install Ruby from https://ruby-lang.org/",
		},
		{
			name:     "go manager",
			manager:  "go",
			expected: "install Go from https://golang.org/",
		},
		{
			name:     "unknown manager",
			manager:  "unknown",
			expected: "check installation instructions for unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getManagerInstallSuggestion(tt.manager)
			if got != tt.expected {
				t.Errorf("getManagerInstallSuggestion(%s) = %v, want %v", tt.manager, got, tt.expected)
			}
		})
	}
}

// Test InstallPackages with temporary directory setup
func TestInstallPackages(t *testing.T) {
	// Save and restore the default executor for package managers
	originalExecutor := defaultExecutor
	defer func() { defaultExecutor = originalExecutor }()

	tests := []struct {
		name            string
		packages        []string
		opts            InstallOptions
		setupMock       func(*MockCommandExecutor)
		expectedResults int
		checkResults    func(t *testing.T, results []resources.OperationResult)
	}{
		{
			name:     "dry run single package",
			packages: []string{"vim"},
			opts: InstallOptions{
				Manager: "brew",
				DryRun:  true,
			},
			setupMock: func(mock *MockCommandExecutor) {
				// No commands should be executed in dry run
			},
			expectedResults: 1,
			checkResults: func(t *testing.T, results []resources.OperationResult) {
				if results[0].Status != "would-add" {
					t.Errorf("Expected status 'would-add', got %s", results[0].Status)
				}
				if results[0].Name != "vim" {
					t.Errorf("Expected name 'vim', got %s", results[0].Name)
				}
			},
		},
		{
			name:     "dry run multiple packages",
			packages: []string{"vim", "git", "curl"},
			opts: InstallOptions{
				Manager: "brew",
				DryRun:  true,
			},
			setupMock: func(mock *MockCommandExecutor) {
				// No commands should be executed in dry run
			},
			expectedResults: 3,
			checkResults: func(t *testing.T, results []resources.OperationResult) {
				for i, pkg := range []string{"vim", "git", "curl"} {
					if results[i].Status != "would-add" {
						t.Errorf("Package %s: expected status 'would-add', got %s", pkg, results[i].Status)
					}
					if results[i].Name != pkg {
						t.Errorf("Expected name '%s', got %s", pkg, results[i].Name)
					}
				}
			},
		},
		{
			name:     "context cancellation",
			packages: []string{"vim", "git", "curl"},
			opts: InstallOptions{
				Manager: "brew",
				DryRun:  true,
			},
			setupMock: func(mock *MockCommandExecutor) {
				// No commands should be executed
			},
			expectedResults: 0, // No packages processed due to immediate cancellation
			checkResults: func(t *testing.T, results []resources.OperationResult) {
				// No specific checks needed
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for config and lock file
			tempDir := t.TempDir()

			// Set up mock command executor
			mock := &MockCommandExecutor{
				Responses: make(map[string]CommandResponse),
			}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}
			SetDefaultExecutor(mock)

			// Create context
			ctx := context.Background()
			if tt.name == "context cancellation" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				// Cancel immediately to test cancellation handling
				cancel()
			}

			// Run InstallPackages
			results, err := InstallPackages(ctx, tempDir, tt.packages, tt.opts)

			// Check error
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check number of results
			if len(results) != tt.expectedResults {
				t.Errorf("Expected %d results, got %d", tt.expectedResults, len(results))
			}

			// Check specific results
			if tt.checkResults != nil {
				tt.checkResults(t, results)
			}
		})
	}
}

// Test UninstallPackages with temporary directory setup
func TestUninstallPackages(t *testing.T) {
	// Save and restore the default executor for package managers
	originalExecutor := defaultExecutor
	defer func() { defaultExecutor = originalExecutor }()

	tests := []struct {
		name            string
		packages        []string
		opts            UninstallOptions
		setupMock       func(*MockCommandExecutor)
		setupLockFile   func(t *testing.T, configDir string)
		expectedResults int
		checkResults    func(t *testing.T, results []resources.OperationResult)
	}{
		{
			name:     "dry run single package",
			packages: []string{"vim"},
			opts: UninstallOptions{
				Manager: "brew",
				DryRun:  true,
			},
			setupMock: func(mock *MockCommandExecutor) {
				// No commands should be executed in dry run
			},
			setupLockFile: func(t *testing.T, configDir string) {
				// No lock file needed for dry run test
			},
			expectedResults: 1,
			checkResults: func(t *testing.T, results []resources.OperationResult) {
				if results[0].Status != "would-remove" {
					t.Errorf("Expected status 'would-remove', got %s", results[0].Status)
				}
				if results[0].Name != "vim" {
					t.Errorf("Expected name 'vim', got %s", results[0].Name)
				}
			},
		},
		{
			name:     "dry run multiple packages",
			packages: []string{"vim", "git", "curl"},
			opts: UninstallOptions{
				Manager: "brew",
				DryRun:  true,
			},
			setupMock: func(mock *MockCommandExecutor) {
				// No commands should be executed in dry run
			},
			setupLockFile: func(t *testing.T, configDir string) {
				// No lock file needed for dry run test
			},
			expectedResults: 3,
			checkResults: func(t *testing.T, results []resources.OperationResult) {
				for i, pkg := range []string{"vim", "git", "curl"} {
					if results[i].Status != "would-remove" {
						t.Errorf("Package %s: expected status 'would-remove', got %s", pkg, results[i].Status)
					}
					if results[i].Name != pkg {
						t.Errorf("Expected name '%s', got %s", pkg, results[i].Name)
					}
				}
			},
		},
		{
			name:     "context cancellation",
			packages: []string{"vim", "git", "curl"},
			opts: UninstallOptions{
				Manager: "brew",
				DryRun:  true,
			},
			setupMock: func(mock *MockCommandExecutor) {
				// No commands should be executed
			},
			setupLockFile: func(t *testing.T, configDir string) {
				// No lock file needed for dry run test
			},
			expectedResults: 0, // No packages processed due to immediate cancellation
			checkResults: func(t *testing.T, results []resources.OperationResult) {
				// No specific checks needed
			},
		},
		{
			name:     "manager detection from lock file",
			packages: []string{"vim"},
			opts: UninstallOptions{
				// No manager specified, should detect from lock file
				DryRun: true,
			},
			setupMock: func(mock *MockCommandExecutor) {
				// No commands in dry run
			},
			setupLockFile: func(t *testing.T, configDir string) {
				// Create a lock file with vim managed by npm
				lockService := lock.NewYAMLLockService(configDir)
				metadata := map[string]interface{}{
					"manager": "npm",
					"name":    "vim",
					"version": "1.0.0",
				}
				err := lockService.AddPackage("npm", "vim", "1.0.0", metadata)
				if err != nil {
					t.Fatalf("Failed to set up lock file: %v", err)
				}
			},
			expectedResults: 1,
			checkResults: func(t *testing.T, results []resources.OperationResult) {
				if results[0].Manager != "npm" {
					t.Errorf("Expected manager 'npm' from lock file, got %s", results[0].Manager)
				}
				if results[0].Status != "would-remove" {
					t.Errorf("Expected status 'would-remove', got %s", results[0].Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for config and lock file
			tempDir := t.TempDir()

			// Set up lock file if needed
			if tt.setupLockFile != nil {
				tt.setupLockFile(t, tempDir)
			}

			// Set up mock command executor
			mock := &MockCommandExecutor{
				Responses: make(map[string]CommandResponse),
			}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}
			SetDefaultExecutor(mock)

			// Create context
			ctx := context.Background()
			if tt.name == "context cancellation" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				// Cancel immediately to test cancellation handling
				cancel()
			}

			// Run UninstallPackages
			results, err := UninstallPackages(ctx, tempDir, tt.packages, tt.opts)

			// Check error
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check number of results
			if len(results) != tt.expectedResults {
				t.Errorf("Expected %d results, got %d", tt.expectedResults, len(results))
			}

			// Check specific results
			if tt.checkResults != nil {
				tt.checkResults(t, results)
			}
		})
	}
}
