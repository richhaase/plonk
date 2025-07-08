// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package config provides configuration management for Plonk, including YAML
// configuration parsing, validation, and generation of shell configuration files.
//
// The package supports loading configuration from plonk.yaml and plonk.local.yaml,
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
	Settings Settings       `yaml:"settings" validate:"required"`
	Backup   BackupConfig   `yaml:"backup,omitempty"`
	Dotfiles []DotfileEntry `yaml:"dotfiles,omitempty" validate:"dive"`
	Homebrew HomebrewConfig `yaml:"homebrew,omitempty"`
	NPM      []NPMPackage   `yaml:"npm,omitempty" validate:"dive"`
}

// Settings contains global configuration settings.
type Settings struct {
	DefaultManager     string `yaml:"default_manager" validate:"required,oneof=homebrew npm"`
	OperationTimeout   int    `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"`        // Timeout in seconds for operations (0 for default, 1-3600 seconds)
	PackageTimeout     int    `yaml:"package_timeout,omitempty" validate:"omitempty,min=0,max=1800"`          // Timeout in seconds for package operations (0 for default, 1-1800 seconds)
	DotfileTimeout     int    `yaml:"dotfile_timeout,omitempty" validate:"omitempty,min=0,max=600"`           // Timeout in seconds for dotfile operations (0 for default, 1-600 seconds)
}

// BackupConfig contains backup configuration settings.
type BackupConfig struct {
	Location  string `yaml:"location,omitempty"`
	KeepCount int    `yaml:"keep_count,omitempty"`
}

// DotfileEntry represents a dotfile configuration entry.
type DotfileEntry struct {
	Source      string `yaml:"source,omitempty" validate:"omitempty,file_path"`
	Destination string `yaml:"destination,omitempty" validate:"omitempty,file_path"`
}

// HomebrewConfig contains homebrew package lists.
type HomebrewConfig struct {
	Brews []HomebrewPackage `yaml:"brews,omitempty" validate:"dive"`
	Casks []HomebrewPackage `yaml:"casks,omitempty" validate:"dive"`
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

// MarshalYAML implements custom marshaling for DotfileEntry.
func (d DotfileEntry) MarshalYAML() (interface{}, error) {
	// If only Destination is set (simple case), marshal as a simple string
	if d.Source == "" && d.Destination != "" {
		return d.Destination, nil
	}
	// Otherwise, marshal as a full object
	type dotfileEntryAlias DotfileEntry
	return dotfileEntryAlias(d), nil
}

// UnmarshalYAML implements custom unmarshaling for DotfileEntry.
func (d *DotfileEntry) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		// Simple string case - treat as destination
		d.Destination = node.Value
		d.Source = "" // Will be inferred later
		return nil
	}

	// Complex object case.
	type dotfileEntryAlias DotfileEntry
	var entry dotfileEntryAlias
	if err := node.Decode(&entry); err != nil {
		return err
	}
	*d = DotfileEntry(entry)
	return nil
}

