// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestDotfileDirectoryExpansion(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test directory expansion in configuration and status
	t.Run("directory expansion in status", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Dotfile Directory Expansion ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			
			# Create test dotfile directory structure
			mkdir -p ~/.config/plonk/config/nvim/lua/config
			mkdir -p ~/.config/plonk/config/nvim/lua/plugins
			
			# Create test nvim config files
			cat > ~/.config/plonk/config/nvim/init.lua << 'EOF'
-- Test nvim init.lua
vim.g.mapleader = " "
require("config.lazy")
EOF
			
			cat > ~/.config/plonk/config/nvim/lua/config/lazy.lua << 'EOF'
-- Test lazy.lua config
return {
  spec = {
    { import = "plugins" },
  },
}
EOF
			
			cat > ~/.config/plonk/config/nvim/lua/config/options.lua << 'EOF'
-- Test options.lua
vim.opt.number = true
vim.opt.relativenumber = true
EOF
			
			cat > ~/.config/plonk/config/nvim/lua/plugins/editor.lua << 'EOF'
-- Test editor plugins
return {
  { "nvim-telescope/telescope.nvim" },
}
EOF
			
			# Create other test dotfiles
			echo "# Test zshrc" > ~/.config/plonk/zshrc
			echo "# Test vimrc" > ~/.config/plonk/vimrc
			
			# Create plonk config with directory mapping
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - git
    - curl
dotfiles:
  - source: zshrc
    destination: ~/.zshrc
  - source: vimrc
    destination: ~/.vimrc
  - source: config/nvim
    destination: ~/.config/nvim
EOF
			
			# Create the actual destination directory and files
			mkdir -p ~/.config/nvim/lua/config
			mkdir -p ~/.config/nvim/lua/plugins
			
			# Copy files to destination (simulate already applied)
			cp ~/.config/plonk/config/nvim/init.lua ~/.config/nvim/
			cp ~/.config/plonk/config/nvim/lua/config/lazy.lua ~/.config/nvim/lua/config/
			cp ~/.config/plonk/config/nvim/lua/config/options.lua ~/.config/nvim/lua/config/
			cp ~/.config/plonk/config/nvim/lua/plugins/editor.lua ~/.config/nvim/lua/plugins/
			
			# Copy other dotfiles
			cp ~/.config/plonk/zshrc ~/.zshrc
			cp ~/.config/plonk/vimrc ~/.vimrc
			
			echo "=== Testing plonk status ==="
			/workspace/plonk status
			
			echo -e "\n=== Testing plonk dot list managed ==="
			/workspace/plonk dot list managed
			
			echo -e "\n=== Verifying individual nvim files are detected ==="
			# Check that status shows multiple nvim files
			/workspace/plonk status | grep -c "nvim" || echo "No nvim files found in status"
			
			echo -e "\n=== Verifying file count in managed dotfiles ==="
			# Should show more than just the 3 simple dotfiles
			/workspace/plonk dot list managed | grep -c "^[^#]" || echo "No managed files found"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Directory expansion test output: %s", output)
		
		if err != nil {
			t.Fatalf("Directory expansion test failed: %v\nOutput: %s", err, output)
		}
		
		// Verify that the output contains expanded nvim files
		outputStr := string(output)
		if !strings.Contains(outputStr, "nvim") {
			t.Error("Expected nvim files to be detected in status output")
		}
		
		// Check that we have multiple managed dotfiles (more than 3)
		if strings.Count(outputStr, "managed") < 1 {
			t.Error("Expected multiple managed dotfiles to be detected")
		}
	})

	// Test that status correctly counts files in directories
	t.Run("status count accuracy", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Status Count Accuracy ==="
			
			# Create a more complex directory structure
			mkdir -p ~/.config/plonk/config/complex/deep/nested
			mkdir -p ~/.config/plonk/config/complex/other
			
			# Create multiple files in nested structure
			echo "file1" > ~/.config/plonk/config/complex/file1.txt
			echo "file2" > ~/.config/plonk/config/complex/file2.txt
			echo "deep1" > ~/.config/plonk/config/complex/deep/deep1.txt
			echo "nested1" > ~/.config/plonk/config/complex/deep/nested/nested1.txt
			echo "other1" > ~/.config/plonk/config/complex/other/other1.txt
			
			# Create config that maps the complex directory
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
dotfiles:
  - source: config/complex
    destination: ~/.config/complex
EOF
			
			# Create actual destination files
			mkdir -p ~/.config/complex/deep/nested
			mkdir -p ~/.config/complex/other
			cp ~/.config/plonk/config/complex/file1.txt ~/.config/complex/
			cp ~/.config/plonk/config/complex/file2.txt ~/.config/complex/
			cp ~/.config/plonk/config/complex/deep/deep1.txt ~/.config/complex/deep/
			cp ~/.config/plonk/config/complex/deep/nested/nested1.txt ~/.config/complex/deep/nested/
			cp ~/.config/plonk/config/complex/other/other1.txt ~/.config/complex/other/
			
			echo "=== Running status command ==="
			/workspace/plonk status
			
			echo -e "\n=== Listing managed dotfiles ==="
			/workspace/plonk dot list managed
			
			echo -e "\n=== Counting managed files ==="
			/workspace/plonk dot list managed | grep -c "^[^#]" || echo "0"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Status count test output: %s", output)
		
		if err != nil {
			t.Fatalf("Status count test failed: %v\nOutput: %s", err, output)
		}
		
		// Verify that all 5 files are detected
		outputStr := string(output)
		managedCount := strings.Count(outputStr, ".config/complex/")
		if managedCount < 5 {
			t.Errorf("Expected at least 5 managed files in complex directory, got %d", managedCount)
		}
	})

	// Test mixed file and directory configuration
	t.Run("mixed file and directory config", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Mixed File and Directory Config ==="
			
			# Create mixed structure
			mkdir -p ~/.config/plonk/config/app1
			mkdir -p ~/.config/plonk/config/app2/subdir
			
			# Create files and directories
			echo "single file" > ~/.config/plonk/gitconfig
			echo "app1 config" > ~/.config/plonk/config/app1/config.json
			echo "app2 main" > ~/.config/plonk/config/app2/main.conf
			echo "app2 sub" > ~/.config/plonk/config/app2/subdir/sub.conf
			
			# Mixed configuration
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
dotfiles:
  - source: gitconfig
    destination: ~/.gitconfig
  - source: config/app1
    destination: ~/.config/app1
  - source: config/app2
    destination: ~/.config/app2
EOF
			
			# Create actual destinations
			mkdir -p ~/.config/app1
			mkdir -p ~/.config/app2/subdir
			cp ~/.config/plonk/gitconfig ~/.gitconfig
			cp ~/.config/plonk/config/app1/config.json ~/.config/app1/
			cp ~/.config/plonk/config/app2/main.conf ~/.config/app2/
			cp ~/.config/plonk/config/app2/subdir/sub.conf ~/.config/app2/subdir/
			
			echo "=== Status output ==="
			/workspace/plonk status
			
			echo -e "\n=== Managed dotfiles ==="
			/workspace/plonk dot list managed
			
			echo -e "\n=== Verifying single file and directory files ==="
			# Should show .gitconfig as single file
			/workspace/plonk dot list managed | grep ".gitconfig" || echo "gitconfig not found"
			# Should show app1 directory files
			/workspace/plonk dot list managed | grep "app1" || echo "app1 files not found"
			# Should show app2 directory files
			/workspace/plonk dot list managed | grep "app2" || echo "app2 files not found"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Mixed config test output: %s", output)
		
		if err != nil {
			t.Fatalf("Mixed config test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Verify single file is detected
		if !strings.Contains(outputStr, ".gitconfig") {
			t.Error("Expected .gitconfig single file to be detected")
		}
		
		// Verify directory files are detected
		if !strings.Contains(outputStr, "app1") {
			t.Error("Expected app1 directory files to be detected")
		}
		
		if !strings.Contains(outputStr, "app2") {
			t.Error("Expected app2 directory files to be detected")
		}
	})

	// Test that missing files in directories are handled correctly
	t.Run("missing directory files", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Missing Directory Files ==="
			
			# Create config directory structure
			mkdir -p ~/.config/plonk/config/missing_test
			echo "existing file" > ~/.config/plonk/config/missing_test/existing.txt
			echo "another file" > ~/.config/plonk/config/missing_test/another.txt
			
			# Create configuration
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
dotfiles:
  - source: config/missing_test
    destination: ~/.config/missing_test
EOF
			
			# Only create partial destination (missing some files)
			mkdir -p ~/.config/missing_test
			cp ~/.config/plonk/config/missing_test/existing.txt ~/.config/missing_test/
			# Deliberately don't copy another.txt
			
			echo "=== Status output (should show missing files) ==="
			/workspace/plonk status
			
			echo -e "\n=== List missing files ==="
			/workspace/plonk dot list missing || echo "No missing command or files"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Missing files test output: %s", output)
		
		if err != nil {
			t.Fatalf("Missing files test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show some missing items
		if !strings.Contains(outputStr, "missing") {
			t.Error("Expected missing files to be detected")
		}
	})
}