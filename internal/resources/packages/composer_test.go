// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
)

func TestComposerManager_parseListOutput(t *testing.T) {
	manager := NewComposerManager()

	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard composer global show JSON output",
			output: []byte(`{
				"installed": [
					{
						"name": "phpunit/phpunit",
						"version": "9.5.28",
						"description": "The PHP Unit Testing framework."
					},
					{
						"name": "friendsofphp/php-cs-fixer",
						"version": "v3.13.0",
						"description": "A tool to automatically fix PHP code style"
					}
				]
			}`),
			want: []string{"friendsofphp/php-cs-fixer", "phpunit/phpunit"},
		},
		{
			name:   "empty JSON object",
			output: []byte(`{}`),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manager.parseListOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseListOutput() = %v, want %v", got, tt.want)
				return
			}
			for i, pkg := range got {
				if pkg != tt.want[i] {
					t.Errorf("parseListOutput()[%d] = %v, want %v", i, pkg, tt.want[i])
				}
			}
		})
	}
}

func TestComposerManager_parseSearchOutput(t *testing.T) {
	manager := NewComposerManager()

	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard composer search JSON output",
			output: []byte(`{
				"phpunit/phpunit": "The PHP Unit Testing framework.",
				"phpunit/php-code-coverage": "Library that provides collection, processing, and rendering functionality for PHP code coverage information.",
				"mockery/mockery": "Mockery is a simple yet flexible PHP mock object framework"
			}`),
			want: []string{"mockery/mockery", "phpunit/php-code-coverage", "phpunit/phpunit"},
		},
		{
			name:   "empty search results",
			output: []byte(`{}`),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "empty JSON array",
			output: []byte(`[]`),
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manager.parseSearchOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
				return
			}
			for i, pkg := range got {
				if pkg != tt.want[i] {
					t.Errorf("parseSearchOutput()[%d] = %v, want %v", i, pkg, tt.want[i])
				}
			}
		})
	}
}

func TestComposerManager_parseInfoOutput(t *testing.T) {
	manager := NewComposerManager()

	tests := []struct {
		name         string
		output       []byte
		packageName  string
		wantName     string
		wantVersion  string
		wantDesc     string
		wantHomepage string
		wantDeps     int // number of dependencies (excluding platform packages)
	}{
		{
			name: "standard composer show JSON output",
			output: []byte(`{
				"name": "phpunit/phpunit",
				"version": "9.5.28",
				"description": "The PHP Unit Testing framework.",
				"homepage": "https://phpunit.de/",
				"keywords": ["testing", "unit"],
				"license": ["BSD-3-Clause"],
				"authors": [
					{
						"name": "Sebastian Bergmann",
						"email": "sebastian@phpunit.de"
					}
				],
				"require": {
					"php": ">=7.3",
					"ext-dom": "*",
					"ext-json": "*",
					"doctrine/instantiator": "^1.3.1",
					"myclabs/deep-copy": "^1.10.0"
				}
			}`),
			packageName:  "phpunit/phpunit",
			wantName:     "phpunit/phpunit",
			wantVersion:  "9.5.28",
			wantDesc:     "The PHP Unit Testing framework.",
			wantHomepage: "https://phpunit.de/",
			wantDeps:     2, // Should exclude php, ext-dom, ext-json
		},
		{
			name: "minimal JSON output",
			output: []byte(`{
				"name": "simple/package",
				"version": "1.0.0"
			}`),
			packageName: "simple/package",
			wantName:    "simple/package",
			wantVersion: "1.0.0",
			wantDesc:    "",
			wantDeps:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manager.parseInfoOutput(tt.output, tt.packageName)

			if got.Name != tt.wantName {
				t.Errorf("parseInfoOutput().Name = %v, want %v", got.Name, tt.wantName)
			}
			if got.Version != tt.wantVersion {
				t.Errorf("parseInfoOutput().Version = %v, want %v", got.Version, tt.wantVersion)
			}
			if got.Description != tt.wantDesc {
				t.Errorf("parseInfoOutput().Description = %v, want %v", got.Description, tt.wantDesc)
			}
			if got.Homepage != tt.wantHomepage {
				t.Errorf("parseInfoOutput().Homepage = %v, want %v", got.Homepage, tt.wantHomepage)
			}
			if len(got.Dependencies) != tt.wantDeps {
				t.Errorf("parseInfoOutput().Dependencies count = %v, want %v (deps: %v)", len(got.Dependencies), tt.wantDeps, got.Dependencies)
			}
		})
	}
}

