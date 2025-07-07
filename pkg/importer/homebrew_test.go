// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"os/exec"
	"testing"

	"plonk/pkg/managers"
)

func TestDiscoverHomebrewPackages(t *testing.T) {
	tests := []struct {
		name         string
		brewOutput   string
		expectedPkgs []string
		expectError  bool
	}{
		{
			name:         "successful brew list",
			brewOutput:   "git\njq\nnode\n",
			expectedPkgs: []string{"git", "jq", "node"},
			expectError:  false,
		},
		{
			name:         "empty brew list",
			brewOutput:   "",
			expectedPkgs: []string{},
			expectError:  false,
		},
		{
			name:         "homebrew not available",
			brewOutput:   "",
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
				mockExecutor.Output = tt.brewOutput
			}

			discoverer := NewHomebrewDiscoverer(mockExecutor)
			packages, err := discoverer.DiscoverPackages()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(packages) != len(tt.expectedPkgs) {
				t.Errorf("Expected %d packages, got %d", len(tt.expectedPkgs), len(packages))
			}

			for i, pkg := range packages {
				if i < len(tt.expectedPkgs) && pkg != tt.expectedPkgs[i] {
					t.Errorf("Expected package %s, got %s", tt.expectedPkgs[i], pkg)
				}
			}
		})
	}
}

// MockCommandExecutor implements managers.CommandExecutor for testing
type MockCommandExecutor struct {
	Output     string
	ShouldFail bool
}

// Verify interface compliance
var _ managers.CommandExecutor = &MockCommandExecutor{}

func (m *MockCommandExecutor) Execute(name string, args ...string) *exec.Cmd {
	if m.ShouldFail {
		return exec.Command("false") // Command that always fails
	}
	return exec.Command("echo", m.Output)
}
