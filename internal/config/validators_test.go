// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidatePackageManager(t *testing.T) {
	tests := []struct {
		name       string
		manager    string
		valid      []string
		shouldPass bool
	}{
		{
			name:       "empty manager is valid",
			manager:    "",
			shouldPass: true,
		},
		{
			name:       "known manager (brew) is valid when registered",
			manager:    "brew",
			valid:      []string{"brew", "npm"},
			shouldPass: true,
		},
		{
			name:       "known manager (npm) is valid when registered",
			manager:    "npm",
			valid:      []string{"brew", "npm"},
			shouldPass: true,
		},
		{
			name:       "test-unavailable is valid when registered (for testing)",
			manager:    "test-unavailable",
			valid:      []string{"test-unavailable"},
			shouldPass: true,
		},
		{
			name:       "unknown manager is invalid",
			manager:    "unknownmanager",
			shouldPass: false,
		},
		{
			name:       "apt is valid when registered (legacy support)",
			manager:    "apt",
			valid:      []string{"apt"},
			shouldPass: true,
		},
		{
			name:       "dynamically registered manager is valid",
			manager:    "custom-manager",
			valid:      []string{"brew", "npm", "custom-manager"},
			shouldPass: true,
		},
		{
			name:       "non-empty manager is invalid when registry is empty",
			manager:    "brew",
			valid:      []string{},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default state
			validManagers = nil
			validManagersDefined = false

			// Simulate dynamic registration if requested
			if tt.valid != nil {
				SetValidManagers(tt.valid)
			}

			v := validator.New()
			err := RegisterValidators(v)
			assert.NoError(t, err)

			type testStruct struct {
				Manager string `validate:"validmanager"`
			}

			s := testStruct{Manager: tt.manager}
			err = v.Struct(s)

			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateListConfig(t *testing.T) {
	v := validator.New()
	err := RegisterValidators(v)
	assert.NoError(t, err)

	type testStruct struct {
		List ListConfig `validate:"listconfig"`
	}

	tests := []struct {
		name       string
		list       ListConfig
		shouldPass bool
	}{
		{
			name:       "lines ok",
			list:       ListConfig{Parse: "lines"},
			shouldPass: true,
		},
		{
			name:       "lines rejects json fields",
			list:       ListConfig{Parse: "lines", JSONField: "name"},
			shouldPass: false,
		},
		{
			name:       "json requires field",
			list:       ListConfig{Parse: "json"},
			shouldPass: false,
		},
		{
			name:       "json ok",
			list:       ListConfig{Parse: "json", JSONField: "name"},
			shouldPass: true,
		},
		{
			name:       "json-map blocks keys_from",
			list:       ListConfig{Parse: "json-map", KeysFrom: "$.deps"},
			shouldPass: false,
		},
		{
			name:       "jsonpath requires selectors",
			list:       ListConfig{Parse: "jsonpath"},
			shouldPass: false,
		},
		{
			name:       "jsonpath ok with keys_from",
			list:       ListConfig{Parse: "jsonpath", KeysFrom: "$.deps"},
			shouldPass: true,
		},
		{
			name:       "jsonpath ok with values_from and lower normalize",
			list:       ListConfig{Parse: "jsonpath", ValuesFrom: "$.pkgs[*].name", Normalize: "lower"},
			shouldPass: true,
		},
		{
			name:       "jsonpath invalid normalize",
			list:       ListConfig{Parse: "jsonpath", KeysFrom: "$.deps", Normalize: "upper"},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{List: tt.list}
			err := v.Struct(s)
			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
