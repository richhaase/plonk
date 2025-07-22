// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package interfaces

import (
	"io"
)

// Forward declare Config to avoid circular dependencies
// The actual Config struct is in internal/config package
type Config interface{}

// ConfigReader provides read operations for configuration
type ConfigReader interface {
	LoadConfig(configDir string) (Config, error)
	LoadConfigFromFile(filePath string) (Config, error)
	LoadConfigFromReader(reader io.Reader) (Config, error)
}

// ConfigWriter provides write operations for configuration
type ConfigWriter interface {
	SaveConfig(configDir string, config Config) error
	SaveConfigToFile(filePath string, config Config) error
	SaveConfigToWriter(writer io.Writer, config Config) error
}

// ConfigValidator provides validation for configuration
type ConfigValidator interface {
	ValidateConfig(config Config) error
	ValidateConfigFromReader(reader io.Reader) error
}

// DomainConfigLoader provides domain-specific configuration loading
type DomainConfigLoader interface {
	GetDotfileTargets() map[string]string
	GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
	GetIgnorePatterns() []string
	GetExpandDirectories() []string
}

// ConfigService combines all configuration capabilities
type ConfigService interface {
	ConfigReader
	ConfigWriter
	ConfigValidator
	DomainConfigLoader
}

// DotfileConfigLoader defines how to load dotfile configuration
type DotfileConfigLoader interface {
	GetDotfileTargets() map[string]string // source -> destination mapping
	GetIgnorePatterns() []string          // ignore patterns for file filtering
	GetExpandDirectories() []string       // directories to expand in dot list
}