// LoadConfig loads configuration from plonk.yaml and optionally plonk.local.yaml.
func LoadConfig(configDir string) (*Config, error) {
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew", // Default value.
		},
	}

	// Load main config file - check both main directory and repo subdirectory.
	mainConfigPath := filepath.Join(configDir, "plonk.yaml")
	repoConfigPath := filepath.Join(configDir, "repo", "plonk.yaml")

	// Try main directory first.
	if err := loadConfigFile(mainConfigPath, config); err != nil {
		if os.IsNotExist(err) {
			// Try repo subdirectory.
			if err := loadConfigFile(repoConfigPath, config); err != nil {
				if os.IsNotExist(err) {
					return nil, errors.ConfigError(errors.ErrConfigNotFound, "load", 
						fmt.Sprintf("config file not found in %s or %s", mainConfigPath, repoConfigPath)).
						WithMetadata("main_path", mainConfigPath).
						WithMetadata("repo_path", repoConfigPath)
				}
				return nil, errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load", 
					"failed to load config from repo directory").
					WithMetadata("path", repoConfigPath)
			}
		} else {
			return nil, errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load", 
				"failed to load config").WithMetadata("path", mainConfigPath)
		}
	}

	// Load local config file if it exists.
	localConfigPath := filepath.Join(configDir, "plonk.local.yaml")
	if _, err := os.Stat(localConfigPath); err == nil {
		if err := loadConfigFile(localConfigPath, config); err != nil {
			return nil, errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "load", 
				"failed to load local config").WithMetadata("path", localConfigPath)
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

// loadConfigFile loads a single YAML config file and merges it into the config.
func loadConfigFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Create a temporary config to decode into.
	var tempConfig Config
	if err := yaml.Unmarshal(data, &tempConfig); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Merge settings.
	if tempConfig.Settings.DefaultManager != "" {
		config.Settings.DefaultManager = tempConfig.Settings.DefaultManager
	}

	// Merge dotfiles.
	config.Dotfiles = append(config.Dotfiles, tempConfig.Dotfiles...)

	// Merge homebrew packages.
	config.Homebrew.Brews = append(config.Homebrew.Brews, tempConfig.Homebrew.Brews...)
	config.Homebrew.Casks = append(config.Homebrew.Casks, tempConfig.Homebrew.Casks...)


	// Merge NPM packages.
	config.NPM = append(config.NPM, tempConfig.NPM...)

	// Merge backup configuration.
	if tempConfig.Backup.Location != "" {
		config.Backup.Location = tempConfig.Backup.Location
	}
	if tempConfig.Backup.KeepCount > 0 {
		config.Backup.KeepCount = tempConfig.Backup.KeepCount
	}

	return nil
}

// GetDotfileTargets returns dotfiles with their target paths.
func (c *Config) GetDotfileTargets() map[string]string {
	result := make(map[string]string)
	for _, dotfile := range c.Dotfiles {
		source := dotfile.Source
		destination := dotfile.Destination
		
		// If source is empty, infer from destination
		if source == "" {
			source = TargetToSource(destination)
		}
		
		// If destination is empty, infer from source
		if destination == "" {
			destination = sourceToTarget(source)
		}
		
		result[source] = destination
	}
	return result
}

// sourceToTarget converts a source path to target path using our convention
// Examples:
//
//	config/nvim/ -> ~/.config/nvim/
//	zshrc -> ~/.zshrc
//	dot_gitconfig -> ~/.gitconfig
func sourceToTarget(source string) string {
	// Handle dot_ prefix convention.
	if len(source) > 4 && source[:4] == "dot_" {
		return "~/." + source[4:]
	}

	// Handle config/ directory.
	if len(source) > 7 && source[:7] == "config/" {
		return "~/." + source
	}

	// Default: add ~/. prefix.
	return "~/." + source
}

// TargetToSource converts a target path to source path using our convention
// Examples:
//
//	~/.config/nvim/ -> config/nvim/
//	~/.zshrc -> zshrc
//	~/.gitconfig -> dot_gitconfig
func TargetToSource(target string) string {
	// Remove ~/ prefix if present
	if len(target) > 2 && target[:2] == "~/" {
		target = target[2:]
	}
	
	// Remove . prefix if present
	if len(target) > 1 && target[:1] == "." {
		target = target[1:]
	}
	
	// Handle .config/ directory
	if len(target) > 7 && target[:7] == "config/" {
		return target
	}
	
	// Default: add dot_ prefix for dotfiles
	return "dot_" + target
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

// LoadConfig loads configuration from a directory containing plonk.yaml and plonk.local.yaml
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
	return c.config.GetDotfileTargets()
}

// GetPackagesForManager returns package names for a specific package manager
func (c *ConfigAdapter) GetPackagesForManager(managerName string) ([]PackageConfigItem, error) {
	var packageNames []string
	
	switch managerName {
	case "homebrew":
		// Get homebrew brews
		for _, brew := range c.config.Homebrew.Brews {
			packageNames = append(packageNames, brew.Name)
		}
		// Get homebrew casks
		for _, cask := range c.config.Homebrew.Casks {
			packageNames = append(packageNames, cask.Name)
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
