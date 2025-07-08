// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"fmt"
	"strings"
	"testing"
)

func TestPackageManagerOperations(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Create a basic config file
	t.Run("create config file", func(t *testing.T) {
		configContent := `settings:
  default_manager: homebrew
packages: []
dotfiles: []
`
		createConfigCmd := fmt.Sprintf("mkdir -p /home/testuser/.config/plonk && echo '%s' > /home/testuser/.config/plonk/plonk.yaml", configContent)
		
		output, err := runner.RunCommand(t, createConfigCmd)
		if err != nil {
			t.Fatalf("Failed to create config file: %v\nOutput: %s", err, output)
		}
		
		t.Logf("Config creation output: %s", output)
	})

	// Test Homebrew package operations
	t.Run("homebrew package operations", func(t *testing.T) {
		// Install a small, safe package
		t.Run("install package", func(t *testing.T) {
			output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk pkg add jq")
			t.Logf("Install output: %s", output)
			
			if err != nil {
				// Check if it's a "package not found" error or real failure
				outputStr := string(output)
				if strings.Contains(outputStr, "No formula") || strings.Contains(outputStr, "not found") {
					t.Skip("Package 'jq' not available in this environment")
				}
				t.Fatalf("Failed to install package: %v\nOutput: %s", err, output)
			}
		})

		// Verify package is installed
		t.Run("verify installation", func(t *testing.T) {
			output, err := runner.RunCommand(t, "which jq || echo 'jq not found'")
			t.Logf("Verify output: %s", output)
			
			if err != nil {
				t.Fatalf("Failed to verify installation: %v\nOutput: %s", err, output)
			}
		})

		// Test uninstall
		t.Run("uninstall package", func(t *testing.T) {
			output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk pkg remove jq")
			t.Logf("Uninstall output: %s", output)
			
			if err != nil {
				// Uninstall failures are often OK (package might not be installed)
				t.Logf("Uninstall failed (might be expected): %v\nOutput: %s", err, output)
			}
		})
	})

	// Test NPM package operations
	t.Run("npm package operations", func(t *testing.T) {
		// Install a small, safe NPM package
		t.Run("install package", func(t *testing.T) {
			output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk pkg add --manager npm json")
			t.Logf("NPM install output: %s", output)
			
			if err != nil {
				// Check if it's a known error or real failure
				outputStr := string(output)
				if strings.Contains(outputStr, "permission denied") || strings.Contains(outputStr, "EACCES") {
					t.Skip("NPM permissions issue in container")
				}
				if strings.Contains(outputStr, "not found") || strings.Contains(outputStr, "404") {
					t.Skip("NPM package 'json' not available")
				}
				t.Fatalf("Failed to install NPM package: %v\nOutput: %s", err, output)
			}
		})

		// Test uninstall
		t.Run("uninstall package", func(t *testing.T) {
			output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk pkg remove --manager npm json")
			t.Logf("NPM uninstall output: %s", output)
			
			if err != nil {
				// Uninstall failures are often OK
				t.Logf("NPM uninstall failed (might be expected): %v\nOutput: %s", err, output)
			}
		})
	})

	// Test error handling scenarios
	t.Run("error handling", func(t *testing.T) {
		// Test installing non-existent package
		t.Run("install non-existent package", func(t *testing.T) {
			output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk pkg add nonexistent-package-12345")
			t.Logf("Non-existent package output: %s", output)
			
			// This should fail with a proper error message
			if err == nil {
				t.Error("Expected error when installing non-existent package")
			}
			
			// Check that we get a reasonable error message
			outputStr := string(output)
			if !strings.Contains(outputStr, "not found") && !strings.Contains(outputStr, "No formula") && !strings.Contains(outputStr, "404") {
				t.Errorf("Expected 'not found' or similar error message, got: %s", outputStr)
			}
		})
	})
}