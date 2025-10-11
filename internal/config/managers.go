// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

// ManagerConfig defines a package manager configuration
type ManagerConfig struct {
	Binary     string        `yaml:"binary,omitempty"`
	List       ListConfig    `yaml:"list,omitempty"`
	Install    CommandConfig `yaml:"install,omitempty"`
	Upgrade    CommandConfig `yaml:"upgrade,omitempty"`
	UpgradeAll CommandConfig `yaml:"upgrade_all,omitempty"`
	Uninstall  CommandConfig `yaml:"uninstall,omitempty"`
}

// ListConfig defines how to list installed packages
type ListConfig struct {
	Command   []string `yaml:"command,omitempty"`
	Parse     string   `yaml:"parse,omitempty"`
	JSONField string   `yaml:"json_field,omitempty"`
}

// CommandConfig defines a package manager command
type CommandConfig struct {
	Command          []string `yaml:"command,omitempty"`
	IdempotentErrors []string `yaml:"idempotent_errors,omitempty"`
}
