// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
	"testing"
)

func TestNpmManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard npm list output",
			output: []byte(`{
  "dependencies": {
    "typescript": {
      "version": "4.9.4"
    },
    "eslint": {
      "version": "8.30.0"
    },
    "prettier": {
      "version": "2.8.1"
    }
  }
}`),
			want: []string{"eslint", "prettier", "typescript"},
		},
		{
			name:   "empty dependencies",
			output: []byte(`{"dependencies": {}}`),
			want:   []string{},
		},
		{
			name:   "no dependencies field",
			output: []byte(`{}`),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "single package",
			output: []byte(`{
  "dependencies": {
    "lodash": {
      "version": "4.17.21"
    }
  }
}`),
			want: []string{"lodash"},
		},
		{
			name: "packages with scoped names",
			output: []byte(`{
  "dependencies": {
    "@types/node": {
      "version": "18.11.18"
    },
    "@typescript-eslint/parser": {
      "version": "5.48.0"
    },
    "react": {
      "version": "18.2.0"
    }
  }
}`),
			want: []string{"@types/node", "@typescript-eslint/parser", "react"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewNpmManager()
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

func TestNpmManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard npm search output",
			output: []byte(`[
  {
    "name": "react",
    "version": "18.2.0",
    "description": "React is a JavaScript library for building user interfaces."
  },
  {
    "name": "react-dom",
    "version": "18.2.0",
    "description": "React package for working with the DOM."
  }
]`),
			want: []string{"react", "react-dom"},
		},
		{
			name:   "empty search results",
			output: []byte(`[]`),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "single result",
			output: []byte(`[
  {
    "name": "lodash",
    "version": "4.17.21",
    "description": "Lodash modular utilities."
  }
]`),
			want: []string{"lodash"},
		},
		{
			name: "scoped packages",
			output: []byte(`[
  {
    "name": "@types/node",
    "version": "18.11.18",
    "description": "TypeScript definitions for Node.js"
  },
  {
    "name": "@typescript-eslint/parser",
    "version": "5.48.0",
    "description": "An ESLint custom parser for TypeScript"
  }
]`),
			want: []string{"@types/node", "@typescript-eslint/parser"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewNpmManager()
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

func TestNpmManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        *PackageInfo
	}{
		{
			name: "standard npm view output",
			output: []byte(`{
  "name": "react",
  "version": "18.2.0",
  "description": "React is a JavaScript library for building user interfaces.",
  "homepage": "https://reactjs.org/"
}`),
			packageName: "react",
			want: &PackageInfo{
				Name:        "react",
				Version:     "18.2.0",
				Description: "React is a JavaScript library for building user interfaces.",
				Homepage:    "https://reactjs.org/",
			},
		},
		{
			name: "minimal npm view output",
			output: []byte(`{
  "name": "lodash",
  "version": "4.17.21"
}`),
			packageName: "lodash",
			want: &PackageInfo{
				Name:        "lodash",
				Version:     "4.17.21",
				Description: "",
				Homepage:    "",
			},
		},
		{
			name: "scoped package",
			output: []byte(`{
  "name": "@types/node",
  "version": "18.11.18",
  "description": "TypeScript definitions for Node.js",
  "homepage": "https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node"
}`),
			packageName: "@types/node",
			want: &PackageInfo{
				Name:        "@types/node",
				Version:     "18.11.18",
				Description: "TypeScript definitions for Node.js",
				Homepage:    "https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node",
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
			manager := NewNpmManager()
			got := manager.parseInfoOutput(tt.output, tt.packageName)
			if !equalPackageInfo(got, tt.want) {
				t.Errorf("parseInfoOutput() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestNpmManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name: "npm is available",
			mockResponses: map[string]CommandResponse{
				"npm --version": {
					Output: []byte("9.2.0"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:          "npm not found",
			mockResponses: map[string]CommandResponse{
				// Empty responses means LookPath will fail
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "npm exists but not functional",
			mockResponses: map[string]CommandResponse{
				"npm": {}, // Makes LookPath succeed
				"npm --version": {
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

			manager := NewNpmManager()
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

func TestNpmManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"npm install -g typescript": {
					Output: []byte("added 1 package in 2s"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"npm install -g nonexistent": {
					Output: []byte("npm ERR! 404 Not Found"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "already installed",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"npm install -g typescript": {
					Output: []byte("up to date in 1s"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "permission denied",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"npm install -g test": {
					Output: []byte("npm ERR! Error: EACCES: permission denied"),
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
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewNpmManager()
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

func TestNpmManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"npm uninstall -g typescript": {
					Output: []byte("removed 1 package in 1s"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"npm uninstall -g nonexistent": {
					Output: []byte("npm ERR! npm uninstall: nonexistent not installed"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Not installed is success for uninstall
		},
		{
			name:        "uninstall with dependencies",
			packageName: "eslint",
			mockResponses: map[string]CommandResponse{
				"npm uninstall -g eslint": {
					Output: []byte("removed 100 packages in 5s"),
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

			manager := NewNpmManager()
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

func TestNpmManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult []string
		expectError    bool
	}{
		{
			name: "list with packages",
			mockResponses: map[string]CommandResponse{
				"npm list -g --depth=0 --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "4.9.4"},
							"eslint": {"version": "8.30.0"},
							"prettier": {"version": "2.8.1"}
						}
					}`),
					Error: nil,
				},
			},
			expectedResult: []string{"eslint", "prettier", "typescript"},
			expectError:    false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"npm list -g --depth=0 --json": {
					Output: []byte(`{"dependencies": {}}`),
					Error:  nil,
				},
			},
			expectedResult: []string{},
			expectError:    false,
		},
		{
			name: "list with warnings (exit code 1)",
			mockResponses: map[string]CommandResponse{
				"npm list -g --depth=0 --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "4.9.4"}
						}
					}`),
					Error: &MockExitError{Code: 1}, // npm uses exit 1 for warnings
				},
			},
			expectedResult: []string{"typescript"},
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

			manager := NewNpmManager()
			result, err := manager.ListInstalled(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) != len(tt.expectedResult) {
				t.Errorf("Expected %d packages but got %d: %v", len(tt.expectedResult), len(result), result)
			} else {
				for i, pkg := range tt.expectedResult {
					if result[i] != pkg {
						t.Errorf("Expected package %s at index %d but got %s", pkg, i, result[i])
					}
				}
			}
		})
	}
}

func TestNpmManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name:        "package is installed",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"npm list -g typescript": {
					Output: []byte("typescript@4.9.4"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"npm list -g nonexistent": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:        "scoped package installed",
			packageName: "@types/node",
			mockResponses: map[string]CommandResponse{
				"npm list -g @types/node": {
					Output: []byte("@types/node@18.11.18"),
					Error:  nil,
				},
			},
			expectedResult: true,
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

			manager := NewNpmManager()
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

func TestNpmManager_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockResponses  map[string]CommandResponse
		expectedResult []string
		expectError    bool
	}{
		{
			name:  "successful search",
			query: "react",
			mockResponses: map[string]CommandResponse{
				"npm search react --json": {
					Output: []byte(`[
						{"name": "react", "version": "18.2.0"},
						{"name": "react-dom", "version": "18.2.0"},
						{"name": "react-router", "version": "6.8.0"}
					]`),
					Error: nil,
				},
			},
			expectedResult: []string{"react", "react-dom", "react-router"},
			expectError:    false,
		},
		{
			name:  "no results",
			query: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"npm search nonexistent --json": {
					Output: []byte("[]"),
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
				"npm search test --json": {
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

			manager := NewNpmManager()
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

func TestNpmManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult string
		expectError    bool
	}{
		{
			name:        "get version of installed package",
			packageName: "typescript",
			mockResponses: map[string]CommandResponse{
				"npm list -g typescript": {
					Output: []byte("typescript@4.9.4"),
					Error:  nil,
				},
				"npm list -g typescript --depth=0 --json": {
					Output: []byte(`{
						"dependencies": {
							"typescript": {"version": "4.9.4"}
						}
					}`),
					Error: nil,
				},
			},
			expectedResult: "4.9.4",
			expectError:    false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"npm list -g nonexistent": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: "",
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

			manager := NewNpmManager()
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
