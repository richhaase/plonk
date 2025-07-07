// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"testing"

	"plonk/pkg/managers"
)

func TestDiscoverNpmPackages(t *testing.T) {
	tests := []struct {
		name         string
		npmOutput    string
		expectedPkgs []string
		expectError  bool
	}{
		{
			name: "successful npm list global packages",
			npmOutput: `/usr/local/lib/node_modules/claude-code
/usr/local/lib/node_modules/npm
/usr/local/lib/node_modules/@angular/cli
/usr/local/lib/node_modules/typescript
`,
			expectedPkgs: []string{
				"claude-code",
				"@angular/cli",
				"typescript",
			},
			expectError: false,
		},
		{
			name:         "empty npm list",
			npmOutput:    "",
			expectedPkgs: []string{},
			expectError:  false,
		},
		{
			name: "single global package",
			npmOutput: `/usr/local/lib/node_modules/claude-code
`,
			expectedPkgs: []string{"claude-code"},
			expectError:  false,
		},
		{
			name: "npm packages in home directory",
			npmOutput: `/Users/user/.npm-global/lib/node_modules/claude-code
/Users/user/.npm-global/lib/node_modules/typescript
`,
			expectedPkgs: []string{
				"claude-code",
				"typescript",
			},
			expectError: false,
		},
		{
			name:         "npm not available",
			npmOutput:    "",
			expectedPkgs: []string{},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{}
			if tt.expectError {
				mockExecutor.ShouldFail = true
			} else {
				mockExecutor.Output = tt.npmOutput
			}

			discoverer := NewNpmDiscoverer(mockExecutor)
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
