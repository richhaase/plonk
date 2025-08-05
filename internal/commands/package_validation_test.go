// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPackageSpecValidator(t *testing.T) {
	tests := []struct {
		name            string
		config          *config.Config
		expectedDefault string
	}{
		{
			name:            "nil config uses default manager",
			config:          nil,
			expectedDefault: packages.DefaultManager,
		},
		{
			name:            "empty config uses default manager",
			config:          &config.Config{},
			expectedDefault: packages.DefaultManager,
		},
		{
			name: "config with default manager uses configured default",
			config: &config.Config{
				DefaultManager: "npm",
			},
			expectedDefault: "npm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewPackageSpecValidator(tt.config)
			assert.Equal(t, tt.config, validator.Config)
			assert.Equal(t, tt.expectedDefault, validator.DefaultManager)
		})
	}
}

func TestPackageSpecValidator_ValidateInstallSpecs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		config       *config.Config
		expectSpecs  int
		expectErrors int
		errorChecks  []func(t *testing.T, err string)
	}{
		{
			name:         "valid package without manager",
			args:         []string{"git"},
			config:       &config.Config{DefaultManager: "brew"},
			expectSpecs:  1,
			expectErrors: 0,
		},
		{
			name:         "valid package with manager",
			args:         []string{"brew:wget"},
			config:       nil,
			expectSpecs:  1,
			expectErrors: 0,
		},
		{
			name:         "multiple valid packages",
			args:         []string{"git", "brew:wget", "npm:prettier"},
			config:       &config.Config{DefaultManager: "brew"},
			expectSpecs:  3,
			expectErrors: 0,
		},
		{
			name:         "empty package specification",
			args:         []string{""},
			config:       nil,
			expectSpecs:  0,
			expectErrors: 1,
			errorChecks: []func(t *testing.T, err string){
				func(t *testing.T, err string) {
					assert.Contains(t, err, "package specification cannot be empty")
				},
			},
		},
		{
			name:         "empty manager prefix",
			args:         []string{":package"},
			config:       nil,
			expectSpecs:  0,
			expectErrors: 1,
			errorChecks: []func(t *testing.T, err string){
				func(t *testing.T, err string) {
					assert.Contains(t, err, "manager prefix cannot be empty")
				},
			},
		},
		{
			name:         "empty package name",
			args:         []string{"brew:"},
			config:       nil,
			expectSpecs:  0,
			expectErrors: 1,
			errorChecks: []func(t *testing.T, err string){
				func(t *testing.T, err string) {
					assert.Contains(t, err, "package name cannot be empty")
				},
			},
		},
		{
			name:         "invalid manager",
			args:         []string{"invalid:package"},
			config:       nil,
			expectSpecs:  0,
			expectErrors: 1,
			errorChecks: []func(t *testing.T, err string){
				func(t *testing.T, err string) {
					assert.Contains(t, err, `unknown package manager "invalid"`)
				},
			},
		},
		{
			name:         "no config default uses system default",
			args:         []string{"git"},
			config:       &config.Config{DefaultManager: ""},
			expectSpecs:  1,
			expectErrors: 0,
		},
		{
			name:         "mixed valid and invalid",
			args:         []string{"git", ":invalid", "brew:wget"},
			config:       &config.Config{DefaultManager: "brew"},
			expectSpecs:  2,
			expectErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewPackageSpecValidator(tt.config)
			specs, errors := validator.ValidateInstallSpecs(tt.args)

			assert.Len(t, specs, tt.expectSpecs)
			assert.Len(t, errors, tt.expectErrors)

			// Check specific error messages if provided
			for i, check := range tt.errorChecks {
				if i < len(errors) {
					check(t, errors[i].Error.Error())
				}
			}

			// Verify valid specs have managers set
			for _, spec := range specs {
				assert.NotEmpty(t, spec.Manager, "spec should have manager set")
				assert.NotEmpty(t, spec.Name, "spec should have name set")
			}
		})
	}
}

func TestPackageSpecValidator_ValidateUninstallSpecs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		config       *config.Config
		expectSpecs  int
		expectErrors int
		errorChecks  []func(t *testing.T, err string)
	}{
		{
			name:         "valid package without manager",
			args:         []string{"git"},
			config:       nil,
			expectSpecs:  1,
			expectErrors: 0,
		},
		{
			name:         "valid package with manager",
			args:         []string{"brew:wget"},
			config:       nil,
			expectSpecs:  1,
			expectErrors: 0,
		},
		{
			name:         "multiple valid packages",
			args:         []string{"git", "brew:wget", "npm:prettier"},
			config:       nil,
			expectSpecs:  3,
			expectErrors: 0,
		},
		{
			name:         "empty package specification",
			args:         []string{""},
			config:       nil,
			expectSpecs:  0,
			expectErrors: 1,
			errorChecks: []func(t *testing.T, err string){
				func(t *testing.T, err string) {
					assert.Contains(t, err, "package specification cannot be empty")
				},
			},
		},
		{
			name:         "invalid manager",
			args:         []string{"invalid:package"},
			config:       nil,
			expectSpecs:  0,
			expectErrors: 1,
			errorChecks: []func(t *testing.T, err string){
				func(t *testing.T, err string) {
					assert.Contains(t, err, `unknown package manager "invalid"`)
				},
			},
		},
		{
			name:         "package without manager is allowed for uninstall",
			args:         []string{"git"},
			config:       &config.Config{DefaultManager: ""},
			expectSpecs:  1,
			expectErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewPackageSpecValidator(tt.config)
			specs, errors := validator.ValidateUninstallSpecs(tt.args)

			assert.Len(t, specs, tt.expectSpecs)
			assert.Len(t, errors, tt.expectErrors)

			// Check specific error messages if provided
			for i, check := range tt.errorChecks {
				if i < len(errors) {
					check(t, errors[i].Error.Error())
				}
			}

			// Verify valid specs have names set (manager may be empty for uninstall)
			for _, spec := range specs {
				assert.NotEmpty(t, spec.Name, "spec should have name set")
			}
		})
	}
}

func TestPackageSpecValidator_ValidateSearchSpec(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid package without manager",
			arg:         "git",
			config:      nil,
			expectError: false,
		},
		{
			name:        "valid package with manager",
			arg:         "brew:wget",
			config:      nil,
			expectError: false,
		},
		{
			name:        "npm scoped package",
			arg:         "npm:@types/node",
			config:      nil,
			expectError: false,
		},
		{
			name:        "empty specification",
			arg:         "",
			config:      nil,
			expectError: true,
			errorMsg:    "package specification cannot be empty",
		},
		{
			name:        "empty manager prefix",
			arg:         ":package",
			config:      nil,
			expectError: true,
			errorMsg:    "manager prefix cannot be empty",
		},
		{
			name:        "empty package name",
			arg:         "brew:",
			config:      nil,
			expectError: true,
			errorMsg:    "package name cannot be empty",
		},
		{
			name:        "invalid manager",
			arg:         "invalid:package",
			config:      nil,
			expectError: true,
			errorMsg:    `unknown package manager "invalid"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewPackageSpecValidator(tt.config)
			spec, err := validator.ValidateSearchSpec(tt.arg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, spec)
			} else {
				require.NoError(t, err)
				require.NotNil(t, spec)
				assert.NotEmpty(t, spec.Name)
				assert.Equal(t, tt.arg, spec.OriginalSpec)
			}
		})
	}
}
