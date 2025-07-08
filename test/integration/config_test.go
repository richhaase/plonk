// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestConfigurationManagement(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test basic configuration operations
	t.Run("config show operations", func(t *testing.T) {
		// Test with no config file
		t.Run("no config file", func(t *testing.T) {
			output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk config show")
			t.Logf("No config output: %s", output)
			
			// Should handle missing config gracefully with exit code 0
			if err != nil {
				t.Errorf("Expected successful exit when no config file exists, got error: %v", err)
			}
			
			outputStr := string(output)
			if !strings.Contains(outputStr, "Configuration file not found") {
				t.Errorf("Expected 'Configuration file not found' message, got: %s", outputStr)
			}
		})

		// Test with valid config file
		t.Run("valid config file", func(t *testing.T) {
			configSetup := `
				cd /home/testuser
				# Create config directory and file
				mkdir -p ~/.config/plonk
				cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - jq
    - curl
npm:
  - lodash
  - prettier
dotfiles:
  - source: dot_zshrc
    destination: ~/.zshrc
EOF
				# Show the config
				/workspace/plonk config show
			`
			
			output, err := runner.RunCommand(t, configSetup)
			t.Logf("Valid config output: %s", output)
			
			if err != nil {
				t.Fatalf("Failed to show config: %v\nOutput: %s", err, output)
			}
			
			// Verify it shows the configuration content
			outputStr := string(output)
			if !strings.Contains(outputStr, "homebrew") {
				t.Errorf("Expected config to contain 'homebrew', got: %s", outputStr)
			}
		})
	})

	// Test configuration validation
	t.Run("config validation", func(t *testing.T) {
		// Test that plonk validates configuration properly
		t.Run("validate config structure", func(t *testing.T) {
			configSetup := `
				cd /home/testuser
				# Create config with various data types
				mkdir -p ~/.config/plonk
				cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - jq
    - curl
    - git
npm:
  - lodash
  - prettier
  - typescript
dotfiles:
  - source: dot_zshrc
    destination: ~/.zshrc
  - source: dot_vimrc
    destination: ~/.vimrc
EOF
				# Validate by showing config
				/workspace/plonk config show
			`
			
			output, err := runner.RunCommand(t, configSetup)
			t.Logf("Config validation output: %s", output)
			
			if err != nil {
				t.Fatalf("Config validation failed: %v\nOutput: %s", err, output)
			}
			
			// Check that it shows all expected elements
			outputStr := string(output)
			expectedItems := []string{"homebrew", "jq", "lodash", "npm", "zshrc", "vimrc"}
			for _, item := range expectedItems {
				if !strings.Contains(outputStr, item) {
					t.Errorf("Expected config to contain '%s', got: %s", item, outputStr)
				}
			}
		})
	})

	// Test configuration file formats
	t.Run("config file formats", func(t *testing.T) {
		// Test JSON output
		t.Run("json output", func(t *testing.T) {
			configSetup := `
				cd /home/testuser
				mkdir -p ~/.config/plonk
				cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - jq
npm:
  - lodash
EOF
				# Show config in JSON format
				/workspace/plonk config show --output json
			`
			
			output, err := runner.RunCommand(t, configSetup)
			t.Logf("JSON config output: %s", output)
			
			if err != nil {
				t.Fatalf("Failed to show config in JSON format: %v\nOutput: %s", err, output)
			}
			
			// Basic JSON validation
			outputStr := string(output)
			if !strings.Contains(outputStr, "{") || !strings.Contains(outputStr, "}") {
				t.Errorf("Expected JSON output to contain braces, got: %s", outputStr)
			}
		})

		// Test YAML output
		t.Run("yaml output", func(t *testing.T) {
			configSetup := `
				cd /home/testuser
				mkdir -p ~/.config/plonk
				cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - jq
EOF
				# Show config in YAML format
				/workspace/plonk config show --output yaml
			`
			
			output, err := runner.RunCommand(t, configSetup)
			t.Logf("YAML config output: %s", output)
			
			if err != nil {
				t.Fatalf("Failed to show config in YAML format: %v\nOutput: %s", err, output)
			}
			
			// Basic YAML validation
			outputStr := string(output)
			if !strings.Contains(outputStr, "homebrew") {
				t.Errorf("Expected YAML output to contain 'homebrew', got: %s", outputStr)
			}
		})
	})
}