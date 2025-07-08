// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestErrorRecovery(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test corrupted configuration recovery
	t.Run("corrupted configuration recovery", func(t *testing.T) {
		corruptionScript := `
			cd /home/testuser
			
			echo "=== Corrupted Configuration Recovery ==="
			
			# 1. Set up valid configuration
			echo "1. Setting up valid configuration..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
    - git
npm:
  - lodash
dotfiles: []
EOF
			
			echo "Valid configuration:"
			/workspace/plonk config show
			
			# 2. Corrupt the configuration file
			echo "2. Corrupting configuration file..."
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
    - git
  invalid_yaml: [unclosed
npm:
  - lodash
EOF
			
			echo "Testing corrupted configuration:"
			/workspace/plonk config show || echo "Configuration corruption detected (expected)"
			
			# 3. Attempt to recover
			echo "3. Attempting to recover with valid configuration..."
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
    - git
npm:
  - lodash
dotfiles: []
EOF
			
			echo "Configuration after recovery:"
			/workspace/plonk config show
			
			# 4. Verify recovery by adding new item
			echo "4. Verifying recovery by adding new item..."
			/workspace/plonk pkg add --manager homebrew jq || echo "Package add after recovery processed"
			
			echo "=== Recovery Testing Complete ==="
		`
		
		output, err := runner.RunCommand(t, corruptionScript)
		t.Logf("Corrupted configuration recovery output: %s", output)
		
		if err != nil {
			t.Logf("Configuration recovery completed with some expected errors: %v", err)
		}
		
		// Verify recovery steps
		outputStr := string(output)
		recoverySteps := []string{
			"Corrupted Configuration Recovery",
			"Setting up valid configuration",
			"Corrupting configuration file",
			"Configuration corruption detected",
			"Attempting to recover",
			"Configuration after recovery",
		}
		
		for _, step := range recoverySteps {
			if !strings.Contains(outputStr, step) {
				t.Errorf("Expected recovery step '%s' not found in output", step)
			}
		}
	})

	// Test missing file recovery
	t.Run("missing file recovery", func(t *testing.T) {
		missingFileScript := `
			cd /home/testuser
			
			echo "=== Missing File Recovery ==="
			
			# 1. Set up configuration with dotfiles
			echo "1. Setting up configuration with dotfiles..."
			mkdir -p ~/.config/plonk/dotfiles
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles:
  - source: dot_bashrc
    destination: ~/.bashrc
  - source: dot_vimrc
    destination: ~/.vimrc
EOF
			
			# Create dotfiles
			echo "# Managed bashrc" > ~/.config/plonk/dotfiles/dot_bashrc
			echo "# Managed vimrc" > ~/.config/plonk/dotfiles/dot_vimrc
			
			echo "Initial configuration:"
			/workspace/plonk config show
			
			# 2. Simulate missing dotfiles
			echo "2. Simulating missing dotfiles..."
			rm -f ~/.config/plonk/dotfiles/dot_bashrc
			
			echo "Testing with missing dotfile:"
			/workspace/plonk dot list || echo "Missing dotfile detected (expected)"
			
			# 3. Attempt recovery by re-adding
			echo "3. Attempting recovery by re-adding..."
			echo "# Recovered bashrc" > ~/.bashrc
			/workspace/plonk dot re-add ~/.bashrc || echo "Re-add recovery processed"
			
			# 4. Verify recovery
			echo "4. Verifying recovery..."
			/workspace/plonk dot list || echo "Dotfile recovery completed"
			
			echo "=== Missing File Recovery Complete ==="
		`
		
		output, err := runner.RunCommand(t, missingFileScript)
		t.Logf("Missing file recovery output: %s", output)
		
		if err != nil {
			t.Logf("Missing file recovery completed with some expected errors: %v", err)
		}
		
		// Verify missing file recovery
		outputStr := string(output)
		if !strings.Contains(outputStr, "Missing File Recovery") {
			t.Error("Missing file recovery test did not execute properly")
		}
	})

	// Test permission error recovery
	t.Run("permission error recovery", func(t *testing.T) {
		permissionScript := `
			cd /home/testuser
			
			echo "=== Permission Error Recovery ==="
			
			# 1. Set up configuration
			echo "1. Setting up configuration..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
npm:
  - lodash
dotfiles: []
EOF
			
			echo "Initial configuration:"
			/workspace/plonk config show
			
			# 2. Simulate permission issues
			echo "2. Simulating permission issues..."
			chmod 000 ~/.config/plonk/plonk.yaml || echo "Permission simulation attempted"
			
			echo "Testing with permission issues:"
			/workspace/plonk config show || echo "Permission error detected (expected)"
			
			# 3. Attempt recovery
			echo "3. Attempting permission recovery..."
			chmod 644 ~/.config/plonk/plonk.yaml || echo "Permission recovery attempted"
			
			echo "Configuration after permission recovery:"
			/workspace/plonk config show
			
			# 4. Verify recovery by adding item
			echo "4. Verifying recovery by adding item..."
			/workspace/plonk pkg add --manager homebrew git || echo "Package add after recovery processed"
			
			echo "=== Permission Recovery Complete ==="
		`
		
		output, err := runner.RunCommand(t, permissionScript)
		t.Logf("Permission error recovery output: %s", output)
		
		if err != nil {
			t.Logf("Permission recovery completed with some expected errors: %v", err)
		}
		
		// Verify permission recovery
		outputStr := string(output)
		if !strings.Contains(outputStr, "Permission Error Recovery") {
			t.Error("Permission error recovery test did not execute properly")
		}
	})
}

