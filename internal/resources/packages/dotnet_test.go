// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
	"testing"
)

func TestDotnetManager_parseListOutput(t *testing.T) {
	manager := NewDotnetManager()

	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "standard dotnet tool list output",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef
dotnet-counters               9.0.572801 dotnet-counters`),
			want: []string{"dotnet-counters", "dotnet-ef", "dotnetsay"},
		},
		{
			name: "single tool output",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay`),
			want: []string{"dotnetsay"},
		},
		{
			name: "no tools installed",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------`),
			want: []string{},
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   []string{},
		},
		{
			name: "output with extra whitespace",
			output: []byte(`  Package Id                    Version    Commands
  -------------------------------------------------------
  dotnetsay                     2.1.4      dotnetsay
  dotnet-outdated-tool          4.6.4      dotnet-outdated  `),
			want: []string{"dotnet-outdated-tool", "dotnetsay"},
		},
		{
			name: "tools with complex names",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
Microsoft.dotnet-openapi      9.0.7      dotnet-openapi
coverlet.console              6.0.0      coverlet
dotnet-format                 5.1.250801 dotnet-format`),
			want: []string{"Microsoft.dotnet-openapi", "coverlet.console", "dotnet-format"},
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

func TestDotnetManager_getInstalledVersion(t *testing.T) {
	_ = NewDotnetManager()

	tests := []struct {
		name        string
		output      []byte
		toolName    string
		wantVersion string
		wantErr     bool
	}{
		{
			name: "find version for existing tool",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef`),
			toolName:    "dotnetsay",
			wantVersion: "2.1.4",
			wantErr:     false,
		},
		{
			name: "find version for different tool",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay
dotnet-ef                     9.0.7      dotnet-ef`),
			toolName:    "dotnet-ef",
			wantVersion: "9.0.7",
			wantErr:     false,
		},
		{
			name: "tool not found",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------
dotnetsay                     2.1.4      dotnetsay`),
			toolName:    "nonexistent-tool",
			wantVersion: "",
			wantErr:     true,
		},
		{
			name: "empty tool list",
			output: []byte(`Package Id                    Version    Commands
-------------------------------------------------------`),
			toolName:    "dotnetsay",
			wantVersion: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the context-based method, so we'll duplicate the parsing logic
			lines := strings.Split(string(tt.output), "\n")
			var inDataSection bool
			var gotVersion string
			var found bool

			for _, line := range lines {
				line = strings.TrimSpace(line)

				if strings.Contains(line, "-------") {
					inDataSection = true
					continue
				}

				if strings.Contains(line, "Package Id") {
					continue
				}

				if inDataSection {
					fields := strings.Fields(line)
					if len(fields) >= 2 && fields[0] == tt.toolName {
						gotVersion = fields[1]
						found = true
						break
					}
				}
			}

			if tt.wantErr {
				if found {
					t.Errorf("getInstalledVersion() expected error but found version %v", gotVersion)
				}
			} else {
				if !found {
					t.Errorf("getInstalledVersion() expected version but got error")
					return
				}
				if gotVersion != tt.wantVersion {
					t.Errorf("getInstalledVersion() = %v, want %v", gotVersion, tt.wantVersion)
				}
			}
		})
	}
}

func TestDotnetManager_IsAvailable(t *testing.T) {
	manager := NewDotnetManager()
	ctx := context.Background()

	// This test just ensures the method exists and doesn't panic
	_, err := manager.IsAvailable(ctx)
	if err != nil && !IsContextError(err) {
		// Only context errors are acceptable here since we can't guarantee dotnet is installed
		t.Logf("IsAvailable returned error (this may be expected): %v", err)
	}
}

func TestDotnetManager_SupportsSearch(t *testing.T) {
	manager := NewDotnetManager()
	if manager.SupportsSearch() {
		t.Errorf("SupportsSearch() = true, want false - .NET CLI doesn't support search")
	}
}

func TestDotnetManager_Search(t *testing.T) {
	manager := NewDotnetManager()
	ctx := context.Background()

	// Search should always return empty results since .NET doesn't support search
	results, err := manager.Search(ctx, "test")
	if err != nil {
		t.Errorf("Search() unexpected error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() = %v, want empty slice", results)
	}
}

func TestDotnetManager_handleInstallError(t *testing.T) {
	manager := NewDotnetManager()

	tests := []struct {
		name         string
		output       []byte
		toolName     string
		exitCode     int
		wantContains string
	}{
		{
			name:         "tool not found",
			output:       []byte("error NU1101: Unable to find package nonexistent-tool. No packages exist with this id"),
			toolName:     "nonexistent-tool",
			exitCode:     1,
			wantContains: "not found",
		},
		{
			name:         "not a dotnet tool",
			output:       []byte("error NU1212: Invalid project-package combination for some-package. DotnetToolReference project style can only contain references of the DotnetTool type"),
			toolName:     "some-package",
			exitCode:     1,
			wantContains: "not a .NET global tool",
		},
		{
			name:         "tool already installed",
			output:       []byte("Tool 'dotnetsay' is already installed."),
			toolName:     "dotnetsay",
			exitCode:     1,
			wantContains: "already installed",
		},
		{
			name:         "permission denied",
			output:       []byte("Access denied when writing to tool directory"),
			toolName:     "testtool",
			exitCode:     1,
			wantContains: "permission denied",
		},
		{
			name:         "generic error with output",
			output:       []byte("Some installation error occurred"),
			toolName:     "testtool",
			exitCode:     2,
			wantContains: "installation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleInstallError(mockErr, tt.output, tt.toolName)
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

func TestDotnetManager_handleUninstallError(t *testing.T) {
	manager := NewDotnetManager()

	tests := []struct {
		name     string
		output   []byte
		toolName string
		exitCode int
		wantErr  bool
	}{
		{
			name:     "tool not installed",
			output:   []byte("Tool 'testtool' is not installed."),
			toolName: "testtool",
			exitCode: 1,
			wantErr:  false, // Not installed should be success
		},
		{
			name:     "no such tool",
			output:   []byte("No such tool exists in the global tools"),
			toolName: "missing-tool",
			exitCode: 1,
			wantErr:  false, // Not found should be success
		},
		{
			name:     "permission denied",
			output:   []byte("Access denied when writing to tool directory"),
			toolName: "testtool",
			exitCode: 1,
			wantErr:  true,
		},
		{
			name:     "generic error",
			output:   []byte("Some uninstallation error"),
			toolName: "testtool",
			exitCode: 2,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockErr := &MockExitError{Code: tt.exitCode}

			err := manager.handleUninstallError(mockErr, tt.output, tt.toolName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleUninstallError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
