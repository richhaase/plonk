// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ConfigShowOutput represents the output structure for config show command
type ConfigShowOutput struct {
	ConfigPath string      `json:"config_path" yaml:"config_path"`
	Config     interface{} `json:"config" yaml:"config"`
	Checker    interface{} `json:"-" yaml:"-"` // Not included in JSON/YAML
	ConfigDir  string      `json:"-" yaml:"-"` // Not included in JSON/YAML
}

// ConfigShowFormatter formats config show output
type ConfigShowFormatter struct {
	Data ConfigShowOutput
}

// NewConfigShowFormatter creates a new formatter
func NewConfigShowFormatter(data ConfigShowOutput) ConfigShowFormatter {
	return ConfigShowFormatter{Data: data}
}

// TableOutput generates human-friendly table output for config show
func (f ConfigShowFormatter) TableOutput() string {
	c := f.Data
	output := fmt.Sprintf("# Configuration for plonk\n")
	output += fmt.Sprintf("# Config file: %s\n\n", c.ConfigPath)

	if c.Config == nil {
		return output + "No configuration loaded\n"
	}

	// Marshal the entire config to YAML
	data, err := yaml.Marshal(c.Config)
	if err != nil {
		return output + "Error formatting configuration\n"
	}

	output += string(data)
	return output
}

// StructuredData returns the structured data for serialization
func (f ConfigShowFormatter) StructuredData() any {
	return f.Data
}
