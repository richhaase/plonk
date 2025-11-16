// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"reflect"
)

// UserDefinedChecker helps identify which configuration values are user-defined
type UserDefinedChecker struct {
	defaults   *Config
	userConfig *Config
}

// NewUserDefinedChecker creates a checker that can identify user-defined values
func NewUserDefinedChecker(configDir string) *UserDefinedChecker {
	// Create a copy of the default config
	defaults := defaultConfig
	userConfig, _ := Load(configDir) // May fail if no user config exists

	return &UserDefinedChecker{
		defaults:   &defaults,
		userConfig: userConfig,
	}
}

// IsFieldUserDefined checks if a specific field value differs from the default
func (c *UserDefinedChecker) IsFieldUserDefined(fieldName string, currentValue interface{}) bool {
	// If we don't have a user config, nothing is user-defined
	if c.userConfig == nil {
		return false
	}

	// Get the default value for comparison
	defaultValue := c.getDefaultFieldValue(fieldName)

	// Use reflect.DeepEqual to compare values
	return !reflect.DeepEqual(currentValue, defaultValue)
}

// GetNonDefaultFields returns a map of only the fields that differ from defaults
func (c *UserDefinedChecker) GetNonDefaultFields(cfg *Config) map[string]interface{} {
	nonDefaults := make(map[string]interface{})

	// Compare each field
	if cfg.DefaultManager != c.defaults.DefaultManager {
		nonDefaults["default_manager"] = cfg.DefaultManager
	}

	if cfg.OperationTimeout != c.defaults.OperationTimeout {
		nonDefaults["operation_timeout"] = cfg.OperationTimeout
	}

	if cfg.PackageTimeout != c.defaults.PackageTimeout {
		nonDefaults["package_timeout"] = cfg.PackageTimeout
	}

	if cfg.DotfileTimeout != c.defaults.DotfileTimeout {
		nonDefaults["dotfile_timeout"] = cfg.DotfileTimeout
	}

	if cfg.DiffTool != c.defaults.DiffTool && cfg.DiffTool != "" {
		nonDefaults["diff_tool"] = cfg.DiffTool
	}

	// For lists, save entire list if ANY element differs
	if !reflect.DeepEqual(cfg.ExpandDirectories, c.defaults.ExpandDirectories) {
		nonDefaults["expand_directories"] = cfg.ExpandDirectories
	}

	if !reflect.DeepEqual(cfg.IgnorePatterns, c.defaults.IgnorePatterns) {
		nonDefaults["ignore_patterns"] = cfg.IgnorePatterns
	}

	// Handle nested structures
	if !reflect.DeepEqual(cfg.Dotfiles, c.defaults.Dotfiles) {
		nonDefaults["dotfiles"] = cfg.Dotfiles
	}

	return nonDefaults
}

// getDefaultFieldValue returns the default value for a specific field
func (c *UserDefinedChecker) getDefaultFieldValue(fieldName string) interface{} {
	switch fieldName {
	case "default_manager":
		return c.defaults.DefaultManager
	case "operation_timeout":
		return c.defaults.OperationTimeout
	case "package_timeout":
		return c.defaults.PackageTimeout
	case "dotfile_timeout":
		return c.defaults.DotfileTimeout
	case "expand_directories":
		return c.defaults.ExpandDirectories
	case "ignore_patterns":
		return c.defaults.IgnorePatterns
	case "dotfiles":
		return c.defaults.Dotfiles
	case "diff_tool":
		return c.defaults.DiffTool
	default:
		return nil
	}
}

// GetNonDefaultManagers returns a map of manager configurations that differ from
// the built-in defaults. Managers that are not part of the defaults are treated
// as custom and always included in the result.
func (c *UserDefinedChecker) GetNonDefaultManagers(cfg *Config) map[string]ManagerConfig {
	result := make(map[string]ManagerConfig)
	if cfg == nil || cfg.Managers == nil {
		return result
	}

	defaults := GetDefaultManagers()

	for name, mgrCfg := range cfg.Managers {
		defaultMgr, ok := defaults[name]
		if !ok {
			// Custom manager not in defaults
			result[name] = mgrCfg
			continue
		}

		if diff, changed := diffManagerConfig(defaultMgr, mgrCfg); changed {
			result[name] = diff
		}
	}

	return result
}

