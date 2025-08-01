// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
	"testing"
)

func TestCargoManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard cargo install --list output",
			output: []byte(`ripgrep v13.0.0:
    rg
bat v0.22.1:
    bat
exa v0.10.1:
    exa`),
			want: []string{"ripgrep", "bat", "exa"},
		},
		{
			name:   "empty list",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "single package",
			output: []byte(`cargo-edit v0.11.6:
    cargo-add
    cargo-rm
    cargo-upgrade`),
			want: []string{"cargo-edit"},
		},
		{
			name: "packages with hyphens",
			output: []byte(`cargo-watch v8.1.2:
    cargo-watch
cargo-update v11.1.2:
    cargo-install-update
    cargo-install-update-config`),
			want: []string{"cargo-watch", "cargo-update"},
		},
		{
			name: "output with extra whitespace",
			output: []byte(`  ripgrep v13.0.0:
    rg

  bat v0.22.1:
    bat
  `),
			want: []string{"ripgrep", "bat"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewCargoManager()
			got := manager.parseListOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
			} else {
				for i, expected := range tt.want {
					if got[i] != expected {
						t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
						break
					}
				}
			}
		})
	}
}

func TestCargoManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard cargo search output",
			output: []byte(`serde = "1.0.193"       # A generic serialization/deserialization framework
serde_json = "1.0.108"  # A JSON serialization file format
serde_yaml = "0.9.27"   # YAML support for Serde`),
			want: []string{"serde", "serde_json", "serde_yaml"},
		},
		{
			name:   "no results",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "single result",
			output: []byte(`ripgrep = "13.0.0"      # ripgrep recursively searches directories for a regex pattern`),
			want:   []string{"ripgrep"},
		},
		{
			name: "output with quotes in description",
			output: []byte(`clap = "4.4.11"         # A simple to use, efficient, and full-featured Command Line Argument Parser
structopt = "0.3.26"    # Parse command line arguments by defining a struct.`),
			want: []string{"clap", "structopt"},
		},
		{
			name: "malformed lines",
			output: []byte(`serde = "1.0.193"
malformed line without equals
another = "1.0.0"`),
			want: []string{"serde", "malformed", "another"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewCargoManager()
			got := manager.parseSearchOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
			} else {
				for i, expected := range tt.want {
					if got[i] != expected {
						t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
						break
					}
				}
			}
		})
	}
}

func TestCargoManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        *PackageInfo
	}{
		{
			name:        "standard search output with description",
			output:      []byte(`serde = "1.0.193"       # A generic serialization/deserialization framework`),
			packageName: "serde",
			want: &PackageInfo{
				Name:        "serde",
				Version:     "1.0.193",
				Description: "A generic serialization/deserialization framework",
			},
		},
		{
			name:        "simple version output",
			output:      []byte(`ripgrep = "13.0.0"`),
			packageName: "ripgrep",
			want: &PackageInfo{
				Name:        "ripgrep",
				Version:     "13.0.0",
				Description: "",
			},
		},
		{
			name:        "package not matching query",
			output:      []byte(`serde = "1.0.193"       # A generic serialization/deserialization framework`),
			packageName: "different",
			want:        nil,
		},
		{
			name:        "empty output",
			output:      []byte(""),
			packageName: "test",
			want:        nil,
		},
		{
			name:        "output with extra spaces",
			output:      []byte(`clap   =   "4.4.11"    # A simple to use Command Line Parser`),
			packageName: "clap",
			want: &PackageInfo{
				Name:        "clap",
				Version:     "4.4.11",
				Description: "A simple to use Command Line Parser",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewCargoManager()
			got := manager.parseInfoOutput(tt.output, tt.packageName)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseInfoOutput() = %+v, want nil", got)
				}
			} else if !equalPackageInfo(got, tt.want) {
				t.Errorf("parseInfoOutput() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCargoManager_extractVersion(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        string
	}{
		{
			name: "standard version format",
			output: []byte(`ripgrep v13.0.0:
    rg
bat v0.22.1:
    bat`),
			packageName: "ripgrep",
			want:        "13.0.0",
		},
		{
			name: "package with hyphen",
			output: []byte(`cargo-edit v0.11.6:
    cargo-add
    cargo-rm`),
			packageName: "cargo-edit",
			want:        "0.11.6",
		},
		{
			name: "package not found",
			output: []byte(`ripgrep v13.0.0:
    rg`),
			packageName: "bat",
			want:        "",
		},
		{
			name:        "empty output",
			output:      []byte(""),
			packageName: "test",
			want:        "",
		},
		{
			name: "version with pre-release",
			output: []byte(`test-crate v1.0.0-alpha.1:
    test`),
			packageName: "test-crate",
			want:        "1.0.0-alpha.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewCargoManager()
			got := manager.extractVersion(tt.output, tt.packageName)
			if got != tt.want {
				t.Errorf("extractVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCargoManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name: "cargo is available",
			mockResponses: map[string]CommandResponse{
				"cargo --version": {
					Output: []byte("cargo 1.74.0 (ecb9851af 2023-10-18)"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:          "cargo not found",
			mockResponses: map[string]CommandResponse{
				// Empty responses means LookPath will fail
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "cargo exists but not functional",
			mockResponses: map[string]CommandResponse{
				"cargo": {}, // Makes LookPath succeed
				"cargo --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewCargoManager()
			result, err := manager.IsAvailable(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestCargoManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"cargo install ripgrep": {
					Output: []byte("    Installed package `ripgrep v13.0.0` (executable `rg`)"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"cargo install nonexistent": {
					Output: []byte("error: could not find `nonexistent` in registry `crates-io`"),
					Error:  &MockExitError{Code: 101},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "already installed",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"cargo install ripgrep": {
					Output: []byte("    Ignored package `ripgrep v13.0.0` is already installed"),
					Error:  &MockExitError{Code: 0},
				},
			},
			expectError: false,
		},
		{
			name:        "permission denied",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"cargo install test": {
					Output: []byte("error: failed to create directory `/usr/local/cargo/bin`: Permission denied"),
					Error:  &MockExitError{Code: 101},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:        "build failed",
			packageName: "broken-crate",
			mockResponses: map[string]CommandResponse{
				"cargo install broken-crate": {
					Output: []byte("error: failed to compile `broken-crate v0.1.0`"),
					Error:  &MockExitError{Code: 101},
				},
			},
			expectError:   true,
			errorContains: "failed to build",
		},
		{
			name:        "network error",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"cargo install test": {
					Output: []byte("error: failed to fetch `https://crates.io/api/v1/crates/test/download`"),
					Error:  &MockExitError{Code: 101},
				},
			},
			expectError:   true,
			errorContains: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewCargoManager()
			err := manager.Install(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && tt.errorContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestCargoManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"cargo uninstall ripgrep": {
					Output: []byte("    Removing /home/user/.cargo/bin/rg"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"cargo uninstall nonexistent": {
					Output: []byte("error: package `nonexistent` is not installed"),
					Error:  &MockExitError{Code: 101},
				},
			},
			expectError: false, // Not installed is success for uninstall
		},
		{
			name:        "permission denied",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"cargo uninstall test": {
					Output: []byte("error: failed to remove `/usr/local/cargo/bin/test`: Permission denied"),
					Error:  &MockExitError{Code: 101},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:        "dependency conflict",
			packageName: "lib-crate",
			mockResponses: map[string]CommandResponse{
				"cargo uninstall lib-crate": {
					Output: []byte("error: package `other-crate` still depends on `lib-crate`"),
					Error:  &MockExitError{Code: 101},
				},
			},
			expectError:   true,
			errorContains: "dependency conflicts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewCargoManager()
			err := manager.Uninstall(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && tt.errorContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestCargoManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult []string
		expectError    bool
	}{
		{
			name: "list with packages",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(`ripgrep v13.0.0:
    rg
bat v0.22.1:
    bat
exa v0.10.1:
    exa`),
					Error: nil,
				},
			},
			expectedResult: []string{"ripgrep", "bat", "exa"},
			expectError:    false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(""),
					Error:  nil,
				},
			},
			expectedResult: []string{},
			expectError:    false,
		},
		{
			name: "list with cargo extensions",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(`cargo-edit v0.11.6:
    cargo-add
    cargo-rm
    cargo-upgrade
cargo-watch v8.1.2:
    cargo-watch`),
					Error: nil,
				},
			},
			expectedResult: []string{"cargo-edit", "cargo-watch"},
			expectError:    false,
		},
		{
			name: "command error",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewCargoManager()
			result, err := manager.ListInstalled(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if len(result) != len(tt.expectedResult) {
					t.Errorf("Expected %d packages but got %d: %v", len(tt.expectedResult), len(result), result)
				} else {
					for i, pkg := range tt.expectedResult {
						if result[i] != pkg {
							t.Errorf("Expected package %s at index %d but got %s", pkg, i, result[i])
						}
					}
				}
			}
		})
	}
}

func TestCargoManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name:        "package is installed",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(`ripgrep v13.0.0:
    rg
bat v0.22.1:
    bat`),
					Error: nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(`ripgrep v13.0.0:
    rg`),
					Error: nil,
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:        "empty list",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(""),
					Error:  nil,
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:        "command error",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewCargoManager()
			result, err := manager.IsInstalled(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestCargoManager_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockResponses  map[string]CommandResponse
		expectedResult []string
		expectError    bool
	}{
		{
			name:  "successful search",
			query: "serde",
			mockResponses: map[string]CommandResponse{
				"cargo search serde": {
					Output: []byte(`serde = "1.0.193"       # A generic serialization/deserialization framework
serde_json = "1.0.108"  # A JSON serialization file format
serde_yaml = "0.9.27"   # YAML support for Serde`),
					Error: nil,
				},
			},
			expectedResult: []string{"serde", "serde_json", "serde_yaml"},
			expectError:    false,
		},
		{
			name:  "no results",
			query: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"cargo search nonexistent": {
					Output: []byte("no crates found for query: nonexistent"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: []string{},
			expectError:    false,
		},
		{
			name:  "search error",
			query: "test",
			mockResponses: map[string]CommandResponse{
				"cargo search test": {
					Output: []byte("error: failed to search"),
					Error:  &MockExitError{Code: 101},
				},
			},
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:  "single result",
			query: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"cargo search ripgrep": {
					Output: []byte(`ripgrep = "13.0.0"      # ripgrep recursively searches directories for a regex pattern`),
					Error:  nil,
				},
			},
			expectedResult: []string{"ripgrep"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewCargoManager()
			result, err := manager.Search(context.Background(), tt.query)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestCargoManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult string
		expectError    bool
	}{
		{
			name:        "get version of installed package",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(`ripgrep v13.0.0:
    rg
bat v0.22.1:
    bat`),
					Error: nil,
				},
			},
			expectedResult: "13.0.0",
			expectError:    false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(`ripgrep v13.0.0:
    rg`),
					Error: nil,
				},
			},
			expectedResult: "",
			expectError:    true,
		},
		{
			name:        "version with pre-release",
			packageName: "test-crate",
			mockResponses: map[string]CommandResponse{
				"cargo install --list": {
					Output: []byte(`test-crate v1.0.0-alpha.1:
    test`),
					Error: nil,
				},
			},
			expectedResult: "1.0.0-alpha.1",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewCargoManager()
			result, err := manager.InstalledVersion(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestCargoManager_Info(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult *PackageInfo
		expectError    bool
	}{
		{
			name:        "successful info for available package",
			packageName: "serde",
			mockResponses: map[string]CommandResponse{
				"cargo search serde --limit 1": {
					Output: []byte(`serde = "1.0.193"       # A generic serialization/deserialization framework`),
					Error:  nil,
				},
				"cargo install --list": {
					Output: []byte(""),
					Error:  nil,
				},
			},
			expectedResult: &PackageInfo{
				Name:        "serde",
				Version:     "1.0.193",
				Description: "A generic serialization/deserialization framework",
				Manager:     "cargo",
				Installed:   false,
			},
			expectError: false,
		},
		{
			name:        "successful info for installed package",
			packageName: "ripgrep",
			mockResponses: map[string]CommandResponse{
				"cargo search ripgrep --limit 1": {
					Output: []byte(`ripgrep = "13.0.0"      # ripgrep recursively searches directories for a regex pattern`),
					Error:  nil,
				},
				"cargo install --list": {
					Output: []byte(`ripgrep v13.0.0:
    rg`),
					Error: nil,
				},
			},
			expectedResult: &PackageInfo{
				Name:        "ripgrep",
				Version:     "13.0.0",
				Description: "ripgrep recursively searches directories for a regex pattern",
				Manager:     "cargo",
				Installed:   true,
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"cargo search nonexistent --limit 1": {
					Output: []byte(""),
					Error:  nil,
				},
			},
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:        "search command error",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"cargo search test --limit 1": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 101},
				},
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewCargoManager()
			result, err := manager.Info(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && !equalPackageInfo(result, tt.expectedResult) {
				t.Errorf("Expected result %+v but got %+v", tt.expectedResult, result)
			}
		})
	}
}
