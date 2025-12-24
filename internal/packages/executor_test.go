// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
)

// executeCommand is a test helper that runs a command using the default executor.
func executeCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return defaultExecutor.Execute(ctx, name, args...)
}

func TestMockExecutor(t *testing.T) {
	// Save and restore original executor
	originalExecutor := defaultExecutor
	defer func() { defaultExecutor = originalExecutor }()

	// Create mock
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"echo hello": {
				Output: []byte("mocked output"),
				Error:  nil,
			},
		},
	}

	// Set mock as default
	SetDefaultExecutor(mock)

	// Test using helper
	output, err := executeCommand(context.Background(), "echo", "hello")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output) != "mocked output" {
		t.Errorf("Expected 'mocked output' but got '%s'", string(output))
	}

	// Verify command was recorded
	if len(mock.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(mock.Commands))
	}
}
