// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"reflect"
	"strings"
)

// GetNonDefaultFields returns a map of only the fields that differ from defaults
func GetNonDefaultFields(cfg *Config) map[string]interface{} {
	nonDefaults := make(map[string]interface{})

	cfgVal := reflect.ValueOf(cfg).Elem()
	defaultVal := reflect.ValueOf(&defaultConfig).Elem()
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
