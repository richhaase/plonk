// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

// ConfigDefaults provides centralized default values for all configuration settings
type ConfigDefaults struct {
	DefaultManager    string
	OperationTimeout  int
	PackageTimeout    int
	DotfileTimeout    int
	ExpandDirectories []string
	IgnorePatterns    []string
}

// GetDefaults returns the default configuration values
// This is the single source of truth for all default values in Plonk
func GetDefaults() ConfigDefaults {
	return ConfigDefaults{
		DefaultManager:   "homebrew",
		OperationTimeout: 300, // 5 minutes
		PackageTimeout:   180, // 3 minutes
		DotfileTimeout:   60,  // 1 minute
		ExpandDirectories: []string{
			".config",
			".ssh",
			".aws",
			".kube",
			".docker",
			".gnupg",
			".local",
		},
		IgnorePatterns: []string{
			".DS_Store",
			".git",
			"*.backup",
			"*.tmp",
			"*.swp",
			"plonk.lock",
		},
	}
}