func diffManagerConfig(base, actual ManagerConfig) (ManagerConfig, bool) {
	var diff ManagerConfig
	modified := false

	if actual.Binary != "" && actual.Binary != base.Binary {
		diff.Binary = actual.Binary
		modified = true
	}

	if listDiff, changed := diffListConfig(base.List, actual.List); changed {
		diff.List = listDiff
		modified = true
	}

	if installDiff, changed := diffCommandConfig(base.Install, actual.Install); changed {
		diff.Install = installDiff
		modified = true
	}

	if upgradeDiff, changed := diffCommandConfig(base.Upgrade, actual.Upgrade); changed {
		diff.Upgrade = upgradeDiff
		modified = true
	}

	if upgradeAllDiff, changed := diffCommandConfig(base.UpgradeAll, actual.UpgradeAll); changed {
		diff.UpgradeAll = upgradeAllDiff
		modified = true
	}

	if uninstallDiff, changed := diffCommandConfig(base.Uninstall, actual.Uninstall); changed {
		diff.Uninstall = uninstallDiff
		modified = true
	}

	if actual.Description != "" && actual.Description != base.Description {
		diff.Description = actual.Description
		modified = true
	}

	if actual.InstallHint != "" && actual.InstallHint != base.InstallHint {
		diff.InstallHint = actual.InstallHint
		modified = true
	}

	if actual.HelpURL != "" && actual.HelpURL != base.HelpURL {
		diff.HelpURL = actual.HelpURL
		modified = true
	}

	if actual.UpgradeTarget != "" && actual.UpgradeTarget != base.UpgradeTarget {
		diff.UpgradeTarget = actual.UpgradeTarget
		modified = true
	}

	if actual.NameTransform != nil {
		if base.NameTransform == nil || !reflect.DeepEqual(actual.NameTransform, base.NameTransform) {
			clone := *actual.NameTransform
			diff.NameTransform = &clone
			modified = true
		}
	}

	if mdDiff, changed := diffMetadataExtractors(base.MetadataExtractors, actual.MetadataExtractors); changed {
		diff.MetadataExtractors = mdDiff
		modified = true
	}

	return diff, modified
}

func diffListConfig(base, actual ListConfig) (ListConfig, bool) {
	var diff ListConfig
	modified := false

	if !reflect.DeepEqual(actual.Command, base.Command) && len(actual.Command) > 0 {
		diff.Command = copyStringSlice(actual.Command)
		modified = true
	}

	if actual.Parse != "" && actual.Parse != base.Parse {
		diff.Parse = actual.Parse
		modified = true
	}

	if actual.ParseStrategy != "" && actual.ParseStrategy != base.ParseStrategy {
		diff.ParseStrategy = actual.ParseStrategy
		modified = true
	}

	if actual.JSONField != "" && actual.JSONField != base.JSONField {
		diff.JSONField = actual.JSONField
		modified = true
	}

	return diff, modified
}

func diffCommandConfig(base, actual CommandConfig) (CommandConfig, bool) {
	var diff CommandConfig
	modified := false

	if !reflect.DeepEqual(actual.Command, base.Command) && len(actual.Command) > 0 {
		diff.Command = copyStringSlice(actual.Command)
		modified = true
	}

	if !reflect.DeepEqual(actual.IdempotentErrors, base.IdempotentErrors) && len(actual.IdempotentErrors) > 0 {
		diff.IdempotentErrors = copyStringSlice(actual.IdempotentErrors)
		modified = true
	}

	return diff, modified
}

func diffMetadataExtractors(base, actual map[string]MetadataExtractorConfig) (map[string]MetadataExtractorConfig, bool) {
	if len(actual) == 0 {
		return nil, false
	}

	result := make(map[string]MetadataExtractorConfig)
	for key, value := range actual {
		if baseVal, ok := base[key]; !ok || !reflect.DeepEqual(value, baseVal) {
			result[key] = value
		}
	}

	if len(result) == 0 {
		return nil, false
	}

	return result, true
}

func copyStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
