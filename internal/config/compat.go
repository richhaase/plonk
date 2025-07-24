// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// This file provides minimal compatibility for the rest of the codebase
// until we can update all references to use the new API directly.

package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/richhaase/plonk/internal/paths"
	"gopkg.in/yaml.v3"
)

// TargetToSource is now provided by the paths package
func TargetToSource(target string) string {
	return paths.TargetToSource(target)
}

// Config type is now defined in config.go

// LoadConfig is an alias to Load for backward compatibility
func LoadConfig(configDir string) (*Config, error) {
	return Load(configDir)
}

// LoadConfigWithDefaults is an alias to LoadWithDefaults for backward compatibility
func LoadConfigWithDefaults(configDir string) *Config {
	return LoadWithDefaults(configDir)
}

// GetDefaultConfigDirectory returns the default config directory
func GetDefaultConfigDirectory() string {
	return paths.GetDefaultConfigDirectory()
}

// ResolvedConfig is an alias to Config for backward compatibility
// In the new system, Config is already resolved
type ResolvedConfig = Config

// GetDefaults returns the default configuration
func GetDefaults() *ResolvedConfig {
	return &defaultConfig
}

// ValidationResult represents the outcome of validation
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// IsValid returns whether the validation passed
func (vr *ValidationResult) IsValid() bool {
	return vr.Valid
}

// GetSummary returns a summary of the validation result
func (vr *ValidationResult) GetSummary() string {
	if vr.Valid {
		return "Configuration is valid"
	}

	summary := "Configuration is invalid:\n"
	for _, err := range vr.Errors {
		summary += fmt.Sprintf("  - %s\n", err)
	}

	if len(vr.Warnings) > 0 {
		summary += "\nWarnings:\n"
		for _, warn := range vr.Warnings {
			summary += fmt.Sprintf("  - %s\n", warn)
		}
	}

	return summary
}

// SimpleValidator provides validation for configuration
type SimpleValidator struct {
	validator *validator.Validate
}

// NewSimpleValidator creates a new validator
func NewSimpleValidator() *SimpleValidator {
	v := validator.New()
	return &SimpleValidator{validator: v}
}

// ValidateConfig validates a parsed config struct
func (v *SimpleValidator) ValidateConfig(config *Config) *ValidationResult {
	// Convert to NewConfig for validation
	resolved := config.Resolve()
	err := v.validator.Struct(resolved)

	result := &ValidationResult{
		Valid:    err == nil,
		Errors:   []string{},
		Warnings: []string{},
	}

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	return result
}

// ValidateConfigFromYAML validates configuration from YAML content
func (v *SimpleValidator) ValidateConfigFromYAML(content []byte) *ValidationResult {
	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return &ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("invalid YAML: %v", err)},
		}
	}

	// Apply defaults
	applyDefaults(&cfg)

	// Validate
	err := v.validator.Struct(&cfg)

	result := &ValidationResult{
		Valid:    err == nil,
		Errors:   []string{},
		Warnings: []string{},
	}

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	return result
}
