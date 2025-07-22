// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package errors provides structured error types for the plonk tool.
// It defines error codes, domains, and utilities for consistent error handling
// across all plonk operations.
package errors

import (
	"fmt"
	"strings"
)

// ErrorCode represents the type of error that occurred
type ErrorCode string

const (
	// Configuration errors
	ErrConfigNotFound     ErrorCode = "CONFIG_NOT_FOUND"
	ErrConfigParseFailure ErrorCode = "CONFIG_PARSE_FAILURE"
	ErrConfigValidation   ErrorCode = "CONFIG_VALIDATION"

	// File system errors
	ErrFileNotFound    ErrorCode = "FILE_NOT_FOUND"
	ErrFileExists      ErrorCode = "FILE_EXISTS"
	ErrFilePermission  ErrorCode = "FILE_PERMISSION"
	ErrFileIO          ErrorCode = "FILE_IO"
	ErrDirectoryCreate ErrorCode = "DIRECTORY_CREATE"
	ErrPathValidation  ErrorCode = "PATH_VALIDATION"

	// Package manager errors
	ErrPackageNotFound    ErrorCode = "PACKAGE_NOT_FOUND"
	ErrPackageInstall     ErrorCode = "PACKAGE_INSTALL"
	ErrPackageUninstall   ErrorCode = "PACKAGE_UNINSTALL"
	ErrManagerUnavailable ErrorCode = "MANAGER_UNAVAILABLE"
	ErrCommandExecution   ErrorCode = "COMMAND_EXECUTION"

	// State management errors
	ErrProviderNotFound ErrorCode = "PROVIDER_NOT_FOUND"
	ErrReconciliation   ErrorCode = "RECONCILIATION"
	ErrItemRetrieval    ErrorCode = "ITEM_RETRIEVAL"

	// General errors
	ErrInvalidInput          ErrorCode = "INVALID_INPUT"
	ErrInternal              ErrorCode = "INTERNAL"
	ErrUnsupported           ErrorCode = "UNSUPPORTED"
	ErrOperationNotSupported ErrorCode = "OPERATION_NOT_SUPPORTED"
)

// Domain represents the subsystem where the error occurred
type Domain string

const (
	DomainConfig   Domain = "config"
	DomainDotfiles Domain = "dotfiles"
	DomainPackages Domain = "packages"
	DomainState    Domain = "state"
	DomainCommands Domain = "commands"
)

// Severity represents the severity level of the error
type Severity string

