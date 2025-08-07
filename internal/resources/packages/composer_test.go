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
