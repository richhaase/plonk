// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestImportLargeDataset(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test with many packages (similar to laptop with 144 packages)
	t.Run("import many packages", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import with Many Packages ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install many packages via homebrew (subset of realistic packages)
			echo "Installing homebrew packages..."
			/home/linuxbrew/.linuxbrew/bin/brew install \
				jq tree wget curl git \
				neovim tmux htop fzf \
				bat ripgrep fd \
				docker kubernetes-cli \
				python3 node \
				go rust
			
			# Install npm packages
			echo "Installing npm packages..."
			npm install -g \
				lodash prettier typescript \
				eslint nodemon \
				express react-scripts \
				@angular/cli
			
			# Run import --packages to see discovery
			echo "=== Running import --packages discovery ==="
			echo "" | timeout 30 /workspace/plonk import --packages 2>&1 | head -100
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import many packages output: %s", output)
		
		if err != nil {
			t.Fatalf("Import many packages test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show packages found
		if !strings.Contains(outputStr, "📦 Found") {
			t.Error("Expected packages to be found")
		}
		
		// Should show some expected packages
		expectedPackages := []string{"jq", "tree", "wget", "curl", "neovim", "tmux", "htop", "fzf", "bat", "ripgrep"}
		foundPackages := 0
		for _, pkg := range expectedPackages {
			if strings.Contains(outputStr, pkg) {
				foundPackages++
			}
		}
		
		if foundPackages < 5 {
			t.Errorf("Expected at least 5 packages to be found, got %d", foundPackages)
		}
		
		// Should show system packages are filtered
		systemPackages := []string{"ca-certificates", "openssl", "sqlite", "zlib", "readline"}
		foundSystemPackages := 0
		for _, pkg := range systemPackages {
			if strings.Contains(outputStr, pkg) {
				foundSystemPackages++
			}
		}
		
		if foundSystemPackages > 2 {
			t.Errorf("Too many system packages found (%d), filtering may not be working", foundSystemPackages)
		}
	})

	// Test with many dotfiles (similar to laptop with 31 dotfiles)
	t.Run("import many dotfiles", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import with Many Dotfiles ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Create many dotfiles (realistic mix)
			echo "Creating dotfiles..."
			
			# Shell configs
			echo "# zshrc" > ~/.zshrc
			echo "# bashrc" > ~/.bashrc
			echo "# bash_profile" > ~/.bash_profile
			echo "# profile" > ~/.profile
			
			# Git configs
			echo "# gitconfig" > ~/.gitconfig
			echo "# gitignore_global" > ~/.gitignore_global
			
			# Editor configs
			echo "# vimrc" > ~/.vimrc
			echo "# editorconfig" > ~/.editorconfig
			
			# Development tools
			echo "# tool-versions" > ~/.tool-versions
			echo "# asdf" > ~/.asdf
			
			# Application configs
			mkdir -p ~/.config/nvim/lua/config
			mkdir -p ~/.config/nvim/lua/plugins
			mkdir -p ~/.config/alacritty
			mkdir -p ~/.config/tmux
			mkdir -p ~/.config/git
			
			echo "-- nvim init" > ~/.config/nvim/init.lua
			echo "-- nvim options" > ~/.config/nvim/lua/config/options.lua
			echo "-- nvim keymaps" > ~/.config/nvim/lua/config/keymaps.lua
			echo "-- nvim lazy" > ~/.config/nvim/lua/config/lazy.lua
			echo "-- nvim plugins" > ~/.config/nvim/lua/plugins/editor.lua
			echo "-- nvim plugins" > ~/.config/nvim/lua/plugins/lsp.lua
			
			echo "# alacritty config" > ~/.config/alacritty/alacritty.yml
			echo "# tmux config" > ~/.config/tmux/tmux.conf
			echo "# git config" > ~/.config/git/config
			
			# SSH and security
			mkdir -p ~/.ssh
			echo "# ssh config" > ~/.ssh/config
			
			# Other dotfiles
			echo "# gemrc" > ~/.gemrc
			echo "# npmrc" > ~/.npmrc
			echo "# curlrc" > ~/.curlrc
			
			# Create system files that should be filtered
			touch ~/.DS_Store
			touch ~/.bash_history
			touch ~/.zsh_history
			mkdir -p ~/.cache
			touch ~/.cache/test
			mkdir -p ~/.npm
			touch ~/.npm/test
			
			# Run import --dotfiles
			echo "=== Running import --dotfiles discovery ==="
			echo "" | timeout 30 /workspace/plonk import --dotfiles 2>&1 | head -100
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import many dotfiles output: %s", output)
		
		if err != nil {
			t.Fatalf("Import many dotfiles test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show dotfiles found
		if !strings.Contains(outputStr, "📄 Found") {
			t.Error("Expected dotfiles to be found")
		}
		
		// Should show some expected dotfiles
		expectedDotfiles := []string{".zshrc", ".bashrc", ".gitconfig", ".vimrc", ".editorconfig", ".tool-versions"}
		foundDotfiles := 0
		for _, dotfile := range expectedDotfiles {
			if strings.Contains(outputStr, dotfile) {
				foundDotfiles++
			}
		}
		
		if foundDotfiles < 4 {
			t.Errorf("Expected at least 4 dotfiles to be found, got %d", foundDotfiles)
		}
		
		// Should show config directory files
		if !strings.Contains(outputStr, ".config") {
			t.Error("Expected .config directory files to be shown")
		}
		
		// Should show nvim files
		if !strings.Contains(outputStr, "nvim") {
			t.Error("Expected nvim files to be shown")
		}
		
		// Should filter out system files
		systemFiles := []string{".DS_Store", ".bash_history", ".zsh_history", ".cache"}
		foundSystemFiles := 0
		for _, file := range systemFiles {
			if strings.Contains(outputStr, file) {
				foundSystemFiles++
			}
		}
		
		if foundSystemFiles > 1 {
			t.Errorf("Too many system files found (%d), filtering may not be working", foundSystemFiles)
		}
	})

	// Test import performance with large dataset
	t.Run("import performance test", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import Performance ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install a reasonable number of packages
			echo "Installing packages for performance test..."
			/home/linuxbrew/.linuxbrew/bin/brew install \
				jq tree wget curl git neovim tmux htop fzf bat ripgrep fd
			
			npm install -g lodash prettier typescript eslint nodemon
			
			# Create many dotfiles
			echo "Creating dotfiles for performance test..."
			for i in {1..20}; do
				echo "# test file $i" > ~/.test_dotfile_$i
			done
			
			# Create config directories
			mkdir -p ~/.config/nvim/lua/config
			mkdir -p ~/.config/nvim/lua/plugins
			for i in {1..10}; do
				echo "-- nvim config $i" > ~/.config/nvim/lua/config/config_$i.lua
				echo "-- nvim plugin $i" > ~/.config/nvim/lua/plugins/plugin_$i.lua
			done
			
			# Time the import discovery
			echo "=== Running timed import discovery ==="
			time (echo "" | timeout 30 /workspace/plonk import 2>&1 | head -50)
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import performance test output: %s", output)
		
		if err != nil {
			t.Fatalf("Import performance test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should complete in reasonable time
		if strings.Contains(outputStr, "Welcome to plonk import") {
			t.Log("Import discovery completed successfully")
		}
		
		// Should find both packages and dotfiles
		if !strings.Contains(outputStr, "📦 Found") {
			t.Error("Expected packages to be found in performance test")
		}
		
		if !strings.Contains(outputStr, "📄 Found") {
			t.Error("Expected dotfiles to be found in performance test")
		}
		
		// Should show reasonable numbers
		if strings.Contains(outputStr, "Found 0") {
			t.Error("Performance test should find some items")
		}
	})

	// Test import --all with large dataset
	t.Run("import all large dataset", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Import All with Large Dataset ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
packages: []
dotfiles: []
EOF
			
			# Install a smaller set for actual import test
			echo "Installing packages for import all test..."
			/home/linuxbrew/.linuxbrew/bin/brew install jq tree wget
			npm install -g lodash prettier
			
			# Create dotfiles
			echo "Creating dotfiles for import all test..."
			echo "# gitconfig" > ~/.gitconfig
			echo "# zshrc" > ~/.zshrc
			echo "# vimrc" > ~/.vimrc
			
			mkdir -p ~/.config/nvim
			echo "-- nvim config" > ~/.config/nvim/init.lua
			
			# Run import --all
			echo "=== Running import --all with large dataset ==="
			timeout 60 /workspace/plonk import --all 2>&1 | head -100
			
			echo -e "\n=== Checking final config ==="
			cat ~/.config/plonk/plonk.yaml | head -50
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Import all large dataset output: %s", output)
		
		if err != nil {
			t.Fatalf("Import all large dataset test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show import summary
		if !strings.Contains(outputStr, "Importing all") {
			t.Error("Expected 'Importing all' message")
		}
		
		// Should show successful import
		if !strings.Contains(outputStr, "Successfully imported") {
			t.Error("Expected successful import message")
		}
		
		// Should show config was updated
		if !strings.Contains(outputStr, "brews:") {
			t.Error("Expected brews section in final config")
		}
		
		if !strings.Contains(outputStr, "dotfiles:") {
			t.Error("Expected dotfiles section in final config")
		}
	})
}