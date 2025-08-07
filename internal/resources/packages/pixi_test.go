// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"
)

func TestPixiManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name:   "standard pixi global list output",
			output: []byte("Global environments as specified in '/Users/user/.pixi/manifests/pixi-global.toml'\n└── hello: 2.12.2 \n    └─ exposes: hello"),
			want:   []string{"hello"},
		},
		{
			name:   "multiple environments",
			output: []byte("Global environments as specified in '/Users/user/.pixi/manifests/pixi-global.toml'\n└── hello: 2.12.2 \n    └─ exposes: hello\n└── ripgrep: 14.1.1\n    └─ exposes: rg"),
			want:   []string{"hello", "ripgrep"},
		},
		{
			name:   "no environments installed",
			output: []byte("No global environments found."),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "single environment with complex name",
			output: []byte("└── python-jupyter: 3.11.0\n    └─ exposes: python, jupyter"),
			want:   []string{"python-jupyter"},
		},
		{
			name:   "environment with version containing spaces",
			output: []byte("└── test-package: 1.2.3 \n    └─ exposes: test"),
			want:   []string{"test-package"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPixiManager()
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

func TestPixiManager_parseSearchOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard pixi search output",
			output: []byte(`ripgrep-14.1.1-h0ef69ab_1 (+ 1 build)
-------------------------------------

Name                ripgrep
Version             14.1.1
Build               h0ef69ab_1
Size                1373159`),
			want: []string{"ripgrep"},
		},
		{
			name: "search with complex package name",
			output: []byte(`python-jupyter-3.11.0-h1234567_0
-------------------------------------

Name                python-jupyter
Version             3.11.0`),
			want: []string{"python-jupyter"},
		},
		{
			name:   "empty search results",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "no results found",
			output: []byte("No packages found matching query."),
			want:   []string{},
		},
		{
			name: "package with underscores",
			output: []byte(`test_package-1.0.0-py311_0
-------------------------------------

Name                test_package`),
			want: []string{"test_package"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPixiManager()
			got := manager.parseSearchOutput(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
			} else {
				for i, expected := range tt.want {
					if i >= len(got) || got[i] != expected {
						t.Errorf("parseSearchOutput() = %v, want %v", got, tt.want)
						break
					}
				}
			}
		})
	}
}

func TestPixiManager_extractPackageName(t *testing.T) {
	tests := []struct {
		name        string
		packageInfo string
		want        string
	}{
		{
			name:        "standard package with version",
			packageInfo: "ripgrep-14.1.1-h0ef69ab_1",
			want:        "ripgrep",
		},
		{
			name:        "package with underscores",
			packageInfo: "python_package-3.11.0-py311_0",
			want:        "python_package",
		},
		{
			name:        "complex package name",
			packageInfo: "jupyter-notebook-6.5.2-pyh6c4a22f_0",
			want:        "jupyter-notebook",
		},
		{
			name:        "simple package name",
			packageInfo: "hello-2.12.2-h0e07e94_0",
			want:        "hello",
		},
		{
			name:        "package without version pattern",
			packageInfo: "simple-package",
			want:        "simple",
		},
		{
			name:        "single word package",
			packageInfo: "git",
			want:        "git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPixiManager()
			got := manager.extractPackageName(tt.packageInfo)
			if got != tt.want {
				t.Errorf("extractPackageName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPixiManager_SupportsSearch(t *testing.T) {
	manager := NewPixiManager()
	if !manager.SupportsSearch() {
		t.Errorf("SupportsSearch() = false, want true - pixi supports search")
	}
}

func TestPixiManager_handleInstallError(t *testing.T) {
	manager := NewPixiManager()

	tests := []struct {
		name         string
		output       []byte
		packageName  string
		exitCode     int
		wantContains string
	}{
		{
			name:         "package not found",
			output:       []byte("No candidates were found for nonexistent"),
			packageName:  "nonexistent",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "dependency resolution failure",
			output:       []byte("failed to solve the environment"),
			packageName:  "testpkg",
			exitCode:     1,
			wantContains: "resolve dependencies",
		},
		{
			name:         "cannot solve request",
			output:       []byte("Cannot solve the request because of: No candidates"),
			packageName:  "missing",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "permission denied",
			output:       []byte("Permission denied accessing directory"),
			packageName:  "testpkg",
			exitCode:     1,
			wantContains: "permission denied",
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

func TestPixiManager_handleUninstallError(t *testing.T) {
	manager := NewPixiManager()

	tests := []struct {
		name            string
		output          []byte
		environmentName string
		exitCode        int
		wantErr         bool
	}{
		{
			name:            "environment not found",
			output:          []byte("No environment named 'notfound' exists"),
			environmentName: "notfound",
			exitCode:        1,
			wantErr:         false, // Should return nil - not an error for uninstall
		},
		{
			name:            "environment does not exist",
			output:          []byte("does not exist"),
			environmentName: "missing",
			exitCode:        1,
			wantErr:         false, // Should return nil - not an error for uninstall
		},
		{
			name:            "permission denied",
			output:          []byte("Permission denied accessing environment directory"),
			environmentName: "testenv",
			exitCode:        1,
			wantErr:         true,
		},
		{
			name:            "generic error",
			output:          []byte("Some uninstallation error"),
			environmentName: "testenv",
			exitCode:        2,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleUninstallError(mockErr, tt.output, tt.environmentName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleUninstallError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Note: stringContains is defined in test_helpers.go
