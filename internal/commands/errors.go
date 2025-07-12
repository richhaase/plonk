// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"

	"github.com/richhaase/plonk/internal/errors"
)

// HandleError processes an error and returns a user-friendly exit code
// Returns 0 for success, 1 for user errors, 2 for system errors
func HandleError(err error) int {
	if err == nil {
		return 0
	}

	// Check if it's a structured PlonkError
	if plonkErr, ok := err.(*errors.PlonkError); ok {
		// Print user-friendly message
		fmt.Fprintf(os.Stderr, "Error: %s\n", plonkErr.UserMessage())

		// Print technical details if verbose or critical
		if plonkErr.Severity == errors.SeverityCritical || os.Getenv("PLONK_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "Technical details: %s\n", plonkErr.Error())
		}

		// Return appropriate exit code based on error type
		switch plonkErr.Code {
		case errors.ErrConfigNotFound, errors.ErrConfigParseFailure, errors.ErrConfigValidation:
			return 1 // User configuration error
		case errors.ErrInvalidInput:
			return 1 // User input error
		case errors.ErrFilePermission:
			return 2 // System permission error
		case errors.ErrManagerUnavailable:
			return 2 // System dependency error
		case errors.ErrInternal:
			return 2 // Internal error
		default:
			return 1 // Default to user error
		}
	}

	// Fallback for non-structured errors
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return 1
}

// WrapCommandError wraps a command error with appropriate context
func WrapCommandError(err error, command string, message string) error {
	if err == nil {
		return nil
	}

	// If it's already a PlonkError, return as-is
	if _, ok := err.(*errors.PlonkError); ok {
		return err
	}

	// Wrap with command context
	return errors.Wrap(err, errors.ErrCommandExecution, errors.DomainCommands, command, message)
}

// NewConfigError creates a configuration-related error
func NewConfigError(code errors.ErrorCode, operation string, message string) error {
	return errors.NewError(code, errors.DomainConfig, operation, message)
}

// NewFileError creates a file-related error
func NewFileError(code errors.ErrorCode, operation string, filename string, message string) error {
	return errors.NewError(code, errors.DomainDotfiles, operation, message).WithItem(filename)
}

// NewPackageError creates a package-related error
func NewPackageError(code errors.ErrorCode, operation string, packageName string, message string) error {
	return errors.NewError(code, errors.DomainPackages, operation, message).WithItem(packageName)
}
