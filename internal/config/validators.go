// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"github.com/go-playground/validator/v10"
)

// validManagers holds the list of valid package managers.
// This is populated at runtime by the packages module via ManagerRegistry
// or explicitly in tests. When unset, we fall back to the default manager
// definitions from GetDefaultManagers.
var (
	validManagers        []string
	validManagersDefined bool
)

// SetValidManagers sets the list of valid package managers for validation.
// In production this is called from the packages module; tests may call it directly.
func SetValidManagers(managers []string) {
	validManagers = managers
	validManagersDefined = true
}

// RegisterValidators registers custom validators for config validation.
func RegisterValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("validmanager", validatePackageManager); err != nil {
		return err
	}
	return v.RegisterValidation("listconfig", validateListConfig)
}

// validatePackageManager validates that a package manager is registered.
// When no managers have been registered yet, any non-empty manager name is treated as invalid.
func validatePackageManager(fl validator.FieldLevel) bool {
	managerName := fl.Field().String()
	if managerName == "" {
		// Empty is valid (will use default).
		return true
	}

	managers := validManagers

	// If no managers have been registered dynamically, fall back to default managers.
	if !validManagersDefined {
		defaults := GetDefaultManagers()
		managers = make([]string, 0, len(defaults))
		for name := range defaults {
			managers = append(managers, name)
		}
	}

	// If managers is still empty (explicitly set to empty), treat any explicit manager as invalid.
	if len(managers) == 0 {
		return false
	}

	for _, valid := range managers {
		if managerName == valid {
			return true
		}
	}

	return false
}

// validateListConfig enforces consistency for ListConfig fields based on parse strategy.
func validateListConfig(fl validator.FieldLevel) bool {
	cfg, ok := fl.Field().Interface().(ListConfig)
	if !ok {
		return false
	}

	// Resolve strategy (parse or parse_strategy)
	strategy := cfg.Parse
	if strategy == "" {
		strategy = cfg.ParseStrategy
	}
	if strategy == "" {
		strategy = "lines"
	}

	switch strategy {
	case "lines":
		if cfg.JSONField != "" || cfg.KeysFrom != "" || cfg.ValuesFrom != "" {
			return false
		}
	case "json":
		if cfg.JSONField == "" {
			return false
		}
		if cfg.KeysFrom != "" || cfg.ValuesFrom != "" {
			return false
		}
	case "json-map":
		if cfg.KeysFrom != "" || cfg.ValuesFrom != "" {
			return false
		}
	case "jsonpath":
		if cfg.KeysFrom == "" && cfg.ValuesFrom == "" {
			return false
		}
	default:
		return false
	}

	// Normalize must be empty, "none", or "lower"
	switch cfg.Normalize {
	case "", "none", "lower":
		return true
	default:
		return false
	}
}
