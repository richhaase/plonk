// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package interfaces

import "io"

// ConfigReader provides read operations for configuration
type ConfigReader interface {
	LoadConfig(configDir string) (interface{}, error)
	LoadConfigFromFile(filePath string) (interface{}, error)
	LoadConfigFromReader(reader io.Reader) (interface{}, error)
}

// ConfigWriter provides write operations for configuration
type ConfigWriter interface {
	SaveConfig(config interface{}, configDir string) error
	SaveConfigToFile(config interface{}, filePath string) error
	SaveConfigToWriter(config interface{}, writer io.Writer) error
}

// ConfigValidator provides validation for configuration
type ConfigValidator interface {
	ValidateConfig(config interface{}) error
	ValidateConfigFromFile(filePath string) error
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
