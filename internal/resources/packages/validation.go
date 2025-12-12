// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"fmt"
)

// ValidationMode defines how strict validation should be
type ValidationMode int

const (
	// ValidationModeInstall requires a package manager (strict)
	ValidationModeInstall ValidationMode = iota
	// ValidationModeUninstall allows optional package manager
	ValidationModeUninstall
	// ValidationModeSearch allows optional package manager
	ValidationModeSearch
)

// ValidationResult represents the outcome of validating a package spec
type ValidationResult struct {
	Spec         *PackageSpec
	OriginalSpec string
	Error        error
}

// BatchValidationResult contains results from validating multiple specs
type BatchValidationResult struct {
	Valid   []*PackageSpec
	Invalid []ValidationResult
}

// ValidateSpecs validates multiple package specifications based on mode
func ValidateSpecs(specs []string, mode ValidationMode, defaultManager string) BatchValidationResult {
	result := BatchValidationResult{
		Valid:   make([]*PackageSpec, 0),
		Invalid: make([]ValidationResult, 0),
	}

	for _, specStr := range specs {
		spec, err := validateSingleSpec(specStr, mode, defaultManager)
		if err != nil {
			result.Invalid = append(result.Invalid, ValidationResult{
				Spec:         spec,
				OriginalSpec: specStr,
				Error:        err,
			})
		} else {
			result.Valid = append(result.Valid, spec)
		}
	}

	return result
}

// validateSingleSpec is the internal validation logic
func validateSingleSpec(specStr string, mode ValidationMode, defaultManager string) (*PackageSpec, error) {
	spec, err := ParsePackageSpec(specStr)
	if err != nil {
		return nil, fmt.Errorf("invalid package specification %q: %w", specStr, err)
	}

	// Apply validation based on mode
	switch mode {
	case ValidationModeInstall:
		// Install requires a manager
		if err := spec.RequireManager(defaultManager); err != nil {
			return spec, err
		}
	case ValidationModeUninstall, ValidationModeSearch:
		// Uninstall and Search only validate if manager is specified
		if err := spec.ValidateManager(); err != nil {
			return spec, err
		}
	default:
		return spec, fmt.Errorf("unknown validation mode: %v", mode)
	}

	return spec, nil
}
