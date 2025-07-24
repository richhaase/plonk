// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import "fmt"

// ErrorHandler provides shared error handling functionality for package managers
type ErrorHandler struct {
	errorMatcher *ErrorMatcher
	managerName  string
}

// NewErrorHandler creates a new error handler with the given error matcher
func NewErrorHandler(errorMatcher *ErrorMatcher, managerName string) *ErrorHandler {
	return &ErrorHandler{
		errorMatcher: errorMatcher,
		managerName:  managerName,
	}
}

// HandleInstallError processes installation errors with common patterns
func (h *ErrorHandler) HandleInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for specific error conditions using ErrorMatcher
	if exitCode, ok := ExtractExitCode(err); ok {
		// Match error patterns
		errorType := h.errorMatcher.MatchError(outputStr)

		switch errorType {
		case ErrorTypeNotFound:
			return fmt.Errorf("package '%s' not found", packageName)

		case ErrorTypeAlreadyInstalled:
			// Package is already installed - this is typically fine
			return nil

		case ErrorTypePermission:
			return fmt.Errorf("permission denied installing %s", packageName)

		case ErrorTypeLocked:
			return fmt.Errorf("package manager database is locked")

		case ErrorTypeNetwork:
			return fmt.Errorf("network error during installation")

		case ErrorTypeBuild:
			return fmt.Errorf("failed to build package '%s'", packageName)

		case ErrorTypeDependency:
			return fmt.Errorf("dependency conflict installing package '%s'", packageName)

		default:
			// Only treat non-zero exit codes as errors
			if exitCode != 0 {
				return fmt.Errorf("package installation failed (exit code %d): %w", exitCode, err)
			}
			// Exit code 0 with no recognized error pattern - success
			return nil
		}
	}

	// Direct error return for other cases
	return err
}

// HandleUninstallError processes uninstallation errors with common patterns
func (h *ErrorHandler) HandleUninstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for specific error conditions using ErrorMatcher
	if exitCode, ok := ExtractExitCode(err); ok {
		// Match error patterns
		errorType := h.errorMatcher.MatchError(outputStr)

		switch errorType {
		case ErrorTypeNotFound, ErrorTypeNotInstalled:
			return fmt.Errorf("package '%s' is not installed", packageName)

		case ErrorTypePermission:
			return fmt.Errorf("permission denied uninstalling %s", packageName)

		case ErrorTypeLocked:
			return fmt.Errorf("package manager database is locked")

		case ErrorTypeDependency:
			return fmt.Errorf("cannot remove '%s' because other packages depend on it", packageName)

		default:
			// Only treat non-zero exit codes as errors
			if exitCode != 0 {
				return fmt.Errorf("package uninstallation failed (exit code %d): %w", exitCode, err)
			}
			// Exit code 0 with no recognized error pattern - success
			return nil
		}
	}

	// Direct error return for other cases
	return err
}

// ClassifyError determines the type of error from output
func (h *ErrorHandler) ClassifyError(output []byte) ErrorType {
	return h.errorMatcher.MatchError(string(output))
}
