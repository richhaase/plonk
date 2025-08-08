// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchesPackage(t *testing.T) {
	tests := []struct {
		name        string
		info        packageMatchInfo
		packageName string
		expected    bool
	}{
		{
			name: "matches by name",
			info: packageMatchInfo{
				Manager: "brew",
				Name:    "git",
			},
			packageName: "git",
			expected:    true,
		},
		{
			name: "matches by Go source path",
			info: packageMatchInfo{
				Manager:    "go",
				Name:       "hey",
				SourcePath: "github.com/rakyll/hey",
			},
			packageName: "github.com/rakyll/hey",
			expected:    true,
		},
		{
			name: "matches Go package by binary name",
			info: packageMatchInfo{
				Manager:    "go",
				Name:       "hey",
				SourcePath: "github.com/rakyll/hey",
			},
			packageName: "hey",
			expected:    true,
		},
		{
			name: "matches npm scoped package by full name",
			info: packageMatchInfo{
				Manager:  "npm",
				Name:     "claude-code",
				FullName: "@anthropic-ai/claude-code",
			},
			packageName: "@anthropic-ai/claude-code",
			expected:    true,
		},
		{
			name: "matches npm scoped package by short name",
			info: packageMatchInfo{
				Manager:  "npm",
				Name:     "claude-code",
				FullName: "@anthropic-ai/claude-code",
			},
			packageName: "claude-code",
			expected:    true,
		},
		{
			name: "no match for different name",
			info: packageMatchInfo{
				Manager: "brew",
				Name:    "git",
			},
			packageName: "htop",
			expected:    false,
		},
		{
			name: "no match for wrong Go source path",
			info: packageMatchInfo{
				Manager:    "go",
				Name:       "hey",
				SourcePath: "github.com/rakyll/hey",
			},
			packageName: "github.com/other/package",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPackage(tt.info, tt.packageName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineUpgradeTarget(t *testing.T) {
	tests := []struct {
		name          string
		info          packageMatchInfo
		requestedName string
		expected      string
	}{
		{
			name: "Go package always uses source path",
			info: packageMatchInfo{
				Manager:    "go",
				Name:       "hey",
				SourcePath: "github.com/rakyll/hey",
			},
			requestedName: "hey",
			expected:      "github.com/rakyll/hey",
		},
		{
			name: "Go package uses source path even when requested by source path",
			info: packageMatchInfo{
				Manager:    "go",
				Name:       "hey",
				SourcePath: "github.com/rakyll/hey",
			},
			requestedName: "github.com/rakyll/hey",
			expected:      "github.com/rakyll/hey",
		},
		{
			name: "npm scoped package uses full name when available",
			info: packageMatchInfo{
				Manager:  "npm",
				Name:     "claude-code",
				FullName: "@anthropic-ai/claude-code",
			},
			requestedName: "claude-code",
			expected:      "@anthropic-ai/claude-code",
		},
		{
			name: "regular package uses name",
			info: packageMatchInfo{
				Manager: "brew",
				Name:    "git",
			},
			requestedName: "git",
			expected:      "git",
		},
		{
			name: "npm package without full name uses name",
			info: packageMatchInfo{
				Manager: "npm",
				Name:    "lodash",
			},
			requestedName: "lodash",
			expected:      "lodash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineUpgradeTarget(tt.info, tt.requestedName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindPackageMatch(t *testing.T) {
	packageInfos := []packageMatchInfo{
		{
			Manager: "brew",
			Name:    "git",
		},
		{
			Manager:    "go",
			Name:       "hey",
			SourcePath: "github.com/rakyll/hey",
		},
		{
			Manager:  "npm",
			Name:     "claude-code",
			FullName: "@anthropic-ai/claude-code",
		},
	}

	tests := []struct {
		name        string
		manager     string
		packageName string
		expected    *string // pointer to expected name, nil if no match
	}{
		{
			name:        "find brew package",
			manager:     "brew",
			packageName: "git",
			expected:    stringPtr("git"),
		},
		{
			name:        "find Go package by binary name",
			manager:     "go",
			packageName: "hey",
			expected:    stringPtr("hey"),
		},
		{
			name:        "find Go package by source path",
			manager:     "go",
			packageName: "github.com/rakyll/hey",
			expected:    stringPtr("hey"),
		},
		{
			name:        "find npm scoped package by full name",
			manager:     "npm",
			packageName: "@anthropic-ai/claude-code",
			expected:    stringPtr("claude-code"),
		},
		{
			name:        "find npm scoped package by short name",
			manager:     "npm",
			packageName: "claude-code",
			expected:    stringPtr("claude-code"),
		},
		{
			name:        "no match for wrong manager",
			manager:     "cargo",
			packageName: "git",
			expected:    nil,
		},
		{
			name:        "no match for wrong package",
			manager:     "brew",
			packageName: "htop",
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findPackageMatch(packageInfos, tt.manager, tt.packageName)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.expected, result.Name)
				assert.Equal(t, tt.manager, result.Manager)
			}
		})
	}
}

func TestParseUpgradeArgs(t *testing.T) {
	// Create mock lock file with various package types
	lockFile := &lock.Lock{
		Resources: []lock.ResourceEntry{
			{
				Type: "package",
				ID:   "brew:git",
				Metadata: map[string]interface{}{
					"manager": "brew",
					"name":    "git",
				},
			},
			{
				Type: "package",
				ID:   "go:hey",
				Metadata: map[string]interface{}{
					"manager":     "go",
					"name":        "hey",
					"source_path": "github.com/rakyll/hey",
				},
			},
			{
				Type: "package",
				ID:   "npm:claude-code",
				Metadata: map[string]interface{}{
					"manager":   "npm",
					"name":      "claude-code",
					"full_name": "@anthropic-ai/claude-code",
				},
			},
		},
	}

	tests := []struct {
		name               string
		args               []string
		expectedUpgradeAll bool
		expectedTargets    map[string][]string
		expectError        bool
		errorContains      string
	}{
		{
			name:               "no args means upgrade all",
			args:               []string{},
			expectedUpgradeAll: true,
			expectedTargets:    map[string][]string{},
			expectError:        false,
		},
		{
			name:               "manager name only",
			args:               []string{"brew"},
			expectedUpgradeAll: false,
			expectedTargets:    map[string][]string{"brew": {"git"}},
			expectError:        false,
		},
		{
			name:               "specific package with manager",
			args:               []string{"brew:git"},
			expectedUpgradeAll: false,
			expectedTargets:    map[string][]string{"brew": {"git"}},
			expectError:        false,
		},
		{
			name:               "Go package by source path",
			args:               []string{"go:github.com/rakyll/hey"},
			expectedUpgradeAll: false,
			expectedTargets:    map[string][]string{"go": {"github.com/rakyll/hey"}}, // Should use source path for upgrade
			expectError:        false,
		},
		{
			name:               "Go package by binary name",
			args:               []string{"go:hey"},
			expectedUpgradeAll: false,
			expectedTargets:    map[string][]string{"go": {"github.com/rakyll/hey"}}, // Should use source path for upgrade
			expectError:        false,
		},
		{
			name:               "npm scoped package by full name",
			args:               []string{"npm:@anthropic-ai/claude-code"},
			expectedUpgradeAll: false,
			expectedTargets:    map[string][]string{"npm": {"@anthropic-ai/claude-code"}}, // Should use full name
			expectError:        false,
		},
		{
			name:               "cross-manager package by name",
			args:               []string{"git"},
			expectedUpgradeAll: false,
			expectedTargets:    map[string][]string{"brew": {"git"}},
			expectError:        false,
		},
		{
			name:               "cross-manager Go package by binary name",
			args:               []string{"hey"},
			expectedUpgradeAll: false,
			expectedTargets:    map[string][]string{"go": {"github.com/rakyll/hey"}}, // Should use source path
			expectError:        false,
		},
		{
			name:          "trailing colon syntax error",
			args:          []string{"brew:"},
			expectError:   true,
			errorContains: "invalid syntax 'brew:' - use 'brew' to upgrade",
		},
		{
			name:          "unknown package",
			args:          []string{"nonexistent"},
			expectError:   true,
			errorContains: "package 'nonexistent' is not managed by plonk",
		},
		{
			name:          "unknown manager:package",
			args:          []string{"unknown:git"},
			expectError:   true,
			errorContains: "package 'git' is not managed by plonk via 'unknown'",
		},
		{
			name:               "multiple packages",
			args:               []string{"brew:git", "go:hey"},
			expectedUpgradeAll: false,
			expectedTargets:    map[string][]string{"brew": {"git"}, "go": {"github.com/rakyll/hey"}},
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parseUpgradeArgs(tt.args, lockFile)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUpgradeAll, spec.UpgradeAll)
				assert.Equal(t, tt.expectedTargets, spec.ManagerTargets)
			}
		})
	}
}

// Helper function to create string pointers for testing
func stringPtr(s string) *string {
	return &s
}
