// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestPlonkBinaryInContainer(t *testing.T) {
	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping integration test")
	}

	// Build Plonk binary inside container
	t.Run("build plonk binary", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm", 
			"-v", ".:/workspace",
			"-w", "/workspace",
			"plonk-test",
			"go", "build", "-buildvcs=false", "-o", "plonk", "./cmd/plonk")
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to build plonk binary: %v\nOutput: %s", err, output)
		}
		
		t.Logf("Build output: %s", output)
	})

	// Test plonk basic command
	t.Run("plonk basic command", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm",
			"-v", ".:/workspace", 
			"-w", "/workspace",
			"plonk-test",
			"./plonk")
		
		output, err := cmd.CombinedOutput()
		// plonk without arguments should show help and exit with code 0
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
		cmd := exec.Command("docker", "run", "--rm",
			"-v", ".:/workspace",
			"-w", "/workspace", 
			"plonk-test",
			"./plonk", "--help")
		
		output, err := cmd.CombinedOutput()
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
		cmd := exec.Command("docker", "run", "--rm",
			"-v", ".:/workspace",
			"-w", "/workspace",
			"plonk-test",
			"/bin/bash", "-c", "cd /home/testuser && /workspace/plonk config show")
		
		output, err := cmd.CombinedOutput()
		// config show may fail if no config exists, that's OK for this test
		t.Logf("Config show output: %s", output)
		t.Logf("Config show error (if any): %v", err)
		
		// Just verify the command runs and produces some output
		outputStr := string(output)
		if len(outputStr) == 0 && err == nil {
			t.Error("Config show produced no output and no error")
		}
	})

	// Clean up any built artifacts
	t.Cleanup(func() {
		if _, err := os.Stat("plonk"); err == nil {
			os.Remove("plonk")
		}
	})
}

func TestContainerIsolation(t *testing.T) {
	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping integration test")
	}

	// Test that each container starts with a fresh environment
	t.Run("fresh environment per container", func(t *testing.T) {
		// Run first container and create a test file
		cmd1 := exec.Command("docker", "run", "--rm",
			"plonk-test",
			"/bin/bash", "-c", "echo 'test content' > /tmp/testfile && cat /tmp/testfile")
		
		output1, err := cmd1.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create test file: %v\nOutput: %s", err, output1)
		}
		
		// Run second container and check that test file doesn't exist
		cmd2 := exec.Command("docker", "run", "--rm",
			"plonk-test",
			"/bin/bash", "-c", "ls /tmp/testfile || echo 'file not found'")
		
		output2, err := cmd2.CombinedOutput()
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
		cmd := exec.Command("docker", "run", "--rm",
			"plonk-test",
			"/bin/bash", "-c", "brew --version && npm --version && apt --version")
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Package managers not available: %v\nOutput: %s", err, output)
		}
		
		t.Logf("Package managers output: %s", output)
	})
}

func TestIntegrationTimeout(t *testing.T) {
	// Set a reasonable timeout for integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Ensure tests don't run indefinitely
	timeout := 5 * time.Minute
	done := make(chan bool)
	
	go func() {
		// Run a simple test that should complete quickly
		cmd := exec.Command("docker", "run", "--rm", "plonk-test", "echo", "timeout test")
		_, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Timeout test failed: %v", err)
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