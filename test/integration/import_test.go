// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestImportCommand(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test import discovery with realistic package environment
	t.Run("import discovery", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import Discovery ==="
			
			# Create plonk config directory with empty config
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install some packages via homebrew
			/home/linuxbrew/.linuxbrew/bin/brew install jq tree wget curl
			
			# Install some npm packages
			npm install -g lodash prettier typescript
			
			# Create test dotfiles
			echo "# Test gitconfig" > ~/.gitconfig
			echo "# Test zshrc" > ~/.zshrc
			echo "# Test vimrc" > ~/.vimrc
			echo "# Test editorconfig" > ~/.editorconfig
			
			# Create config directory structure
			mkdir -p ~/.config/nvim
			echo "-- nvim config" > ~/.config/nvim/init.lua
			
			# Create some system files that should be filtered
			touch ~/.DS_Store
			touch ~/.cache
			mkdir -p ~/.npm
			touch ~/.npm/test
			
			# Run import discovery (non-interactive - should show what would be imported)
			echo "=== Running import discovery ==="
			echo "" | /workspace/plonk import --packages 2>&1 | head -50
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import discovery output: %s", output)
		
		if err != nil {
			t.Fatalf("Import discovery test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Verify packages are discovered
		if !strings.Contains(outputStr, "jq") {
			t.Error("Expected jq to be discovered as untracked package")
		}
		
		if !strings.Contains(outputStr, "tree") {
			t.Error("Expected tree to be discovered as untracked package")
		}
		
		if !strings.Contains(outputStr, "wget") {
			t.Error("Expected wget to be discovered as untracked package")
		}
		
		// Verify system packages are filtered out
		if strings.Contains(outputStr, "ca-certificates") {
			t.Error("System package ca-certificates should be filtered out")
		}
		
		if strings.Contains(outputStr, "openssl") {
			t.Error("System package openssl should be filtered out")
		}
	})

	// Test import with dotfiles
	t.Run("import dotfiles discovery", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Dotfile Import Discovery ==="
			
			# Create plonk config directory with empty config
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Create test dotfiles
			echo "# Test gitconfig" > ~/.gitconfig
			echo "# Test zshrc" > ~/.zshrc
			echo "# Test vimrc" > ~/.vimrc
			echo "# Test editorconfig" > ~/.editorconfig
			
			# Create config directory structure
			mkdir -p ~/.config/nvim/lua/config
			echo "-- nvim init" > ~/.config/nvim/init.lua
			echo "-- nvim config" > ~/.config/nvim/lua/config/options.lua
			
			# Create some system files that should be filtered
			touch ~/.DS_Store
			touch ~/.cache
			mkdir -p ~/.npm
			touch ~/.npm/test
			touch ~/.bash_history
			
			# Run import discovery for dotfiles
			echo "=== Running dotfile import discovery ==="
			echo "" | /workspace/plonk import --dotfiles 2>&1 | head -50
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Dotfile import discovery output: %s", output)
		
		if err != nil {
			t.Fatalf("Dotfile import discovery test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Verify dotfiles are discovered
		if !strings.Contains(outputStr, ".gitconfig") {
			t.Error("Expected .gitconfig to be discovered as untracked dotfile")
		}
		
		if !strings.Contains(outputStr, ".zshrc") {
			t.Error("Expected .zshrc to be discovered as untracked dotfile")
		}
		
		if !strings.Contains(outputStr, ".vimrc") {
			t.Error("Expected .vimrc to be discovered as untracked dotfile")
		}
		
		// Verify system dotfiles are filtered out
		if strings.Contains(outputStr, ".DS_Store") {
			t.Error("System dotfile .DS_Store should be filtered out")
		}
		
		if strings.Contains(outputStr, ".cache") {
			t.Error("System dotfile .cache should be filtered out")
		}
		
		if strings.Contains(outputStr, ".bash_history") {
			t.Error("System dotfile .bash_history should be filtered out")
		}
	})

	// Test import --all mode
	t.Run("import all mode", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import All Mode ==="
			
			# Create plonk config directory with empty config
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install minimal packages
			/home/linuxbrew/.linuxbrew/bin/brew install jq tree
			
			# Create test dotfiles
			echo "# Test gitconfig" > ~/.gitconfig
			echo "# Test zshrc" > ~/.zshrc
			
			# Run import --all (should import everything without interaction)
			echo "=== Running import --all ==="
			/workspace/plonk import --all 2>&1 | head -30
			
			echo -e "\n=== Checking config after import ==="
			cat ~/.config/plonk/plonk.yaml
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import all mode output: %s", output)
		
		if err != nil {
			t.Fatalf("Import all mode test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Verify import happened
		if !strings.Contains(outputStr, "Successfully imported") {
			t.Error("Expected successful import message")
		}
		
		// Verify packages were added to config
		if !strings.Contains(outputStr, "jq") {
			t.Error("Expected jq to be in imported packages")
		}
		
		if !strings.Contains(outputStr, "tree") {
			t.Error("Expected tree to be in imported packages")
		}
		
		// Verify dotfiles were added to config
		if !strings.Contains(outputStr, ".gitconfig") {
			t.Error("Expected .gitconfig to be in imported dotfiles")
		}
		
		if !strings.Contains(outputStr, ".zshrc") {
			t.Error("Expected .zshrc to be in imported dotfiles")
		}
		
		// Verify config file was updated
		if !strings.Contains(outputStr, "brews:") {
			t.Error("Expected brews section in config")
		}
		
		if !strings.Contains(outputStr, "dotfiles:") {
			t.Error("Expected dotfiles section in config")
		}
	})

	// Test import with existing config
	t.Run("import with existing config", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import with Existing Config ==="
			
			# Create plonk config directory with some existing items
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - git
    - curl
dotfiles:
  - source: gitconfig
    destination: ~/.gitconfig
EOF
			
			# Install packages (some already managed, some not)
			/home/linuxbrew/.linuxbrew/bin/brew install git curl jq tree
			
			# Create dotfiles (some already managed, some not)
			echo "# Test gitconfig" > ~/.gitconfig
			echo "# Test zshrc" > ~/.zshrc
			
			# Run import discovery
			echo "=== Running import with existing config ==="
			echo "" | /workspace/plonk import --packages 2>&1 | head -30
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import with existing config output: %s", output)
		
		if err != nil {
			t.Fatalf("Import with existing config test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Verify only untracked items are shown
		if !strings.Contains(outputStr, "jq") {
			t.Error("Expected jq to be shown as untracked")
		}
		
		if !strings.Contains(outputStr, "tree") {
			t.Error("Expected tree to be shown as untracked")
		}
		
		// Verify already managed items are not shown
		if strings.Contains(outputStr, "git") {
			t.Error("Already managed package git should not be shown")
		}
		
		if strings.Contains(outputStr, "curl") {
			t.Error("Already managed package curl should not be shown")
		}
	})

	// Test import error handling
	t.Run("import error handling", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import Error Handling ==="
			
			# Test with no config directory
			rm -rf ~/.config/plonk
			
			# Run import (should create new config)
			echo "=== Running import without config ==="
			echo "" | /workspace/plonk import --packages 2>&1 | head -20
			
			# Test with read-only config directory
			mkdir -p ~/.config/plonk
			chmod 444 ~/.config/plonk
			
			echo "=== Running import with read-only config ==="
			echo "" | /workspace/plonk import --packages 2>&1 | head -20 || echo "Expected error occurred"
			
			# Restore permissions
			chmod 755 ~/.config/plonk
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import error handling output: %s", output)
		
		// Error handling test - some errors are expected
		if err != nil {
			t.Logf("Import error handling test completed with expected errors: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should handle missing config gracefully
		if strings.Contains(outputStr, "Welcome to plonk import") {
			t.Log("Successfully handled missing config by creating new one")
		}
	})

	// Test import with no untracked items
	t.Run("import with no untracked items", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import with No Untracked Items ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - git
    - curl
dotfiles:
  - source: gitconfig
    destination: ~/.gitconfig
EOF
			
			# Install only packages that are already managed
			/home/linuxbrew/.linuxbrew/bin/brew install git curl
			
			# Create only dotfiles that are already managed
			echo "# Test gitconfig" > ~/.gitconfig
			
			# Run import
			echo "=== Running import with no untracked items ==="
			echo "" | /workspace/plonk import 2>&1
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import with no untracked items output: %s", output)
		
		if err != nil {
			t.Fatalf("Import with no untracked items test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show success message
		if !strings.Contains(outputStr, "No untracked items found") {
			t.Error("Expected 'No untracked items found' message")
		}
	})
}