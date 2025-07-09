// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestConfigCommands(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test config validate command
	t.Run("config validate", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Config Validate Command ==="
			
			# Create plonk config directory with valid config
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - git
    - curl
dotfiles:
  - zshrc
  - gitconfig
EOF
			
			# Test validate with valid config
			echo "=== Running config validate on valid config ==="
			/workspace/plonk config validate
			
			# Test validate with invalid config
			echo "=== Testing config validate on invalid config ==="
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: invalid_manager
homebrew:
  brews:
    - ""
dotfiles:
  - ""
EOF
			
			echo "=== Running config validate on invalid config ==="
			/workspace/plonk config validate || echo "Expected validation failure"
			
			# Test validate with missing config
			echo "=== Testing config validate with missing config ==="
			rm -f ~/.config/plonk/plonk.yaml
			/workspace/plonk config validate || echo "Expected missing config error"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Config validate output: %s", output)
		
		if err != nil {
			t.Fatalf("Config validate test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show valid config success
		if !strings.Contains(outputStr, "Configuration is valid") {
			t.Error("Expected valid configuration message")
		}
		
		// Should show invalid config error
		if !strings.Contains(outputStr, "Configuration has 2 errors") {
			t.Error("Expected validation failure with error count")
		}
		
		// Should show missing config error
		if !strings.Contains(outputStr, "Configuration file not found") {
			t.Error("Expected missing config error message")
		}
	})

	// Test env command
	t.Run("env command", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Env Command ==="
			
			# Create plonk config directory with valid config
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
dotfiles: []
EOF
			
			# Test env command
			echo "=== Running env command ==="
			/workspace/plonk env
			
			echo "=== Testing env command with JSON output ==="
			/workspace/plonk env -o json | head -20
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Env command output: %s", output)
		
		if err != nil {
			t.Fatalf("Env command test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show system information
		if !strings.Contains(outputStr, "System Information") {
			t.Error("Expected system information section")
		}
		
		// Should show configuration information
		if !strings.Contains(outputStr, "Configuration") {
			t.Error("Expected configuration section")
		}
		
		// Should show package managers
		if !strings.Contains(outputStr, "Package Managers") {
			t.Error("Expected package managers section")
		}
		
		// Should show environment variables
		if !strings.Contains(outputStr, "Environment Variables") {
			t.Error("Expected environment variables section")
		}
		
		// Should show paths
		if !strings.Contains(outputStr, "Paths") {
			t.Error("Expected paths section")
		}
		
		// Should show Linux OS
		if !strings.Contains(outputStr, "OS: linux") {
			t.Error("Expected OS: linux")
		}
	})

	// Test config show command integration
	t.Run("config show with new commands", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Config Show Integration ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - git
    - tree
dotfiles:
  - zshrc
  - gitconfig
EOF
			
			# Test show command
			echo "=== Running config show ==="
			/workspace/plonk config show
			
			# Test validate then show
			echo "=== Validate then show ==="
			/workspace/plonk config validate
			/workspace/plonk config show | head -20
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Config show integration output: %s", output)
		
		if err != nil {
			t.Fatalf("Config show integration test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show config content
		if !strings.Contains(outputStr, "default_manager: homebrew") {
			t.Error("Expected config content to be shown")
		}
		
		// Should show validation success
		if !strings.Contains(outputStr, "Configuration is valid") {
			t.Error("Expected validation success message")
		}
	})
}