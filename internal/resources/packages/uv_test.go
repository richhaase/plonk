// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"
)

func TestUvManager_parseListOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name:   "standard uv tool list output",
			output: []byte("cowsay v6.1\n- cowsay\nruff v0.1.0\n- ruff"),
			want:   []string{"cowsay", "ruff"},
		},
		{
			name:   "single tool",
			output: []byte("black v23.1.0\n- black"),
			want:   []string{"black"},
		},
		{
			name:   "no tools installed",
			output: []byte("No tools installed"),
			want:   []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name:   "tool with path info",
			output: []byte("pytest v7.2.0 (/home/user/.local/share/uv/tools/pytest)\n- pytest"),
			want:   []string{"pytest"},
		},
		{
			name:   "multiple tools with executables",
			output: []byte("cowsay v6.1\n- cowsay\n- cowthink\nruff v0.1.0\n- ruff\npytest v7.2.0\n- pytest\n- py.test"),
			want:   []string{"cowsay", "ruff", "pytest"},
		},
		{
			name:   "tools with complex names",
			output: []byte("black-formatter v1.0.0\n- black\npython-lsp-server v1.7.1\n- pylsp"),
			want:   []string{"black-formatter", "python-lsp-server"},
		},
		{
			name:   "tools with hyphenated names",
			output: []byte("pre-commit v2.20.0\n- pre-commit\nblack-formatter v1.0.0\n- black"),
			want:   []string{"pre-commit", "black-formatter"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewUvManager()
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

func TestUvManager_SupportsSearch(t *testing.T) {
	manager := NewUvManager()
	if manager.SupportsSearch() {
		t.Errorf("SupportsSearch() = true, want false - UV does not support search")
	}
}

func TestUvManager_handleInstallError(t *testing.T) {
	manager := NewUvManager()

	tests := []struct {
		name         string
		output       []byte
		packageName  string
		exitCode     int
		wantContains string
	}{
		{
			name:         "package not found",
			output:       []byte("No such package 'nonexistent'"),
			packageName:  "nonexistent",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "permission denied",
			output:       []byte("Permission denied accessing /usr/local"),
			packageName:  "testpkg",
			exitCode:     1,
			wantContains: "permission denied",
		},
		{
			name:         "404 error",
			output:       []byte("404: Package not found in registry"),
			packageName:  "missing",
			exitCode:     1,
			wantContains: "not found",
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

func TestUvManager_handleUninstallError(t *testing.T) {
	manager := NewUvManager()

	tests := []struct {
		name        string
		output      []byte
		packageName string
		exitCode    int
		wantErr     bool
	}{
		{
			name:        "tool not installed",
			output:      []byte("No tool named 'notinstalled' found"),
			packageName: "notinstalled",
			exitCode:    1,
			wantErr:     false, // Should return nil - not an error for uninstall
		},
		{
			name:        "tool not found",
			output:      []byte("not found"),
			packageName: "missing",
			exitCode:    1,
			wantErr:     false, // Should return nil - not an error for uninstall
		},
		{
			name:        "permission denied",
			output:      []byte("Permission denied accessing tool directory"),
			packageName: "testpkg",
			exitCode:    1,
			wantErr:     true,
		},
		{
			name:        "generic error",
			output:      []byte("Some uninstallation error"),
			packageName: "testpkg",
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

// Note: Uses MockExitError from executor.go and stringContains from test_helpers.go
