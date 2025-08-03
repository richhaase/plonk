// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/stretchr/testify/assert"
)

func TestValidatePackageSpec(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		manager     string
		packageName string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid brew spec",
			spec:        "brew:wget",
			manager:     "brew",
			packageName: "wget",
			wantErr:     false,
		},
		{
			name:        "valid npm spec",
			spec:        "npm:prettier",
			manager:     "npm",
			packageName: "prettier",
			wantErr:     false,
		},
		{
			name:        "no manager prefix",
			spec:        "wget",
			manager:     "",
			packageName: "wget",
			wantErr:     false,
		},
		{
			name:        "empty package name",
			spec:        "brew:",
			manager:     "brew",
			packageName: "",
			wantErr:     true,
			errContains: "package name cannot be empty",
		},
		{
			name:        "empty manager with colon",
			spec:        ":wget",
			manager:     "",
			packageName: "wget",
			wantErr:     true,
			errContains: "manager prefix cannot be empty",
		},
		{
			name:        "empty spec",
			spec:        "",
			manager:     "",
			packageName: "",
			wantErr:     true,
			errContains: "package name cannot be empty",
		},
		{
			name:        "complex go package",
			spec:        "go:golang.org/x/tools/cmd/gopls",
			manager:     "go",
			packageName: "golang.org/x/tools/cmd/gopls",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePackageSpec(tt.spec, tt.manager, tt.packageName)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResolvePackageManager(t *testing.T) {
	tests := []struct {
		name            string
		manager         string
		cfg             *config.Config
		expectedManager string
	}{
		{
			name:            "explicit manager takes precedence",
			manager:         "npm",
			cfg:             &config.Config{DefaultManager: "brew"},
			expectedManager: "npm",
		},
		{
			name:            "use config default when no manager specified",
			manager:         "",
			cfg:             &config.Config{DefaultManager: "brew"},
			expectedManager: "brew",
		},
		{
			name:            "use system default when no config default",
			manager:         "",
			cfg:             &config.Config{},
			expectedManager: packages.DefaultManager,
		},
		{
			name:            "nil config uses system default",
			manager:         "",
			cfg:             nil,
			expectedManager: packages.DefaultManager,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePackageManager(tt.manager, tt.cfg)
			assert.Equal(t, tt.expectedManager, result)
		})
	}
}

func TestParseAndValidatePackageSpecs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		cfg           *config.Config
		wantSpecCount int
		wantErrCount  int
		validateSpecs func(t *testing.T, specs []ValidatedPackageSpec)
		validateErrs  func(t *testing.T, errs []resources.OperationResult)
	}{
		{
			name:          "all valid specs",
			args:          []string{"brew:wget", "npm:prettier", "pip:black"},
			cfg:           &config.Config{},
			wantSpecCount: 3,
			wantErrCount:  0,
			validateSpecs: func(t *testing.T, specs []ValidatedPackageSpec) {
				assert.Equal(t, "brew", specs[0].Manager)
				assert.Equal(t, "wget", specs[0].PackageName)
				assert.Equal(t, "npm", specs[1].Manager)
				assert.Equal(t, "prettier", specs[1].PackageName)
				assert.Equal(t, "pip", specs[2].Manager)
				assert.Equal(t, "black", specs[2].PackageName)
			},
		},
		{
			name:          "mix of valid and invalid",
			args:          []string{"brew:wget", ":invalid", "npm:prettier", "brew:"},
			cfg:           &config.Config{},
			wantSpecCount: 2,
			wantErrCount:  2,
			validateSpecs: func(t *testing.T, specs []ValidatedPackageSpec) {
				assert.Equal(t, "brew", specs[0].Manager)
				assert.Equal(t, "wget", specs[0].PackageName)
				assert.Equal(t, "npm", specs[1].Manager)
				assert.Equal(t, "prettier", specs[1].PackageName)
			},
			validateErrs: func(t *testing.T, errs []resources.OperationResult) {
				assert.Equal(t, "failed", errs[0].Status)
				assert.Contains(t, errs[0].Error.Error(), "manager prefix cannot be empty")
				assert.Equal(t, "failed", errs[1].Status)
				assert.Contains(t, errs[1].Error.Error(), "package name cannot be empty")
			},
		},
		{
			name:          "use default manager",
			args:          []string{"wget", "htop"},
			cfg:           &config.Config{DefaultManager: "brew"},
			wantSpecCount: 2,
			wantErrCount:  0,
			validateSpecs: func(t *testing.T, specs []ValidatedPackageSpec) {
				assert.Equal(t, "brew", specs[0].Manager)
				assert.Equal(t, "wget", specs[0].PackageName)
				assert.Equal(t, "brew", specs[1].Manager)
				assert.Equal(t, "htop", specs[1].PackageName)
			},
		},
		{
			name:          "invalid manager",
			args:          []string{"invalid:package"},
			cfg:           &config.Config{},
			wantSpecCount: 0,
			wantErrCount:  1,
			validateErrs: func(t *testing.T, errs []resources.OperationResult) {
				assert.Equal(t, "failed", errs[0].Status)
				assert.Contains(t, errs[0].Error.Error(), "unknown package manager")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs, errs := parseAndValidatePackageSpecs(tt.args, tt.cfg)

			assert.Len(t, specs, tt.wantSpecCount)
			assert.Len(t, errs, tt.wantErrCount)

			if tt.validateSpecs != nil {
				tt.validateSpecs(t, specs)
			}

			if tt.validateErrs != nil {
				tt.validateErrs(t, errs)
			}
		})
	}
}

