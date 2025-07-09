// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package config provides configuration management interfaces for plonk.
// These interfaces enable loose coupling between configuration sources and consumers.
package config

import (
	"io"
)

// ConfigReader provides methods for loading configuration from various sources
type ConfigReader interface {
	// LoadConfig loads configuration from a directory containing plonk.yaml
	LoadConfig(configDir string) (*Config, error)
	
	// LoadConfigFromFile loads configuration from a specific file path
	LoadConfigFromFile(filePath string) (*Config, error)
	
	// LoadConfigFromReader loads configuration from an io.Reader
	LoadConfigFromReader(reader io.Reader) (*Config, error)
}

// ConfigWriter provides methods for saving configuration to various destinations
type ConfigWriter interface {
	// SaveConfig saves configuration to a directory as plonk.yaml
	SaveConfig(configDir string, config *Config) error
	
	// SaveConfigToFile saves configuration to a specific file path
	SaveConfigToFile(filePath string, config *Config) error
	
	// SaveConfigToWriter saves configuration to an io.Writer
	SaveConfigToWriter(writer io.Writer, config *Config) error
}

// ConfigReadWriter combines reading and writing capabilities
type ConfigReadWriter interface {
	ConfigReader
	ConfigWriter
}

// DotfileConfigReader provides methods for reading dotfile configuration
type DotfileConfigReader interface {
	// GetDotfileTargets returns a map of source -> destination paths for dotfiles
	GetDotfileTargets() map[string]string
}

// PackageConfigReader provides methods for reading package configuration
type PackageConfigReader interface {
	// GetPackagesForManager returns package names for a specific package manager
	GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
}

// PackageConfigItem represents a package from configuration
type PackageConfigItem struct {
	Name string
}

// ConfigValidator provides methods for validating configuration
type ConfigValidator interface {
	// ValidateConfig validates a configuration object
	ValidateConfig(config *Config) error
	
	// ValidateConfigFromReader validates configuration from an io.Reader
	ValidateConfigFromReader(reader io.Reader) error
}

// ConfigService combines all configuration interfaces for a complete configuration service
type ConfigService interface {
	ConfigReadWriter
	DotfileConfigReader
	PackageConfigReader
	ConfigValidator
}