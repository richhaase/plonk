// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorHandling_StandardErrorMessages(t *testing.T) {
	tests := []struct {
		name           string
		errorFunc      func() error
		expectedPrefix string
		expectedFormat string
	}{
		{
			name: "config load error",
			errorFunc: func() error {
				return WrapConfigError(errors.New("file not found"))
			},
			expectedPrefix: "failed to load configuration",
			expectedFormat: "failed to load configuration: %s",
		},
		{
			name: "package manager unavailable error",
			errorFunc: func() error {
				return WrapPackageManagerError("homebrew", errors.New("command not found"))
			},
			expectedPrefix: "package manager 'homebrew' is not available",
			expectedFormat: "package manager 'homebrew' is not available: %s",
		},
		{
			name: "installation error",
			errorFunc: func() error {
				return WrapInstallError("git", errors.New("network error"))
			},
			expectedPrefix: "failed to install package 'git'",
			expectedFormat: "failed to install package 'git': %s",
		},
		{
			name: "file operation error",
			errorFunc: func() error {
				return WrapFileError("write", "/path/to/file", errors.New("permission denied"))
			},
			expectedPrefix: "failed to write file '/path/to/file'",
			expectedFormat: "failed to write file '/path/to/file': %s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			errMsg := err.Error()
			if !strings.HasPrefix(errMsg, tt.expectedPrefix) {
				t.Errorf("Error message should start with %q, got %q", tt.expectedPrefix, errMsg)
			}

			// Verify that the error is properly wrapped (contains original error)
			if !strings.Contains(errMsg, ":") {
				t.Error("Error should be wrapped with original error")
			}
		})
	}
}

func TestErrorHandling_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		validator   func([]string) error
		expectedErr string
	}{
		{
			name: "no args command with args",
			args: []string{"unexpected"},
			validator: func(args []string) error {
				return ValidateNoArgs("test", args)
			},
			expectedErr: "command 'test' takes no arguments",
		},
		{
			name: "exact args command with wrong count",
			args: []string{},
			validator: func(args []string) error {
				return ValidateExactArgs("test", 1, args)
			},
			expectedErr: "command 'test' requires exactly 1 argument",
		},
		{
			name: "exact args command with multiple wrong count",
			args: []string{"one", "two", "three"},
			validator: func(args []string) error {
				return ValidateExactArgs("test", 2, args)
			},
			expectedErr: "command 'test' requires exactly 2 arguments",
		},
		{
			name: "max args command with too many",
			args: []string{"one", "two", "three"},
			validator: func(args []string) error {
				return ValidateMaxArgs("test", 2, args)
			},
			expectedErr: "command 'test' accepts at most 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator(tt.args)
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}

			if err.Error() != tt.expectedErr {
				t.Errorf("Expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestErrorHandling_ArgumentValidationSuccess(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		validator func([]string) error
	}{
		{
			name: "no args command with no args",
			args: []string{},
			validator: func(args []string) error {
				return ValidateNoArgs("test", args)
			},
		},
		{
			name: "exact args command with correct count",
			args: []string{"arg1"},
			validator: func(args []string) error {
				return ValidateExactArgs("test", 1, args)
			},
		},
		{
			name: "max args command within limit",
			args: []string{"arg1"},
			validator: func(args []string) error {
				return ValidateMaxArgs("test", 2, args)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator(tt.args)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestErrorHandling_ErrorWrappingConsistency(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name        string
		wrapFunc    func(error) error
		expectsWrap bool
	}{
		{
			name: "wrap config error",
			wrapFunc: func(err error) error {
				return WrapConfigError(err)
			},
			expectsWrap: true,
		},
		{
			name: "wrap install error",
			wrapFunc: func(err error) error {
				return WrapInstallError("package", err)
			},
			expectsWrap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrappedErr := tt.wrapFunc(originalErr)

			if tt.expectsWrap {
				// Verify the error is properly wrapped
				if !errors.Is(wrappedErr, originalErr) {
					t.Error("Error should be properly wrapped with errors.Is compatibility")
				}

				// Verify error chain can be unwrapped
				if errors.Unwrap(wrappedErr) == nil {
					t.Error("Error should be unwrappable")
				}
			}
		})
	}
}
