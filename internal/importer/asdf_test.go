// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"os"
	"path/filepath"
	"testing"

	"plonk/internal/managers"
)

func TestDiscoverAsdfPackages(t *testing.T) {
	tests := []struct {
		name         string
		toolVersions string
		expectedPkgs []string
		expectError  bool
	}{
		{
			name: "successful parsing of global .tool-versions",
			toolVersions: `nodejs 20.0.0
python 3.11.3
ruby 3.0.0
`,
			expectedPkgs: []string{
				"nodejs 20.0.0",
				"python 3.11.3",
				"ruby 3.0.0",
			},
			expectError: false,
		},
		{
			name:         "empty .tool-versions file",
			toolVersions: "",
			expectedPkgs: []string{},
			expectError:  false,
		},
		{
			name: "single global tool",
			toolVersions: `nodejs 18.16.0
`,
			expectedPkgs: []string{"nodejs 18.16.0"},
			expectError:  false,
		},
		{
			name: "tools with comments and empty lines",
			toolVersions: `# My global tools
nodejs 20.0.0

python 3.11.3
# Another comment
ruby 3.0.0
`,
			expectedPkgs: []string{
				"nodejs 20.0.0",
				"python 3.11.3",
				"ruby 3.0.0",
			},
			expectError: false,
		},
		{
			name:         "no .tool-versions file",
			toolVersions: "",
			expectedPkgs: []string{},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary home directory
			tempHome := t.TempDir()
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempHome)
			defer os.Setenv("HOME", originalHome)

			// Create .tool-versions file if content provided
			if tt.toolVersions != "" {
				toolVersionsPath := filepath.Join(tempHome, ".tool-versions")
				err := os.WriteFile(toolVersionsPath, []byte(tt.toolVersions), 0644)
				if err != nil {
					t.Fatalf("Failed to create test .tool-versions file: %v", err)
				}
			}

			mockExecutor := &MockCommandExecutor{}
			discoverer := NewAsdfDiscoverer(mockExecutor)
			packages, err := discoverer.DiscoverPackages()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(packages) != len(tt.expectedPkgs) {
				t.Errorf("Expected %d packages, got %d", len(tt.expectedPkgs), len(packages))
				t.Errorf("Expected: %v", tt.expectedPkgs)
				t.Errorf("Got: %v", packages)
			}

			for i, pkg := range packages {
				if i < len(tt.expectedPkgs) && pkg != tt.expectedPkgs[i] {
					t.Errorf("Expected package %s, got %s", tt.expectedPkgs[i], pkg)
				}
			}
		})
	}
}

// Verify interface compliance
var _ managers.CommandExecutor = &MockCommandExecutor{}
