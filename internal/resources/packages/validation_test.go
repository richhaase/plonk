// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSpec(t *testing.T) {
	tests := []struct {
		name           string
		spec           string
		mode           ValidationMode
		defaultManager string
		wantSpec       *PackageSpec
		wantErr        bool
		errContains    string
	}{
		// Install mode tests
		{
			name:           "install with manager specified",
			spec:           "brew:git",
			mode:           ValidationModeInstall,
			defaultManager: "npm",
			wantSpec:       &PackageSpec{Name: "git", Manager: "brew", OriginalSpec: "brew:git"},
			wantErr:        false,
		},
		{
			name:           "install without manager uses default",
			spec:           "git",
			mode:           ValidationModeInstall,
			defaultManager: "brew",
			wantSpec:       &PackageSpec{Name: "git", Manager: "brew", OriginalSpec: "git"},
			wantErr:        false,
		},
		{
			name:           "install without manager and no default fails",
			spec:           "git",
			mode:           ValidationModeInstall,
			defaultManager: "",
			wantSpec:       nil,
			wantErr:        true,
			errContains:    "no package manager specified",
		},
		{
			name:           "install with invalid manager",
			spec:           "invalid:git",
			mode:           ValidationModeInstall,
			defaultManager: "brew",
			wantSpec:       nil,
			wantErr:        true,
			errContains:    "unknown package manager",
		},
		// Uninstall mode tests
		{
			name:           "uninstall with manager specified",
			spec:           "npm:lodash",
			mode:           ValidationModeUninstall,
			defaultManager: "",
			wantSpec:       &PackageSpec{Name: "lodash", Manager: "npm", OriginalSpec: "npm:lodash"},
			wantErr:        false,
		},
		{
			name:           "uninstall without manager is valid",
			spec:           "lodash",
			mode:           ValidationModeUninstall,
			defaultManager: "brew",
			wantSpec:       &PackageSpec{Name: "lodash", Manager: "", OriginalSpec: "lodash"},
			wantErr:        false,
		},
		{
			name:           "uninstall with invalid manager fails",
			spec:           "invalid:lodash",
			mode:           ValidationModeUninstall,
			defaultManager: "",
			wantSpec:       nil,
			wantErr:        true,
			errContains:    "unknown package manager",
		},
		// Search mode tests
		{
			name:           "search with manager specified",
			spec:           "pip:requests",
			mode:           ValidationModeSearch,
			defaultManager: "",
			wantSpec:       &PackageSpec{Name: "requests", Manager: "pip", OriginalSpec: "pip:requests"},
			wantErr:        false,
		},
		{
			name:           "search without manager is valid",
			spec:           "requests",
			mode:           ValidationModeSearch,
			defaultManager: "npm",
			wantSpec:       &PackageSpec{Name: "requests", Manager: "", OriginalSpec: "requests"},
			wantErr:        false,
		},
		// Invalid spec tests
		{
			name:           "empty spec",
			spec:           "",
			mode:           ValidationModeInstall,
			defaultManager: "brew",
			wantSpec:       nil,
			wantErr:        true,
			errContains:    "invalid package specification",
		},
		{
			name:           "spec with empty name",
			spec:           "brew:",
			mode:           ValidationModeInstall,
			defaultManager: "brew",
			wantSpec:       nil,
			wantErr:        true,
			errContains:    "package name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := ValidateSpec(tt.spec, tt.mode, tt.defaultManager)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.wantSpec != nil {
				assert.Equal(t, tt.wantSpec.Name, spec.Name)
				assert.Equal(t, tt.wantSpec.Manager, spec.Manager)
				assert.Equal(t, tt.wantSpec.OriginalSpec, spec.OriginalSpec)
			}
		})
	}
}

func TestValidateSpecs(t *testing.T) {
	tests := []struct {
		name           string
		specs          []string
		mode           ValidationMode
		defaultManager string
		wantValid      int
		wantInvalid    int
		checkValid     []string
		checkInvalid   []string
	}{
		{
			name:           "all valid install specs",
			specs:          []string{"brew:git", "npm:lodash", "pip:requests"},
			mode:           ValidationModeInstall,
			defaultManager: "brew",
			wantValid:      3,
			wantInvalid:    0,
			checkValid:     []string{"git", "lodash", "requests"},
		},
		{
			name:           "mixed valid and invalid install specs",
			specs:          []string{"git", "invalid:package", "npm:lodash", ""},
			mode:           ValidationModeInstall,
			defaultManager: "brew",
			wantValid:      2,
			wantInvalid:    2,
			checkValid:     []string{"git", "lodash"},
			checkInvalid:   []string{"invalid:package", ""},
		},
		{
			name:           "uninstall specs without managers",
			specs:          []string{"git", "lodash", "requests"},
			mode:           ValidationModeUninstall,
			defaultManager: "",
			wantValid:      3,
			wantInvalid:    0,
			checkValid:     []string{"git", "lodash", "requests"},
		},
		{
			name:           "empty specs list",
			specs:          []string{},
			mode:           ValidationModeInstall,
			defaultManager: "brew",
			wantValid:      0,
			wantInvalid:    0,
		},
		{
			name:           "install without default manager",
			specs:          []string{"git", "brew:wget", "lodash"},
			mode:           ValidationModeInstall,
			defaultManager: "",
			wantValid:      1,
			wantInvalid:    2,
			checkValid:     []string{"wget"},
			checkInvalid:   []string{"git", "lodash"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateSpecs(tt.specs, tt.mode, tt.defaultManager)

			assert.Len(t, result.Valid, tt.wantValid)
			assert.Len(t, result.Invalid, tt.wantInvalid)

			// Check specific valid specs
			for _, expected := range tt.checkValid {
				found := false
				for _, spec := range result.Valid {
					if spec.Name == expected {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find valid spec with name %s", expected)
			}

			// Check specific invalid specs
			for _, expected := range tt.checkInvalid {
				found := false
				for _, invalid := range result.Invalid {
					if invalid.OriginalSpec == expected {
						found = true
						assert.Error(t, invalid.Error)
						break
					}
				}
				assert.True(t, found, "Expected to find invalid spec %s", expected)
			}
		})
	}
}

func TestBatchValidationResult_Helpers(t *testing.T) {
	t.Run("empty result", func(t *testing.T) {
		result := BatchValidationResult{}
		assert.True(t, result.AllValid())
		assert.False(t, result.HasErrors())
	})

	t.Run("all valid", func(t *testing.T) {
		result := BatchValidationResult{
			Valid: []*PackageSpec{
				{Name: "git", Manager: "brew"},
				{Name: "lodash", Manager: "npm"},
			},
		}
		assert.True(t, result.AllValid())
		assert.False(t, result.HasErrors())
	})

	t.Run("has errors", func(t *testing.T) {
		result := BatchValidationResult{
			Valid: []*PackageSpec{
				{Name: "git", Manager: "brew"},
			},
			Invalid: []ValidationResult{
				{OriginalSpec: "invalid:pkg", Error: errors.New("test error")},
			},
		}
		assert.False(t, result.AllValid())
		assert.True(t, result.HasErrors())
	})
}

// Helper methods for BatchValidationResult
func (r *BatchValidationResult) AllValid() bool {
	return len(r.Invalid) == 0
}

func (r *BatchValidationResult) HasErrors() bool {
	return len(r.Invalid) > 0
}
