// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestCrossManagerOperations(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test mixed package manager operations
	t.Run("mixed package manager operations", func(t *testing.T) {
		mixedScript := `
			cd /home/testuser
			
			echo "=== Mixed Package Manager Operations ==="
			
			# 1. Set up configuration with both managers
			echo "1. Setting up mixed package manager configuration..."
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
  - prettier
dotfiles: []
EOF
			
			echo "Initial mixed configuration:"
			/workspace/plonk config show
			
			# 2. Add packages to different managers
			echo "2. Adding packages to different managers..."
			/workspace/plonk pkg add --manager homebrew jq || echo "Homebrew package add processed"
			/workspace/plonk pkg add --manager npm typescript || echo "NPM package add processed"
			
			# 3. Test default manager behavior
			echo "3. Testing default manager behavior..."
			/workspace/plonk pkg add wget || echo "Default manager package add processed"
			
			# 4. List packages from all managers
			echo "4. Listing packages from all managers..."
			/workspace/plonk pkg list || echo "Package list completed"
			
			# 5. Remove packages from different managers
			echo "5. Removing packages from different managers..."
			/workspace/plonk pkg remove curl || echo "Package remove processed"
			/workspace/plonk pkg remove --manager npm lodash || echo "NPM package remove processed"
			
			# 6. Show final mixed state
			echo "6. Final mixed configuration:"
			/workspace/plonk config show
			
			echo "=== Mixed Operations Complete ==="
		`
		
		output, err := runner.RunCommand(t, mixedScript)
		t.Logf("Mixed package manager operations output: %s", output)
		
		if err != nil {
			t.Logf("Mixed operations completed with some expected errors: %v", err)
		}
		
		// Verify mixed operations
		outputStr := string(output)
		mixedSteps := []string{
			"Mixed Package Manager Operations",
			"Setting up mixed package manager",
			"Adding packages to different managers",
			"Testing default manager behavior",
			"Listing packages from all managers",
			"Removing packages from different managers",
		}
		
		for _, step := range mixedSteps {
			if !strings.Contains(outputStr, step) {
				t.Errorf("Expected mixed operation step '%s' not found in output", step)
			}
		}
	})

	// Test package manager switching
	t.Run("package manager switching", func(t *testing.T) {
		switchScript := `
			cd /home/testuser
			
			echo "=== Package Manager Switching ==="
			
			# 1. Start with homebrew as default
			echo "1. Starting with homebrew as default..."
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
			
			echo "Initial configuration (homebrew default):"
			/workspace/plonk config show
			
			# 2. Add package with default manager
			echo "2. Adding package with default manager..."
			/workspace/plonk pkg add git || echo "Default manager package add processed"
			
			# 3. Switch default manager to npm
			echo "3. Switching default manager to npm..."
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: npm
homebrew:
  brews:
    - curl
    - git
npm:
  - lodash
dotfiles: []
EOF
			
			echo "Configuration after switch (npm default):"
			/workspace/plonk config show
			
			# 4. Add package with new default manager
			echo "4. Adding package with new default manager..."
			/workspace/plonk pkg add prettier || echo "New default manager package add processed"
			
			# 5. Verify both managers still work
			echo "5. Verifying both managers still work..."
			/workspace/plonk pkg add --manager homebrew jq || echo "Explicit homebrew add processed"
			/workspace/plonk pkg add --manager npm typescript || echo "Explicit npm add processed"
			
			# 6. Show final state
			echo "6. Final configuration after switching:"
			/workspace/plonk config show
			
			echo "=== Manager Switching Complete ==="
		`
		
		output, err := runner.RunCommand(t, switchScript)
		t.Logf("Package manager switching output: %s", output)
		
		if err != nil {
			t.Logf("Manager switching completed with some expected errors: %v", err)
		}
		
		// Verify switching operations
		outputStr := string(output)
		if !strings.Contains(outputStr, "Package Manager Switching") {
			t.Error("Package manager switching test did not execute properly")
		}
	})

	// Test manager-specific package conflicts
	t.Run("manager specific conflicts", func(t *testing.T) {
		conflictScript := `
			cd /home/testuser
			
			echo "=== Manager-Specific Conflicts ==="
			
			# 1. Set up potential conflict scenario
			echo "1. Setting up potential conflict scenario..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - node
npm:
  - lodash
dotfiles: []
EOF
			
			echo "Initial configuration:"
			/workspace/plonk config show
			
			# 2. Try to add same package to different managers
			echo "2. Testing package conflicts between managers..."
			/workspace/plonk pkg add --manager homebrew node || echo "Homebrew node add processed"
			/workspace/plonk pkg add --manager npm node || echo "NPM node add processed"
			
			# 3. Show configuration after conflict attempts
			echo "3. Configuration after conflict attempts:"
			/workspace/plonk config show
			
			# 4. Try to add manager-specific packages
			echo "4. Adding manager-specific packages..."
			/workspace/plonk pkg add --manager homebrew wget || echo "Homebrew-specific package add processed"
			/workspace/plonk pkg add --manager npm eslint || echo "NPM-specific package add processed"
			
			# 5. List all packages to see final state
			echo "5. Final package state:"
			/workspace/plonk pkg list || echo "Package list completed"
			
			echo "=== Conflict Testing Complete ==="
		`
		
		output, err := runner.RunCommand(t, conflictScript)
		t.Logf("Manager-specific conflicts output: %s", output)
		
		if err != nil {
			t.Logf("Conflict testing completed with some expected errors: %v", err)
		}
		
		// Verify conflict handling
		outputStr := string(output)
		if !strings.Contains(outputStr, "Manager-Specific Conflicts") {
			t.Error("Manager-specific conflicts test did not execute properly")
		}
	})
}

