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

	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/errors"
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

// shouldSkipDotfile determines if a file/directory should be skipped during auto-discovery
func shouldSkipDotfile(relPath string, info os.FileInfo, ignorePatterns []string) bool {
	// Always skip plonk config file
	if relPath == "plonk.yaml" {
		return true
	}

	// Check against configured ignore patterns
	for _, pattern := range ignorePatterns {
		// Check exact match for file/directory name
		if pattern == info.Name() || pattern == relPath {
			return true
		}

		// Check glob pattern match
		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}

		// Handle directory patterns like .git/
		if strings.HasSuffix(pattern, "/") && info.IsDir() {
			dirPattern := strings.TrimSuffix(pattern, "/")
			if dirPattern == info.Name() || dirPattern == relPath {
				return true
			}
		}

		// Handle prefix patterns for directories
		if strings.HasPrefix(relPath, pattern+"/") {
			return true
		}
	}

	return false
}

// sourceToTarget converts a source path to target path using our convention
// Prepends ~/. to make all files/directories hidden
// Examples:
//
//	config/nvim/ -> ~/.config/nvim/
//	zshrc -> ~/.zshrc
//	editorconfig -> ~/.editorconfig
func sourceToTarget(source string) string {
	return "~/." + source
}

// TargetToSource converts a target path to source path using our convention
// Removes the ~/. prefix
// Examples:
//
//	~/.config/nvim/ -> config/nvim/
//	~/.zshrc -> zshrc
//	~/.editorconfig -> editorconfig
func TargetToSource(target string) string {
	// Remove ~/. prefix if present
	if len(target) > 3 && target[:3] == "~/." {
		return target[3:]
	}
	// Remove ~/ prefix if present (shouldn't happen with our convention)
	if len(target) > 2 && target[:2] == "~/" {
		return target[2:]
	}
	return target
}

// Legacy timeout methods removed - use ResolvedConfig.Get*Timeout() instead

// GetDefaultConfigDirectory returns the default config directory, checking PLONK_DIR environment variable first
func GetDefaultConfigDirectory() string {
	// Check for PLONK_DIR environment variable
	if envDir := os.Getenv("PLONK_DIR"); envDir != "" {
		// Expand ~ if present
		if strings.HasPrefix(envDir, "~/") {
			return filepath.Join(os.Getenv("HOME"), envDir[2:])
		}
		return envDir
	}

	// Default location
	return filepath.Join(os.Getenv("HOME"), ".config", "plonk")
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
	result := make(map[string]string)

	// Auto-discover dotfiles from configured directory
	configDir := GetDefaultConfigDirectory()
	resolvedConfig := c.config.Resolve()
	ignorePatterns := resolvedConfig.GetIgnorePatterns()

	// Walk the directory to find all files
	_ = filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		// Get relative path from config dir
		relPath, err := filepath.Rel(configDir, path)
		if err != nil {
			return nil
		}

		// Skip certain files and directories
		if shouldSkipDotfile(relPath, info, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories themselves (we'll get the files inside)
		if info.IsDir() {
			return nil
		}

		// Add to results with proper mapping
		source := relPath
		target := sourceToTarget(source)
		result[source] = target

		return nil
	})

	return result
}

// GetPackagesForManager returns package names for a specific package manager
// NOTE: Packages are now managed by the lock file, so this always returns empty
func (c *ConfigAdapter) GetPackagesForManager(managerName string) ([]PackageConfigItem, error) {
	// Validate manager name
	switch managerName {
	case "homebrew", "npm", "cargo":
		// Return empty slice - packages are now in lock file
		return []PackageConfigItem{}, nil
	default:
		return nil, errors.NewError(errors.ErrInvalidInput, errors.DomainConfig, "get-packages",
			fmt.Sprintf("unknown package manager: %s", managerName)).WithItem(managerName)
	}
}
