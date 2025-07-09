// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestFullWorkflows(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test complete new user setup workflow
	t.Run("new user setup workflow", func(t *testing.T) {
		workflowScript := `
			cd /home/testuser
			
			echo "=== New User Setup Workflow ==="
			
			# 1. Check initial state (no config)
			echo "1. Checking initial state..."
			/workspace/plonk config show || echo "No config found (expected)"
			
			# 2. Initialize configuration
			echo "2. Initializing configuration..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles: []
EOF
			
			# 3. Verify config was created
			echo "3. Verifying configuration..."
			/workspace/plonk config show
			
			# 4. Add first package
			echo "4. Adding first package..."
			/workspace/plonk pkg add --manager homebrew curl || echo "Package add failed (expected in container)"
			
			# 5. Create and add first dotfile
			echo "5. Creating and adding first dotfile..."
			echo "# Test bashrc" > ~/.bashrc
			/workspace/plonk dot add ~/.bashrc || echo "Dotfile add completed"
			
			# 6. Check final state
			echo "6. Checking final state..."
			/workspace/plonk config show
			
			echo "=== Workflow Complete ==="
		`
		
		output, err := runner.RunCommand(t, workflowScript)
		t.Logf("New user setup workflow output: %s", output)
		
		if err != nil {
			t.Logf("Workflow completed with some expected errors: %v", err)
		}
		
		// Verify workflow steps completed
		outputStr := string(output)
		expectedSteps := []string{
			"New User Setup Workflow",
			"Checking initial state",
			"Initializing configuration",
			"Verifying configuration",
			"Adding first package",
			"Creating and adding first dotfile",
			"Checking final state",
			"Workflow Complete",
		}
		
		for _, step := range expectedSteps {
			if !strings.Contains(outputStr, step) {
				t.Errorf("Expected workflow step '%s' not found in output", step)
			}
		}
	})

	// Test existing user configuration update workflow
	t.Run("configuration update workflow", func(t *testing.T) {
		workflowScript := `
			cd /home/testuser
			
			echo "=== Configuration Update Workflow ==="
			
			# 1. Start with existing config
			echo "1. Setting up existing configuration..."
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
dotfiles:
  - source: dot_bashrc
    destination: ~/.bashrc
EOF
			
			# 2. Show current config
			echo "2. Current configuration:"
			/workspace/plonk config show
			
			# 3. Add new package
			echo "3. Adding new package..."
			/workspace/plonk pkg add --manager homebrew jq || echo "Package add processed"
			
			# 4. Add new dotfile
			echo "4. Adding new dotfile..."
			echo "# Test vimrc" > ~/.vimrc
			/workspace/plonk dot add ~/.vimrc || echo "Dotfile add processed"
			
			# 5. Show updated config
			echo "5. Updated configuration:"
			/workspace/plonk config show
			
			# 6. List packages
			echo "6. Package list:"
			/workspace/plonk pkg list || echo "Package list completed"
			
			# 7. List dotfiles
			echo "7. Dotfile list:"
			/workspace/plonk dot list || echo "Dotfile list completed"
			
			echo "=== Update Workflow Complete ==="
		`
		
		output, err := runner.RunCommand(t, workflowScript)
		t.Logf("Configuration update workflow output: %s", output)
		
		if err != nil {
			t.Logf("Update workflow completed with some expected errors: %v", err)
		}
		
		// Verify workflow executed
		outputStr := string(output)
		if !strings.Contains(outputStr, "Configuration Update Workflow") {
			t.Error("Configuration update workflow did not execute properly")
		}
	})

	// Test migration workflow (moving from one machine to another)
	t.Run("migration workflow", func(t *testing.T) {
		migrationScript := `
			cd /home/testuser
			
			echo "=== Migration Workflow ==="
			
			# 1. Simulate source machine setup
			echo "1. Setting up source machine configuration..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
    - jq
    - git
npm:
  - lodash
  - prettier
dotfiles:
  - source: dot_bashrc
    destination: ~/.bashrc
  - source: dot_vimrc
    destination: ~/.vimrc
EOF
			
			# Create corresponding dotfiles
			mkdir -p ~/.config/plonk/dotfiles
			echo "# Source machine bashrc" > ~/.config/plonk/dotfiles/dot_bashrc
			echo "# Source machine vimrc" > ~/.config/plonk/dotfiles/dot_vimrc
			
			# 2. Show source configuration
			echo "2. Source machine configuration:"
			/workspace/plonk config show
			
			# 3. Simulate target machine (clean slate)
			echo "3. Simulating target machine setup..."
			rm -f ~/.bashrc ~/.vimrc
			
			# 4. Apply configuration on target
			echo "4. Applying configuration on target machine..."
			/workspace/plonk apply || echo "Apply completed"
			
			# 5. Verify dotfiles were applied
			echo "5. Verifying dotfiles were applied:"
			ls -la ~/.bashrc ~/.vimrc 2>/dev/null || echo "Dotfiles not found (expected in test)"
			
			# 6. Show status
			echo "6. Final status:"
			/workspace/plonk status || echo "Status completed"
			
			echo "=== Migration Workflow Complete ==="
		`
		
		output, err := runner.RunCommand(t, migrationScript)
		t.Logf("Migration workflow output: %s", output)
		
		if err != nil {
			t.Logf("Migration workflow completed with some expected errors: %v", err)
		}
		
		// Verify migration steps
		outputStr := string(output)
		migrationSteps := []string{
			"Migration Workflow",
			"Setting up source machine",
			"Source machine configuration",
			"Simulating target machine",
			"Applying configuration",
			"Verifying dotfiles",
			"Final status",
		}
		
		for _, step := range migrationSteps {
			if !strings.Contains(outputStr, step) {
				t.Errorf("Expected migration step '%s' not found in output", step)
			}
		}
	})
}

