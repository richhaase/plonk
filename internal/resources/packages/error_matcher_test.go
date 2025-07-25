// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"
)

func TestErrorMatcher_MatchError(t *testing.T) {
	matcher := NewCommonErrorMatcher()

	tests := []struct {
		name     string
		output   string
		expected ErrorType
	}{
		// Not Found errors
		{
			name:     "package not found - pip",
			output:   "ERROR: Could not find a version that satisfies the requirement foobar",
			expected: ErrorTypeNotFound,
		},
		{
			name:     "package not found - apt",
			output:   "E: Unable to locate package foobar",
			expected: ErrorTypeNotFound,
		},
		{
			name:     "package not found - homebrew",
			output:   "Error: No available formula with the name \"foobar\"",
			expected: ErrorTypeNotFound,
		},

		// Permission errors
		{
			name:     "permission denied - generic",
			output:   "Permission denied: /usr/local/lib",
			expected: ErrorTypePermission,
		},
		{
			name:     "permission denied - apt",
			output:   "E: Could not open lock file /var/lib/dpkg/lock-frontend - open (13: Permission denied)",
			expected: ErrorTypePermission,
		},
		{
			name:     "permission denied - are you root",
			output:   "error: are you root?",
			expected: ErrorTypePermission,
		},

		// Already installed
		{
			name:     "already installed - pip",
			output:   "Requirement already satisfied: requests in /usr/local/lib/python3.9/site-packages",
			expected: ErrorTypeAlreadyInstalled,
		},
		{
			name:     "already installed - apt",
			output:   "curl is already the newest version (7.68.0-1ubuntu2.7).",
			expected: ErrorTypeAlreadyInstalled,
		},
		{
			name:     "already installed - homebrew",
			output:   "Warning: wget 1.21.2 is already installed and up-to-date.",
			expected: ErrorTypeAlreadyInstalled,
		},

		// Not installed
		{
			name:     "not installed - pip",
			output:   "WARNING: Skipping foobar as it is not installed.",
			expected: ErrorTypeNotInstalled,
		},
		{
			name:     "not installed - apt",
			output:   "Package 'foobar' is not installed, so not removed",
			expected: ErrorTypeNotInstalled,
		},

		// Locked
		{
			name:     "locked - apt",
			output:   "E: Could not get lock /var/lib/dpkg/lock-frontend",
			expected: ErrorTypeLocked,
		},
		{
			name:     "locked - dpkg interrupted",
			output:   "E: dpkg was interrupted, you must manually run 'sudo dpkg --configure -a'",
			expected: ErrorTypeLocked,
		},

		// Unknown
		{
			name:     "unknown error",
			output:   "Some random error message that doesn't match any pattern",
			expected: ErrorTypeUnknown,
		},

		// Case insensitive
		{
			name:     "case insensitive matching",
			output:   "ERROR: PACKAGE NOT FOUND IN REPOSITORY",
			expected: ErrorTypeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.MatchError(tt.output)
			if result != tt.expected {
				t.Errorf("MatchError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorMatcher_AddPattern(t *testing.T) {
	matcher := NewCommonErrorMatcher()

	// Add a new pattern to an existing error type
	matcher.AddPattern(ErrorTypeNotFound, "404 not found", "no such file")

	// Test the new patterns work
	tests := []struct {
		output   string
		expected ErrorType
	}{
		{
			output:   "Error: 404 not found",
			expected: ErrorTypeNotFound,
		},
		{
			output:   "no such file or directory",
			expected: ErrorTypeNotFound,
		},
	}

	for _, tt := range tests {
		result := matcher.MatchError(tt.output)
		if result != tt.expected {
			t.Errorf("MatchError() after AddPattern = %v, want %v", result, tt.expected)
		}
	}

	// Add a completely new error type
	const ErrorTypeNetwork ErrorType = "network"
	matcher.AddPattern(ErrorTypeNetwork, "connection refused", "timeout", "network unreachable")

	networkTests := []struct {
		output   string
		expected ErrorType
	}{
		{
			output:   "Error: connection refused",
			expected: ErrorTypeNetwork,
		},
		{
			output:   "request timeout",
			expected: ErrorTypeNetwork,
		},
		{
			output:   "network unreachable error",
			expected: ErrorTypeNetwork,
		},
	}

	for _, tt := range networkTests {
		result := matcher.MatchError(tt.output)
		if result != tt.expected {
			t.Errorf("MatchError(%q) for network errors = %v, want %v", tt.output, result, tt.expected)
		}
	}
}

func TestErrorMatcher_IsSuccess(t *testing.T) {
	matcher := NewCommonErrorMatcher()

	tests := []struct {
		name      string
		output    string
		errorType ErrorType
		expected  bool
	}{
		{
			name:      "already installed is success",
			output:    "Package already installed",
			errorType: ErrorTypeAlreadyInstalled,
			expected:  true,
		},
		{
			name:      "not installed is success for uninstall",
			output:    "Package not installed",
			errorType: ErrorTypeNotInstalled,
			expected:  true,
		},
		{
			name:      "not found is not success",
			output:    "Package not found",
			errorType: ErrorTypeNotFound,
			expected:  false,
		},
		{
			name:      "permission error is not success",
			output:    "Permission denied",
			errorType: ErrorTypePermission,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.IsSuccess(tt.output, tt.errorType)
			if result != tt.expected {
				t.Errorf("IsSuccess() = %v, want %v", result, tt.expected)
			}
		})
	}
}
