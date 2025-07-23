// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// This file provides the old pointer-based Config struct for backward compatibility
// until all commands can be updated to use the new direct struct API.

package config

// OldConfig represents the old config structure with pointers
type OldConfig struct {
	DefaultManager    *string   `yaml:"default_manager,omitempty"`
	OperationTimeout  *int      `yaml:"operation_timeout,omitempty"`
	PackageTimeout    *int      `yaml:"package_timeout,omitempty"`
	DotfileTimeout    *int      `yaml:"dotfile_timeout,omitempty"`
	ExpandDirectories *[]string `yaml:"expand_directories,omitempty"`
	IgnorePatterns    []string  `yaml:"ignore_patterns,omitempty"`
}

// Resolve merges user configuration with defaults to produce final configuration values
func (c *OldConfig) Resolve() *ResolvedConfig {
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

// Helper methods for OldConfig
func (c *OldConfig) getDefaultManager(defaultValue string) string {
	if c.DefaultManager != nil {
		return *c.DefaultManager
	}
	return defaultValue
}

func (c *OldConfig) getOperationTimeout(defaultValue int) int {
	if c.OperationTimeout != nil {
		return *c.OperationTimeout
	}
	return defaultValue
}

func (c *OldConfig) getPackageTimeout(defaultValue int) int {
	if c.PackageTimeout != nil {
		return *c.PackageTimeout
	}
	return defaultValue
}

func (c *OldConfig) getDotfileTimeout(defaultValue int) int {
	if c.DotfileTimeout != nil {
		return *c.DotfileTimeout
	}
	return defaultValue
}

func (c *OldConfig) getExpandDirectories(defaultValue []string) []string {
	if c.ExpandDirectories != nil {
		return *c.ExpandDirectories
	}
	return defaultValue
}

func (c *OldConfig) getIgnorePatterns(defaultValue []string) []string {
	if len(c.IgnorePatterns) > 0 {
		return c.IgnorePatterns
	}
	return defaultValue
}

// ConvertNewToOld converts NewConfig to old OldConfig type
func ConvertNewToOld(nc *NewConfig) *OldConfig {
	if nc == nil {
		return &OldConfig{}
	}

	return &OldConfig{
		DefaultManager:    &nc.DefaultManager,
		OperationTimeout:  &nc.OperationTimeout,
		PackageTimeout:    &nc.PackageTimeout,
		DotfileTimeout:    &nc.DotfileTimeout,
		ExpandDirectories: &nc.ExpandDirectories,
		IgnorePatterns:    nc.IgnorePatterns,
	}
}

// LoadConfigOld loads configuration and returns the old pointer-based structure
func LoadConfigOld(configDir string) (*OldConfig, error) {
	nc, err := LoadNew(configDir)
	if err != nil {
		return nil, err
	}
	return ConvertNewToOld(nc), nil
}

// LoadConfigWithDefaultsOld loads configuration or returns defaults in old format
func LoadConfigWithDefaultsOld(configDir string) *OldConfig {
	nc := LoadNewWithDefaults(configDir)
	return ConvertNewToOld(nc)
}

// GetIgnorePatterns returns ignore patterns from OldConfig
func (c *OldConfig) GetIgnorePatterns() []string {
	resolved := c.Resolve()
	return resolved.GetIgnorePatterns()
}

// GetExpandDirectories returns expand directories from OldConfig
func (c *OldConfig) GetExpandDirectories() []string {
	resolved := c.Resolve()
	return resolved.GetExpandDirectories()
}
