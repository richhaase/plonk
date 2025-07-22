// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"strings"
)

// ErrorType represents the type of error detected from command output
type ErrorType string

const (
	// ErrorTypeNotFound indicates package was not found
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypePermission indicates permission denied
	ErrorTypePermission ErrorType = "permission"
	// ErrorTypeLocked indicates resource is locked (e.g., apt lock)
	ErrorTypeLocked ErrorType = "locked"
	// ErrorTypeAlreadyInstalled indicates package is already installed
	ErrorTypeAlreadyInstalled ErrorType = "already_installed"
	// ErrorTypeNotInstalled indicates package is not installed (for uninstall)
	ErrorTypeNotInstalled ErrorType = "not_installed"
	// ErrorTypeNetwork indicates network connectivity issues
	ErrorTypeNetwork ErrorType = "network"
	// ErrorTypeBuild indicates build/compilation failures
	ErrorTypeBuild ErrorType = "build"
	// ErrorTypeDependency indicates dependency conflicts or issues
	ErrorTypeDependency ErrorType = "dependency"
	// ErrorTypeUnknown indicates an unrecognized error
	ErrorTypeUnknown ErrorType = "unknown"
)

// ErrorPattern defines a pattern to match against command output
type ErrorPattern struct {
	Type     ErrorType
	Patterns []string
}

// ErrorMatcher matches command output against known error patterns
type ErrorMatcher struct {
	patterns []ErrorPattern
}

// NewCommonErrorMatcher creates an error matcher with common patterns across package managers
func NewCommonErrorMatcher() *ErrorMatcher {
	return &ErrorMatcher{
		patterns: []ErrorPattern{
			{
				Type: ErrorTypeNotFound,
				Patterns: []string{
					"not found",
					"unable to locate",
					"no such package",
					"could not find",
					"has no installation candidate",
					"no matching distribution",
					"unable to find",
					"package not found",
					"no packages found",
					"no available formula",
				},
			},
			{
				Type: ErrorTypePermission,
				Patterns: []string{
					"permission denied",
					"are you root",
					"could not open lock file",
					"access is denied",
					"access denied",
					"requires sudo",
					"operation not permitted",
				},
			},
			{
				Type: ErrorTypeAlreadyInstalled,
				Patterns: []string{
					"already installed",
					"is already the newest version",
					"already satisfied",
					"up to date",
					"already up-to-date",
					"nothing to install",
				},
			},
			{
				Type: ErrorTypeNotInstalled,
				Patterns: []string{
					"not installed",
					"cannot uninstall",
					"is not installed",
					"no such package installed",
					"package not installed",
				},
			},
			{
				Type: ErrorTypeLocked,
				Patterns: []string{
					"could not get lock",
					"unable to lock",
					"database is locked",
					"lock file exists",
					"waiting for cache lock",
					"dpkg was interrupted",
				},
			},
		},
	}
}

// MatchError analyzes output and returns the detected error type
func (m *ErrorMatcher) MatchError(output string) ErrorType {
	lowerOutput := strings.ToLower(output)

	for _, errorPattern := range m.patterns {
		for _, pattern := range errorPattern.Patterns {
			if strings.Contains(lowerOutput, strings.ToLower(pattern)) {
				return errorPattern.Type
			}
		}
	}

	return ErrorTypeUnknown
}

// AddPattern adds a custom error pattern to the matcher
func (m *ErrorMatcher) AddPattern(errorType ErrorType, patterns ...string) {
	// Check if we already have this error type
	for i := range m.patterns {
		if m.patterns[i].Type == errorType {
			m.patterns[i].Patterns = append(m.patterns[i].Patterns, patterns...)
			return
		}
	}

	// Add new error type
	m.patterns = append(m.patterns, ErrorPattern{
		Type:     errorType,
		Patterns: patterns,
	})
}

// IsSuccess checks if the output indicates a successful operation
func (m *ErrorMatcher) IsSuccess(output string, errorType ErrorType) bool {
	// For some error types, certain outputs indicate success
	switch errorType {
	case ErrorTypeAlreadyInstalled:
		// Already installed is success for install operations
		return true
	case ErrorTypeNotInstalled:
		// Not installed is success for uninstall operations
		return true
	default:
		return false
	}
}
