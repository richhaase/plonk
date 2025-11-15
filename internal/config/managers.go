// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

// ManagerConfig defines a package manager configuration
type ManagerConfig struct {
	Binary             string                             `yaml:"binary,omitempty"`
	List               ListConfig                         `yaml:"list,omitempty"`
	Install            CommandConfig                      `yaml:"install,omitempty"`
	Upgrade            CommandConfig                      `yaml:"upgrade,omitempty"`
	UpgradeAll         CommandConfig                      `yaml:"upgrade_all,omitempty"`
	Uninstall          CommandConfig                      `yaml:"uninstall,omitempty"`
	Description        string                             `yaml:"description,omitempty"`
	InstallHint        string                             `yaml:"install_hint,omitempty"`
	HelpURL            string                             `yaml:"help_url,omitempty"`
	NameTransform      *NameTransformConfig               `yaml:"name_transform,omitempty"`
	MetadataExtractors map[string]MetadataExtractorConfig `yaml:"metadata_extractors,omitempty"`
}

// ListConfig defines how to list installed packages
type ListConfig struct {
	Command       []string `yaml:"command,omitempty"`
	Parse         string   `yaml:"parse,omitempty"`
	ParseStrategy string   `yaml:"parse_strategy,omitempty"`
	JSONField     string   `yaml:"json_field,omitempty"`
}

// CommandConfig defines a package manager command
type CommandConfig struct {
	Command          []string `yaml:"command,omitempty"`
	IdempotentErrors []string `yaml:"idempotent_errors,omitempty"`
}

// NameTransformConfig defines how to normalize package names for a manager.
// Typical example: strip npm scopes or apply a regex rewrite.
type NameTransformConfig struct {
	// Type controls which transform to apply (e.g. "regex").
	Type string `yaml:"type,omitempty"`

	// Pattern is a regular expression used by regex-based transforms.
	Pattern string `yaml:"pattern,omitempty"`

	// Replacement is the replacement pattern for regex-based transforms.
	Replacement string `yaml:"replacement,omitempty"`
}

// MetadataExtractorConfig defines how to derive additional metadata fields
// (like scope, version, full_name) from parsed package names or JSON fields.
type MetadataExtractorConfig struct {
	// Pattern is an optional regular expression used to capture groups.
	Pattern string `yaml:"pattern,omitempty"`

	// Group selects which capturing group should be used for simple extractors.
	Group int `yaml:"group,omitempty"`

	// Source indicates where to read data from (e.g. "json_field", "name").
	Source string `yaml:"source,omitempty"`

	// Field is the JSON field name when Source is "json_field".
	Field string `yaml:"field,omitempty"`
}
