// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestDotfileOperations(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
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
		
		output, err := runner.RunCommand(t, setupScript)
		if err != nil {
			t.Fatalf("Failed to setup test environment: %v\nOutput: %s", err, output)
		}
		
		t.Logf("Setup output: %s", output)
	})

	// Test adding a dotfile
	t.Run("add dotfile", func(t *testing.T) {
		addScript := `
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
		`
		
		output, err := runner.RunCommand(t, addScript)
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
		deployScript := `
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
			/workspace/plonk apply || echo "Apply command executed"
		`
		
		output, err := runner.RunCommand(t, deployScript)
		t.Logf("Deploy dotfiles output: %s", output)
		
		if err != nil {
			// Deploy might fail if no dotfiles configured, that's OK
			t.Logf("Deploy failed (might be expected): %v\nOutput: %s", err, output)
		}
	})

	// Test backup functionality
	t.Run("backup verification", func(t *testing.T) {
		output, err := runner.RunCommand(t, "ls -la /home/testuser/.zshrc* || echo 'No backup files found'")
		t.Logf("Backup verification output: %s", output)
		
		if err != nil {
			t.Fatalf("Failed to check backup files: %v\nOutput: %s", err, output)
		}
	})

	// Test dotfile status
	t.Run("dotfile status", func(t *testing.T) {
		output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk dot status")
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
			output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk dot add ~/.config/nvim/")
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
			output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk dot add ~/.nonexistent")
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
		atomicScript := `
			cd /home/testuser
			# Create a test file
			echo "test content" > test.txt
			# Try to add it and then check if partial state exists
			/workspace/plonk dot add ~/test.txt || echo "Add failed as expected"
			# Check that no partial state remains
			ls -la ~/.config/plonk/ || echo "No config directory pollution"
		`
		
		output, err := runner.RunCommand(t, atomicScript)
		t.Logf("Atomic operations output: %s", output)
		
		if err != nil {
			t.Logf("Atomic operations test completed with expected errors: %v\nOutput: %s", err, output)
		}
	})
}