func TestParseAndValidateUninstallSpecs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantSpecCount int
		wantErrCount  int
		validateSpecs func(t *testing.T, specs []ValidatedPackageSpec)
		validateErrs  func(t *testing.T, errs []resources.OperationResult)
	}{
		{
			name:          "all valid specs with managers",
			args:          []string{"brew:wget", "npm:prettier", "pip:black"},
			wantSpecCount: 3,
			wantErrCount:  0,
			validateSpecs: func(t *testing.T, specs []ValidatedPackageSpec) {
				assert.Equal(t, "brew", specs[0].Manager)
				assert.Equal(t, "wget", specs[0].PackageName)
				assert.Equal(t, "npm", specs[1].Manager)
				assert.Equal(t, "prettier", specs[1].PackageName)
				assert.Equal(t, "pip", specs[2].Manager)
				assert.Equal(t, "black", specs[2].PackageName)
			},
		},
		{
			name:          "specs without managers (let uninstall determine)",
			args:          []string{"wget", "htop", "tree"},
			wantSpecCount: 3,
			wantErrCount:  0,
			validateSpecs: func(t *testing.T, specs []ValidatedPackageSpec) {
				assert.Equal(t, "", specs[0].Manager)
				assert.Equal(t, "wget", specs[0].PackageName)
				assert.Equal(t, "", specs[1].Manager)
				assert.Equal(t, "htop", specs[1].PackageName)
				assert.Equal(t, "", specs[2].Manager)
				assert.Equal(t, "tree", specs[2].PackageName)
			},
		},
		{
			name:          "mix of valid and invalid",
			args:          []string{"brew:wget", ":invalid", "npm:prettier", "brew:"},
			wantSpecCount: 2,
			wantErrCount:  2,
			validateSpecs: func(t *testing.T, specs []ValidatedPackageSpec) {
				assert.Equal(t, "brew", specs[0].Manager)
				assert.Equal(t, "wget", specs[0].PackageName)
				assert.Equal(t, "npm", specs[1].Manager)
				assert.Equal(t, "prettier", specs[1].PackageName)
			},
			validateErrs: func(t *testing.T, errs []resources.OperationResult) {
				assert.Equal(t, "failed", errs[0].Status)
				assert.Contains(t, errs[0].Error.Error(), "manager prefix cannot be empty")
				assert.Equal(t, "failed", errs[1].Status)
				assert.Contains(t, errs[1].Error.Error(), "package name cannot be empty")
			},
		},
		{
			name:          "invalid manager specified",
			args:          []string{"invalid:package"},
			wantSpecCount: 0,
			wantErrCount:  1,
			validateErrs: func(t *testing.T, errs []resources.OperationResult) {
				assert.Equal(t, "failed", errs[0].Status)
				assert.Contains(t, errs[0].Error.Error(), "unknown package manager")
			},
		},
		{
			name:          "mix of with and without managers",
			args:          []string{"brew:git", "htop", "npm:eslint"},
			wantSpecCount: 3,
			wantErrCount:  0,
			validateSpecs: func(t *testing.T, specs []ValidatedPackageSpec) {
				assert.Equal(t, "brew", specs[0].Manager)
				assert.Equal(t, "git", specs[0].PackageName)
				assert.Equal(t, "", specs[1].Manager)
				assert.Equal(t, "htop", specs[1].PackageName)
				assert.Equal(t, "npm", specs[2].Manager)
				assert.Equal(t, "eslint", specs[2].PackageName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs, errs := parseAndValidateUninstallSpecs(tt.args)

			assert.Len(t, specs, tt.wantSpecCount)
			assert.Len(t, errs, tt.wantErrCount)

			if tt.validateSpecs != nil {
				tt.validateSpecs(t, specs)
			}

			if tt.validateErrs != nil {
				tt.validateErrs(t, errs)
			}
		})
	}
}

func TestValidateSearchSpec(t *testing.T) {
	tests := []struct {
		name        string
		manager     string
		packageName string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid search with manager",
			manager:     "brew",
			packageName: "git",
			wantErr:     false,
		},
		{
			name:        "valid search without manager",
			manager:     "",
			packageName: "git",
			wantErr:     false,
		},
		{
			name:        "empty package name",
			manager:     "brew",
			packageName: "",
			wantErr:     true,
			errContains: "package name cannot be empty",
		},
		{
			name:        "invalid manager",
			manager:     "invalid",
			packageName: "git",
			wantErr:     true,
			errContains: "unknown package manager",
		},
		{
			name:        "empty package name without manager",
			manager:     "",
			packageName: "",
			wantErr:     true,
			errContains: "package name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSearchSpec(tt.manager, tt.packageName)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
