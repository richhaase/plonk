// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"github.com/go-playground/validator/v10"
)

// validManagers holds the list of valid package managers.
// This is populated at runtime by the packages module via ManagerRegistry
// or explicitly in tests.
var validManagers []string

// SetValidManagers sets the list of valid package managers for validation.
// In production this is called from the packages module; tests may call it directly.
func SetValidManagers(managers []string) {
	validManagers = managers
}

// RegisterValidators registers custom validators for config validation.
func RegisterValidators(v *validator.Validate) error {
	return v.RegisterValidation("validmanager", validatePackageManager)
}

// validatePackageManager validates that a package manager is registered.
// Valid managers are set by the packages registry during initialization.
func validatePackageManager(fl validator.FieldLevel) bool {
	managerName := fl.Field().String()
	if managerName == "" {
		// Empty is valid (will use default).
		return true
	}

	// If no managers have been registered, treat any explicit manager as invalid.
	if len(validManagers) == 0 {
		return false
	}

	for _, valid := range validManagers {
		if managerName == valid {
			return true
		}
	}

	return false
}

