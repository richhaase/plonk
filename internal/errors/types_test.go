// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestPlonkError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *PlonkError
		expected string
	}{
		{
			name: "error with item",
			err: &PlonkError{
				Code:      ErrFileNotFound,
				Domain:    DomainDotfiles,
				Operation: "copy",
				Item:      "zshrc",
				Message:   "source file missing",
			},
			expected: "plonk copy dotfiles [zshrc]: source file missing",
		},
		{
			name: "error without item",
			err: &PlonkError{
				Code:      ErrConfigValidation,
				Domain:    DomainConfig,
				Operation: "load",
				Message:   "invalid YAML syntax",
			},
			expected: "plonk load config: invalid YAML syntax",
		},
		{
			name: "error with cause",
			err: &PlonkError{
				Code:      ErrFileIO,
				Domain:    DomainDotfiles,
				Operation: "copy",
				Item:      "gitconfig",
				Message:   "failed to write file",
				Cause:     fmt.Errorf("permission denied"),
			},
			expected: "plonk copy dotfiles [gitconfig]: failed to write file: caused by: permission denied",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.err.Error()
			if result != test.expected {
				t.Errorf("Error() = %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestPlonkError_UserMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *PlonkError
		expected string
	}{
		{
			name: "config not found",
			err: &PlonkError{
				Code:   ErrConfigNotFound,
				Domain: DomainConfig,
			},
			expected: "Configuration file not found. Please run 'plonk config init' to create one.",
		},
		{
			name: "file not found with item",
			err: &PlonkError{
				Code:   ErrFileNotFound,
				Domain: DomainDotfiles,
				Item:   "/home/user/.zshrc",
			},
			expected: "File not found: /home/user/.zshrc",
		},
		{
			name: "package install failure",
			err: &PlonkError{
				Code:   ErrPackageInstall,
				Domain: DomainPackages,
				Item:   "neovim",
			},
			expected: "Failed to install package: neovim",
		},
		{
			name: "manager unavailable",
			err: &PlonkError{
				Code:   ErrManagerUnavailable,
				Domain: DomainPackages,
				Item:   "homebrew",
			},
			expected: "Package manager 'homebrew' is not available",
		},
		{
			name: "default message",
			err: &PlonkError{
				Code:    ErrInternal,
				Domain:  DomainCommands,
				Message: "unexpected internal error",
			},
			expected: "unexpected internal error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.err.UserMessage()
			if result != test.expected {
				t.Errorf("UserMessage() = %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestPlonkError_WithMetadata(t *testing.T) {
	err := &PlonkError{
		Code:     ErrFileNotFound,
		Domain:   DomainDotfiles,
		Metadata: make(map[string]interface{}),
	}

	result := err.WithMetadata("path", "/home/user/.zshrc")

	if result != err {
		t.Error("WithMetadata should return the same error instance")
	}

	if err.Metadata["path"] != "/home/user/.zshrc" {
		t.Errorf("Metadata[\"path\"] = %v, expected /home/user/.zshrc", err.Metadata["path"])
	}
}

func TestPlonkError_WithCause(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := &PlonkError{
		Code:   ErrFileIO,
		Domain: DomainDotfiles,
	}

	result := err.WithCause(originalErr)

	if result != err {
		t.Error("WithCause should return the same error instance")
	}

	if err.Cause != originalErr {
		t.Errorf("Cause = %v, expected %v", err.Cause, originalErr)
	}
}

func TestPlonkError_WithItem(t *testing.T) {
	err := &PlonkError{
		Code:   ErrPackageInstall,
		Domain: DomainPackages,
	}

	result := err.WithItem("neovim")

	if result != err {
		t.Error("WithItem should return the same error instance")
	}

	if err.Item != "neovim" {
		t.Errorf("Item = %v, expected neovim", err.Item)
	}
}

func TestPlonkError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := &PlonkError{
		Code:   ErrFileIO,
		Domain: DomainDotfiles,
		Cause:  originalErr,
	}

	result := err.Unwrap()
	if result != originalErr {
		t.Errorf("Unwrap() = %v, expected %v", result, originalErr)
	}
}

func TestPlonkError_Is(t *testing.T) {
	err1 := &PlonkError{
		Code:   ErrFileNotFound,
		Domain: DomainDotfiles,
	}

	err2 := &PlonkError{
		Code:   ErrFileNotFound,
		Domain: DomainDotfiles,
	}

	err3 := &PlonkError{
		Code:   ErrFileNotFound,
		Domain: DomainConfig,
	}

	regularErr := fmt.Errorf("regular error")

	if !err1.Is(err2) {
		t.Error("err1.Is(err2) should be true for same code and domain")
	}

	if err1.Is(err3) {
		t.Error("err1.Is(err3) should be false for different domain")
	}

	if err1.Is(regularErr) {
		t.Error("err1.Is(regularErr) should be false for non-PlonkError")
	}
}

func TestNewError(t *testing.T) {
	err := NewError(ErrConfigValidation, DomainConfig, "load", "invalid syntax")

	if err.Code != ErrConfigValidation {
		t.Errorf("Code = %v, expected %v", err.Code, ErrConfigValidation)
	}

	if err.Domain != DomainConfig {
		t.Errorf("Domain = %v, expected %v", err.Domain, DomainConfig)
	}

	if err.Operation != "load" {
		t.Errorf("Operation = %v, expected load", err.Operation)
	}

	if err.Message != "invalid syntax" {
		t.Errorf("Message = %v, expected invalid syntax", err.Message)
	}

	if err.Severity != SeverityWarning {
		t.Errorf("Severity = %v, expected %v", err.Severity, SeverityWarning)
	}

	if err.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestNewError_SeverityMapping(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected Severity
	}{
		{ErrConfigValidation, SeverityWarning},
		{ErrInternal, SeverityCritical},
		{ErrFileNotFound, SeverityError},
	}

	for _, test := range tests {
		err := NewError(test.code, DomainConfig, "test", "test message")
		if err.Severity != test.expected {
			t.Errorf("NewError(%v) severity = %v, expected %v", test.code, err.Severity, test.expected)
		}
	}
}

func TestWrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := Wrap(originalErr, ErrFileIO, DomainDotfiles, "copy", "failed to copy file")

	if err.Code != ErrFileIO {
		t.Errorf("Code = %v, expected %v", err.Code, ErrFileIO)
	}

	if err.Cause != originalErr {
		t.Errorf("Cause = %v, expected %v", err.Cause, originalErr)
	}

	if err.Message != "failed to copy file" {
		t.Errorf("Message = %v, expected failed to copy file", err.Message)
	}
}

func TestWrapWithItem(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := WrapWithItem(originalErr, ErrPackageInstall, DomainPackages, "install", "neovim", "installation failed")

	if err.Code != ErrPackageInstall {
		t.Errorf("Code = %v, expected %v", err.Code, ErrPackageInstall)
	}

	if err.Item != "neovim" {
		t.Errorf("Item = %v, expected neovim", err.Item)
	}

	if err.Cause != originalErr {
		t.Errorf("Cause = %v, expected %v", err.Cause, originalErr)
	}
}

func TestErrorCollection(t *testing.T) {
	collection := NewErrorCollection()

	if collection.HasErrors() {
		t.Error("New collection should not have errors")
	}

	if collection.ErrorOrNil() != nil {
		t.Error("Empty collection should return nil")
	}

	err1 := NewError(ErrFileNotFound, DomainDotfiles, "copy", "file1 not found")
	err2 := NewError(ErrFileNotFound, DomainDotfiles, "copy", "file2 not found")

	collection.Add(err1)
	collection.Add(err2)

	if !collection.HasErrors() {
		t.Error("Collection should have errors after adding")
	}

	if collection.ErrorOrNil() == nil {
		t.Error("Collection with errors should return error")
	}

	if len(collection.Errors) != 2 {
		t.Errorf("Collection should have 2 errors, got %d", len(collection.Errors))
	}
}

func TestErrorCollection_Error(t *testing.T) {
	collection := NewErrorCollection()

	// Empty collection
	if collection.Error() != "no errors" {
		t.Errorf("Empty collection error = %q, expected 'no errors'", collection.Error())
	}

	// Single error
	err1 := NewError(ErrFileNotFound, DomainDotfiles, "copy", "file not found")
	collection.Add(err1)

	if collection.Error() != err1.Error() {
		t.Errorf("Single error collection should return the error's message")
	}

	// Multiple errors
	err2 := NewError(ErrConfigValidation, DomainConfig, "load", "invalid config")
	collection.Add(err2)

	result := collection.Error()
	if !strings.HasPrefix(result, "multiple errors:") {
		t.Errorf("Multiple errors should start with 'multiple errors:', got %q", result)
	}

	if !strings.Contains(result, err1.Error()) {
		t.Errorf("Multiple errors should contain first error message")
	}

	if !strings.Contains(result, err2.Error()) {
		t.Errorf("Multiple errors should contain second error message")
	}
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func() *PlonkError
		expected Domain
	}{
		{
			name:     "ConfigError",
			function: func() *PlonkError { return ConfigError(ErrConfigNotFound, "load", "not found") },
			expected: DomainConfig,
		},
		{
			name:     "DotfileError",
			function: func() *PlonkError { return DotfileError(ErrFileNotFound, "copy", "not found") },
			expected: DomainDotfiles,
		},
		{
			name:     "PackageError",
			function: func() *PlonkError { return PackageError(ErrPackageInstall, "install", "failed") },
			expected: DomainPackages,
		},
		{
			name:     "StateError",
			function: func() *PlonkError { return StateError(ErrReconciliation, "reconcile", "failed") },
			expected: DomainState,
		},
		{
			name:     "CommandError",
			function: func() *PlonkError { return CommandError(ErrInvalidInput, "execute", "invalid") },
			expected: DomainCommands,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.function()
			if err.Domain != test.expected {
				t.Errorf("%s created error with domain %v, expected %v", test.name, err.Domain, test.expected)
			}
		})
	}
}

func TestErrorsIsIntegration(t *testing.T) {
	// Test that PlonkError works with errors.Is
	originalErr := fmt.Errorf("original error")
	plonkErr := Wrap(originalErr, ErrFileIO, DomainDotfiles, "copy", "failed to copy")

	// Should find the PlonkError itself
	if !errors.Is(plonkErr, plonkErr) {
		t.Error("errors.Is should find PlonkError itself")
	}

	// Should find the original error through unwrapping
	if !errors.Is(plonkErr, originalErr) {
		t.Error("errors.Is should find original error through unwrapping")
	}

	// Should find errors with same code and domain
	sameErr := &PlonkError{Code: ErrFileIO, Domain: DomainDotfiles}
	if !errors.Is(plonkErr, sameErr) {
		t.Error("errors.Is should find errors with same code and domain")
	}
}

func TestErrorMetadataPreservation(t *testing.T) {
	err := NewError(ErrFileNotFound, DomainDotfiles, "copy", "file not found")
	err.WithMetadata("path", "/home/user/.zshrc")
	err.WithMetadata("operation", "copy")

	if err.Metadata["path"] != "/home/user/.zshrc" {
		t.Error("Metadata should preserve path information")
	}

	if err.Metadata["operation"] != "copy" {
		t.Error("Metadata should preserve operation information")
	}
}