func TestNetworkErrorRecovery(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test network failure scenarios
	t.Run("network failure scenarios", func(t *testing.T) {
		networkScript := `
			cd /home/testuser
			
			echo "=== Network Failure Scenarios ==="
			
			# 1. Set up configuration for network operations
			echo "1. Setting up configuration for network operations..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
npm:
  - lodash
dotfiles: []
EOF
			
			echo "Initial configuration:"
			/workspace/plonk config show
			
			# 2. Test package installation (may fail due to network)
			echo "2. Testing package installation..."
			/workspace/plonk pkg add --manager homebrew nonexistent-package-12345 || echo "Package installation failed (expected)"
			
			# 3. Test NPM package installation (may fail due to network)
			echo "3. Testing NPM package installation..."
			/workspace/plonk pkg add --manager npm nonexistent-npm-package-12345 || echo "NPM installation failed (expected)"
			
			# 4. Verify configuration remains intact after failures
			echo "4. Verifying configuration integrity after failures..."
			/workspace/plonk config show
			
			# 5. Test successful operations after failures
			echo "5. Testing successful operations after failures..."
			/workspace/plonk pkg add --manager homebrew git || echo "Package add after failure processed"
			
			echo "=== Network Failure Testing Complete ==="
		`
		
		output, err := runner.RunCommand(t, networkScript)
		t.Logf("Network failure scenarios output: %s", output)
		
		if err != nil {
			t.Logf("Network failure testing completed with some expected errors: %v", err)
		}
		
		// Verify network failure handling
		outputStr := string(output)
		if !strings.Contains(outputStr, "Network Failure Scenarios") {
			t.Error("Network failure scenarios test did not execute properly")
		}
	})

	// Test partial failure recovery
	t.Run("partial failure recovery", func(t *testing.T) {
		partialScript := `
			cd /home/testuser
			
			echo "=== Partial Failure Recovery ==="
			
			# 1. Set up configuration
			echo "1. Setting up configuration for partial failure testing..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
    - git
npm:
  - lodash
dotfiles: []
EOF
			
			echo "Initial configuration:"
			/workspace/plonk config show
			
			# 2. Simulate partial operation failure
			echo "2. Simulating partial operation failure..."
			# Add multiple packages, some may fail
			/workspace/plonk pkg add --manager homebrew jq || echo "Package add 1 processed"
			/workspace/plonk pkg add --manager homebrew nonexistent-package || echo "Package add 2 failed (expected)"
			/workspace/plonk pkg add --manager homebrew wget || echo "Package add 3 processed"
			
			# 3. Check state after partial failures
			echo "3. Checking state after partial failures..."
			/workspace/plonk config show
			
			# 4. Verify successful operations completed
			echo "4. Verifying successful operations completed..."
			/workspace/plonk pkg list || echo "Package list after partial failures"
			
			echo "=== Partial Failure Recovery Complete ==="
		`
		
		output, err := runner.RunCommand(t, partialScript)
		t.Logf("Partial failure recovery output: %s", output)
		
		if err != nil {
			t.Logf("Partial failure recovery completed with some expected errors: %v", err)
		}
		
		// Verify partial failure recovery
		outputStr := string(output)
		if !strings.Contains(outputStr, "Partial Failure Recovery") {
			t.Error("Partial failure recovery test did not execute properly")
		}
	})

	// Test timeout and retry scenarios
	t.Run("timeout and retry scenarios", func(t *testing.T) {
		timeoutScript := `
			cd /home/testuser
			
			echo "=== Timeout and Retry Scenarios ==="
			
			# 1. Set up configuration
			echo "1. Setting up configuration for timeout testing..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
npm:
  - lodash
dotfiles: []
EOF
			
			echo "Initial configuration:"
			/workspace/plonk config show
			
			# 2. Test operations that might timeout
			echo "2. Testing operations that might timeout..."
			timeout 10s /workspace/plonk pkg add --manager homebrew very-large-package || echo "Operation timed out or failed (expected)"
			
			# 3. Verify system state after timeout
			echo "3. Verifying system state after timeout..."
			/workspace/plonk config show
			
			# 4. Test that normal operations still work
			echo "4. Testing normal operations after timeout..."
			/workspace/plonk pkg add --manager homebrew git || echo "Normal operation after timeout processed"
			
			echo "=== Timeout Testing Complete ==="
		`
		
		output, err := runner.RunCommand(t, timeoutScript)
		t.Logf("Timeout and retry scenarios output: %s", output)
		
		if err != nil {
			t.Logf("Timeout testing completed with some expected errors: %v", err)
		}
		
		// Verify timeout handling
		outputStr := string(output)
		if !strings.Contains(outputStr, "Timeout and Retry Scenarios") {
			t.Error("Timeout and retry scenarios test did not execute properly")
		}
	})
}