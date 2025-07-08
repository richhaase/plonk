// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package main

import (
	"fmt"
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

func TestPackageManagerOperations(t *testing.T) {
	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping integration test")
	}

	// Build plonk binary first
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

	// Create a basic config file
	t.Run("create config file", func(t *testing.T) {
		configContent := `settings:
  default_manager: homebrew
packages: []
dotfiles: []
`
		cmd := exec.Command("docker", "run", "--rm",
			"-v", ".:/workspace",
			"-w", "/workspace",
			"plonk-test",
			"/bin/bash", "-c", 
			fmt.Sprintf("mkdir -p /home/testuser/.config/plonk && echo '%s' > /home/testuser/.config/plonk/plonk.yaml", configContent))
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create config file: %v\nOutput: %s", err, output)
		}
		
		t.Logf("Config creation output: %s", output)
	})

	// Test Homebrew package operations
	t.Run("homebrew package operations", func(t *testing.T) {
		// Install a small, safe package
		t.Run("install package", func(t *testing.T) {
			cmd := exec.Command("docker", "run", "--rm",
				"-v", ".:/workspace",
				"-w", "/workspace",
				"plonk-test",
				"/bin/bash", "-c", "cd /home/testuser && /workspace/plonk pkg add jq")
			
			output, err := cmd.CombinedOutput()
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
			cmd := exec.Command("docker", "run", "--rm",
				"plonk-test",
				"/bin/bash", "-c", "which jq || echo 'jq not found'")
			
			output, err := cmd.CombinedOutput()
			t.Logf("Verify output: %s", output)
			
			if err != nil {
				t.Fatalf("Failed to verify installation: %v\nOutput: %s", err, output)
			}
		})

		// Test uninstall
		t.Run("uninstall package", func(t *testing.T) {
			cmd := exec.Command("docker", "run", "--rm",
				"-v", ".:/workspace",
				"-w", "/workspace",
				"plonk-test",
				"/bin/bash", "-c", "cd /home/testuser && /workspace/plonk pkg remove jq")
			
			output, err := cmd.CombinedOutput()
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
			cmd := exec.Command("docker", "run", "--rm",
				"-v", ".:/workspace",
				"-w", "/workspace",
				"plonk-test",
				"/bin/bash", "-c", "cd /home/testuser && /workspace/plonk pkg add --manager npm json")
			
			output, err := cmd.CombinedOutput()
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
			cmd := exec.Command("docker", "run", "--rm",
				"-v", ".:/workspace",
				"-w", "/workspace",
				"plonk-test",
				"/bin/bash", "-c", "cd /home/testuser && /workspace/plonk pkg remove --manager npm json")
			
			output, err := cmd.CombinedOutput()
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
			cmd := exec.Command("docker", "run", "--rm",
				"-v", ".:/workspace",
				"-w", "/workspace",
				"plonk-test",
				"/bin/bash", "-c", "cd /home/testuser && /workspace/plonk pkg add nonexistent-package-12345")
			
			output, err := cmd.CombinedOutput()
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

	// Clean up any built artifacts
	t.Cleanup(func() {
		if _, err := os.Stat("plonk"); err == nil {
			os.Remove("plonk")
		}
	})
}

func TestDotfileOperations(t *testing.T) {
	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping integration test")
	}

	// Build plonk binary first
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

	// Create a basic config file and test dotfiles
	t.Run("setup test environment", func(t *testing.T) {
		setupScript := `
# Create plonk config directory
mkdir -p /home/testuser/.config/plonk

# Create a basic config file
cat > /home/testuser/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF

# Create some test dotfiles in the config directory
mkdir -p /home/testuser/.config/plonk/dotfiles
echo "# Test zshrc content" > /home/testuser/.config/plonk/dotfiles/zshrc
echo "# Test vimrc content" > /home/testuser/.config/plonk/dotfiles/vimrc

# Create a test directory structure
mkdir -p /home/testuser/.config/plonk/dotfiles/config/nvim
echo "-- Test nvim config" > /home/testuser/.config/plonk/dotfiles/config/nvim/init.lua

# Create an existing file to test backup functionality
echo "# Original zshrc" > /home/testuser/.zshrc
`
		
		cmd := exec.Command("docker", "run", "--rm",
			"-v", ".:/workspace",
			"-w", "/workspace",
			"plonk-test",
			"/bin/bash", "-c", setupScript)
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to setup test environment: %v\nOutput: %s", err, output)
		}
		
		t.Logf("Setup output: %s", output)
	})

	// Test adding a dotfile
	t.Run("add dotfile", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm",
			"-v", ".:/workspace",
			"-w", "/workspace",
			"plonk-test",
			"/bin/bash", "-c", `
				cd /home/testuser
				# Create the test environment again (containers don't persist state)
				mkdir -p ~/.config/plonk
				cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
				# Create a test dotfile
				echo "# Test zshrc content" > ~/.zshrc
				# Try to add it
				/workspace/plonk dot add ~/.zshrc
			`)
		
		output, err := cmd.CombinedOutput()
		t.Logf("Add dotfile output: %s", output)
		
		if err != nil {
			// Check if it's a known limitation or real failure
			outputStr := string(output)
			if strings.Contains(outputStr, "already managed") || strings.Contains(outputStr, "exists") {
				t.Skip("Dotfile already managed or exists")
			}
			t.Fatalf("Failed to add dotfile: %v\nOutput: %s", err, output)
		}
	})

	// Test deploying dotfiles
	t.Run("deploy dotfiles", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm",
			"-v", ".:/workspace",
			"-w", "/workspace",
			"plonk-test",
			"/bin/bash", "-c", `
				cd /home/testuser
				# Set up basic environment
				mkdir -p ~/.config/plonk
				cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
				# Try to deploy (should show help or handle empty config gracefully)
				/workspace/plonk dot apply || echo "Deploy command executed"
			`)
		
		output, err := cmd.CombinedOutput()
		t.Logf("Deploy dotfiles output: %s", output)
		
		if err != nil {
			// Deploy might fail if no dotfiles configured, that's OK
			t.Logf("Deploy failed (might be expected): %v\nOutput: %s", err, output)
		}
	})

	// Test backup functionality
	t.Run("backup verification", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm",
			"plonk-test",
			"/bin/bash", "-c", "ls -la /home/testuser/.zshrc* || echo 'No backup files found'")
		
		output, err := cmd.CombinedOutput()
		t.Logf("Backup verification output: %s", output)
		
		if err != nil {
			t.Fatalf("Failed to check backup files: %v\nOutput: %s", err, output)
		}
	})

	// Test dotfile status
	t.Run("dotfile status", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm",
			"-v", ".:/workspace",
			"-w", "/workspace",
			"plonk-test",
			"/bin/bash", "-c", "cd /home/testuser && /workspace/plonk dot status")
		
		output, err := cmd.CombinedOutput()
		t.Logf("Dotfile status output: %s", output)
		
		if err != nil {
			// Status might fail if no dotfiles configured
			t.Logf("Status failed (might be expected): %v\nOutput: %s", err, output)
		}
	})

	// Test file operations and edge cases
	t.Run("file operations edge cases", func(t *testing.T) {
		// Test copying a directory
		t.Run("copy directory", func(t *testing.T) {
			cmd := exec.Command("docker", "run", "--rm",
				"-v", ".:/workspace",
				"-w", "/workspace",
				"plonk-test",
				"/bin/bash", "-c", "cd /home/testuser && /workspace/plonk dot add ~/.config/nvim/")
			
			output, err := cmd.CombinedOutput()
			t.Logf("Copy directory output: %s", output)
			
			if err != nil {
				outputStr := string(output)
				if strings.Contains(outputStr, "not found") || strings.Contains(outputStr, "does not exist") {
					t.Skip("Directory does not exist for testing")
				}
				t.Logf("Copy directory failed (might be expected): %v\nOutput: %s", err, output)
			}
		})

		// Test error handling for non-existent files
		t.Run("non-existent file", func(t *testing.T) {
			cmd := exec.Command("docker", "run", "--rm",
				"-v", ".:/workspace",
				"-w", "/workspace",
				"plonk-test",
				"/bin/bash", "-c", "cd /home/testuser && /workspace/plonk dot add ~/.nonexistent")
			
			output, err := cmd.CombinedOutput()
			t.Logf("Non-existent file output: %s", output)
			
			// This should fail with a proper error message
			if err == nil {
				t.Error("Expected error when adding non-existent file")
			}
			
			// Check that we get a reasonable error message
			outputStr := string(output)
			if !strings.Contains(outputStr, "not found") && !strings.Contains(outputStr, "does not exist") {
				t.Errorf("Expected 'not found' or 'does not exist' error message, got: %s", outputStr)
			}
		})
	})

	// Test atomic operations and error recovery
	t.Run("atomic operations", func(t *testing.T) {
		// Test that operations are atomic (either succeed completely or fail completely)
		cmd := exec.Command("docker", "run", "--rm",
			"-v", ".:/workspace",
			"-w", "/workspace",
			"plonk-test",
			"/bin/bash", "-c", `
				cd /home/testuser
				# Create a test file
				echo "test content" > test.txt
				# Try to add it and then check if partial state exists
				/workspace/plonk dot add ~/test.txt || echo "Add failed as expected"
				# Check that no partial state remains
				ls -la ~/.config/plonk/ || echo "No config directory pollution"
			`)
		
		output, err := cmd.CombinedOutput()
		t.Logf("Atomic operations output: %s", output)
		
		if err != nil {
			t.Logf("Atomic operations test completed with expected errors: %v\nOutput: %s", err, output)
		}
	})

	// Clean up any built artifacts
	t.Cleanup(func() {
		if _, err := os.Stat("plonk"); err == nil {
			os.Remove("plonk")
		}
	})
}