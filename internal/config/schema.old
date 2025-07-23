// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"encoding/json"
	"strings"

	"github.com/richhaase/plonk/internal/constants"
)

// SchemaProperty represents a JSON schema property
type SchemaProperty struct {
	Type        string                     `json:"type,omitempty"`
	Description string                     `json:"description,omitempty"`
	Items       *SchemaProperty            `json:"items,omitempty"`
	Properties  map[string]*SchemaProperty `json:"properties,omitempty"`
	Required    []string                   `json:"required,omitempty"`
	Default     interface{}                `json:"default,omitempty"`
	Examples    []interface{}              `json:"examples,omitempty"`
}

// Schema represents a JSON schema
type Schema struct {
	Schema      string                     `json:"$schema"`
	Type        string                     `json:"type"`
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Properties  map[string]*SchemaProperty `json:"properties"`
	Required    []string                   `json:"required,omitempty"`
}

// GenerateConfigSchema generates a JSON schema for the Config struct
func GenerateConfigSchema() *Schema {
	schema := &Schema{
		Schema:      "https://json-schema.org/draft/2020-12/schema",
		Type:        "object",
		Title:       "Plonk Configuration",
		Description: "Configuration file for plonk package and dotfile manager",
		Properties:  make(map[string]*SchemaProperty),
	}

	// Add properties based on the Config struct
	schema.Properties["default_manager"] = &SchemaProperty{
		Type:        "string",
		Description: "Default package manager to use when none is specified",
		Examples:    stringSliceToInterface(constants.SupportedManagers),
	}

	schema.Properties["operation_timeout"] = &SchemaProperty{
		Type:        "integer",
		Description: "Timeout in seconds for general operations (0 for unlimited, 1-3600 seconds)",
		Default:     60,
		Examples:    []interface{}{60, 120, 300},
	}

	schema.Properties["package_timeout"] = &SchemaProperty{
		Type:        "integer",
		Description: "Timeout in seconds for package operations (0 for unlimited, 1-1800 seconds)",
		Default:     300,
		Examples:    []interface{}{120, 300, 600},
	}

	schema.Properties["dotfile_timeout"] = &SchemaProperty{
		Type:        "integer",
		Description: "Timeout in seconds for dotfile operations (0 for unlimited, 1-600 seconds)",
		Default:     30,
		Examples:    []interface{}{30, 60, 120},
	}

	schema.Properties["expand_directories"] = &SchemaProperty{
		Type:        "array",
		Description: "Directories to expand in dot list output",
		Items: &SchemaProperty{
			Type: "string",
		},
		Examples: []interface{}{[]string{"~/.config", "~/.local"}},
	}

	schema.Properties["ignore_patterns"] = &SchemaProperty{
		Type:        "array",
		Description: "Glob patterns for files to ignore during dotfile discovery",
		Items: &SchemaProperty{
			Type: "string",
		},
		Default:  []string{".git", ".DS_Store", "*.tmp"},
		Examples: []interface{}{[]string{".git", ".DS_Store", "*.tmp", "node_modules"}},
	}

	return schema
}

// ToJSON converts the schema to JSON
func (s *Schema) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// GetConfigFieldDocumentation returns human-readable documentation for config fields
func GetConfigFieldDocumentation() map[string]string {
	return map[string]string{
		"default_manager": `The default package manager to use when installing packages without specifying a manager.`,

		"operation_timeout": `Timeout in seconds for general operations.
Set to 0 for unlimited timeout, or 1-3600 seconds.
Default: 60 seconds
Example: operation_timeout: 120`,

		"package_timeout": `Timeout in seconds for package manager operations.
Set to 0 for unlimited timeout, or 1-1800 seconds.
Default: 300 seconds
Example: package_timeout: 600`,

		"dotfile_timeout": `Timeout in seconds for dotfile operations.
Set to 0 for unlimited timeout, or 1-600 seconds.
Default: 30 seconds
Example: dotfile_timeout: 60`,

		"expand_directories": `Directories to expand when listing dotfiles.
Useful for organizing dotfile output by directory.
Example:
expand_directories:
  - ~/.config
  - ~/.local`,

		"ignore_patterns": `Glob patterns for files and directories to ignore during dotfile discovery.
Useful for excluding temporary files, version control directories, etc.
Default: [".git", ".DS_Store", "*.tmp"]
Example:
ignore_patterns:
  - .git
  - .DS_Store
  - "*.tmp"
  - node_modules`,
	}
}

// ValidateAgainstSchema validates a config against the generated schema (basic validation)
func ValidateAgainstSchema(cfg *Config) []string {
	var errors []string

	// Validate default_manager
	if cfg.DefaultManager != nil {
		if !contains(constants.SupportedManagers, *cfg.DefaultManager) {
			errors = append(errors, "default_manager must be one of: "+strings.Join(constants.SupportedManagers, ", "))
		}
	}

	// Validate timeouts
	if cfg.OperationTimeout != nil && (*cfg.OperationTimeout < 0 || *cfg.OperationTimeout > 3600) {
		errors = append(errors, "operation_timeout must be between 0 and 3600 seconds")
	}
	if cfg.PackageTimeout != nil && (*cfg.PackageTimeout < 0 || *cfg.PackageTimeout > 1800) {
		errors = append(errors, "package_timeout must be between 0 and 1800 seconds")
	}
	if cfg.DotfileTimeout != nil && (*cfg.DotfileTimeout < 0 || *cfg.DotfileTimeout > 600) {
		errors = append(errors, "dotfile_timeout must be between 0 and 600 seconds")
	}

	return errors
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// stringSliceToInterface converts []string to []interface{}
func stringSliceToInterface(slice []string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, s := range slice {
		result[i] = s
	}
	return result
}
