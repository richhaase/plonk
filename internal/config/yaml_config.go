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

	"plonk/internal/dotfiles"
	"plonk/internal/errors"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure.
type Config struct {
	Settings        Settings          `yaml:"settings" validate:"required"`
	IgnorePatterns  []string          `yaml:"ignore_patterns,omitempty"`
	Homebrew        []HomebrewPackage `yaml:"homebrew,omitempty" validate:"dive"`
	NPM             []NPMPackage      `yaml:"npm,omitempty" validate:"dive"`
}

// Settings contains global configuration settings.
type Settings struct {
	DefaultManager     string `yaml:"default_manager" validate:"required,oneof=homebrew npm"`
	OperationTimeout   int    `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"`        // Timeout in seconds for operations (0 for default, 1-3600 seconds)
	PackageTimeout     int    `yaml:"package_timeout,omitempty" validate:"omitempty,min=0,max=1800"`          // Timeout in seconds for package operations (0 for default, 1-1800 seconds)
	DotfileTimeout     int    `yaml:"dotfile_timeout,omitempty" validate:"omitempty,min=0,max=600"`           // Timeout in seconds for dotfile operations (0 for default, 1-600 seconds)
}




// HomebrewPackage can be a simple string or complex object.
type HomebrewPackage struct {
	Name   string `yaml:"name,omitempty" validate:"required,package_name"`
	Config string `yaml:"config,omitempty" validate:"omitempty,file_path"`
}



// NPMPackage represents an NPM package configuration.
type NPMPackage struct {
	Name    string `yaml:"name,omitempty" validate:"omitempty,package_name"`
	Package string `yaml:"package,omitempty" validate:"omitempty,package_name"` // If different from name.
	Config  string `yaml:"config,omitempty" validate:"omitempty,file_path"`
}

// MarshalYAML implements custom marshaling for HomebrewPackage.
func (h HomebrewPackage) MarshalYAML() (interface{}, error) {
	// If only Name is set, marshal as a simple string
	if h.Config == "" {
		return h.Name, nil
	}
	// Otherwise, marshal as a full object
	type homebrewPackageAlias HomebrewPackage
	return homebrewPackageAlias(h), nil
}

// UnmarshalYAML implements custom unmarshaling for HomebrewPackage.
func (h *HomebrewPackage) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		// Simple string case.
		h.Name = node.Value
		return nil
	}

	// Complex object case.
	type homebrewPackageAlias HomebrewPackage
	var pkg homebrewPackageAlias
	if err := node.Decode(&pkg); err != nil {
		return err
	}
	*h = HomebrewPackage(pkg)
	return nil
}

// MarshalYAML implements custom marshaling for NPMPackage.
func (n NPMPackage) MarshalYAML() (interface{}, error) {
	// If only Name is set (no Package or Config), marshal as a simple string
	if n.Package == "" && n.Config == "" {
		return n.Name, nil
	}
	// Otherwise, marshal as a full object
	type npmPackageAlias NPMPackage
	return npmPackageAlias(n), nil
}

// UnmarshalYAML implements custom unmarshaling for NPMPackage.
func (n *NPMPackage) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		// Simple string case.
		n.Name = node.Value
		return nil
	}

	// Complex object case.
	type npmPackageAlias NPMPackage
	var pkg npmPackageAlias
	if err := node.Decode(&pkg); err != nil {
		return err
	}
	*n = NPMPackage(pkg)
	return nil
}

// LoadConfig loads configuration from plonk.yaml.
func LoadConfig(configDir string) (*Config, error) {
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew", // Default value.
		},
	}

	// Load config file.
	configPath := filepath.Join(configDir, "plonk.yaml")

	if err := loadConfigFile(configPath, config); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.ConfigError(errors.ErrConfigNotFound, "load", 
				fmt.Sprintf("config file not found: %s", configPath)).
				WithMetadata("config_path", configPath)
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

// GetOperationTimeout returns the operation timeout in seconds with a default of 300 seconds (5 minutes)
func (s *Settings) GetOperationTimeout() int {
	if s.OperationTimeout <= 0 {
		return 300 // Default 5 minutes
	}
	return s.OperationTimeout
}

// GetPackageTimeout returns the package timeout in seconds with a default of 180 seconds (3 minutes)
func (s *Settings) GetPackageTimeout() int {
	if s.PackageTimeout <= 0 {
		return 180 // Default 3 minutes
	}
	return s.PackageTimeout
}

// GetDotfileTimeout returns the dotfile timeout in seconds with a default of 60 seconds (1 minute)
func (s *Settings) GetDotfileTimeout() int {
	if s.DotfileTimeout <= 0 {
		return 60 // Default 1 minute
	}
	return s.DotfileTimeout
}


// GetIgnorePatterns returns the ignore patterns with sensible defaults
func (c *Config) GetIgnorePatterns() []string {
	if len(c.IgnorePatterns) == 0 {
		return []string{
			".DS_Store",
			".git", 
			"*.backup",
			"*.tmp",
			"*.swp",
		}
	}
	return c.IgnorePatterns
}

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
	
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew", // Default value
		},
	}
	
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load", 
			"failed to parse YAML")
	}
	
	// Validate configuration
	result := y.validator.ValidateConfig(config)
	if !result.IsValid() {
		return nil, errors.ConfigError(errors.ErrConfigValidation, "validate", 
			fmt.Sprintf("config validation failed: %s", strings.Join(result.Errors, "; "))).
			WithMetadata("errors", result.Errors)
	}
	
	return config, nil
}

// SaveConfig saves configuration to a directory as plonk.yaml
func (y *YAMLConfigService) SaveConfig(configDir string, config *Config) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
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
func (y *YAMLConfigService) ValidateConfig(config *Config) error {
	result := y.validator.ValidateConfig(config)
	if !result.IsValid() {
		return errors.ConfigError(errors.ErrConfigValidation, "validate", 
			fmt.Sprintf("config validation failed: %s", strings.Join(result.Errors, "; "))).
			WithMetadata("errors", result.Errors)
	}
	return nil
}

// ValidateConfigFromReader validates configuration from an io.Reader
func (y *YAMLConfigService) ValidateConfigFromReader(reader io.Reader) error {
	config, err := y.LoadConfigFromReader(reader)
	if err != nil {
		return err
	}
	return y.ValidateConfig(config)
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
	ignorePatterns := c.config.GetIgnorePatterns()
	
	// Walk the directory to find all files
	filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
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
func (c *ConfigAdapter) GetPackagesForManager(managerName string) ([]PackageConfigItem, error) {
	var packageNames []string
	
	switch managerName {
	case "homebrew":
		// Get homebrew packages (unified)
		for _, pkg := range c.config.Homebrew {
			packageNames = append(packageNames, pkg.Name)
		}
	case "npm":
		// Get NPM packages
		for _, pkg := range c.config.NPM {
			packageNames = append(packageNames, pkg.Name)
		}
	default:
		return nil, errors.NewError(errors.ErrInvalidInput, errors.DomainConfig, "get-packages", 
			fmt.Sprintf("unknown package manager: %s", managerName)).WithItem(managerName)
	}
	
	items := make([]PackageConfigItem, len(packageNames))
	for i, name := range packageNames {
		items[i] = PackageConfigItem{Name: name}
	}
	
	return items, nil
}
