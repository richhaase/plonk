// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommandExecutor abstracts command execution for testability
type CommandExecutor interface {
	// Execute runs a command and returns stdout
	Execute(ctx context.Context, name string, args ...string) ([]byte, error)

	// CombinedOutput runs a command and returns combined stdout/stderr
	CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error)

	// LookPath searches for an executable in PATH
	LookPath(name string) (string, error)
}

// RealCommandExecutor implements CommandExecutor using os/exec
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

func (r *RealCommandExecutor) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

func (r *RealCommandExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// Package-level default executor
var defaultExecutor CommandExecutor = &RealCommandExecutor{}

// SetDefaultExecutor allows tests to override the executor
func SetDefaultExecutor(executor CommandExecutor) {
	defaultExecutor = executor
}

// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
	// Commands records all executed commands
	Commands []ExecutedCommand

	// Responses maps command patterns to responses
	Responses map[string]CommandResponse

	// DefaultResponse is used when no pattern matches
	DefaultResponse CommandResponse
}

type ExecutedCommand struct {
	Name    string
	Args    []string
	Context context.Context
}

type CommandResponse struct {
	Output []byte
	Error  error
}

// MockExitError implements the minimal interface that package managers check for
type MockExitError struct {
	Code int
}

func (e *MockExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

func (e *MockExitError) ExitCode() int {
	return e.Code
}

func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	m.Commands = append(m.Commands, ExecutedCommand{Name: name, Args: args, Context: ctx})

	// Find matching response using simple string matching
	// This is intentionally simple - exact matches only for v1.0
	key := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	if resp, ok := m.Responses[key]; ok {
		return resp.Output, resp.Error
	}

	return m.DefaultResponse.Output, m.DefaultResponse.Error
}

func (m *MockCommandExecutor) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	// For mock, Execute and CombinedOutput behave the same
	return m.Execute(ctx, name, args...)
}

func (m *MockCommandExecutor) LookPath(name string) (string, error) {
	// Simple implementation: check if we have any responses for this command
	for key := range m.Responses {
		if strings.HasPrefix(key, name+" ") || key == name {
			return "/usr/bin/" + name, nil // Mock path
		}
	}
	return "", &exec.Error{Name: name, Err: exec.ErrNotFound}
}
