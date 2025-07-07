// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// SimpleValidator uses the go-playground/validator library with custom validators.
type SimpleValidator struct {
	validator *validator.Validate
}

// ValidationResult represents the outcome of validation
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// NewSimpleValidator creates a new validator with custom validation functions
func NewSimpleValidator() *SimpleValidator {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("package_name", validatePackageName)
	v.RegisterValidation("file_path", validateFilePath)

	return &SimpleValidator{validator: v}
}

// ValidateConfig validates a parsed config struct
func (v *SimpleValidator) ValidateConfig(config *Config) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Validate the struct
	if err := v.validator.Struct(config); err != nil {
		result.Valid = false

		// Convert validator errors to friendly messages
		for _, err := range err.(validator.ValidationErrors) {
			result.Errors = append(result.Errors, v.formatValidationError(err))
		}
	}

	// Add warnings for best practices
	if config.Settings.DefaultManager == "npm" {
		result.Warnings = append(result.Warnings,
			"Using npm as default manager may be slower than homebrew for most packages")
	}

	return result
}

// ValidateConfigFromYAML validates YAML content and returns structured result
func (v *SimpleValidator) ValidateConfigFromYAML(content []byte) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Step 1: Validate YAML syntax
	if err := ValidateYAML(content); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("YAML syntax error: %s", err.Error()))
		return result
	}

	// Step 2: Parse into config struct
	var config Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse config: %s", err.Error()))
		return result
	}

	// Step 3: Validate the struct
	structResult := v.ValidateConfig(&config)
	result.Valid = structResult.Valid
	result.Errors = append(result.Errors, structResult.Errors...)
	result.Warnings = append(result.Warnings, structResult.Warnings...)

	return result
}

// formatValidationError converts a validator.FieldError to a friendly message
func (v *SimpleValidator) formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	value := err.Value()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s (got: %v)", field, err.Param(), value)
	case "package_name":
		return fmt.Sprintf("%s contains invalid package name: %v", field, value)
	case "file_path":
		return fmt.Sprintf("%s contains invalid file path: %v", field, value)
	case "required_without":
		return fmt.Sprintf("%s is required when %s is not provided", field, err.Param())
	default:
		return fmt.Sprintf("%s validation failed for tag '%s': %v", field, tag, value)
	}
}

// Custom validation functions

// validatePackageName validates package names
func validatePackageName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	return ValidatePackageName(name) == nil
}

// validateFilePath validates file paths
func validateFilePath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	return ValidateFilePath(path) == nil
}

// IsValid returns true if there are no validation errors
func (r *ValidationResult) IsValid() bool {
	return r.Valid
}

// GetSummary returns a human-readable summary
func (r *ValidationResult) GetSummary() string {
	if r.Valid {
		return "Configuration is valid"
	}

	var parts []string
	if len(r.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", len(r.Errors)))
	}
	if len(r.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("%d warnings", len(r.Warnings)))
	}

	return fmt.Sprintf("Configuration has %s", strings.Join(parts, ", "))
}

// GetMessages returns all validation messages
func (r *ValidationResult) GetMessages() []string {
	var messages []string

	for _, err := range r.Errors {
		messages = append(messages, fmt.Sprintf("ERROR: %s", err))
	}

	for _, warning := range r.Warnings {
		messages = append(messages, fmt.Sprintf("WARNING: %s", warning))
	}

	return messages
}
