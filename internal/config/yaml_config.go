// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package config provides configuration management for Plonk, including YAML
// configuration parsing, validation, and generation of shell configuration files.
//
// The package supports loading configuration from plonk.yaml,
// validating package definitions and file paths, and generating shell-specific
// configuration files like .zshrc, .zshenv, and .gitconfig.
package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/constants"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/paths"
	"gopkg.in/yaml.v3"
)

// Config represents the user's configuration overrides.
// All fields are optional - defaults will be used for nil values.
type Config struct {
	DefaultManager    *string   `yaml:"default_manager,omitempty" validate:"omitempty,oneof=homebrew npm cargo"`
	OperationTimeout  *int      `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"` // Timeout in seconds for operations (0 for unlimited, 1-3600 seconds)
	PackageTimeout    *int      `yaml:"package_timeout,omitempty" validate:"omitempty,min=0,max=1800"`   // Timeout in seconds for package operations (0 for unlimited, 1-1800 seconds)
	DotfileTimeout    *int      `yaml:"dotfile_timeout,omitempty" validate:"omitempty,min=0,max=600"`    // Timeout in seconds for dotfile operations (0 for unlimited, 1-600 seconds)
	ExpandDirectories *[]string `yaml:"expand_directories,omitempty"`                                    // Directories to expand in dot list output
	IgnorePatterns    []string  `yaml:"ignore_patterns,omitempty"`
}

// Package types removed - packages now managed in lock file

// Package marshal/unmarshal methods removed - packages now managed in lock file

// Resolve merges user configuration with defaults to produce final configuration values
func (c *Config) Resolve() *ResolvedConfig {
	defaults := GetDefaults()

	return &ResolvedConfig{
		DefaultManager:    c.getDefaultManager(defaults.DefaultManager),
		OperationTimeout:  c.getOperationTimeout(defaults.OperationTimeout),
		PackageTimeout:    c.getPackageTimeout(defaults.PackageTimeout),
		DotfileTimeout:    c.getDotfileTimeout(defaults.DotfileTimeout),
		ExpandDirectories: c.getExpandDirectories(defaults.ExpandDirectories),
		IgnorePatterns:    c.getIgnorePatterns(defaults.IgnorePatterns),
	}
}

// getDefaultManager returns the user's default manager or the default value
func (c *Config) getDefaultManager(defaultValue string) string {
	if c.DefaultManager != nil {
		return *c.DefaultManager
	}
	return defaultValue
}

// getOperationTimeout returns the user's operation timeout or the default value
func (c *Config) getOperationTimeout(defaultValue int) int {
	if c.OperationTimeout != nil {
		return *c.OperationTimeout
	}
	return defaultValue
}

// getPackageTimeout returns the user's package timeout or the default value
func (c *Config) getPackageTimeout(defaultValue int) int {
	if c.PackageTimeout != nil {
		return *c.PackageTimeout
	}
	return defaultValue
}

// getDotfileTimeout returns the user's dotfile timeout or the default value
func (c *Config) getDotfileTimeout(defaultValue int) int {
	if c.DotfileTimeout != nil {
		return *c.DotfileTimeout
	}
	return defaultValue
}

// getExpandDirectories returns the user's expand directories or the default value
func (c *Config) getExpandDirectories(defaultValue []string) []string {
	if c.ExpandDirectories != nil {
		return *c.ExpandDirectories
	}
	return defaultValue
}

// getIgnorePatterns returns the user's ignore patterns or the default value
func (c *Config) getIgnorePatterns(defaultValue []string) []string {
	if len(c.IgnorePatterns) > 0 {
		return c.IgnorePatterns
	}
	return defaultValue
}

