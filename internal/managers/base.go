// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"context"
	"fmt"
	"os/exec"
)

// ManagerConfig defines the configuration for a package manager
type ManagerConfig struct {
	// BinaryName is the primary binary name (e.g., "pip", "apt", "brew")
	BinaryName string

	// FallbackBinaries are alternative binaries to try (e.g., ["pip3"] for pip)
	FallbackBinaries []string

	// VersionArgs are the arguments to verify the binary works
	VersionArgs []string

	// Command builders for common operations
	ListArgs      func() []string
	InstallArgs   func(pkg string) []string
	UninstallArgs func(pkg string) []string

	// Output format preferences
	PreferJSON bool
	JSONFlag   string // e.g., "--json" or "--format=json"
}

// BaseManager provides common functionality for all package managers
type BaseManager struct {
	Config       ManagerConfig
	ErrorMatcher *ErrorMatcher
	binaryCache  string // Cache the binary name after first check
}

// NewBaseManager creates a new base manager with the given configuration
func NewBaseManager(config ManagerConfig) *BaseManager {
	return &BaseManager{
		Config:       config,
		ErrorMatcher: NewCommonErrorMatcher(),
	}
}

// IsAvailable checks if the package manager is installed and accessible
func (b *BaseManager) IsAvailable(ctx context.Context) (bool, error) {
	// Try primary binary
	if _, err := exec.LookPath(b.Config.BinaryName); err == nil {
		if verifyErr := b.verifyBinary(ctx, b.Config.BinaryName); verifyErr == nil {
			b.binaryCache = b.Config.BinaryName
			return true, nil
		} else {
			// Check for context cancellation
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			// Also check if the error itself is a context error
			if verifyErr == context.Canceled || verifyErr == context.DeadlineExceeded {
				return false, verifyErr
			}
		}
	}

	// Try fallback binaries
	for _, fallback := range b.Config.FallbackBinaries {
		if _, err := exec.LookPath(fallback); err == nil {
			if verifyErr := b.verifyBinary(ctx, fallback); verifyErr == nil {
				b.binaryCache = fallback
				return true, nil
			} else {
				// Check for context cancellation
				if ctx.Err() != nil {
					return false, ctx.Err()
				}
				// Also check if the error itself is a context error
				if verifyErr == context.Canceled || verifyErr == context.DeadlineExceeded {
					return false, verifyErr
				}
			}
		}
	}

	// No binary found - this is not an error condition
	return false, nil
}

// GetBinary returns the cached binary name or the primary binary name
func (b *BaseManager) GetBinary() string {
	if b.binaryCache != "" {
		return b.binaryCache
	}
	return b.Config.BinaryName
}

// verifyBinary verifies that a binary is functional by running the version command
func (b *BaseManager) verifyBinary(ctx context.Context, binary string) error {
	args := b.Config.VersionArgs
	if len(args) == 0 {
		args = []string{"--version"} // Default version argument
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	_, err := cmd.Output()
	if err != nil {
		// Check for context cancellation - return directly without wrapping
		if err == context.Canceled || err == context.DeadlineExceeded {
			return err
		}
		// Return the error wrapped with context
		return fmt.Errorf("%s binary found but not functional: %w", binary, err)
	}
	return nil
}

// ExecuteList runs the list command with proper error handling
func (b *BaseManager) ExecuteList(ctx context.Context) ([]byte, error) {
	if b.Config.ListArgs == nil {
		return nil, fmt.Errorf("list command not configured for this manager")
	}

	binary := b.GetBinary()
	args := b.Config.ListArgs()

	// Add JSON flag if configured
	if b.Config.PreferJSON && b.Config.JSONFlag != "" {
		args = append(args, b.Config.JSONFlag)
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, b.wrapCommandError(err, "list", "failed to execute list command")
	}

	return output, nil
}

// ExecuteInstall runs the install command with proper error handling
func (b *BaseManager) ExecuteInstall(ctx context.Context, packageName string) error {
	if b.Config.InstallArgs == nil {
		return fmt.Errorf("install command not configured for this manager")
	}

	binary := b.GetBinary()
	args := b.Config.InstallArgs(packageName)

	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return b.handleInstallError(err, output, packageName)
	}

	return nil
}

// ExecuteUninstall runs the uninstall command with proper error handling
func (b *BaseManager) ExecuteUninstall(ctx context.Context, packageName string) error {
	if b.Config.UninstallArgs == nil {
		return fmt.Errorf("uninstall command not configured for this manager")
	}

	binary := b.GetBinary()
	args := b.Config.UninstallArgs(packageName)

	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return b.handleUninstallError(err, output, packageName)
	}

	return nil
}

// handleInstallError processes install command errors using ErrorMatcher
func (b *BaseManager) handleInstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for specific error conditions using ErrorMatcher
	if execErr, ok := err.(interface{ ExitCode() int }); ok {
		errorType := b.ErrorMatcher.MatchError(outputStr)

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
			if execErr.ExitCode() != 0 {
				return fmt.Errorf("package installation failed (exit code %d): %w", execErr.ExitCode(), err)
			}
			// Exit code 0 with no recognized error pattern - success
			return nil
		}
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return fmt.Errorf("failed to execute install command: %w", err)
}

// handleUninstallError processes uninstall command errors using ErrorMatcher
func (b *BaseManager) handleUninstallError(err error, output []byte, packageName string) error {
	outputStr := string(output)

	// Check for specific error conditions using ErrorMatcher
	if execErr, ok := err.(interface{ ExitCode() int }); ok {
		errorType := b.ErrorMatcher.MatchError(outputStr)

		switch errorType {
		case ErrorTypeNotInstalled:
			// Package is not installed - this is typically fine for uninstall
			return nil

		case ErrorTypePermission:
			return fmt.Errorf("permission denied uninstalling %s", packageName)

		case ErrorTypeLocked:
			return fmt.Errorf("package manager database is locked")

		case ErrorTypeDependency:
			return fmt.Errorf("cannot uninstall package '%s' due to dependency conflicts", packageName)

		default:
			// Only treat non-zero exit codes as errors
			if execErr.ExitCode() != 0 {
				return fmt.Errorf("package uninstallation failed (exit code %d): %w", execErr.ExitCode(), err)
			}
			// Exit code 0 with no recognized error pattern - success
			return nil
		}
	}

	// Non-exit errors (command not found, context cancellation, etc.)
	return fmt.Errorf("failed to execute uninstall command: %w", err)
}

// wrapCommandError wraps a command error with appropriate context
func (b *BaseManager) wrapCommandError(err error, operation, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// SupportsSearch returns true by default as most package managers support search.
// Managers that don't support search (like Go, Pip) should override this method.
func (b *BaseManager) SupportsSearch() bool {
	return true
}
