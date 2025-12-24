// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"github.com/go-playground/validator/v10"
)

// ManagerChecker is a function that checks if a manager name is valid.
// This is set by the packages module during initialization.
var ManagerChecker func(string) bool

// RegisterValidators registers custom validators for config validation.
func RegisterValidators(v *validator.Validate) error {
	return v.RegisterValidation("validmanager", validatePackageManager)
}

// validatePackageManager validates that a package manager is supported.
func validatePackageManager(fl validator.FieldLevel) bool {
	managerName := fl.Field().String()
	if managerName == "" {
		// Empty is valid (will use default).
		return true
	}

	if ManagerChecker == nil {
		return false
	}

	return ManagerChecker(managerName)
}