// LoadConfig loads configuration from plonk.yaml.
func LoadConfig(configDir string) (*Config, error) {
	config := &Config{}

	// Load config file.
	configPath := filepath.Join(configDir, "plonk.yaml")

	if err := loadConfigFile(configPath, config); err != nil {
		if os.IsNotExist(err) {
			// Zero-config: Return empty config when file doesn't exist
			// This will use all defaults when resolved
			config = &Config{}
		} else {
			return nil, errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load",
				"failed to load config").WithMetadata("path", configPath)
		}
	}

	// Validate configuration with new unified validator.
	validator := NewSimpleValidator()
	result := validator.ValidateConfig(config)
	if !result.IsValid() {
		return nil, errors.ConfigError(errors.ErrConfigValidation, "validate",
			fmt.Sprintf("config validation failed: %s", strings.Join(result.Errors, "; "))).
			WithMetadata("errors", result.Errors).
			WithMetadata("config_dir", configDir)
	}

	return config, nil
}

// GetOrCreateConfig loads configuration if it exists, or creates a default config
// with the specified directory if it doesn't exist. Useful for commands that need
// to save configuration.
func GetOrCreateConfig(configDir string) (*Config, error) {
	cfg, err := LoadConfig(configDir)
	if err != nil {
		// LoadConfig only returns errors for parse/validation failures now,
		// not for missing files (zero-config behavior)
		return nil, err
	}

	// Ensure config directory exists for future saves
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return nil, errors.Wrap(err, errors.ErrDirectoryCreate, errors.DomainConfig, "create",
			"failed to create config directory").WithItem(configDir)
	}

	return cfg, nil
}

// loadConfigFile loads a single YAML config file.
func loadConfigFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}

// TargetToSource is now provided by the paths package
func TargetToSource(target string) string {
	return paths.TargetToSource(target)
}

// Legacy timeout methods removed - use ResolvedConfig.Get*Timeout() instead

// GetDefaultConfigDirectory returns the default config directory, checking PLONK_DIR environment variable first
func GetDefaultConfigDirectory() string {
	// Delegate to the centralized path resolution in paths package
	return paths.GetDefaultConfigDirectory()
}

// YAMLConfigService implements all configuration interfaces for YAML-based configuration
type YAMLConfigService struct {
	validator    *SimpleValidator
	atomicWriter *dotfiles.AtomicFileWriter
}

// NewYAMLConfigService creates a new YAML configuration service
func NewYAMLConfigService() *YAMLConfigService {
	return &YAMLConfigService{
		validator:    NewSimpleValidator(),
		atomicWriter: dotfiles.NewAtomicFileWriter(),
	}
}

// LoadConfig loads configuration from a directory containing plonk.yaml
func (y *YAMLConfigService) LoadConfig(configDir string) (*Config, error) {
	return LoadConfig(configDir)
}

// LoadConfigFromFile loads configuration from a specific file path
func (y *YAMLConfigService) LoadConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load",
			"failed to read config file").WithItem(filePath)
	}

	return y.LoadConfigFromReader(strings.NewReader(string(data)))
}

// LoadConfigFromReader loads configuration from an io.Reader
func (y *YAMLConfigService) LoadConfigFromReader(reader io.Reader) (*Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load",
			"failed to read config data")
	}

	config := &Config{}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load",
			"failed to parse YAML")
	}

	// Validate configuration
	result := y.ValidateConfig(config)
	if !result.IsValid() {
		return nil, errors.ConfigError(errors.ErrConfigValidation, "validate",
			fmt.Sprintf("config validation failed: %s", strings.Join(result.Errors, "; "))).
			WithMetadata("errors", result.Errors)
	}

	return config, nil
}

// SaveConfig saves configuration to a directory as plonk.yaml
func (y *YAMLConfigService) SaveConfig(configDir string, config *Config) error {
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return errors.Wrap(err, errors.ErrDirectoryCreate, errors.DomainConfig, "save",
			"failed to create config directory").WithItem(configDir)
	}

	filePath := filepath.Join(configDir, "plonk.yaml")
	return y.SaveConfigToFile(filePath, config)
}

