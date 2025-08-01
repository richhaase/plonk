// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
	"testing"
)

func TestPipManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard pip list JSON output",
			output: []byte(`[
  {
    "name": "numpy",
    "version": "1.24.1"
  },
  {
    "name": "pandas",
    "version": "1.5.2"
  },
  {
    "name": "Django",
    "version": "4.1.5"
  }
]`),
			want: []string{"numpy", "pandas", "django"},
		},
		{
			name:   "empty list",
			output: []byte(`[]`),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "single package",
			output: []byte(`[
  {
    "name": "requests",
    "version": "2.28.1"
  }
]`),
			want: []string{"requests"},
		},
		{
			name: "packages with hyphen and underscore",
			output: []byte(`[
  {
    "name": "scikit-learn",
    "version": "1.2.0"
  },
  {
    "name": "python_dateutil",
    "version": "2.8.2"
  }
]`),
			want: []string{"scikit-learn", "python_dateutil"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPipManager()
			got, err := manager.parseListOutput(tt.output)
			if err != nil {
				t.Errorf("parseListOutput() error = %v", err)
				return
			}
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

func TestPipManager_parseListOutputPlainText(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard pip list plain text output",
			output: []byte(`Package                Version
---------------------- --------
numpy                  1.24.1
pandas                 1.5.2
Django                 4.1.5`),
			want: []string{"numpy", "pandas", "django"},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "output with separator line",
			output: []byte(`Package    Version
---------- -------
requests   2.28.1`),
			want: []string{"requests"},
		},
		{
			name: "output with extra whitespace",
			output: []byte(`Package                  Version
------------------------ -----------
scikit-learn             1.2.0
python-dateutil          2.8.2`),
			want: []string{"scikit-learn", "python-dateutil"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPipManager()
			got := manager.parseListOutputPlainText(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseListOutputPlainText() = %v, want %v", got, tt.want)
			} else {
				for i, expected := range tt.want {
					if got[i] != expected {
						t.Errorf("parseListOutputPlainText() = %v, want %v", got, tt.want)
						break
					}
				}
			}
		})
	}
}

func TestPipManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard pip search output",
			output: []byte(`django-rest-framework (0.1.0)  - alias.
djangorestframework (3.14.0)   - Web APIs for Django, made easy.
django-rest-framework-json-api (6.0.0) - JSON:API support for Django REST framework`),
			want: []string{"django-rest-framework", "djangorestframework", "django-rest-framework-json-api"},
		},
		{
			name:   "empty search results",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "single result",
			output: []byte(`requests (2.28.1) - Python HTTP for Humans.`),
			want:   []string{"requests"},
		},
		{
			name: "results with version numbers",
			output: []byte(`numpy (1.24.1)     - Fundamental package for array computing in Python
numpy-financial (1.0.0) - Simple financial functions
numpy-quaternion (2022.4.3) - Add built-in support for quaternions to NumPy`),
			want: []string{"numpy", "numpy-financial", "numpy-quaternion"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPipManager()
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

func TestPipManager_parseInfoOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		packageName string
		want        *PackageInfo
	}{
		{
			name: "standard pip show output",
			output: []byte(`Name: Django
Version: 4.1.5
Summary: A high-level Python web framework that encourages rapid development and clean, pragmatic design.
Home-page: https://www.djangoproject.com/
Author: Django Software Foundation
Author-email: foundation@djangoproject.com
License: BSD-3-Clause
Location: /usr/local/lib/python3.9/site-packages
Requires: asgiref, sqlparse
Required-by: djangorestframework`),
			packageName: "Django",
			want: &PackageInfo{
				Name:         "Django",
				Version:      "4.1.5",
				Description:  "A high-level Python web framework that encourages rapid development and clean, pragmatic design.",
				Homepage:     "https://www.djangoproject.com/",
				Dependencies: []string{"asgiref", "sqlparse"},
			},
		},
		{
			name: "minimal pip show output",
			output: []byte(`Name: requests
Version: 2.28.1
Summary: Python HTTP for Humans.`),
			packageName: "requests",
			want: &PackageInfo{
				Name:        "requests",
				Version:     "2.28.1",
				Description: "Python HTTP for Humans.",
				Homepage:    "",
			},
		},
		{
			name: "package with no dependencies",
			output: []byte(`Name: six
Version: 1.16.0
Summary: Python 2 and 3 compatibility utilities
Home-page: https://github.com/benjaminp/six
Requires:`),
			packageName: "six",
			want: &PackageInfo{
				Name:        "six",
				Version:     "1.16.0",
				Description: "Python 2 and 3 compatibility utilities",
				Homepage:    "https://github.com/benjaminp/six",
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
			manager := NewPipManager()
			got := manager.parseInfoOutput(tt.output, tt.packageName)
			if !equalPackageInfo(got, tt.want) {
				t.Errorf("parseInfoOutput() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPipManager_normalizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercase conversion",
			input: "Django",
			want:  "django",
		},
		{
			name:  "hyphen to underscore",
			input: "scikit-learn",
			want:  "scikit_learn",
		},
		{
			name:  "mixed case with hyphen",
			input: "Django-REST-framework",
			want:  "django_rest_framework",
		},
		{
			name:  "already normalized",
			input: "numpy",
			want:  "numpy",
		},
		{
			name:  "underscore preserved",
			input: "python_dateutil",
			want:  "python_dateutil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPipManager()
			got := manager.normalizeName(tt.input)
			if got != tt.want {
				t.Errorf("normalizeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name: "pip3 is available",
			mockResponses: map[string]CommandResponse{
				"pip3 --version": {
					Output: []byte("pip 22.3.1 from /usr/lib/python3/dist-packages/pip (python 3.10)"),
					Error:  nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:          "pip3 not found",
			mockResponses: map[string]CommandResponse{
				// Empty responses means LookPath will fail
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "pip3 exists but not functional",
			mockResponses: map[string]CommandResponse{
				"pip3": {}, // Makes LookPath succeed
				"pip3 --version": {
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

			manager := NewPipManager()
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

func TestPipManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "requests",
			mockResponses: map[string]CommandResponse{
				"pip3 install --user --break-system-packages requests": {
					Output: []byte("Successfully installed requests-2.28.1"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pip3 install --user --break-system-packages nonexistent": {
					Output: []byte("ERROR: Could not find a version that satisfies the requirement nonexistent"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "already installed",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"pip3 install --user --break-system-packages numpy": {
					Output: []byte("Requirement already satisfied: numpy in /usr/local/lib/python3.9/site-packages"),
					Error:  &MockExitError{Code: 0},
				},
			},
			expectError: false,
		},
		{
			name:        "permission denied",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"pip3 install --user --break-system-packages test": {
					Output: []byte("ERROR: Could not install packages due to an OSError: [Errno 13] Permission denied"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "permission denied",
		},
		{
			name:        "fallback to no --user flag",
			packageName: "requests",
			mockResponses: map[string]CommandResponse{
				"pip3 install --user --break-system-packages requests": {
					Output: []byte("ERROR: Can not perform a '--user' install"),
					Error:  &MockExitError{Code: 1},
				},
				"pip3 install --break-system-packages requests": {
					Output: []byte("Successfully installed requests-2.28.1"),
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

			manager := NewPipManager()
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

func TestPipManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "requests",
			mockResponses: map[string]CommandResponse{
				"pip3 uninstall -y --break-system-packages requests": {
					Output: []byte("Successfully uninstalled requests-2.28.1"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pip3 uninstall -y --break-system-packages nonexistent": {
					Output: []byte("WARNING: Skipping nonexistent as it is not installed."),
					Error:  &MockExitError{Code: 0},
				},
			},
			expectError: false, // Not installed is success for uninstall
		},
		{
			name:        "permission denied",
			packageName: "test",
			mockResponses: map[string]CommandResponse{
				"pip3 uninstall -y --break-system-packages test": {
					Output: []byte("ERROR: Cannot uninstall 'test'. It is a distutils installed project"),
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
			defer func() { defaultExecutor = originalExecutor }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := NewPipManager()
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

func TestPipManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  map[string]CommandResponse
		expectedResult []string
		expectError    bool
	}{
		{
			name: "list with JSON output",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[
						{"name": "numpy", "version": "1.24.1"},
						{"name": "pandas", "version": "1.5.2"},
						{"name": "Django", "version": "4.1.5"}
					]`),
					Error: nil,
				},
			},
			expectedResult: []string{"numpy", "pandas", "django"},
			expectError:    false,
		},
		{
			name: "empty list",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[]`),
					Error:  nil,
				},
			},
			expectedResult: []string{},
			expectError:    false,
		},
		{
			name: "fallback to plain text",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte("pip: error: no such option: --format"),
					Error:  &MockExitError{Code: 2},
				},
				"pip3 list --user": {
					Output: []byte(`Package    Version
---------- -------
numpy      1.24.1
pandas     1.5.2`),
					Error: nil,
				},
			},
			expectedResult: []string{"numpy", "pandas"},
			expectError:    false,
		},
		{
			name: "fallback to no --user flag",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte("pip: error: unknown option: --user"),
					Error:  &MockExitError{Code: 2},
				},
				"pip3 list --user": {
					Output: []byte("pip: error: unknown option: --user"),
					Error:  &MockExitError{Code: 2},
				},
				"pip3 list": {
					Output: []byte(`Package    Version
---------- -------
numpy      1.24.1`),
					Error: nil,
				},
			},
			expectedResult: []string{"numpy"},
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

			manager := NewPipManager()
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

func TestPipManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult bool
		expectError    bool
	}{
		{
			name:        "package is installed",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[
						{"name": "numpy", "version": "1.24.1"},
						{"name": "pandas", "version": "1.5.2"}
					]`),
					Error: nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:        "package not installed",
			packageName: "requests",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[
						{"name": "numpy", "version": "1.24.1"},
						{"name": "pandas", "version": "1.5.2"}
					]`),
					Error: nil,
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:        "normalized name match",
			packageName: "scikit-learn",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[
						{"name": "scikit_learn", "version": "1.2.0"}
					]`),
					Error: nil,
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:        "case insensitive match",
			packageName: "Django",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[
						{"name": "django", "version": "4.1.5"}
					]`),
					Error: nil,
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

			manager := NewPipManager()
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

func TestPipManager_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockResponses  map[string]CommandResponse
		expectedResult []string
		expectError    bool
		errorContains  string
	}{
		{
			name:  "successful search",
			query: "django",
			mockResponses: map[string]CommandResponse{
				"pip3 search django": {
					Output: []byte(`django-rest-framework (0.1.0)  - alias.
djangorestframework (3.14.0)   - Web APIs for Django, made easy.
django-rest-framework-json-api (6.0.0) - JSON:API support for Django REST framework`),
					Error: nil,
				},
			},
			expectedResult: []string{"django-rest-framework", "djangorestframework", "django-rest-framework-json-api"},
			expectError:    false,
		},
		{
			name:  "no results",
			query: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pip3 search nonexistent": {
					Output: []byte(""),
					Error:  nil,
				},
			},
			expectedResult: []string{},
			expectError:    false,
		},
		{
			name:  "search API disabled",
			query: "test",
			mockResponses: map[string]CommandResponse{
				"pip3 search test": {
					Output: []byte("ERROR: XMLRPC request failed [code: -32500]"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectedResult: nil,
			expectError:    true,
			errorContains:  "PyPI search API is currently disabled",
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

			manager := NewPipManager()
			result, err := manager.Search(context.Background(), tt.query)

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
			if !tt.expectError && !stringSlicesEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestPipManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult string
		expectError    bool
	}{
		{
			name:        "get version of installed package",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[
						{"name": "numpy", "version": "1.24.1"},
						{"name": "pandas", "version": "1.5.2"}
					]`),
					Error: nil,
				},
				"pip3 show numpy": {
					Output: []byte(`Name: numpy
Version: 1.24.1
Summary: Fundamental package for array computing in Python
Home-page: https://numpy.org`),
					Error: nil,
				},
			},
			expectedResult: "1.24.1",
			expectError:    false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[]`),
					Error:  nil,
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

			manager := NewPipManager()
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

func TestPipManager_Info(t *testing.T) {
	tests := []struct {
		name           string
		packageName    string
		mockResponses  map[string]CommandResponse
		expectedResult *PackageInfo
		expectError    bool
	}{
		{
			name:        "successful info for installed package",
			packageName: "Django",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[{"name": "Django", "version": "4.1.5"}]`),
					Error:  nil,
				},
				"pip3 show Django": {
					Output: []byte(`Name: Django
Version: 4.1.5
Summary: A high-level Python web framework that encourages rapid development and clean, pragmatic design.
Home-page: https://www.djangoproject.com/
Requires: asgiref, sqlparse`),
					Error: nil,
				},
			},
			expectedResult: &PackageInfo{
				Name:         "Django",
				Version:      "4.1.5",
				Description:  "A high-level Python web framework that encourages rapid development and clean, pragmatic design.",
				Homepage:     "https://www.djangoproject.com/",
				Dependencies: []string{"asgiref", "sqlparse"},
				Manager:      "pip",
				Installed:    true,
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"pip3 list --user --format=json": {
					Output: []byte(`[]`),
					Error:  nil,
				},
				"pip3 show nonexistent": {
					Output: []byte("WARNING: Package(s) not found: nonexistent"),
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

			manager := NewPipManager()
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

func TestPipManager_extractVersion(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   string
	}{
		{
			name: "standard version format",
			output: []byte(`Name: numpy
Version: 1.24.1
Summary: Fundamental package for array computing in Python`),
			want: "1.24.1",
		},
		{
			name: "version with spaces",
			output: []byte(`Name: django
Version:    4.1.5
Summary: A high-level Python web framework`),
			want: "4.1.5",
		},
		{
			name:   "no version found",
			output: []byte(`Name: test`),
			want:   "",
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPipManager()
			got := manager.extractVersion(tt.output)
			if got != tt.want {
				t.Errorf("extractVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
