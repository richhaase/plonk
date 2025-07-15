// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import "fmt"

// ExitError is a test helper that simulates an exec.ExitError
type ExitError struct {
	Code   int
	Stderr []byte
}

// Error implements the error interface
func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

// ExitCode returns the exit code
func (e *ExitError) ExitCode() int {
	return e.Code
}
