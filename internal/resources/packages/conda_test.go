// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
)

func TestCondaManager_ListInstalled(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name: "list with packages",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[
						{
							"name": "numpy",
							"version": "1.24.3",
							"build": "py311h08b1b3b_0",
							"channel": "conda-forge"
						},
						{
							"name": "pandas",
							"version": "2.0.3",
							"build": "py311hd9cd6c9_0",
							"channel": "conda-forge"
						}
					]`),
					Error: nil,
				},
			},
			expected:    []string{"numpy", "pandas"}, // sorted
			expectError: false,
		},
		{
			name: "empty environment",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[]`),
					Error:  nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name: "command error",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 1},
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

			manager := &CondaManager{binary: "mamba", useMamba: true}
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

func TestCondaManager_Install(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful install",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"mamba install -n base -y numpy": {
					Output: []byte("Collecting package metadata (current_repodata.json): done\nSolving environment: done\n\n==> WARNING: A newer version of conda exists. <==\n\n## Package Plan ##\n\n  environment location: /opt/conda\n\n  added / updated specs:\n    - numpy\n\n\nThe following packages will be downloaded:\n\n    package                    |            build\n    ---------------------------|-----------------\n    numpy-1.24.3               |   py311h08b1b3b_0         6.6 MB  conda-forge\n    ------------------------------------------------------------\n                                           Total:         6.6 MB\n\nThe following NEW packages will be INSTALLED:\n\n  numpy              conda-forge/linux-64::numpy-1.24.3-py311h08b1b3b_0\n\n\n\nDownloading and Extracting Packages\n                                                                                \rnumpy-1.24.3         | 6.6 MB    | ############################################# | 100% \n\nPreparing transaction: done\nVerifying transaction: done\nExecuting transaction: done"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"mamba install -n base -y nonexistent": {
					Output: []byte("PackageNotFoundError: The following packages are not available from current channels:\n\n  - nonexistent\n\nCurrent channels:\n\n  - https://conda.anaconda.org/conda-forge/linux-64\n  - https://conda.anaconda.org/conda-forge/noarch\n  - https://repo.anaconda.com/pkgs/main/linux-64\n  - https://repo.anaconda.com/pkgs/main/noarch\n  - https://repo.anaconda.com/pkgs/r/linux-64\n  - https://repo.anaconda.com/pkgs/r/noarch\n\nTo search for alternate channels that may provide the conda package you're\nlooking for, navigate to\n\n    https://anaconda.org\n\nand use the search bar at the top of the page."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "already installed",
			packageName: "python",
			mockResponses: map[string]CommandResponse{
				"mamba install -n base -y python": {
					Output: []byte("# All requested packages already installed."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Should return nil for already installed
		},
		{
			name:        "dependency conflict",
			packageName: "conflicting-package",
			mockResponses: map[string]CommandResponse{
				"mamba install -n base -y conflicting-package": {
					Output: []byte("UnsatisfiableError: The following specifications were found to be incompatible with the existing python installation in your environment:\n\nspecifications:\n\n  - conflicting-package -> python[version='>=3.9,<3.10.0a0']\n\nYour python: python=3.11\n\nIf python is on the left-most side of the chain, that's the version you've asked for.\nWhen python appears to the right, that indicates that the thing on the left is somehow\nnot available for the python version you are constrained to. Note that conda will not\nchange your python version to a different minor version unless you explicitly specify\nthat."),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "dependency conflicts",
		},
		{
			name:        "network error",
			packageName: "matplotlib",
			mockResponses: map[string]CommandResponse{
				"mamba install -n base -y matplotlib": {
					Output: []byte("CondaHTTPError: HTTP 000 CONNECTION FAILED for url <https://conda.anaconda.org/conda-forge/linux-64/repodata.json>\nElapsed: -\n\nAn HTTP error occurred when trying to retrieve this URL.\nHTTP errors are often intermittent, and a simple retry will get you on your way."),
					Error:  &MockExitError{Code: 1},
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
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := &CondaManager{binary: "mamba", useMamba: true}
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

func TestCondaManager_Uninstall(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful uninstall",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"mamba remove -n base -y numpy": {
					Output: []byte("Collecting package metadata (repodata.json): done\nSolving environment: done\n\n## Package Plan ##\n\n  environment location: /opt/conda\n\n  removed specs:\n    - numpy\n\n\nThe following packages will be REMOVED:\n\n  numpy-1.24.3-py311h08b1b3b_0\n\n\nPreparing transaction: done\nVerifying transaction: done\nExecuting transaction: done"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"mamba remove -n base -y nonexistent": {
					Output: []byte("PackageNotInstalledError: Package is not installed in environment"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Should return nil for not installed
		},
		{
			name:        "environment locked",
			packageName: "python",
			mockResponses: map[string]CommandResponse{
				"mamba remove -n base -y python": {
					Output: []byte("EnvironmentLockedError: The environment is locked and cannot be modified"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "locked",
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

			manager := &CondaManager{binary: "mamba", useMamba: true}
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

func TestCondaManager_IsInstalled(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      bool
		expectError   bool
	}{
		{
			name:        "package is installed",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[
						{
							"name": "numpy",
							"version": "1.24.3",
							"build": "py311h08b1b3b_0",
							"channel": "conda-forge"
						},
						{
							"name": "pandas",
							"version": "2.0.3",
							"build": "py311hd9cd6c9_0",
							"channel": "conda-forge"
						}
					]`),
					Error: nil,
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[
						{
							"name": "numpy",
							"version": "1.24.3",
							"build": "py311h08b1b3b_0",
							"channel": "conda-forge"
						}
					]`),
					Error: nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "empty environment",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[]`),
					Error:  nil,
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:        "command error",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 1},
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

			manager := &CondaManager{binary: "mamba", useMamba: true}
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

func TestCondaManager_InstalledVersion(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expected      string
		expectError   bool
	}{
		{
			name:        "get version of installed package",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[
						{
							"name": "numpy",
							"version": "1.24.3",
							"build": "py311h08b1b3b_0",
							"channel": "conda-forge"
						},
						{
							"name": "pandas",
							"version": "2.0.3",
							"build": "py311hd9cd6c9_0",
							"channel": "conda-forge"
						}
					]`),
					Error: nil,
				},
			},
			expected:    "1.24.3",
			expectError: false,
		},
		{
			name:        "package not installed",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[
						{
							"name": "numpy",
							"version": "1.24.3",
							"build": "py311h08b1b3b_0",
							"channel": "conda-forge"
						}
					]`),
					Error: nil,
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name:        "command error on first call",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte("error: command failed"),
					Error:  &MockExitError{Code: 1},
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

			manager := &CondaManager{binary: "mamba", useMamba: true}
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

func TestCondaManager_IsAvailable(t *testing.T) {
	tests := []struct {
		name          string
		binary        string
		mockResponses map[string]CommandResponse
		expected      bool
	}{
		{
			name:   "mamba is available",
			binary: "mamba",
			mockResponses: map[string]CommandResponse{
				"mamba --version": {
					Output: []byte("mamba 1.5.1"),
					Error:  nil,
				},
			},
			expected: true,
		},
		{
			name:   "conda is available",
			binary: "conda",
			mockResponses: map[string]CommandResponse{
				"conda --version": {
					Output: []byte("conda 23.7.4"),
					Error:  nil,
				},
			},
			expected: true,
		},
		{
			name:   "binary not found",
			binary: "conda",
			mockResponses: map[string]CommandResponse{
				"conda --version": {
					Output: []byte(""),
					Error:  &MockExitError{Code: 127},
				},
			},
			expected: false,
		},
		{
			name:   "binary exists but not functional",
			binary: "conda",
			mockResponses: map[string]CommandResponse{
				"conda --version": {
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

			manager := &CondaManager{binary: tt.binary}
			result, _ := manager.IsAvailable(context.Background())

			if result != tt.expected {
				t.Errorf("Expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestCondaManager_Search(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		mockResponses map[string]CommandResponse
		expected      []string
		expectError   bool
	}{
		{
			name:  "successful search",
			query: "numpy",
			mockResponses: map[string]CommandResponse{
				"mamba search numpy --json": {
					Output: []byte(`{
						"numpy": [
							{
								"name": "numpy",
								"version": "1.24.3",
								"build": "py311h08b1b3b_0",
								"channel": "conda-forge"
							}
						],
						"numpy-base": [
							{
								"name": "numpy-base",
								"version": "1.24.3",
								"build": "py311h973a644_0",
								"channel": "conda-forge"
							}
						]
					}`),
					Error: nil,
				},
			},
			expected:    []string{"numpy", "numpy-base"},
			expectError: false,
		},
		{
			name:  "no results found",
			query: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"mamba search nonexistent --json": {
					Output: []byte("PackagesNotFoundError: The following packages are not available from current channels:\n\n  - nonexistent"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expected:    []string{},
			expectError: false,
		},
		{
			name:  "empty search results",
			query: "empty",
			mockResponses: map[string]CommandResponse{
				"mamba search empty --json": {
					Output: []byte(`{}`),
					Error:  nil,
				},
			},
			expected:    []string{},
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

			manager := &CondaManager{binary: "mamba", useMamba: true}
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

func TestCondaManager_Info(t *testing.T) {
	tests := []struct {
		name          string
		packageName   string
		mockResponses map[string]CommandResponse
		expectError   bool
		expectInfo    bool
	}{
		{
			name:        "info for installed package",
			packageName: "numpy",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[
						{
							"name": "numpy",
							"version": "1.24.3",
							"build": "py311h08b1b3b_0",
							"channel": "conda-forge"
						}
					]`),
					Error: nil,
				},
				"mamba search numpy --info --json": {
					Output: []byte(`{
						"numpy": [
							{
								"name": "numpy",
								"version": "1.24.3",
								"build": "py311h08b1b3b_0",
								"channel": "conda-forge",
								"depends": ["python >=3.11,<3.12.0a0", "libblas >=3.9.0,<4.0a0"],
								"license": "BSD-3-Clause",
								"summary": "The fundamental package for scientific computing with Python",
								"description": "NumPy is the fundamental package for array computing with Python.",
								"home": "https://www.numpy.org/",
								"size": 6945542
							}
						]
					}`),
					Error: nil,
				},
			},
			expectError: false,
			expectInfo:  true,
		},
		{
			name:        "info for non-installed package",
			packageName: "pandas",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[]`),
					Error:  nil,
				},
				"mamba search pandas --info --json": {
					Output: []byte(`{
						"pandas": [
							{
								"name": "pandas",
								"version": "2.0.3",
								"build": "py311hd9cd6c9_0",
								"channel": "conda-forge",
								"depends": ["numpy >=1.21.0", "python >=3.11,<3.12.0a0"],
								"license": "BSD-3-Clause",
								"summary": "Powerful data structures for data analysis, time series, and statistics",
								"description": "pandas is a fast, powerful, flexible and easy to use open source data analysis and manipulation tool",
								"home": "https://pandas.pydata.org/",
								"size": 12234567
							}
						]
					}`),
					Error: nil,
				},
			},
			expectError: false,
			expectInfo:  true,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[]`),
					Error:  nil,
				},
				"mamba search nonexistent --info --json": {
					Output: []byte("PackagesNotFoundError: The following packages are not available from current channels:\n\n  - nonexistent"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: true,
			expectInfo:  false,
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

			manager := &CondaManager{binary: "mamba", useMamba: true}
			result, err := manager.Info(context.Background(), tt.packageName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectInfo && result == nil {
				t.Errorf("Expected PackageInfo but got nil")
			}
			if tt.expectInfo && result != nil {
				if result.Name != tt.packageName {
					t.Errorf("Expected name %s but got %s", tt.packageName, result.Name)
				}
				if result.Manager != "conda" {
					t.Errorf("Expected manager 'conda' but got '%s'", result.Manager)
				}
			}
		})
	}
}

func TestCondaManager_Upgrade(t *testing.T) {
	tests := []struct {
		name          string
		packages      []string
		mockResponses map[string]CommandResponse
		expectError   bool
		errorContains string
	}{
		{
			name:     "upgrade specific packages",
			packages: []string{"numpy", "pandas"},
			mockResponses: map[string]CommandResponse{
				"mamba update -n base -y numpy": {
					Output: []byte("# All requested packages already installed.\n\n## Package Plan ##\n\n  environment location: /opt/conda\n\nThe following packages will be downloaded:\n\n    package                    |            build\n    ---------------------------|-----------------\n    numpy-1.25.0               |   py311h08b1b3b_0         6.8 MB  conda-forge\n    ------------------------------------------------------------\n                                           Total:         6.8 MB\n\nThe following packages will be UPDATED:\n\n  numpy              conda-forge/linux-64::numpy-1.24.3-py311h08b1b3b_0 --> conda-forge/linux-64::numpy-1.25.0-py311h08b1b3b_0\n\n\nPreparing transaction: done\nVerifying transaction: done\nExecuting transaction: done"),
					Error:  nil,
				},
				"mamba update -n base -y pandas": {
					Output: []byte("Package upgraded successfully"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade all packages",
			packages: []string{}, // empty means all packages
			mockResponses: map[string]CommandResponse{
				"mamba list -n base --json": {
					Output: []byte(`[
						{
							"name": "numpy",
							"version": "1.24.3",
							"build": "py311h08b1b3b_0",
							"channel": "conda-forge"
						},
						{
							"name": "pandas",
							"version": "2.0.3",
							"build": "py311hd9cd6c9_0",
							"channel": "conda-forge"
						}
					]`),
					Error: nil,
				},
				"mamba update -n base -y numpy": {
					Output: []byte("Package upgraded successfully"),
					Error:  nil,
				},
				"mamba update -n base -y pandas": {
					Output: []byte("Package upgraded successfully"),
					Error:  nil,
				},
			},
			expectError: false,
		},
		{
			name:     "upgrade with package not found error",
			packages: []string{"nonexistent"},
			mockResponses: map[string]CommandResponse{
				"mamba update -n base -y nonexistent": {
					Output: []byte("PackageNotFoundError: The following packages are not available from current channels:\n\n  - nonexistent"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:     "upgrade already up-to-date",
			packages: []string{"numpy"},
			mockResponses: map[string]CommandResponse{
				"mamba update -n base -y numpy": {
					Output: []byte("# All requested packages already installed.\n"),
					Error:  &MockExitError{Code: 1},
				},
			},
			expectError: false, // Should return nil for already up-to-date
		},
		{
			name:     "upgrade with dependency conflict",
			packages: []string{"conflicting-package"},
			mockResponses: map[string]CommandResponse{
				"mamba update -n base -y conflicting-package": {
					Output: []byte("UnsatisfiableError: The following specifications were found to be incompatible with each other:\n\nOutput in format: Requested package -> Available versions"),
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
			defer func() { SetDefaultExecutor(originalExecutor) }()

			mock := &MockCommandExecutor{
				Responses: tt.mockResponses,
			}
			SetDefaultExecutor(mock)

			manager := &CondaManager{binary: "mamba", useMamba: true}
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

func TestCondaManager_DetectVariant(t *testing.T) {
	tests := []struct {
		name           string
		availableTools []string // Mock available binaries
		expectedBinary string
		expectedMamba  bool
	}{
		{
			name:           "mamba available",
			availableTools: []string{"mamba"},
			expectedBinary: "mamba",
			expectedMamba:  true,
		},
		{
			name:           "conda only",
			availableTools: []string{"conda"},
			expectedBinary: "conda",
			expectedMamba:  false,
		},
		{
			name:           "both available - prefer mamba",
			availableTools: []string{"mamba", "conda"},
			expectedBinary: "mamba",
			expectedMamba:  true,
		},
		{
			name:           "neither available",
			availableTools: []string{},
			expectedBinary: "conda",
			expectedMamba:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original executor
			originalExecutor := defaultExecutor
			defer func() { SetDefaultExecutor(originalExecutor) }()

			// Create mock responses for version checks
			mockResponses := make(map[string]CommandResponse)
			for _, tool := range tt.availableTools {
				mockResponses[tool+" --version"] = CommandResponse{
					Output: []byte(tool + " 1.0.0"),
					Error:  nil,
				}
			}

			mock := &MockCommandExecutor{
				Responses: mockResponses,
			}
			SetDefaultExecutor(mock)

			binary, useMamba := detectCondaVariant()

			if binary != tt.expectedBinary {
				t.Errorf("Expected binary %s but got %s", tt.expectedBinary, binary)
			}
			if useMamba != tt.expectedMamba {
				t.Errorf("Expected useMamba %v but got %v", tt.expectedMamba, useMamba)
			}
		})
	}
}

func TestCondaManager_Dependencies(t *testing.T) {
	manager := NewCondaManager()
	deps := manager.Dependencies()

	expected := []string{"brew"}
	if !stringSlicesEqual(deps, expected) {
		t.Errorf("Expected dependencies %v but got %v", expected, deps)
	}
}

// Note: Uses MockExitError from executor.go and stringContains/stringSlicesEqual from test_helpers.go
