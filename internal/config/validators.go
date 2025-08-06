// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"github.com/go-playground/validator/v10"
)

// validManagers holds the list of valid package managers
// This is populated at runtime by the packages module
var validManagers []string

// SetValidManagers sets the list of valid package managers for validation
func SetValidManagers(managers []string) {
	validManagers = managers
}

// knownManagers is a fallback list of known package managers
// used when the dynamic list hasn't been populated yet
var knownManagers = []string{"apt", "brew", "npm", "pip", "gem", "go", "cargo", "test-unavailable"}

// RegisterValidators registers custom validators for config validation
func RegisterValidators(v *validator.Validate) error {
	return v.RegisterValidation("validmanager", validatePackageManager)
}

// validatePackageManager validates that a package manager is registered
func validatePackageManager(fl validator.FieldLevel) bool {
	managerName := fl.Field().String()
	if managerName == "" {
		// Empty is valid (will use default)
		return true
	}

	// Check against the valid managers list
	managers := validManagers
	if len(managers) == 0 {
		// Fallback to known managers if dynamic list not populated yet
		managers = knownManagers
	}

	for _, valid := range managers {
		if managerName == valid {
			return true
		}
	}

	return false
}
