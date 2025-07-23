// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/errors"
)

// ConfigLoader provides a centralized and consistent way to load configuration
// across all commands, supporting zero-config behavior.
type ConfigLoader struct {
	configDir string
	validator *SimpleValidator
}

// NewConfigLoader creates a new ConfigLoader for the specified config directory
func NewConfigLoader(configDir string) *ConfigLoader {
	return &ConfigLoader{
		configDir: configDir,
		validator: NewSimpleValidator(),
	}
}

// Load attempts to load the configuration, returning an error only for
// parse or validation failures. Missing config files return empty Config.
func (l *ConfigLoader) Load() (*Config, error) {
	return LoadConfig(l.configDir)
}

// LoadOrDefault loads the configuration, returning an empty Config if
// the file doesn't exist or if there are any errors. This ensures
// zero-config behavior across all commands.
func (l *ConfigLoader) LoadOrDefault() *Config {
	cfg, err := l.Load()
	if err != nil {
		// Log the error for debugging but continue with defaults
		// This maintains consistent zero-config behavior
		return &Config{}
	}
	return cfg
}

// LoadOrCreate loads the configuration if it exists, or creates the
// config directory if it doesn't. Returns error only for parse/validation
// failures or if directory creation fails.
func (l *ConfigLoader) LoadOrCreate() (*Config, error) {
	return GetOrCreateConfig(l.configDir)
}

// EnsureConfigDir ensures the configuration directory exists
func (l *ConfigLoader) EnsureConfigDir() error {
	return os.MkdirAll(l.configDir, 0750)
}

// ConfigPath returns the full path to the plonk.yaml file
func (l *ConfigLoader) ConfigPath() string {
	return filepath.Join(l.configDir, "plonk.yaml")
}

// Exists checks if the configuration file exists
func (l *ConfigLoader) Exists() bool {
	_, err := os.Stat(l.ConfigPath())
	return err == nil
}

// Validate validates the configuration at the current path
func (l *ConfigLoader) Validate() (*ValidationResult, error) {
	cfg, err := l.Load()
	if err != nil {
		// Return validation result with the error
		return &ValidationResult{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	return l.validator.ValidateConfig(cfg), nil
}

// DefaultConfigLoader creates a ConfigLoader for the default config directory
func DefaultConfigLoader() *ConfigLoader {
	return NewConfigLoader(GetDefaultConfigDirectory())
}

// LoadConfigWithDefaults is a convenience function that loads config
// from the specified directory and returns empty config on any error.
// This is the recommended pattern for commands that need zero-config behavior.
func LoadConfigWithDefaults(configDir string) *Config {
	loader := NewConfigLoader(configDir)
	return loader.LoadOrDefault()
}

// ConfigManager provides high-level configuration management operations.
// It wraps ConfigLoader and YAMLConfigService to provide a unified interface.
type ConfigManager struct {
	loader  *ConfigLoader
	service *YAMLConfigService
}

// NewConfigManager creates a new ConfigManager for the specified directory
func NewConfigManager(configDir string) *ConfigManager {
	return &ConfigManager{
		loader:  NewConfigLoader(configDir),
		service: NewYAMLConfigService(),
	}
}

// Load loads the configuration, returning error only for parse/validation failures
func (m *ConfigManager) Load() (*Config, error) {
	return m.loader.Load()
}

// LoadOrDefault loads the configuration, returning empty config on any error
func (m *ConfigManager) LoadOrDefault() *Config {
	return m.loader.LoadOrDefault()
}

// LoadOrCreate ensures the config directory exists and loads the configuration
func (m *ConfigManager) LoadOrCreate() (*Config, error) {
	return m.loader.LoadOrCreate()
}

// Save saves the configuration to the config directory
func (m *ConfigManager) Save(cfg *Config) error {
	if err := m.loader.EnsureConfigDir(); err != nil {
		return errors.Wrap(err, errors.ErrDirectoryCreate, errors.DomainConfig, "save",
			"failed to create config directory").WithItem(m.loader.configDir)
	}

	return m.service.SaveConfig(m.loader.configDir, cfg)
}

// CreateDefault creates a default configuration file
func (m *ConfigManager) CreateDefault() error {
	cfg := &Config{}
	return m.Save(cfg)
}

// DefaultConfigManager creates a ConfigManager for the default config directory
func DefaultConfigManager() *ConfigManager {
	return NewConfigManager(GetDefaultConfigDirectory())
}
