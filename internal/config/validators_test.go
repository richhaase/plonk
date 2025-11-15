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
