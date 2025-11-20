// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

// ManagerConfig defines a package manager configuration
type ManagerConfig struct {
	Binary             string                             `yaml:"binary,omitempty"`
	List               ListConfig                         `yaml:"list,omitempty" validate:"listconfig"`
	Install            CommandConfig                      `yaml:"install,omitempty"`
	Upgrade            CommandConfig                      `yaml:"upgrade,omitempty"`
	UpgradeAll         CommandConfig                      `yaml:"upgrade_all,omitempty"`
	Uninstall          CommandConfig                      `yaml:"uninstall,omitempty"`
	Description        string                             `yaml:"description,omitempty"`
	InstallHint        string                             `yaml:"install_hint,omitempty"`
	HelpURL            string                             `yaml:"help_url,omitempty"`
	UpgradeTarget      string                             `yaml:"upgrade_target,omitempty"`
	NameTransform      *NameTransformConfig               `yaml:"name_transform,omitempty"`
	MetadataExtractors map[string]MetadataExtractorConfig `yaml:"metadata_extractors,omitempty"`
}

// ListConfig defines how to list installed packages
type ListConfig struct {
	Command       []string `yaml:"command,omitempty"`
	Parse         string   `yaml:"parse,omitempty"`
	ParseStrategy string   `yaml:"parse_strategy,omitempty"`
	JSONField     string   `yaml:"json_field,omitempty"`
	KeysFrom      string   `yaml:"keys_from,omitempty"`
	ValuesFrom    string   `yaml:"values_from,omitempty"`
	Normalize     string   `yaml:"normalize,omitempty"`
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

// MergeManagerConfig merges a base manager configuration with an override.
// Non-empty fields in override take precedence; empty fields inherit from base.
func MergeManagerConfig(base, override ManagerConfig) ManagerConfig {
	result := base

	if override.Binary != "" {
		result.Binary = override.Binary
	}

	result.List = mergeListConfig(base.List, override.List)

	result.Install = mergeCommandConfig(base.Install, override.Install)
	result.Upgrade = mergeCommandConfig(base.Upgrade, override.Upgrade)
	result.UpgradeAll = mergeCommandConfig(base.UpgradeAll, override.UpgradeAll)
	result.Uninstall = mergeCommandConfig(base.Uninstall, override.Uninstall)

	if override.Description != "" {
		result.Description = override.Description
	}
	if override.InstallHint != "" {
		result.InstallHint = override.InstallHint
	}
	if override.HelpURL != "" {
		result.HelpURL = override.HelpURL
	}
	if override.UpgradeTarget != "" {
		result.UpgradeTarget = override.UpgradeTarget
	}

	if override.NameTransform != nil {
		result.NameTransform = override.NameTransform
	}

	result.MetadataExtractors = mergeMetadataExtractors(base.MetadataExtractors, override.MetadataExtractors)

	return result
}

func mergeListConfig(base, override ListConfig) ListConfig {
	result := base

	if len(override.Command) > 0 {
		result.Command = override.Command
	}

	if override.Parse != "" {
		result.Parse = override.Parse
	}

	if override.ParseStrategy != "" {
		result.ParseStrategy = override.ParseStrategy
	}

	if override.JSONField != "" {
		result.JSONField = override.JSONField
	}

	if override.KeysFrom != "" {
		result.KeysFrom = override.KeysFrom
	}

	if override.ValuesFrom != "" {
		result.ValuesFrom = override.ValuesFrom
	}

	if override.Normalize != "" {
		result.Normalize = override.Normalize
	}

	return result
}

func mergeCommandConfig(base, override CommandConfig) CommandConfig {
	result := base

	if len(override.Command) > 0 {
		result.Command = override.Command
	}

	if len(override.IdempotentErrors) > 0 {
		result.IdempotentErrors = override.IdempotentErrors
	}

	return result
}

func mergeMetadataExtractors(base, override map[string]MetadataExtractorConfig) map[string]MetadataExtractorConfig {
	if base == nil && override == nil {
		return nil
	}

	// If no override is provided, return a copy of base to avoid aliasing.
	if override == nil {
		result := make(map[string]MetadataExtractorConfig, len(base))
		for k, v := range base {
			result[k] = v
		}
		return result
	}

	result := make(map[string]MetadataExtractorConfig, len(base)+len(override))

	for k, v := range base {
		result[k] = v
	}

	for k, v := range override {
		result[k] = v
	}

	return result
}
