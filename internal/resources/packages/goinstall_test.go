// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"
)

func TestParseModulePath(t *testing.T) {
	tests := []struct {
		name        string
		pkg         string
		wantModule  string
		wantVersion string
	}{
		{
			name:        "simple module path",
			pkg:         "github.com/user/repo",
			wantModule:  "github.com/user/repo",
			wantVersion: "latest",
		},
		{
			name:        "module path with version",
			pkg:         "github.com/user/repo@v1.2.3",
			wantModule:  "github.com/user/repo",
			wantVersion: "v1.2.3",
		},
		{
			name:        "module path with latest",
			pkg:         "github.com/user/repo@latest",
			wantModule:  "github.com/user/repo",
			wantVersion: "latest",
		},
		{
			name:        "module path with commit hash",
			pkg:         "github.com/user/repo@abc123",
			wantModule:  "github.com/user/repo",
			wantVersion: "abc123",
		},
		{
			name:        "module path with subpath",
			pkg:         "github.com/user/repo/cmd/tool@v1.0.0",
			wantModule:  "github.com/user/repo/cmd/tool",
			wantVersion: "v1.0.0",
		},
		{
			name:        "complex module path",
			pkg:         "go.uber.org/zap@v1.24.0",
			wantModule:  "go.uber.org/zap",
			wantVersion: "v1.24.0",
		},
		{
			name:        "module without version",
			pkg:         "golang.org/x/tools/cmd/goimports",
			wantModule:  "golang.org/x/tools/cmd/goimports",
			wantVersion: "latest",
		},
		{
			name:        "empty package",
			pkg:         "",
			wantModule:  "",
			wantVersion: "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModule, gotVersion := parseModulePath(tt.pkg)
			if gotModule != tt.wantModule {
				t.Errorf("parseModulePath() module = %v, want %v", gotModule, tt.wantModule)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("parseModulePath() version = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

func TestGoInstallManager_Configuration(t *testing.T) {
	manager := NewGoInstallManager()

	if manager.binary != "go" {
		t.Errorf("binary = %v, want go", manager.binary)
	}

	if manager.errorMatcher == nil {
		t.Error("errorMatcher not initialized")
	}
}

func TestGoInstallManager_SupportsSearch(t *testing.T) {
	manager := NewGoInstallManager()
	if manager.SupportsSearch() {
		t.Error("GoInstallManager should not support search")
	}
}
