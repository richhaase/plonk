// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"reflect"
	"strings"
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

	cfgVal := reflect.ValueOf(cfg).Elem()
	defaultVal := reflect.ValueOf(c.defaults).Elem()
	t := cfgVal.Type()

	for i := 0; i < t.NumField(); i++ {
		tag := strings.Split(t.Field(i).Tag.Get("yaml"), ",")[0]
		if tag == "" {
			continue
		}
		currentField := cfgVal.Field(i).Interface()
		defaultField := defaultVal.Field(i).Interface()
		if !reflect.DeepEqual(currentField, defaultField) {
			nonDefaults[tag] = currentField
		}
	}

	return nonDefaults
}

// getDefaultFieldValue returns the default value for a specific field
func (c *UserDefinedChecker) getDefaultFieldValue(fieldName string) interface{} {
	v := reflect.ValueOf(c.defaults).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		tag := strings.Split(t.Field(i).Tag.Get("yaml"), ",")[0]
		if tag == fieldName {
			return v.Field(i).Interface()
		}
	}
	return nil
}


