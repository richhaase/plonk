// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
	"time"
)

func TestPlonkBinaryInContainer(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build Plonk binary inside container
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test plonk basic command
	t.Run("plonk basic command", func(t *testing.T) {
		output, err := runner.RunCommand(t, "./plonk")
		if err != nil {
			t.Fatalf("Failed to run plonk: %v\nOutput: %s", err, output)
		}
		
		t.Logf("Basic command output: %s", output)
		
		// Basic validation that output contains expected content
		outputStr := string(output)
		if len(outputStr) == 0 {
			t.Error("Command output is empty")
		}
		if !strings.Contains(outputStr, "plonk") {
			t.Error("Output doesn't contain 'plonk'")
		}
	})

	// Test plonk --help
	t.Run("plonk help", func(t *testing.T) {
		output, err := runner.RunCommand(t, "./plonk --help")
		if err != nil {
			t.Fatalf("Failed to run plonk --help: %v\nOutput: %s", err, output)
		}
		
		t.Logf("Help output: %s", output)
		
		// Basic validation that help output contains expected content
		outputStr := string(output)
		if len(outputStr) == 0 {
			t.Error("Help output is empty")
		}
	})

	// Test basic plonk configuration command
	t.Run("plonk config show", func(t *testing.T) {
		output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk config show")
		
		// config show may fail if no config exists, that's OK for this test
		t.Logf("Config show output: %s", output)
		t.Logf("Config show error (if any): %v", err)
		
		// Just verify the command runs and produces some output
		outputStr := string(output)
		if len(outputStr) == 0 && err == nil {
			t.Error("Config show produced no output and no error")
		}
	})
}

func TestContainerIsolation(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()

	// Test that each container starts with a fresh environment
	t.Run("fresh environment per container", func(t *testing.T) {
		// Run first container and create a test file
		output1, err := runner.RunCommand(t, "echo 'test content' > /tmp/testfile && cat /tmp/testfile")
		if err != nil {
			t.Fatalf("Failed to create test file: %v\nOutput: %s", err, output1)
		}
		
		// Run second container and check that test file doesn't exist
		output2, err := runner.RunCommand(t, "ls /tmp/testfile || echo 'file not found'")
		if err != nil {
			t.Fatalf("Failed to check test file: %v\nOutput: %s", err, output2)
		}
		
		// Check that the output indicates file not found (either explicit message or ls error)
		output2Str := string(output2)
		if !strings.Contains(output2Str, "file not found") && !strings.Contains(output2Str, "No such file or directory") {
			t.Errorf("Expected file not found indication, got: %s", output2Str)
		}
		
		t.Logf("Container 1 output: %s", output1)
		t.Logf("Container 2 output: %s", output2)
	})

	// Test that package managers are available and functional
	t.Run("package managers available", func(t *testing.T) {
		output, err := runner.RunCommand(t, "brew --version && npm --version && apt --version")
		if err != nil {
			t.Fatalf("Package managers not available: %v\nOutput: %s", err, output)
		}
		
		t.Logf("Package managers output: %s", output)
	})
}

func TestIntegrationTimeout(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	
	// Set a reasonable timeout for integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Ensure tests don't run indefinitely
	timeout := 5 * time.Minute
	done := make(chan bool)
	
	go func() {
		// Run a simple test that should complete quickly
		output, err := runner.RunCommand(t, "echo 'timeout test'")
		if err != nil {
			t.Errorf("Timeout test failed: %v\nOutput: %s", err, output)
		}
		done <- true
	}()
	
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(timeout):
		t.Fatalf("Integration test timed out after %v", timeout)
	}
}