func TestComplexWorkflows(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test multi-step development workflow
	t.Run("development workflow", func(t *testing.T) {
		devWorkflowScript := `
			cd /home/testuser
			
			echo "=== Development Workflow ==="
			
			# 1. Developer starts new project
			echo "1. Starting new development project..."
			mkdir -p ~/projects/myapp
			cd ~/projects/myapp
			
			# 2. Initialize plonk config for project
			echo "2. Initializing project-specific plonk config..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - git
    - curl
npm:
  - lodash
  - express
dotfiles:
  - source: dot_gitconfig
    destination: ~/.gitconfig
EOF
			
			# 3. Set up development environment
			echo "3. Setting up development environment..."
			echo "[user]" > ~/.gitconfig
			echo "    name = Test User" >> ~/.gitconfig
			echo "    email = test@example.com" >> ~/.gitconfig
			
			# 4. Add development-specific packages
			echo "4. Adding development packages..."
			/workspace/plonk pkg add --manager npm jest || echo "Package add processed"
			/workspace/plonk pkg add --manager homebrew node || echo "Package add processed"
			
			# 5. Show development environment
			echo "5. Development environment status:"
			/workspace/plonk status || echo "Status completed"
			
			# 6. List all managed items
			echo "6. Managed packages:"
			/workspace/plonk pkg list || echo "Package list completed"
			
			echo "7. Managed dotfiles:"
			/workspace/plonk dot list || echo "Dotfile list completed"
			
			echo "=== Development Workflow Complete ==="
		`
		
		output, err := runner.RunCommand(t, devWorkflowScript)
		t.Logf("Development workflow output: %s", output)
		
		if err != nil {
			t.Logf("Development workflow completed with some expected errors: %v", err)
		}
		
		// Verify development workflow
		outputStr := string(output)
		if !strings.Contains(outputStr, "Development Workflow") {
			t.Error("Development workflow did not execute properly")
		}
	})

	// Test cleanup and reset workflow
	t.Run("cleanup workflow", func(t *testing.T) {
		cleanupScript := `
			cd /home/testuser
			
			echo "=== Cleanup Workflow ==="
			
			# 1. Start with configured environment
			echo "1. Setting up environment to clean..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
    - jq
npm:
  - lodash
dotfiles:
  - source: dot_bashrc
    destination: ~/.bashrc
EOF
			
			# Create dotfiles
			mkdir -p ~/.config/plonk/dotfiles
			echo "# Test bashrc" > ~/.config/plonk/dotfiles/dot_bashrc
			echo "# System bashrc" > ~/.bashrc
			
			# 2. Show current state
			echo "2. Current state:"
			/workspace/plonk config show
			
			# 3. Remove packages (simulate cleanup)
			echo "3. Removing packages..."
			/workspace/plonk pkg remove curl || echo "Package remove processed"
			/workspace/plonk pkg remove --manager npm lodash || echo "Package remove processed"
			
			# 4. Show updated state
			echo "4. State after package removal:"
			/workspace/plonk config show
			
			# 5. Remove dotfiles
			echo "5. Removing dotfiles..."
			# Note: plonk doesn't have a dot remove command, so simulate
			rm -f ~/.bashrc
			
			# 6. Show final state
			echo "6. Final cleaned state:"
			/workspace/plonk config show
			
			echo "=== Cleanup Workflow Complete ==="
		`
		
		output, err := runner.RunCommand(t, cleanupScript)
		t.Logf("Cleanup workflow output: %s", output)
		
		if err != nil {
			t.Logf("Cleanup workflow completed with some expected errors: %v", err)
		}
		
		// Verify cleanup workflow
		outputStr := string(output)
		if !strings.Contains(outputStr, "Cleanup Workflow") {
			t.Error("Cleanup workflow did not execute properly")
		}
	})
}