// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
	"testing"
)

func TestGemManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard gem list output",
			output: []byte(`*** LOCAL GEMS ***

bundler
minitest
power_assert
rake
test-unit`),
			want: []string{"bundler", "minitest", "power_assert", "rake", "test-unit"},
		},
		{
			name:   "empty list",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "list with header only",
			output: []byte(`*** LOCAL GEMS ***

`),
			want: []string{},
		},
		{
			name: "single gem",
			output: []byte(`*** LOCAL GEMS ***

rails`),
			want: []string{"rails"},
		},
		{
			name: "gems with hyphens and underscores",
			output: []byte(`*** LOCAL GEMS ***

ruby-debug
test_framework
json-schema`),
			want: []string{"ruby-debug", "test_framework", "json-schema"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewGemManager()
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

func TestGemManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard gem search output",
			output: []byte(`rails (7.0.4.3)
rails-api (0.4.1)
rails-i18n (7.0.6)
rails_admin (3.1.1)`),
			want: []string{"rails", "rails-api", "rails-i18n", "rails_admin"},
		},
		{
			name:   "no results",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "single result",
			output: []byte(`sinatra (3.0.5)`),
			want:   []string{"sinatra"},
		},
		{
			name: "gems with multiple versions",
			output: []byte(`nokogiri (1.14.2)
nokogiri-diff (0.2.0)
nokogiri-happymapper (0.9.0)`),
			want: []string{"nokogiri", "nokogiri-diff", "nokogiri-happymapper"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewGemManager()
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

func TestGemManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        *PackageInfo
	}{
		{
			name: "standard gem specification output",
			output: []byte(`--- !ruby/object:Gem::Specification
name: rails
version: !ruby/object:Gem::Version
  version: 7.0.4.3
platform: ruby
authors:
- David Heinemeier Hansson
summary: Full-stack web application framework.
homepage: https://rubyonrails.org
license: MIT`),
			packageName: "rails",
			want: &PackageInfo{
				Name:        "rails",
				Version:     "7.0.4.3",
				Description: "Full-stack web application framework.",
				Homepage:    "https://rubyonrails.org",
			},
		},
		{
			name: "minimal specification",
			output: []byte(`--- !ruby/object:Gem::Specification
name: minitest
version: !ruby/object:Gem::Version
  version: 5.18.0`),
			packageName: "minitest",
			want: &PackageInfo{
				Name:        "minitest",
				Version:     "5.18.0",
				Description: "",
				Homepage:    "",
			},
		},
		{
			name: "version with quotes",
			output: []byte(`--- !ruby/object:Gem::Specification
name: json
version: !ruby/object:Gem::Version
  version: "2.6.3"
summary: "JSON Implementation for Ruby"
homepage: 'https://github.com/ruby/json'`),
			packageName: "json",
			want: &PackageInfo{
				Name:        "json",
				Version:     "2.6.3",
				Description: "JSON Implementation for Ruby",
				Homepage:    "https://github.com/ruby/json",
			},
		},
		{
			name:        "empty output",
			output:      []byte(""),
			packageName: "unknown",
			want: &PackageInfo{
				Name:        "unknown",
				Version:     "",
				Description: "",
				Homepage:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewGemManager()
			got := manager.parseInfoOutput(tt.output, tt.packageName)
			if !equalPackageInfo(got, tt.want) {
				t.Errorf("parseInfoOutput() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGemManager_extractVersion(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        string
	}{
		{
			name:        "single version",
			output:      []byte("rails (7.0.4.3)"),
			packageName: "rails",
			want:        "7.0.4.3",
		},
		{
			name:        "multiple versions",
			output:      []byte("bundler (2.4.7, 2.3.26)"),
			packageName: "bundler",
			want:        "2.4.7",
		},
		{
			name:        "version with pre-release",
			output:      []byte("test-gem (1.0.0.pre)"),
			packageName: "test-gem",
			want:        "1.0.0.pre",
		},
		{
			name:        "no version found",
			output:      []byte("some other output"),
			packageName: "test",
			want:        "",
		},
		{
			name:        "empty output",
			output:      []byte(""),
			packageName: "test",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewGemManager()
			got := manager.extractVersion(tt.output, tt.packageName)
			if got != tt.want {
				t.Errorf("extractVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGemManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name: "gem is available",
			mockResponses: map[string]CommandResponse{
				"gem --version": {
					Output: []byte("3.4.6"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:          "gem not found",
			mockResponses: map[string]CommandResponse{
				// Empty responses means LookPath will fail
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "gem exists but not functional",
			mockResponses: map[string]CommandResponse{
				"gem": {}, // Makes LookPath succeed
				"gem --version": {
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

			manager := NewGemManager()
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

func TestGemManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install with --user-install",
			packageName: "sinatra",
			mockResponses: map[string]CommandResponse{
				"gem install sinatra --user-install": {
					Output: []byte("Successfully installed sinatra-3.0.5\n1 gem installed"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"gem install nonexistent --user-install": {
					Output: []byte("ERROR:  Could not find a valid gem 'nonexistent' (>= 0) in any repository"),
					Error:  &MockExitError{Code: 2},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "already installed",
			packageName: "bundler",
			mockResponses: map[string]CommandResponse{
				"gem install bundler --user-install": {
					Output: []byte("bundler-2.4.7 already installed"),
					Error:  &MockExitError{Code: 0},
				},
			},
			expectError: false,
		},
		{
			name:        "permission denied",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"gem install test --user-install": {
					Output: []byte("ERROR:  While executing gem ... (Gem::FilePermissionError)\n    You don't have write permissions for the /usr/local directory."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:        "ruby version mismatch",
			packageName: "rails",
			mockResponses: map[string]CommandResponse{
				"gem install rails --user-install": {
					Output: []byte("ERROR:  Error installing rails:\n\trails requires Ruby version >= 2.7.0. The current ruby version is 2.6.10."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "requires a different Ruby version",
		},
		{
			name:        "build error",
			packageName: "nokogiri",
			mockResponses: map[string]CommandResponse{
				"gem install nokogiri --user-install": {
					Output: []byte("Building native extensions. This could take a while...\nERROR:  Error installing nokogiri:\n\tERROR: Failed to build gem native extension."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "failed to build",
		},
		{
			name:        "retry without --user-install",
			packageName: "sinatra",
			mockResponses: map[string]CommandResponse{
				"gem install sinatra --user-install": {
					Output: []byte("ERROR: Use --user-install instead of --local"),
					Error:  &MockExitError{Code: 1},
				},
				"gem install sinatra": {
					Output: []byte("Successfully installed sinatra-3.0.5"),
					Error:  nil,
				},
			},
			expectError: false,
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

			manager := NewGemManager()
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

func TestGemManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "sinatra",
			mockResponses: map[string]CommandResponse{
				"gem uninstall sinatra -x -a -I": {
					Output: []byte("Successfully uninstalled sinatra-3.0.5"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"gem uninstall nonexistent -x -a -I": {
					Output: []byte("Gem 'nonexistent' is not installed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not installed is success for uninstall
		},
		{
			name:        "permission denied",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"gem uninstall test -x -a -I": {
					Output: []byte("ERROR:  While executing gem ... (Gem::FilePermissionError)\n    You don't have write permissions for the /usr/local directory."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:        "dependency conflict",
			packageName: "rack",
			mockResponses: map[string]CommandResponse{
				"gem uninstall rack -x -a -I": {
					Output: []byte("ERROR: rack is depended upon by sinatra-3.0.5"),
					Error:  &MockExitError{Code: 1},
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

			manager := NewGemManager()
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

func TestGemManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult []string
		expectError    bool
	}{
		{
			name: "list with gems",
			mockResponses: map[string]CommandResponse{
				"gem list --local --no-versions": {
					Output: []byte(`*** LOCAL GEMS ***

bundler
minitest
power_assert
rake
test-unit`),
					Error: nil,
				},
			},
			expectedResult: []string{"bundler", "minitest", "power_assert", "rake", "test-unit"},
			expectError:    false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"gem list --local --no-versions": {
					Output: []byte(`*** LOCAL GEMS ***

`),
					Error: nil,
				},
			},
			expectedResult: []string{},
			expectError:    false,
		},
		{
			name: "command error",
			mockResponses: map[string]CommandResponse{
				"gem list --local --no-versions": {
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

			manager := NewGemManager()
			result, err := manager.ListInstalled(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if len(result) != len(tt.expectedResult) {
					t.Errorf("Expected %d gems but got %d: %v", len(tt.expectedResult), len(result), result)
				} else {
					for i, pkg := range tt.expectedResult {
						if result[i] != pkg {
							t.Errorf("Expected gem %s at index %d but got %s", pkg, i, result[i])
						}
					}
				}
			}
		})
	}
}

func TestGemManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name:        "gem is installed",
			packageName: "bundler",
			mockResponses: map[string]CommandResponse{
				"gem list --local bundler": {
					Output: []byte("bundler (2.4.7, 2.3.26)"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:        "gem not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"gem list --local nonexistent": {
					Output: []byte(""),
					Error:  nil,
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:        "exact name match",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"gem list --local test": {
					Output: []byte("test (1.0.0)"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:        "command error",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"gem list --local test": {
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

			manager := NewGemManager()
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

func TestGemManager_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockResponses  map[string]CommandResponse
		expectedResult []string
		expectError    bool
	}{
		{
			name:  "successful search",
			query: "rails",
			mockResponses: map[string]CommandResponse{
				"gem search rails": {
					Output: []byte(`rails (7.0.4.3)
rails-api (0.4.1)
rails-i18n (7.0.6)
rails_admin (3.1.1)`),
					Error: nil,
				},
			},
			expectedResult: []string{"rails", "rails-api", "rails-i18n", "rails_admin"},
			expectError:    false,
		},
		{
			name:  "no results",
			query: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"gem search nonexistent": {
					Output: []byte(""),
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
				"gem search test": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 2},
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

			manager := NewGemManager()
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

func TestGemManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult string
		expectError    bool
	}{
		{
			name:        "get version of installed gem",
			packageName: "bundler",
			mockResponses: map[string]CommandResponse{
				"gem list --local bundler": {
					Output: []byte("bundler (2.4.7, 2.3.26)"),
					Error:  nil,
				},
			},
			expectedResult: "2.4.7",
			expectError:    false,
		},
		{
			name:        "gem not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"gem list --local nonexistent": {
					Output: []byte(""),
					Error:  nil,
				},
			},
			expectedResult: "",
			expectError:    true,
		},
		{
			name:        "single version",
			packageName: "rake",
			mockResponses: map[string]CommandResponse{
				"gem list --local rake": {
					Output: []byte("rake (13.0.6)"),
					Error:  nil,
				},
			},
			expectedResult: "13.0.6",
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

			manager := NewGemManager()
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

func TestGemManager_Info(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult *PackageInfo
		expectError    bool
	}{
		{
			name:        "successful info for installed gem",
			packageName: "rails",
			mockResponses: map[string]CommandResponse{
				"gem list --local rails": {
					Output: []byte("rails (7.0.4.3)"),
					Error:  nil,
				},
				"gem specification rails": {
					Output: []byte(`--- !ruby/object:Gem::Specification
name: rails
version: !ruby/object:Gem::Version
  version: 7.0.4.3
summary: Full-stack web application framework.
homepage: https://rubyonrails.org`),
					Error: nil,
				},
				"gem dependency rails": {
					Output: []byte(`Gem rails-7.0.4.3
  actioncable (= 7.0.4.3)
  actionmailbox (= 7.0.4.3)
  actionmailer (= 7.0.4.3)`),
					Error: nil,
				},
			},
			expectedResult: &PackageInfo{
				Name:         "rails",
				Version:      "7.0.4.3",
				Description:  "Full-stack web application framework.",
				Homepage:     "https://rubyonrails.org",
				Dependencies: []string{"actioncable", "actionmailbox", "actionmailer"},
				Manager:      "gem",
				Installed:    true,
			},
			expectError: false,
		},
		{
			name:        "gem not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"gem list --local nonexistent": {
					Output: []byte(""),
					Error:  nil,
				},
				"gem specification nonexistent": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:        "info for non-installed gem",
			packageName: "sinatra",
			mockResponses: map[string]CommandResponse{
				"gem list --local sinatra": {
					Output: []byte(""),
					Error:  nil,
				},
				"gem specification sinatra": {
					Output: []byte(`--- !ruby/object:Gem::Specification
name: sinatra
version: !ruby/object:Gem::Version
  version: 3.0.5
summary: Classy web-development dressed in a DSL
homepage: http://sinatrarb.com/`),
					Error: nil,
				},
			},
			expectedResult: &PackageInfo{
				Name:        "sinatra",
				Version:     "3.0.5",
				Description: "Classy web-development dressed in a DSL",
				Homepage:    "http://sinatrarb.com/",
				Manager:     "gem",
				Installed:   false,
			},
			expectError: false,
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

			manager := NewGemManager()
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
