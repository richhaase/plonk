// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import "context"

// ManagerConfig holds the configuration for creating a package manager
type ManagerConfig struct {
	Name           string
	BinaryName     string
	BinaryOptions  []string
	ErrorPatterns  map[ErrorType][]string
	SearchBinaries []string // For managers that have multiple possible binary names
}

// StandardManager represents the common fields shared by all managers
type StandardManager struct {
	Binary       string
	ErrorMatcher *ErrorMatcher
	ErrorHandler *ErrorHandler
}

// NewStandardManager creates a standardized manager configuration
func NewStandardManager(config ManagerConfig) *StandardManager {
	// Create base error matcher with common patterns
	errorMatcher := NewCommonErrorMatcher()

	// Add manager-specific error patterns
	for errorType, patterns := range config.ErrorPatterns {
		for _, pattern := range patterns {
			errorMatcher.AddPattern(errorType, pattern)
		}
	}

	// Find the available binary
	binary := config.BinaryName
	if len(config.SearchBinaries) > 0 {
		foundBinary := FindAvailableBinary(context.Background(), config.SearchBinaries, []string{"--version"})
		if foundBinary != "" {
			binary = foundBinary
		}
	}

	// Create error handler
	errorHandler := NewErrorHandler(errorMatcher, config.Name)

	return &StandardManager{
		Binary:       binary,
		ErrorMatcher: errorMatcher,
		ErrorHandler: errorHandler,
	}
}

// GetNpmConfig returns the configuration for NPM manager
func GetNpmConfig() ManagerConfig {
	return ManagerConfig{
		Name:       "npm",
		BinaryName: "npm",
		ErrorPatterns: map[ErrorType][]string{
			ErrorTypeNotFound:     {"404", "E404", "Not found"},
			ErrorTypePermission:   {"EACCES"},
			ErrorTypeNotInstalled: {"ENOENT", "cannot remove"},
		},
	}
}

// GetPipConfig returns the configuration for Pip manager
func GetPipConfig() ManagerConfig {
	return ManagerConfig{
		Name:           "pip",
		BinaryName:     "pip",
		SearchBinaries: []string{"pip", "pip3"},
		ErrorPatterns: map[ErrorType][]string{
			ErrorTypeNotFound:         {"Could not find", "No matching distribution", "ERROR: No matching distribution"},
			ErrorTypeAlreadyInstalled: {"Requirement already satisfied", "already satisfied"},
			ErrorTypeNotInstalled:     {"WARNING: Skipping", "not installed", "Cannot uninstall"},
			ErrorTypePermission:       {"Permission denied", "access is denied"},
		},
	}
}

// GetGemConfig returns the configuration for Gem manager
func GetGemConfig() ManagerConfig {
	return ManagerConfig{
		Name:       "gem",
		BinaryName: "gem",
		ErrorPatterns: map[ErrorType][]string{
			ErrorTypeNotFound:         {"Could not find a valid gem", "ERROR:  Could not find"},
			ErrorTypeAlreadyInstalled: {"already installed"},
			ErrorTypeNotInstalled:     {"is not installed"},
			ErrorTypePermission:       {"Errno::EACCES", "Gem::FilePermissionError"},
			ErrorTypeDependency:       {"requires Ruby version", "ruby version is"},
		},
	}
}

// GetCargoConfig returns the configuration for Cargo manager
func GetCargoConfig() ManagerConfig {
	return ManagerConfig{
		Name:       "cargo",
		BinaryName: "cargo",
		ErrorPatterns: map[ErrorType][]string{
			ErrorTypeNotFound:         {"no crates found", "could not find"},
			ErrorTypeAlreadyInstalled: {"binary `", "` already exists"},
			ErrorTypeNotInstalled:     {"not installed"},
		},
	}
}

// GetHomebrewConfig returns the configuration for Homebrew manager
func GetHomebrewConfig() ManagerConfig {
	return ManagerConfig{
		Name:       "homebrew",
		BinaryName: "brew",
		ErrorPatterns: map[ErrorType][]string{
			ErrorTypeNotFound:         {"No available formula", "No formulae found"},
			ErrorTypeAlreadyInstalled: {"already installed"},
			ErrorTypeNotInstalled:     {"No such keg", "not installed"},
			ErrorTypeDependency:       {"because it is required by", "still has dependents"},
		},
	}
}

// GetGoInstallConfig returns the configuration for Go Install manager
func GetGoInstallConfig() ManagerConfig {
	return ManagerConfig{
		Name:       "go",
		BinaryName: "go",
		ErrorPatterns: map[ErrorType][]string{
			ErrorTypeNotFound: {"cannot find module", "no matching versions", "malformed module path"},
			ErrorTypeNetwork:  {"connection", "timeout"},
			ErrorTypeBuild:    {"build failed", "compilation"},
		},
	}
}
