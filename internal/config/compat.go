// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// This file provides minimal compatibility for the rest of the codebase
// until we can update all references to use the new API directly.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/paths"
	"gopkg.in/yaml.v3"
)

// ConfigManager manages configuration loading and saving
type ConfigManager struct {
	configDir string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configDir string) *ConfigManager {
	return &ConfigManager{configDir: configDir}
}

// LoadOrCreate loads existing configuration or creates new one with defaults
func (cm *ConfigManager) LoadOrCreate() (*OldConfig, error) {
	return LoadConfigOld(cm.configDir)
}

// Save saves the configuration to disk
func (cm *ConfigManager) Save(config *OldConfig) error {
	configPath := filepath.Join(cm.configDir, "plonk.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(configPath, data, 0644)
}

// ConfigAdapter adapts a loaded Config to provide domain-specific interfaces
type ConfigAdapter struct {
	config *NewConfig
}

// NewConfigAdapter creates a new config adapter
func NewConfigAdapter(config *Config) *ConfigAdapter {
	// Convert OldConfig to NewConfig if needed
	if config == nil {
		return &ConfigAdapter{config: &defaultConfig}
	}
	resolved := config.Resolve()
	return &ConfigAdapter{config: resolved}
}

// GetDotfileTargets returns a map of source -> destination paths for dotfiles
func (c *ConfigAdapter) GetDotfileTargets() map[string]string {
	// Use PathResolver to expand config directory and generate paths
	resolver, err := paths.NewPathResolverFromDefaults()
	if err != nil {
		// Handle error, log it, and return empty map
		return make(map[string]string)
	}

	// Get ignore patterns from config
	ignorePatterns := c.config.GetIgnorePatterns()

	// Delegate to PathResolver for directory expansion and path mapping
	result, err := resolver.ExpandConfigDirectory(ignorePatterns)
	if err != nil {
		// Handle error, log it, and return empty map
		return make(map[string]string)
	}

	return result
}

// GetPackagesForManager returns package names for a specific package manager
// NOTE: Packages are now managed by the lock file, so this always returns empty
func (c *ConfigAdapter) GetPackagesForManager(managerName string) ([]managers.PackageConfigItem, error) {
	// Validate manager name
	for _, supported := range managers.SupportedManagers {
		if managerName == supported {
			// Return empty slice - packages are now in lock file
			return []managers.PackageConfigItem{}, nil
		}
	}

	return nil, fmt.Errorf("unknown package manager: %s", managerName)
}

// StateDotfileConfigAdapter bridges the config package's ConfigAdapter to the state
// package's DotfileConfigLoader interface. This adapter prevents circular dependencies
// between the config and state packages, allowing the state package to consume
// dotfile configuration without directly importing the config package.
type StateDotfileConfigAdapter struct {
	configAdapter *ConfigAdapter
}

// NewStateDotfileConfigAdapter creates a new adapter for state dotfile interfaces
func NewStateDotfileConfigAdapter(configAdapter *ConfigAdapter) *StateDotfileConfigAdapter {
	return &StateDotfileConfigAdapter{configAdapter: configAdapter}
}

// GetDotfileTargets implements interfaces.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetDotfileTargets() map[string]string {
	return s.configAdapter.GetDotfileTargets()
}

// GetIgnorePatterns implements interfaces.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetIgnorePatterns() []string {
	return s.configAdapter.config.GetIgnorePatterns()
}

// GetExpandDirectories implements interfaces.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetExpandDirectories() []string {
	return s.configAdapter.config.GetExpandDirectories()
}

// TargetToSource is now provided by the paths package
func TargetToSource(target string) string {
	return paths.TargetToSource(target)
}

// Config is an alias to OldConfig for backward compatibility
type Config = OldConfig

// LoadConfig loads configuration and returns the old pointer-based structure
func LoadConfig(configDir string) (*Config, error) {
	return LoadConfigOld(configDir)
}

// LoadConfigWithDefaults loads configuration or returns defaults in old format
func LoadConfigWithDefaults(configDir string) *Config {
	return LoadConfigWithDefaultsOld(configDir)
}

// GetDefaultConfigDirectory returns the default config directory
func GetDefaultConfigDirectory() string {
	return paths.GetDefaultConfigDirectory()
}

// ResolvedConfig is an alias to NewConfig for backward compatibility
// In the new system, Config is already resolved
type ResolvedConfig = NewConfig

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
	var cfg NewConfig
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