// SaveConfigToFile saves configuration to a specific file path atomically
func (y *YAMLConfigService) SaveConfigToFile(filePath string, config *Config) error {
	// Marshal configuration to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "save",
			"failed to marshal config to YAML")
	}

	// Write atomically
	if err := y.atomicWriter.WriteFile(filePath, data, 0644); err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainConfig, "save",
			"failed to write config file").WithItem(filePath)
	}

	return nil
}

// SaveConfigToWriter saves configuration to an io.Writer
func (y *YAMLConfigService) SaveConfigToWriter(writer io.Writer, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "save",
			"failed to marshal config to YAML")
	}

	if _, err := writer.Write(data); err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainConfig, "save",
			"failed to write config data")
	}

	return nil
}

// GetDotfileTargets returns a map of source -> destination paths for dotfiles
func (y *YAMLConfigService) GetDotfileTargets() map[string]string {
	// This method requires a Config instance, so we'll need to load it first
	// This is a limitation of the current design - we'll address it in the implementation
	panic("GetDotfileTargets requires a Config instance - use DotfileConfigAdapter instead")
}

// GetPackagesForManager returns package names for a specific package manager
func (y *YAMLConfigService) GetPackagesForManager(managerName string) ([]PackageConfigItem, error) {
	// This method requires a Config instance, so we'll need to load it first
	// This is a limitation of the current design - we'll address it in the implementation
	panic("GetPackagesForManager requires a Config instance - use PackageConfigAdapter instead")
}

// ValidateConfig validates a configuration object
func (y *YAMLConfigService) ValidateConfig(config *Config) *ValidationResult {
	return y.validator.ValidateConfig(config)
}

// ValidateConfigFromReader validates configuration from an io.Reader
func (y *YAMLConfigService) ValidateConfigFromReader(reader io.Reader) error {
	config, err := y.LoadConfigFromReader(reader)
	if err != nil {
		return err
	}
	result := y.ValidateConfig(config)
	if !result.IsValid() {
		return errors.ConfigError(errors.ErrConfigValidation, "validate",
			fmt.Sprintf("config validation failed: %s", strings.Join(result.Errors, "; "))).
			WithMetadata("errors", result.Errors)
	}
	return nil
}

// GetDefaultConfig returns a default configuration with sensible defaults
func (y *YAMLConfigService) GetDefaultConfig() *Config {
	defaults := GetDefaults()
	return &Config{
		DefaultManager:    &defaults.DefaultManager,
		OperationTimeout:  &defaults.OperationTimeout,
		PackageTimeout:    &defaults.PackageTimeout,
		DotfileTimeout:    &defaults.DotfileTimeout,
		ExpandDirectories: &defaults.ExpandDirectories,
		IgnorePatterns:    defaults.IgnorePatterns,
	}
}

// ConfigAdapter adapts a loaded Config to provide domain-specific interfaces
type ConfigAdapter struct {
	config *Config
}

// NewConfigAdapter creates a new config adapter
func NewConfigAdapter(config *Config) *ConfigAdapter {
	return &ConfigAdapter{config: config}
}

// GetDotfileTargets returns a map of source -> destination paths for dotfiles
func (c *ConfigAdapter) GetDotfileTargets() map[string]string {
	// Use PathResolver to expand config directory and generate paths
	resolver, err := paths.NewPathResolverFromDefaults()
	if err != nil {
		// Handle error, log it, and return empty map
		return make(map[string]string)
	}

	// Get ignore patterns from resolved config
	resolvedConfig := c.config.Resolve()
	ignorePatterns := resolvedConfig.GetIgnorePatterns()

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
func (c *ConfigAdapter) GetPackagesForManager(managerName string) ([]PackageConfigItem, error) {
	// Validate manager name
	for _, supported := range constants.SupportedManagers {
		if managerName == supported {
			// Return empty slice - packages are now in lock file
			return []PackageConfigItem{}, nil
		}
	}

	return nil, errors.NewError(errors.ErrInvalidInput, errors.DomainConfig, "get-packages",
		fmt.Sprintf("unknown package manager: %s", managerName)).WithItem(managerName)
}
