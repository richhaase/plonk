package commands

import (
	"fmt"
)

// Standard error wrapping functions for consistent error handling

// WrapConfigError wraps configuration-related errors with standard context
func WrapConfigError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to load configuration: %w", err)
}

// WrapPackageManagerError wraps package manager availability errors
func WrapPackageManagerError(managerName string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("package manager '%s' is not available: %w", managerName, err)
}

// WrapInstallError wraps package installation errors
func WrapInstallError(packageName string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to install package '%s': %w", packageName, err)
}

// WrapFileError wraps file operation errors
func WrapFileError(operation, filePath string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to %s file '%s': %w", operation, filePath, err)
}

// Standard argument validation functions

// ValidateNoArgs validates that no arguments were provided
func ValidateNoArgs(commandName string, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("command '%s' takes no arguments", commandName)
	}
	return nil
}

// ValidateExactArgs validates that exactly the expected number of arguments were provided
func ValidateExactArgs(commandName string, expected int, args []string) error {
	if len(args) != expected {
		if expected == 1 {
			return fmt.Errorf("command '%s' requires exactly %d argument", commandName, expected)
		}
		return fmt.Errorf("command '%s' requires exactly %d arguments", commandName, expected)
	}
	return nil
}

// ValidateMaxArgs validates that at most the maximum number of arguments were provided
func ValidateMaxArgs(commandName string, max int, args []string) error {
	if len(args) > max {
		return fmt.Errorf("command '%s' accepts at most %d arguments", commandName, max)
	}
	return nil
}