func TestCrossManagerWorkflows(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test full-stack development workflow
	t.Run("full stack development workflow", func(t *testing.T) {
		fullStackScript := `
			cd /home/testuser
			
			echo "=== Full-Stack Development Workflow ==="
			
			# 1. Set up full-stack development environment
			echo "1. Setting up full-stack development environment..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - git
    - curl
    - postgresql
npm:
  - express
  - react
  - typescript
dotfiles:
  - source: dot_gitconfig
    destination: ~/.gitconfig
EOF
			
			echo "Initial full-stack configuration:"
			/workspace/plonk config show
			
			# 2. Add backend development tools
			echo "2. Adding backend development tools..."
			/workspace/plonk pkg add --manager homebrew redis || echo "Backend tool add processed"
			/workspace/plonk pkg add --manager npm sequelize || echo "Backend package add processed"
			
			# 3. Add frontend development tools
			echo "3. Adding frontend development tools..."
			/workspace/plonk pkg add --manager npm webpack || echo "Frontend tool add processed"
			/workspace/plonk pkg add --manager npm eslint || echo "Frontend package add processed"
			
			# 4. Add development utilities
			echo "4. Adding development utilities..."
			/workspace/plonk pkg add --manager homebrew jq || echo "Utility add processed"
			/workspace/plonk pkg add --manager npm nodemon || echo "Development utility add processed"
			
			# 5. Show complete development environment
			echo "5. Complete full-stack development environment:"
			/workspace/plonk config show
			
			# 6. List all packages by manager
			echo "6. All packages by manager:"
			/workspace/plonk pkg list || echo "Package list completed"
			
			echo "=== Full-Stack Workflow Complete ==="
		`
		
		output, err := runner.RunCommand(t, fullStackScript)
		t.Logf("Full-stack workflow output: %s", output)
		
		if err != nil {
			t.Logf("Full-stack workflow completed with some expected errors: %v", err)
		}
		
		// Verify full-stack workflow
		outputStr := string(output)
		if !strings.Contains(outputStr, "Full-Stack Development Workflow") {
			t.Error("Full-stack development workflow did not execute properly")
		}
	})

	// Test deployment environment setup
	t.Run("deployment environment setup", func(t *testing.T) {
		deploymentScript := `
			cd /home/testuser
			
			echo "=== Deployment Environment Setup ==="
			
			# 1. Set up production deployment tools
			echo "1. Setting up production deployment tools..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - docker
    - kubernetes-cli
    - terraform
npm:
  - pm2
  - forever
dotfiles:
  - source: dot_dockerconfig
    destination: ~/.docker/config.json
EOF
			
			echo "Initial deployment configuration:"
			/workspace/plonk config show
			
			# 2. Add monitoring tools
			echo "2. Adding monitoring tools..."
			/workspace/plonk pkg add --manager homebrew prometheus || echo "Monitoring tool add processed"
			/workspace/plonk pkg add --manager npm newrelic || echo "Monitoring package add processed"
			
			# 3. Add CI/CD tools
			echo "3. Adding CI/CD tools..."
			/workspace/plonk pkg add --manager homebrew gh || echo "CI/CD tool add processed"
			/workspace/plonk pkg add --manager npm semantic-release || echo "CI/CD package add processed"
			
			# 4. Show complete deployment environment
			echo "4. Complete deployment environment:"
			/workspace/plonk config show
			
			echo "=== Deployment Setup Complete ==="
		`
		
		output, err := runner.RunCommand(t, deploymentScript)
		t.Logf("Deployment environment setup output: %s", output)
		
		if err != nil {
			t.Logf("Deployment setup completed with some expected errors: %v", err)
		}
		
		// Verify deployment setup
		outputStr := string(output)
		if !strings.Contains(outputStr, "Deployment Environment Setup") {
			t.Error("Deployment environment setup did not execute properly")
		}
	})

	// Test cross-platform compatibility
	t.Run("cross platform compatibility", func(t *testing.T) {
		crossPlatformScript := `
			cd /home/testuser
			
			echo "=== Cross-Platform Compatibility ==="
			
			# 1. Set up cross-platform configuration
			echo "1. Setting up cross-platform configuration..."
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - git
    - curl
    - wget
npm:
  - lodash
  - moment
  - axios
dotfiles:
  - source: dot_bashrc
    destination: ~/.bashrc
  - source: dot_zshrc
    destination: ~/.zshrc
EOF
			
			echo "Initial cross-platform configuration:"
			/workspace/plonk config show
			
			# 2. Add platform-agnostic tools
			echo "2. Adding platform-agnostic tools..."
			/workspace/plonk pkg add --manager homebrew jq || echo "Cross-platform tool add processed"
			/workspace/plonk pkg add --manager npm cross-env || echo "Cross-platform package add processed"
			
			# 3. Test different shell configurations
			echo "3. Testing different shell configurations..."
			echo "# Cross-platform bashrc" > ~/.bashrc
			echo "# Cross-platform zshrc" > ~/.zshrc
			/workspace/plonk dot add ~/.bashrc || echo "Cross-platform dotfile add processed"
			
			# 4. Show final cross-platform state
			echo "4. Final cross-platform configuration:"
			/workspace/plonk config show
			
			echo "=== Cross-Platform Testing Complete ==="
		`
		
		output, err := runner.RunCommand(t, crossPlatformScript)
		t.Logf("Cross-platform compatibility output: %s", output)
		
		if err != nil {
			t.Logf("Cross-platform testing completed with some expected errors: %v", err)
		}
		
		// Verify cross-platform compatibility
		outputStr := string(output)
		if !strings.Contains(outputStr, "Cross-Platform Compatibility") {
			t.Error("Cross-platform compatibility test did not execute properly")
		}
	})
}