const (
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// ErrorSuggestion represents a helpful suggestion for resolving an error
type ErrorSuggestion struct {
	Message string `json:"message"`
	Command string `json:"command,omitempty"`
}

// PlonkError represents a structured error with context and metadata
type PlonkError struct {
	Code       ErrorCode              `json:"code"`
	Domain     Domain                 `json:"domain"`
	Operation  string                 `json:"operation"`
	Item       string                 `json:"item,omitempty"`
	Message    string                 `json:"message"`
	Severity   Severity               `json:"severity"`
	Suggestion *ErrorSuggestion       `json:"suggestion,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Cause      error                  `json:"-"` // Original error, not serialized
}

// Error implements the error interface
func (e *PlonkError) Error() string {
	var parts []string

	if e.Item != "" {
		parts = append(parts, fmt.Sprintf("plonk %s %s [%s]", e.Operation, e.Domain, e.Item))
	} else {
		parts = append(parts, fmt.Sprintf("plonk %s %s", e.Operation, e.Domain))
	}

	parts = append(parts, e.Message)

	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("caused by: %v", e.Cause))
	}

	return strings.Join(parts, ": ")
}

// Unwrap returns the underlying error for error wrapping
func (e *PlonkError) Unwrap() error {
	return e.Cause
}

// Is supports error comparison for errors.Is
func (e *PlonkError) Is(target error) bool {
	if pe, ok := target.(*PlonkError); ok {
		return e.Code == pe.Code && e.Domain == pe.Domain
	}
	return false
}

// UserMessage returns a user-friendly error message with suggestions
func (e *PlonkError) UserMessage() string {
	var message string

	switch e.Code {
	case ErrConfigNotFound:
		message = "Configuration file not found. Please run 'plonk config init' to create one."
	case ErrConfigParseFailure:
		message = "Configuration file has invalid syntax. Please check the YAML format."
	case ErrConfigValidation:
		message = fmt.Sprintf("Configuration is invalid: %s", e.Message)
	case ErrFileNotFound:
		if e.Item != "" {
			message = fmt.Sprintf("File not found: %s", e.Item)
		} else {
			message = "Required file not found"
		}
	case ErrFileExists:
		if e.Item != "" {
			message = fmt.Sprintf("File already exists: %s", e.Item)
		} else {
			message = "File already exists"
		}
	case ErrFilePermission:
		message = fmt.Sprintf("Permission denied accessing file: %s", e.Item)
	case ErrPackageNotFound:
		message = fmt.Sprintf("Package not found: %s", e.Item)
	case ErrPackageInstall:
		if e.Message != "" {
			message = e.Message
		} else {
			message = fmt.Sprintf("Failed to install package: %s", e.Item)
		}
	case ErrManagerUnavailable:
		message = fmt.Sprintf("Package manager '%s' is not available", e.Item)
	case ErrProviderNotFound:
		message = fmt.Sprintf("No provider found for domain: %s", e.Item)
	case ErrOperationNotSupported:
		if e.Item != "" {
			message = fmt.Sprintf("Operation '%s' is not supported by %s", e.Operation, e.Item)
		} else {
			message = fmt.Sprintf("Operation '%s' is not supported", e.Operation)
		}
	default:
		message = e.Message
	}

	// Add suggestion if present
	if e.Suggestion != nil {
		if e.Suggestion.Command != "" {
			message += fmt.Sprintf("\n     Try: %s", e.Suggestion.Command)
		} else {
			message += fmt.Sprintf("\n     %s", e.Suggestion.Message)
		}
	}

	return message
}

// WithMetadata adds metadata to the error
func (e *PlonkError) WithMetadata(key string, value interface{}) *PlonkError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithCause sets the underlying cause
func (e *PlonkError) WithCause(cause error) *PlonkError {
	e.Cause = cause
	return e
}

// WithItem sets the item being operated on
func (e *PlonkError) WithItem(item string) *PlonkError {
	e.Item = item
	return e
}

// WithSuggestion adds a helpful suggestion to the error
func (e *PlonkError) WithSuggestion(suggestion ErrorSuggestion) *PlonkError {
	e.Suggestion = &suggestion
	return e
}

// WithSuggestionCommand adds a command suggestion to the error
func (e *PlonkError) WithSuggestionCommand(command string) *PlonkError {
	e.Suggestion = &ErrorSuggestion{Command: command}
	return e
}

// WithSuggestionMessage adds a message suggestion to the error
func (e *PlonkError) WithSuggestionMessage(message string) *PlonkError {
	e.Suggestion = &ErrorSuggestion{Message: message}
	return e
}

// NewError creates a new PlonkError with the specified parameters
func NewError(code ErrorCode, domain Domain, operation string, message string) *PlonkError {
	severity := SeverityError
	if code == ErrConfigValidation {
		severity = SeverityWarning
	}
	if code == ErrInternal {
		severity = SeverityCritical
	}

	return &PlonkError{
		Code:       code,
		Domain:     domain,
		Operation:  operation,
		Message:    message,
		Severity:   severity,
		Suggestion: nil,
		Metadata:   make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with plonk context
func Wrap(err error, code ErrorCode, domain Domain, operation string, message string) *PlonkError {
	return NewError(code, domain, operation, message).WithCause(err)
}

// WrapWithItem wraps an error with item context
func WrapWithItem(err error, code ErrorCode, domain Domain, operation string, item string, message string) *PlonkError {
	return NewError(code, domain, operation, message).WithItem(item).WithCause(err)
}

// ErrorCollection holds multiple errors
type ErrorCollection struct {
	Errors []*PlonkError `json:"errors"`
}

// Error implements the error interface for ErrorCollection
func (ec *ErrorCollection) Error() string {
	if len(ec.Errors) == 0 {
		return "no errors"
	}
	if len(ec.Errors) == 1 {
		return ec.Errors[0].Error()
	}

	var messages []string
	for _, err := range ec.Errors {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple errors: %s", strings.Join(messages, "; "))
}

// Add adds an error to the collection
func (ec *ErrorCollection) Add(err *PlonkError) {
	ec.Errors = append(ec.Errors, err)
}

// HasErrors returns true if there are any errors
func (ec *ErrorCollection) HasErrors() bool {
	return len(ec.Errors) > 0
}

// ErrorOrNil returns the collection as an error if it has errors, nil otherwise
func (ec *ErrorCollection) ErrorOrNil() error {
	if ec.HasErrors() {
		return ec
	}
	return nil
}

// NewErrorCollection creates a new error collection
func NewErrorCollection() *ErrorCollection {
	return &ErrorCollection{
		Errors: make([]*PlonkError, 0),
	}
}

// Helper functions for common error patterns

// ConfigError creates a configuration-related error
func ConfigError(code ErrorCode, operation string, message string) *PlonkError {
	return NewError(code, DomainConfig, operation, message)
}

// DotfileError creates a dotfile-related error
func DotfileError(code ErrorCode, operation string, message string) *PlonkError {
	return NewError(code, DomainDotfiles, operation, message)
}

// PackageError creates a package-related error
func PackageError(code ErrorCode, operation string, message string) *PlonkError {
	return NewError(code, DomainPackages, operation, message)
}

// StateError creates a state-related error
func StateError(code ErrorCode, operation string, message string) *PlonkError {
	return NewError(code, DomainState, operation, message)
}

// CommandError creates a command-related error
func CommandError(code ErrorCode, operation string, message string) *PlonkError {
	return NewError(code, DomainCommands, operation, message)
}
