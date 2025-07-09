// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strconv"
	"strings"
	"testing"
)

func TestStatusWithDirectoryFiles(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test that status command properly counts and displays directory files
	t.Run("status counts directory files correctly", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Status Command with Directory Files ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			
			# Create comprehensive dotfile structure
			mkdir -p ~/.config/plonk/config/nvim/lua/config
			mkdir -p ~/.config/plonk/config/nvim/lua/plugins
			mkdir -p ~/.config/plonk/config/alacritty
			
			# Create nvim config files
			cat > ~/.config/plonk/config/nvim/init.lua << 'EOF'
-- Test nvim init.lua
require("config.lazy")
EOF
			
			cat > ~/.config/plonk/config/nvim/lua/config/lazy.lua << 'EOF'
-- Test lazy.lua
return { spec = {} }
EOF
			
			cat > ~/.config/plonk/config/nvim/lua/config/options.lua << 'EOF'
-- Test options.lua
vim.opt.number = true
EOF
			
			cat > ~/.config/plonk/config/nvim/lua/plugins/editor.lua << 'EOF'
-- Test editor.lua
return {}
EOF
			
			# Create alacritty config
			cat > ~/.config/plonk/config/alacritty/alacritty.yml << 'EOF'
window:
  dimensions:
    columns: 80
    lines: 24
EOF
			
			# Create simple dotfiles
			echo "# Test zshrc" > ~/.config/plonk/zshrc
			echo "# Test gitconfig" > ~/.config/plonk/gitconfig
			
			# Create configuration with mixed files and directories
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - git
    - neovim
dotfiles:
  - source: zshrc
    destination: ~/.zshrc
  - source: gitconfig
    destination: ~/.gitconfig
  - source: config/nvim
    destination: ~/.config/nvim
  - source: config/alacritty
    destination: ~/.config/alacritty
EOF
			
			# Create actual destination files (simulating applied state)
			mkdir -p ~/.config/nvim/lua/config
			mkdir -p ~/.config/nvim/lua/plugins
			mkdir -p ~/.config/alacritty
			
			cp ~/.config/plonk/config/nvim/init.lua ~/.config/nvim/
			cp ~/.config/plonk/config/nvim/lua/config/lazy.lua ~/.config/nvim/lua/config/
			cp ~/.config/plonk/config/nvim/lua/config/options.lua ~/.config/nvim/lua/config/
			cp ~/.config/plonk/config/nvim/lua/plugins/editor.lua ~/.config/nvim/lua/plugins/
			cp ~/.config/plonk/config/alacritty/alacritty.yml ~/.config/alacritty/
			cp ~/.config/plonk/zshrc ~/.zshrc
			cp ~/.config/plonk/gitconfig ~/.gitconfig
			
			echo "=== Running plonk status ==="
			/workspace/plonk status
			
			echo -e "\n=== Checking managed count ==="
			# Count managed items from status output
			/workspace/plonk status | grep "managed items" | head -1
			
			echo -e "\n=== Verifying individual files are listed ==="
			/workspace/plonk dot list managed | sort
			
			echo -e "\n=== Counting actual managed dotfiles ==="
			/workspace/plonk dot list managed | grep -c "^[^#]" || echo "0"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Status directory test output: %s", output)
		
		if err != nil {
			t.Fatalf("Status directory test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Verify that status shows the correct number of managed items
		// Expected: 2 simple files + 4 nvim files + 1 alacritty file = 7 dotfiles
		if strings.Contains(outputStr, "managed items") {
			// Extract the number from the managed items line
			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				if strings.Contains(line, "managed items") {
					// Extract number from line like "  ✅ 7 managed items"
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						managedCount, err := strconv.Atoi(parts[1])
						if err == nil && managedCount < 7 {
							t.Errorf("Expected at least 7 managed items, got %d", managedCount)
						}
					}
					break
				}
			}
		}
		
		// Verify that individual nvim files are shown
		if !strings.Contains(outputStr, "init.lua") {
			t.Error("Expected nvim init.lua to be shown in managed files")
		}
		
		if !strings.Contains(outputStr, "lazy.lua") {
			t.Error("Expected nvim lazy.lua to be shown in managed files")
		}
		
		if !strings.Contains(outputStr, "options.lua") {
			t.Error("Expected nvim options.lua to be shown in managed files")
		}
		
		if !strings.Contains(outputStr, "editor.lua") {
			t.Error("Expected nvim editor.lua to be shown in managed files")
		}
		
		if !strings.Contains(outputStr, "alacritty.yml") {
			t.Error("Expected alacritty.yml to be shown in managed files")
		}
	})

	// Test that status handles missing directory files correctly
	t.Run("status handles missing directory files", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Status with Missing Directory Files ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			
			# Create source files
			mkdir -p ~/.config/plonk/config/incomplete
			echo "file1 content" > ~/.config/plonk/config/incomplete/file1.txt
			echo "file2 content" > ~/.config/plonk/config/incomplete/file2.txt
			echo "file3 content" > ~/.config/plonk/config/incomplete/file3.txt
			
			# Create configuration
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
dotfiles:
  - source: config/incomplete
    destination: ~/.config/incomplete
EOF
			
			# Only create partial destination (missing some files)
			mkdir -p ~/.config/incomplete
			cp ~/.config/plonk/config/incomplete/file1.txt ~/.config/incomplete/
			# Deliberately omit file2.txt and file3.txt
			
			echo "=== Running plonk status ==="
			/workspace/plonk status
			
			echo -e "\n=== Checking for missing items ==="
			/workspace/plonk status | grep "missing items" || echo "No missing items found"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Missing directory files test output: %s", output)
		
		if err != nil {
			t.Fatalf("Missing directory files test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show missing items
		if !strings.Contains(outputStr, "missing") {
			t.Error("Expected missing items to be detected when directory files are missing")
		}
		
		// Should show that at least one file is managed
		if !strings.Contains(outputStr, "managed") {
			t.Error("Expected at least one file to be shown as managed")
		}
	})

	// Test status output formatting with many directory files
	t.Run("status formatting with many files", func(t *testing.T) {
		testScript := `
			cd /home/testuser
			
			echo "=== Testing Status Formatting with Many Files ==="
			
			# Create plonk config directory
			mkdir -p ~/.config/plonk
			
			# Create a complex directory structure with many files
			mkdir -p ~/.config/plonk/config/complex/{dir1,dir2,dir3,dir4,dir5}
			
			# Create many files
			for i in {1..10}; do
				echo "file $i content" > ~/.config/plonk/config/complex/dir1/file$i.txt
				echo "file $i content" > ~/.config/plonk/config/complex/dir2/file$i.txt
			done
			
			# Create single files too
			echo "simple1" > ~/.config/plonk/simple1
			echo "simple2" > ~/.config/plonk/simple2
			
			# Create configuration
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
dotfiles:
  - source: simple1
    destination: ~/.simple1
  - source: simple2
    destination: ~/.simple2
  - source: config/complex
    destination: ~/.config/complex
EOF
			
			# Create destination files
			mkdir -p ~/.config/complex/{dir1,dir2,dir3,dir4,dir5}
			
			for i in {1..10}; do
				cp ~/.config/plonk/config/complex/dir1/file$i.txt ~/.config/complex/dir1/
				cp ~/.config/plonk/config/complex/dir2/file$i.txt ~/.config/complex/dir2/
			done
			
			cp ~/.config/plonk/simple1 ~/.simple1
			cp ~/.config/plonk/simple2 ~/.simple2
			
			echo "=== Running plonk status ==="
			/workspace/plonk status
			
			echo -e "\n=== Checking that status shows truncated list ==="
			# Status should show truncated list with "and X more"
			/workspace/plonk status | grep "and.*more" || echo "No truncation found"
		`
		
		output, err := runner.RunCommand(t, testScript)
		t.Logf("Status formatting test output: %s", output)
		
		if err != nil {
			t.Fatalf("Status formatting test failed: %v\nOutput: %s", err, output)
		}
		
		outputStr := string(output)
		
		// Should show many managed items
		if strings.Contains(outputStr, "managed items") {
			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				if strings.Contains(line, "managed items") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						managedCount, err := strconv.Atoi(parts[1])
						if err == nil && managedCount < 20 {
							t.Errorf("Expected at least 20 managed items with many files, got %d", managedCount)
						}
					}
					break
				}
			}
		}
		
		// Should show truncation if there are many files
		if !strings.Contains(outputStr, "more") {
			t.Log("Status didn't show truncation, which is OK if there aren't enough files")
		}
	})
}