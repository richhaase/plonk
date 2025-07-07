// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package managers provides unified interfaces for managing different package managers
// including Homebrew, ASDF, and NPM plugins. Each manager implements the
// PackageManager interface to provide consistent operations for checking availability,
// listing installed packages, and installing new packages.
//
// The package uses dependency injection through the CommandExecutor interface
// to allow for testing with mocked command execution.
package managers

import (
	"bytes"
	"os/exec"
	"time"
)

// CommandExecutor defines an interface for executing system commands with dependency injection.
type CommandExecutor interface {
	Execute(name string, args ...string) *exec.Cmd
}

// PackageStatus represents the state of a package
type PackageStatus int

const (
	PackageInstalled PackageStatus = iota // Package is installed
	PackageAvailable                      // Package is available but not installed
	PackageUnknown                        // Package status unknown
)

// String returns the string representation of PackageStatus
func (s PackageStatus) String() string {
	switch s {
	case PackageInstalled:
		return "installed"
	case PackageAvailable:
		return "available"
	case PackageUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// PackageInfo contains metadata about a package
type PackageInfo struct {
	Name        string        // Package name
	Version     string        // Installed version (if available)
	Status      PackageStatus // Current status
	Manager     string        // Package manager name (homebrew, asdf, npm)
	Description string        // Package description (if available)
	InstallDate time.Time     // Installation date (if available)
}

// PackageManager defines the common operations for all package managers.
type PackageManager interface {
	IsAvailable() bool
	ListInstalled() ([]string, error)
	ListInstalledPackages() ([]PackageInfo, error) // Enhanced method returning rich objects
}

// ExtendedPackageManager defines additional operations for package managers that support them.
type ExtendedPackageManager interface {
	PackageManager
	Search(query string) ([]string, error)
	Info(packageName string) (string, error)
	Update(packageName string) error
	UpdateAll() error
	IsInstalled(packageName string) bool
}

// PackageManagerInfo holds a package manager and its display name.
type PackageManagerInfo struct {
	Name    string
	Manager PackageManager
}

// CommandRunner provides common command execution functionality.
type CommandRunner struct {
	executor    CommandExecutor
	commandName string
}

// NewCommandRunner creates a new command runner for a specific command.
func NewCommandRunner(executor CommandExecutor, commandName string) *CommandRunner {
	return &CommandRunner{
		executor:    executor,
		commandName: commandName,
	}
}

// RunCommandWithOutput executes a command and returns output + error.
func (c *CommandRunner) RunCommandWithOutput(args ...string) (string, error) {
	cmd := c.executor.Execute(c.commandName, args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err
}

// RunCommand executes a command and returns success/error (ignores output).
func (c *CommandRunner) RunCommand(args ...string) error {
	_, err := c.RunCommandWithOutput(args...)
	return err
}
