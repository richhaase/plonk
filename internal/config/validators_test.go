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
		setupFunc  func()
		shouldPass bool
	}{
		{
			name:       "empty manager is valid",
			manager:    "",
			shouldPass: true,
		},
		{
			name:       "known manager (brew) is valid",
			manager:    "brew",
			shouldPass: true,
		},
		{
			name:       "known manager (npm) is valid",
			manager:    "npm",
			shouldPass: true,
		},
		{
			name:       "test-unavailable is valid (for testing)",
			manager:    "test-unavailable",
			shouldPass: true,
		},
		{
			name:       "unknown manager is invalid",
			manager:    "unknownmanager",
			shouldPass: false,
		},
		{
			name:       "apt is valid (legacy support)",
			manager:    "apt",
			shouldPass: true,
		},
		{
			name:    "dynamically registered manager is valid",
			manager: "custom-manager",
			setupFunc: func() {
				// Simulate dynamic registration
				SetValidManagers([]string{"brew", "npm", "custom-manager"})
			},
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default state
			validManagers = nil

			if tt.setupFunc != nil {
				tt.setupFunc()
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
