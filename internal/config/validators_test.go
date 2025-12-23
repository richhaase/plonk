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
		checker    func(string) bool
		shouldPass bool
	}{
		{
			name:       "empty manager is valid",
			manager:    "",
			shouldPass: true,
		},
		{
			name:    "known manager (brew) is valid when checker accepts it",
			manager: "brew",
			checker: func(name string) bool {
				return name == "brew" || name == "npm"
			},
			shouldPass: true,
		},
		{
			name:    "known manager (npm) is valid when checker accepts it",
			manager: "npm",
			checker: func(name string) bool {
				return name == "brew" || name == "npm"
			},
			shouldPass: true,
		},
		{
			name:    "unknown manager is invalid",
			manager: "unknownmanager",
			checker: func(name string) bool {
				return name == "brew" || name == "npm"
			},
			shouldPass: false,
		},
		{
			name:       "non-empty manager is invalid when checker is nil",
			manager:    "brew",
			checker:    nil,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the checker for this test
			ManagerChecker = tt.checker

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
