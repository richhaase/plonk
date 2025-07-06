// Package managers provides unified interfaces for managing different package managers
// including Homebrew, ASDF, NPM, and ZSH plugins. Each manager implements the
// PackageManager interface to provide consistent operations for checking availability,
// listing installed packages, and installing new packages.
//
// The package uses dependency injection through the CommandExecutor interface
// to allow for testing with mocked command execution.
package managers

import (
	"bytes"
	"os/exec"
)

// CommandExecutor defines an interface for executing system commands with dependency injection.
type CommandExecutor interface {
	Execute(name string, args ...string) *exec.Cmd
}

// PackageManager defines the common operations for all package managers.
type PackageManager interface {
	IsAvailable() bool
	ListInstalled() ([]string, error)
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
