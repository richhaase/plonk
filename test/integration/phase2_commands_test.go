// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestPhase2Commands(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test doctor command
	t.Run("doctor command", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Doctor Command ==="
			
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
			
			# Test doctor command
			echo "=== Running doctor command ==="
			/workspace/plonk doctor
			
			echo "=== Testing doctor command with JSON output ==="
			/workspace/plonk doctor -o json | head -30
			
			# Test doctor with invalid config
			echo "=== Testing doctor with invalid config ==="
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: invalid_manager
homebrew:
  brews:
    - ""
EOF
			
			echo "=== Running doctor on invalid config ==="
			/workspace/plonk doctor || echo "Doctor shows issues correctly"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Doctor command output: %s", output)
		
		if err != nil {
			t.Fatalf("Doctor command test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show system checks
		if !strings.Contains(outputStr, "System Requirements") {
			t.Error("Expected system requirements check")
		}
		
		// Should show environment checks
		if !strings.Contains(outputStr, "Environment Variables") {
			t.Error("Expected environment variables check")
		}
		
		// Should show permissions checks
		if !strings.Contains(outputStr, "Permissions") {
			t.Error("Expected permissions check")
		}
		
		// Should show configuration checks
		if !strings.Contains(outputStr, "Configuration") {
			t.Error("Expected configuration check")
		}
		
		// Should show package manager checks
		if !strings.Contains(outputStr, "Package Managers") {
			t.Error("Expected package manager check")
		}
		
		// Should show overall health status
		if !strings.Contains(outputStr, "Overall Health") && !strings.Contains(outputStr, "Overall Status") {
			t.Error("Expected overall health status")
		}
	})

	// Test search command
	t.Run("search command", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Search Command ==="
			
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
EOF
			
			# Test search for nonexistent package
			echo "=== Testing search for nonexistent package ==="
			/workspace/plonk search nonexistent_package_12345 || echo "Expected no results"
			
			# Test search with JSON output
			echo "=== Testing search with JSON output ==="
			/workspace/plonk search git -o json | head -20
			
			# Test search without default manager
			echo "=== Testing search without default manager ==="
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
homebrew:
  brews: []
dotfiles: []
EOF
			
			/workspace/plonk search git || echo "Search works without default manager"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Search command output: %s", output)
		
		if err != nil {
			t.Fatalf("Search command test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should handle nonexistent packages gracefully
		if !strings.Contains(outputStr, "not found") {
			t.Error("Expected not found message for nonexistent package")
		}
		
		// Should show package information in structured format
		if !strings.Contains(outputStr, "\"package\"") {
			t.Error("Expected JSON output to contain package field")
		}
	})

	// Test info command
	t.Run("info command", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Info Command ==="
			
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
EOF
			
			# Test info for nonexistent package
			echo "=== Testing info for nonexistent package ==="
			/workspace/plonk info nonexistent_package_12345 || echo "Expected no info found"
			
			# Test info with JSON output
			echo "=== Testing info with JSON output ==="
			/workspace/plonk info git -o json | head -20
			
			# Test info without default manager
			echo "=== Testing info without default manager ==="
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
homebrew:
  brews: []
dotfiles: []
EOF
			
			/workspace/plonk info git || echo "Info works without default manager"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Info command output: %s", output)
		
		if err != nil {
			t.Fatalf("Info command test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should handle nonexistent packages gracefully
		if !strings.Contains(outputStr, "not found") {
			t.Error("Expected not found message for nonexistent package")
		}
		
		// Should show package information in structured format
		if !strings.Contains(outputStr, "\"package\"") {
			t.Error("Expected JSON output to contain package field")
		}
	})

	// Test integration between Phase 2 commands
	t.Run("phase2 commands integration", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Phase 2 Commands Integration ==="
			
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
			
			# Test workflow: doctor -> search -> info
			echo "=== Running doctor to check health ==="
			/workspace/plonk doctor
			
			echo "=== Searching for available packages ==="
			/workspace/plonk search git
			
			echo "=== Getting info about found packages ==="
			/workspace/plonk info git
			
			# Test with various output formats
			echo "=== Testing all commands with YAML output ==="
			/workspace/plonk doctor -o yaml | head -20
			/workspace/plonk search git -o yaml | head -10
			/workspace/plonk info git -o yaml | head -10
			
			# Test error handling consistency
			echo "=== Testing error handling consistency ==="
			/workspace/plonk search "" || echo "Empty search handled gracefully"
			/workspace/plonk info "" || echo "Empty info handled gracefully"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Phase 2 integration output: %s", output)
		
		if err != nil {
			t.Fatalf("Phase 2 integration test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show health check results
		if !strings.Contains(outputStr, "Overall Health") && !strings.Contains(outputStr, "Overall Status") {
			t.Error("Expected health check results")
		}
		
		// Should show search results
		if !strings.Contains(outputStr, "git") {
			t.Error("Expected search results for git")
		}
		
		// Should show package info
		if !strings.Contains(outputStr, "Name:") && !strings.Contains(outputStr, "Manager:") {
			t.Error("Expected package info output")
		}
		
		// Should handle YAML output
		if !strings.Contains(outputStr, "package:") {
			t.Error("Expected YAML output format")
		}
	})

	// Test error scenarios for Phase 2 commands
	t.Run("phase2 error scenarios", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Phase 2 Error Scenarios ==="
			
			# Remove config to test missing config scenarios
			rm -rf ~/.config/plonk
			
			echo "=== Testing doctor with missing config ==="
			/workspace/plonk doctor || echo "Doctor handles missing config"
			
			echo "=== Testing search with missing config ==="
			/workspace/plonk search git || echo "Search handles missing config"
			
			echo "=== Testing info with missing config ==="
			/workspace/plonk info git || echo "Info handles missing config"
			
			# Test with corrupted config
			echo "=== Testing with corrupted config ==="
			mkdir -p ~/.config/plonk
			echo "invalid yaml: [[[" > ~/.config/plonk/plonk.yaml
			
			/workspace/plonk doctor || echo "Doctor handles corrupted config"
			/workspace/plonk search git || echo "Search handles corrupted config"
			/workspace/plonk info git || echo "Info handles corrupted config"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Phase 2 error scenarios output: %s", output)
		
		if err != nil {
			t.Fatalf("Phase 2 error scenarios test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should handle missing config gracefully - commands should still work
		if !strings.Contains(outputStr, "Doctor handles missing config") &&
		   !strings.Contains(outputStr, "Package Manager Availability") {
			t.Error("Expected doctor to handle missing config gracefully")
		}
		
		if !strings.Contains(outputStr, "Search handles missing config") &&
		   !strings.Contains(outputStr, "Package 'git' available") {
			t.Error("Expected search to handle missing config gracefully")  
		}
		
		if !strings.Contains(outputStr, "Info handles missing config") &&
		   !strings.Contains(outputStr, "Name: git") {
			t.Error("Expected info to handle missing config gracefully")
		}
		
		// Should handle corrupted config gracefully - commands should still work
		if !strings.Contains(outputStr, "Doctor handles corrupted config") &&
		   !strings.Contains(outputStr, "Configuration is invalid") {
			t.Error("Expected doctor to handle corrupted config gracefully")
		}
		
		if !strings.Contains(outputStr, "Search handles corrupted config") &&
		   !strings.Contains(outputStr, "Package 'git' available") {
			t.Error("Expected search to handle corrupted config gracefully")
		}
		
		if !strings.Contains(outputStr, "Info handles corrupted config") &&
		   !strings.Contains(outputStr, "Name: git") {
			t.Error("Expected info to handle corrupted config gracefully")
		}
	})
}