func TestComposerManager_extractValueAfterColon(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "standard key-value pair",
			line: "version : v3.13.0",
			want: "v3.13.0",
		},
		{
			name: "key-value with quotes",
			line: `description : "A tool to automatically fix PHP code style"`,
			want: `"A tool to automatically fix PHP code style"`,
		},
		{
			name: "no colon",
			line: "no colon here",
			want: "",
		},
		{
			name: "multiple colons",
			line: "url : https://example.com:8080/path",
			want: "https://example.com:8080/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractValueAfterColon(tt.line)
			if got != tt.want {
				t.Errorf("extractValueAfterColon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComposerManager_cleanValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "quoted string",
			value: `"hello world"`,
			want:  "hello world",
		},
		{
			name:  "single quoted string",
			value: "'hello world'",
			want:  "hello world",
		},
		{
			name:  "whitespace around value",
			value: "  hello world  ",
			want:  "hello world",
		},
		{
			name:  "quotes and whitespace",
			value: `  "hello world"  `,
			want:  "hello world",
		},
		{
			name:  "no quotes or whitespace",
			value: "hello",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanValue(tt.value)
			if got != tt.want {
				t.Errorf("cleanValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComposerManager_IsAvailable(t *testing.T) {
	manager := NewComposerManager()
	ctx := context.Background()

	// This test just ensures the method exists and doesn't panic
	_, err := manager.IsAvailable(ctx)
	if err != nil && !IsContextError(err) {
		// Only context errors are acceptable here since we can't guarantee composer is installed
		t.Logf("IsAvailable returned error (this may be expected): %v", err)
	}
}

func TestComposerManager_SupportsSearch(t *testing.T) {
	manager := NewComposerManager()
	if !manager.SupportsSearch() {
		t.Errorf("SupportsSearch() = false, want true - composer supports search")
	}
}

func TestComposerManager_handleInstallError(t *testing.T) {
	manager := NewComposerManager()

	tests := []struct {
		name         string
		output       []byte
		packageName  string
		exitCode     int
		wantContains string
	}{
		{
			name:         "package not found",
			output:       []byte("could not find package nonexistent/package"),
			packageName:  "nonexistent/package",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "permission denied",
			output:       []byte("permission denied writing to composer directory"),
			packageName:  "testpkg/test",
			exitCode:     1,
			wantContains: "permission denied",
		},
		{
			name:         "dependency resolution failure",
			output:       []byte("Your requirements could not be resolved to an installable set of packages"),
			packageName:  "incompatible/package",
			exitCode:     2,
			wantContains: "dependency resolution failed",
		},
		{
			name:         "generic error with output",
			output:       []byte("Some installation error occurred"),
			packageName:  "testpkg",
			exitCode:     2,
			wantContains: "installation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleInstallError(mockErr, tt.output, tt.packageName)
			if err == nil {
				t.Errorf("handleInstallError() = nil, want error")
				return
			}

			if tt.wantContains != "" && !stringContains(err.Error(), tt.wantContains) {
				t.Errorf("handleInstallError() error = %v, want to contain %s", err, tt.wantContains)
			}
		})
	}
}

func TestComposerManager_handleUninstallError(t *testing.T) {
	manager := NewComposerManager()

	tests := []struct {
		name        string
		output      []byte
		packageName string
		exitCode    int
		wantErr     bool
	}{
		{
			name:        "package not installed",
			output:      []byte("Package testpkg/test is not installed"),
			packageName: "testpkg/test",
			exitCode:    1,
			wantErr:     false, // Not installed should be success
		},
		{
			name:        "package does not exist",
			output:      []byte("Package does not exist in global composer.json"),
			packageName: "missing/package",
			exitCode:    1,
			wantErr:     false, // Does not exist should be success
		},
		{
			name:        "permission denied",
			output:      []byte("permission denied writing to composer directory"),
			packageName: "testpkg",
			exitCode:    1,
			wantErr:     true,
		},
		{
			name:        "generic error",
			output:      []byte("Some uninstallation error"),
			packageName: "testenv",
			exitCode:    2,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleUninstallError(mockErr, tt.output, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleUninstallError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComposerManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "phpunit/phpunit",
			mockResponses: map[string]CommandResponse{
				"composer global require phpunit/phpunit": {
					Output: []byte("Changed current directory to /home/user/.config/composer\nUsing version ^9.5 for phpunit/phpunit\n./composer.json has been updated\n./composer.lock has been updated\nInstalling dependencies from lock file (including require-dev)\n- Installing phpunit/phpunit (9.5.28): Loading from cache"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent/package",
			mockResponses: map[string]CommandResponse{
				"composer global require nonexistent/package": {
					Output: []byte("could not find package nonexistent/package"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "dependency resolution failure",
			packageName: "incompatible/package",
			mockResponses: map[string]CommandResponse{
				"composer global require incompatible/package": {
					Output: []byte("Your requirements could not be resolved to an installable set of packages"),
					Error:  &MockExitError{Code: 2},
				},
			},
			expectError:   true,
			errorContains: "dependency resolution failed",
		},
		{
			name:        "permission denied",
			packageName: "some/package",
			mockResponses: map[string]CommandResponse{
				"composer global require some/package": {
					Output: []byte("permission denied writing to composer directory"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewComposerManager()
			err := manager.Install(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !stringContains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestComposerManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "phpunit/phpunit",
			mockResponses: map[string]CommandResponse{
				"composer global remove phpunit/phpunit": {
					Output: []byte("Changed current directory to /home/user/.config/composer\n./composer.json has been updated\n./composer.lock has been updated\nUninstalling phpunit/phpunit (9.5.28)\n- Removing phpunit/phpunit (9.5.28)"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent/package",
			mockResponses: map[string]CommandResponse{
				"composer global remove nonexistent/package": {
					Output: []byte("Package nonexistent/package is not installed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not installed should be success for uninstall
		},
		{
			name:        "package does not exist in composer.json",
			packageName: "missing/package",
			mockResponses: map[string]CommandResponse{
				"composer global remove missing/package": {
					Output: []byte("Package does not exist in global composer.json"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Does not exist should be success for uninstall
		},
		{
			name:        "permission denied",
			packageName: "some/package",
			mockResponses: map[string]CommandResponse{
				"composer global remove some/package": {
					Output: []byte("permission denied writing to composer directory"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewComposerManager()
			err := manager.Uninstall(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !stringContains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestComposerManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name: "list with packages",
			mockResponses: map[string]CommandResponse{
				"composer global show --format=json": {
					Output: []byte(`{
						"installed": [
							{
								"name": "phpunit/phpunit",
								"version": "9.5.28",
								"description": "The PHP Unit Testing framework."
							},
							{
								"name": "friendsofphp/php-cs-fixer",
								"version": "v3.13.0",
								"description": "A tool to automatically fix PHP code style"
							}
						]
					}`),
					Error: nil,
				},
			},
			expected:    []string{"friendsofphp/php-cs-fixer", "phpunit/phpunit"},
			expectError: false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"composer global show --format=json": {
					Output: []byte(`{}`),
					Error:  nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name: "command error",
			mockResponses: map[string]CommandResponse{
				"composer global show --format=json": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    []string{}, // Exit code 1 is handled as no packages
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewComposerManager()
			result, err := manager.ListInstalled(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && !stringSlicesEqual(result, tt.expected) {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestComposerManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      bool
		expectError   bool
	}{
		{
			name:        "package is installed",
			packageName: "phpunit/phpunit",
			mockResponses: map[string]CommandResponse{
				"composer global show phpunit/phpunit": {
					Output: []byte("phpunit/phpunit 9.5.28 The PHP Unit Testing framework."),
					Error:  nil,
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent/package",
			mockResponses: map[string]CommandResponse{
				"composer global show nonexistent/package": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "empty list",
			packageName: "symfony/console",
			mockResponses: map[string]CommandResponse{
				"composer global show symfony/console": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "command error",
			packageName: "test/package",
			mockResponses: map[string]CommandResponse{
				"composer global show test/package": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 2},
				},
			},
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewComposerManager()
			result, err := manager.IsInstalled(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestComposerManager_Search(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name:  "successful search",
			query: "phpunit",
			mockResponses: map[string]CommandResponse{
				"composer search phpunit --format=json": {
					Output: []byte(`{
						"phpunit/phpunit": "The PHP Unit Testing framework.",
						"phpunit/php-code-coverage": "Library that provides collection, processing, and rendering functionality for PHP code coverage information.",
						"mockery/mockery": "Mockery is a simple yet flexible PHP mock object framework"
					}`),
					Error: nil,
				},
			},
			expected:    []string{"mockery/mockery", "phpunit/php-code-coverage", "phpunit/phpunit"},
			expectError: false,
		},
		{
			name:  "no results found",
			query: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"composer search nonexistent --format=json": {
					Output: []byte(`{}`),
					Error:  nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name:  "command error",
			query: "test",
			mockResponses: map[string]CommandResponse{
				"composer search test --format=json": {
					Output: []byte("error: search failed"),
					Error:  &MockExitError{Code: 2}, // Exit code 1 means no results, 2+ means real error
				},
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewComposerManager()
			result, err := manager.Search(context.Background(), tt.query)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && !stringSlicesEqual(result, tt.expected) {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestComposerManager_Info(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		checkName     string
		checkVersion  string
	}{
		{
			name:        "info for available package",
			packageName: "phpunit/phpunit",
			mockResponses: map[string]CommandResponse{
				"composer global show phpunit/phpunit": {
					Output: []byte("phpunit/phpunit 9.5.28 The PHP Unit Testing framework."),
					Error:  nil,
				},
				"composer show phpunit/phpunit --format=json": {
					Output: []byte(`{
						"name": "phpunit/phpunit",
						"version": "9.5.28",
						"description": "The PHP Unit Testing framework.",
						"homepage": "https://phpunit.de/",
						"keywords": ["testing", "unit"],
						"license": ["BSD-3-Clause"],
						"authors": [
							{
								"name": "Sebastian Bergmann",
								"email": "sebastian@phpunit.de"
							}
						],
						"require": {
							"php": ">=7.3",
							"ext-dom": "*",
							"ext-json": "*",
							"doctrine/instantiator": "^1.3.1",
							"myclabs/deep-copy": "^1.10.0"
						}
					}`),
					Error: nil,
				},
			},
			expectError:  false,
			checkName:    "phpunit/phpunit",
			checkVersion: "9.5.28",
		},
		{
			name:        "info for not installed package",
			packageName: "notinstalled/package",
			mockResponses: map[string]CommandResponse{
				"composer global show notinstalled/package": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
				"composer show notinstalled/package --format=json": {
					Output: []byte(`{
						"name": "notinstalled/package",
						"version": "1.0.0",
						"description": "A test package"
					}`),
					Error: nil,
				},
			},
			expectError:  false,
			checkName:    "notinstalled/package",
			checkVersion: "1.0.0",
		},
		{
			name:        "command error on info fetch",
			packageName: "test/package",
			mockResponses: map[string]CommandResponse{
				"composer global show test/package": {
					Output: []byte("error: package not found"),
					Error:  &MockExitError{Code: 1},
				},
				"composer show test/package --format=json": {
					Output: []byte("error: package not found"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewComposerManager()
			result, err := manager.Info(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result == nil {
				t.Errorf("Expected non-nil PackageInfo but got nil")
			}
			if !tt.expectError && result != nil {
				if result.Name != tt.checkName {
					t.Errorf("Expected package name '%s' but got '%s'", tt.checkName, result.Name)
				}
				if tt.checkVersion != "" && result.Version != tt.checkVersion {
					t.Errorf("Expected package version '%s' but got '%s'", tt.checkVersion, result.Version)
				}
			}
		})
	}
}

func TestComposerManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      string
		expectError   bool
	}{
		{
			name:        "get version of installed package",
			packageName: "phpunit/phpunit",
			mockResponses: map[string]CommandResponse{
				"composer global show phpunit/phpunit": {
					Output: []byte("phpunit/phpunit 9.5.28 The PHP Unit Testing framework."),
					Error:  nil,
				},
				"composer global show phpunit/phpunit --format=json": {
					Output: []byte(`{"version":"9.5.28"}`),
					Error:  nil,
				},
			},
			expected:    "9.5.28",
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent/package",
			mockResponses: map[string]CommandResponse{
				"composer global show nonexistent/package": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name:        "command error",
			packageName: "phpunit/phpunit",
			mockResponses: map[string]CommandResponse{
				"composer global show phpunit/phpunit": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 2},
				},
			},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewComposerManager()
			result, err := manager.InstalledVersion(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected '%s' but got '%s'", tt.expected, result)
			}
		})
	}
}

func TestComposerManager_Upgrade(t *testing.T) {
	tests := []struct {
		name          string
		packages      []string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:     "upgrade specific packages",
			packages: []string{"phpunit/phpunit", "friendsofphp/php-cs-fixer"},
			mockResponses: map[string]CommandResponse{
				"composer global update phpunit/phpunit friendsofphp/php-cs-fixer": {
					Output: []byte("Changed current directory to /home/user/.config/composer\nUpdating dependencies\nLock file operations: 2 installs, 0 updates, 0 removals\n- Upgrading phpunit/phpunit (9.5.27 => 9.5.28)\n- Upgrading friendsofphp/php-cs-fixer (v3.12.0 => v3.13.0)"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade all packages",
			packages: []string{}, // empty means all packages
			mockResponses: map[string]CommandResponse{
				"composer global update": {
					Output: []byte("Changed current directory to /home/user/.config/composer\nUpdating dependencies\nLock file operations: 2 installs, 0 updates, 0 removals\n- Upgrading phpunit/phpunit (9.5.27 => 9.5.28)\n- Upgrading friendsofphp/php-cs-fixer (v3.12.0 => v3.13.0)"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade with dependency resolution error",
			packages: []string{"problematic/package"},
			mockResponses: map[string]CommandResponse{
				"composer global update problematic/package": {
					Output: []byte("Your requirements could not be resolved to an installable set of packages"),
					Error:  &MockExitError{Code: 2},
				},
			},
			expectError:   true,
			errorContains: "dependency resolution failed",
		},
		{
			name:     "upgrade with permission error",
			packages: []string{"some/package"},
			mockResponses: map[string]CommandResponse{
				"composer global update some/package": {
					Output: []byte("permission denied writing to composer directory"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:     "upgrade with generic error",
			packages: []string{"failing/package"},
			mockResponses: map[string]CommandResponse{
				"composer global update failing/package": {
					Output: []byte("Some upgrade error occurred"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "upgrade failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewComposerManager()
			err := manager.Upgrade(context.Background(), tt.packages)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !stringContains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s' but got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestComposerManager_IsAvailableWithMock(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      bool
	}{
		{
			name: "composer is available",
			mockResponses: map[string]CommandResponse{
				"composer --version": {
					Output: []byte("Composer version 2.5.1 2023-02-09 11:52:44"),
					Error:  nil,
				},
			},
			expected: true,
		},
		{
			name: "composer not found",
			mockResponses: map[string]CommandResponse{
				"composer --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 127},
				},
			},
			expected: false,
		},
		{
			name: "composer exists but not functional",
			mockResponses: map[string]CommandResponse{
				"composer --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewComposerManager()
			result, _ := manager.IsAvailable(context.Background())

			if result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

// Note: Uses MockExitError from executor.go and stringContains, stringSlicesEqual from test_helpers.go
