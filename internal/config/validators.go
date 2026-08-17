// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"strconv"

	"github.com/go-playground/validator/v10"
)

// ManagerChecker is a function that checks if a manager name is valid.
// This is set by the packages module during initialization.
var ManagerChecker func(string) bool

// RegisterValidators registers custom validators for config validation.
func RegisterValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("validmanager", validatePackageManager); err != nil {
		return err
	}
	return v.RegisterValidation("filemode", validateFileMode)
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

// validateFileMode validates that a deploy mode is octal file permissions in
// the standard 0000-0777 range (digits 0-7 only, using the requested permission
// bits). Empty is valid (means "no explicit mode").
func validateFileMode(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}

	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '7' {
			return false
		}
	}

	v, err := strconv.ParseUint(s, 8, 16)
	if err != nil {
		return false
	}
	// Permission bits only: 0000-0777.
	return v <= 0o777
}
