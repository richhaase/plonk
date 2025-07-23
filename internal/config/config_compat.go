// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// This file provides compatibility functions to ease migration from old to new config.
// It will be used in Phase 2 to allow atomic switch.

package config

// ConvertNewToOld converts NewConfig to old Config type
// This is used during migration to maintain compatibility
func ConvertNewToOld(nc *NewConfig) *Config {
	if nc == nil {
		return &Config{}
	}

	return &Config{
		DefaultManager:    &nc.DefaultManager,
		OperationTimeout:  &nc.OperationTimeout,
		PackageTimeout:    &nc.PackageTimeout,
		DotfileTimeout:    &nc.DotfileTimeout,
		ExpandDirectories: &nc.ExpandDirectories,
		IgnorePatterns:    nc.IgnorePatterns,
	}
}

// ConvertNewToResolvedConfig converts NewConfig to ResolvedConfig
// Since NewConfig is already resolved, this is a simple type conversion
func ConvertNewToResolvedConfig(nc *NewConfig) *ResolvedConfig {
	if nc == nil {
		defaults := GetDefaults()
		return &ResolvedConfig{
			DefaultManager:    defaults.DefaultManager,
			OperationTimeout:  defaults.OperationTimeout,
			PackageTimeout:    defaults.PackageTimeout,
			DotfileTimeout:    defaults.DotfileTimeout,
			ExpandDirectories: defaults.ExpandDirectories,
			IgnorePatterns:    defaults.IgnorePatterns,
		}
	}

	return &ResolvedConfig{
		DefaultManager:    nc.DefaultManager,
		OperationTimeout:  nc.OperationTimeout,
		PackageTimeout:    nc.PackageTimeout,
		DotfileTimeout:    nc.DotfileTimeout,
		ExpandDirectories: nc.ExpandDirectories,
		IgnorePatterns:    nc.IgnorePatterns,
	}
}

// MakeNewConfigResolve adds a Resolve method that returns ResolvedConfig
// This allows NewConfig to be used where Config.Resolve() is expected
func MakeNewConfigResolve(nc *NewConfig) *ResolvedConfig {
	return ConvertNewToResolvedConfig(nc)
}
