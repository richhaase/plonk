// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestImportModes(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test packages-only mode
	t.Run("packages only mode", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Packages Only Mode ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install packages and create dotfiles
			/home/linuxbrew/.linuxbrew/bin/brew install jq tree
			npm install -g lodash prettier
			echo "# Test gitconfig" > ~/.gitconfig
			echo "# Test zshrc" > ~/.zshrc
			
			# Run import --packages (should only show packages)
			echo "=== Running import --packages ==="
			echo "" | /workspace/plonk import --packages 2>&1 | head -30
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Packages only mode output: %s", output)
		
		if err != nil {
			t.Fatalf("Packages only mode test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show packages
		if !strings.Contains(outputStr, "📦 Found") {
			t.Error("Expected packages to be found")
		}
		
		// Should NOT show dotfiles section
		if strings.Contains(outputStr, "📄 Found") {
			t.Error("Should not show dotfiles in packages-only mode")
		}
		
		// Should show some expected packages
		if !strings.Contains(outputStr, "jq") {
			t.Error("Expected jq package to be shown")
		}
		
		if !strings.Contains(outputStr, "tree") {
			t.Error("Expected tree package to be shown")
		}
	})

	// Test dotfiles-only mode
	t.Run("dotfiles only mode", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Dotfiles Only Mode ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install packages and create dotfiles
			/home/linuxbrew/.linuxbrew/bin/brew install jq tree
			echo "# Test gitconfig" > ~/.gitconfig
			echo "# Test zshrc" > ~/.zshrc
			echo "# Test vimrc" > ~/.vimrc
			
			# Create config directory
			mkdir -p ~/.config/nvim
			echo "-- nvim config" > ~/.config/nvim/init.lua
			
			# Run import --dotfiles (should only show dotfiles)
			echo "=== Running import --dotfiles ==="
			echo "" | /workspace/plonk import --dotfiles 2>&1 | head -30
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Dotfiles only mode output: %s", output)
		
		if err != nil {
			t.Fatalf("Dotfiles only mode test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show dotfiles
		if !strings.Contains(outputStr, "📄 Found") {
			t.Error("Expected dotfiles to be found")
		}
		
		// Should NOT show packages section
		if strings.Contains(outputStr, "📦 Found") {
			t.Error("Should not show packages in dotfiles-only mode")
		}
		
		// Should show some expected dotfiles
		if !strings.Contains(outputStr, ".gitconfig") {
			t.Error("Expected .gitconfig dotfile to be shown")
		}
		
		if !strings.Contains(outputStr, ".zshrc") {
			t.Error("Expected .zshrc dotfile to be shown")
		}
		
		if !strings.Contains(outputStr, ".vimrc") {
			t.Error("Expected .vimrc dotfile to be shown")
		}
	})

	// Test mixed mode (default)
	t.Run("mixed mode default", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Mixed Mode (Default) ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install packages and create dotfiles
			/home/linuxbrew/.linuxbrew/bin/brew install jq tree
			npm install -g lodash
			echo "# Test gitconfig" > ~/.gitconfig
			echo "# Test zshrc" > ~/.zshrc
			
			# Run import (default mode - should show both)
			echo "=== Running import (default mixed mode) ==="
			echo "" | /workspace/plonk import 2>&1 | head -40
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Mixed mode output: %s", output)
		
		if err != nil {
			t.Fatalf("Mixed mode test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show both packages and dotfiles
		if !strings.Contains(outputStr, "📦 Found") {
			t.Error("Expected packages to be found in mixed mode")
		}
		
		if !strings.Contains(outputStr, "📄 Found") {
			t.Error("Expected dotfiles to be found in mixed mode")
		}
		
		// Should show some expected packages
		if !strings.Contains(outputStr, "jq") {
			t.Error("Expected jq package to be shown in mixed mode")
		}
		
		// Should show some expected dotfiles
		if !strings.Contains(outputStr, ".gitconfig") {
			t.Error("Expected .gitconfig dotfile to be shown in mixed mode")
		}
	})

	// Test all mode with confirmation
	t.Run("all mode behavior", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing All Mode Behavior ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install a few packages and create dotfiles
			/home/linuxbrew/.linuxbrew/bin/brew install jq tree
			echo "# Test gitconfig" > ~/.gitconfig
			echo "# Test zshrc" > ~/.zshrc
			
			# Run import --all (should show what will be imported, then import)
			echo "=== Running import --all ==="
			/workspace/plonk import --all 2>&1 | head -40
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("All mode behavior output: %s", output)
		
		if err != nil {
			t.Fatalf("All mode behavior test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show what will be imported
		if !strings.Contains(outputStr, "Importing all") {
			t.Error("Expected 'Importing all' message in --all mode")
		}
		
		// Should show packages and dotfiles being imported
		if !strings.Contains(outputStr, "📦") {
			t.Error("Expected package symbols in --all mode")
		}
		
		if !strings.Contains(outputStr, "📄") {
			t.Error("Expected dotfile symbols in --all mode")
		}
		
		// Should show success message
		if !strings.Contains(outputStr, "Successfully imported") {
			t.Error("Expected success message in --all mode")
		}
	})

	// Test output formatting
	t.Run("import with different output formats", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import Output Formats ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install packages and create dotfiles
			/home/linuxbrew/.linuxbrew/bin/brew install jq
			echo "# Test gitconfig" > ~/.gitconfig
			
			# Test JSON output (this would be non-interactive discovery)
			echo "=== Testing JSON output ==="
			echo "" | /workspace/plonk import --packages -o json 2>&1 | head -20 || echo "JSON test completed"
			
			# Test YAML output
			echo "=== Testing YAML output ==="
			echo "" | /workspace/plonk import --packages -o yaml 2>&1 | head -20 || echo "YAML test completed"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import output formats output: %s", output)
		
		if err != nil {
			t.Fatalf("Import output formats test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should contain some structured output indicators
		if strings.Contains(outputStr, "JSON test completed") {
			t.Log("JSON output test completed")
		}
		
		if strings.Contains(outputStr, "YAML test completed") {
			t.Log("YAML output test completed")
		}
	})

	// Test import with complex directory structure
	t.Run("import complex directory structure", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import with Complex Directory Structure ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Create complex dotfile structure
			mkdir -p ~/.config/nvim/lua/config
			mkdir -p ~/.config/nvim/lua/plugins
			mkdir -p ~/.config/alacritty
			
			echo "-- nvim init" > ~/.config/nvim/init.lua
			echo "-- nvim options" > ~/.config/nvim/lua/config/options.lua
			echo "-- nvim keymaps" > ~/.config/nvim/lua/config/keymaps.lua
			echo "-- nvim plugins" > ~/.config/nvim/lua/plugins/editor.lua
			
			echo "# alacritty config" > ~/.config/alacritty/alacritty.yml
			
			# Regular dotfiles
			echo "# Test gitconfig" > ~/.gitconfig
			echo "# Test zshrc" > ~/.zshrc
			
			# Run import --dotfiles
			echo "=== Running import with complex structure ==="
			echo "" | /workspace/plonk import --dotfiles 2>&1 | head -50
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Complex directory structure output: %s", output)
		
		if err != nil {
			t.Fatalf("Complex directory structure test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show directory-based dotfiles
		if !strings.Contains(outputStr, ".config") {
			t.Error("Expected .config directory files to be shown")
		}
		
		// Should show regular dotfiles
		if !strings.Contains(outputStr, ".gitconfig") {
			t.Error("Expected .gitconfig to be shown")
		}
		
		if !strings.Contains(outputStr, ".zshrc") {
			t.Error("Expected .zshrc to be shown")
		}
		
		// Should show some nvim files
		if !strings.Contains(outputStr, "nvim") {
			t.Error("Expected nvim files to be shown")
		}
	})
}