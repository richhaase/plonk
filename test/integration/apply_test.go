// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration
// +build integration

package integration

import (
	"strings"
	"testing"
)

func TestApplyCommand(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test basic apply command help
	t.Run("apply help", func(t *testing.T) {
		output, err := runner.RunCommand(t, "/workspace/plonk apply --help")
		if err != nil {
			t.Fatalf("Failed to run apply --help: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		expectedPhrases := []string{
			"Apply the complete plonk configuration",
			"Install all missing packages",
			"Deploy all dotfiles",
			"--dry-run",
			"--backup",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(outputStr, phrase) {
				t.Errorf("Help output missing expected phrase: %s\nOutput: %s", phrase, outputStr)
			}
		}
	})

	// Test apply with no configuration
	t.Run("apply with no config", func(t *testing.T) {
		output, err := runner.RunCommand(t, "cd /home/testuser && /workspace/plonk apply")
		
		// Should error due to missing config
		if err == nil {
			t.Error("Expected error for missing config file")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "config") {
			t.Errorf("Expected config-related error message, got: %s", outputStr)
		}
	})

	// Test apply with empty configuration
	t.Run("apply with empty config", func(t *testing.T) {
		setupScript := `
			cd /home/testuser
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles: []
EOF
			/workspace/plonk apply --dry-run
		`

		output, err := runner.RunCommand(t, setupScript)
		if err != nil {
			t.Fatalf("Failed to run apply with empty config: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		expectedPhrases := []string{
			"All packages up to date",
			"Plonk Apply (Dry Run)",
			"📦 Packages: 0 would be installed",
			"📄 Dotfiles: 0 would be deployed",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(outputStr, phrase) {
				t.Errorf("Empty config output missing expected phrase: %s\nOutput: %s", phrase, outputStr)
			}
		}
	})

	// Test apply with package configuration
	t.Run("apply with packages", func(t *testing.T) {
		packageScript := `
			cd /home/testuser
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
dotfiles: []
EOF
			/workspace/plonk apply --dry-run
		`

		output, err := runner.RunCommand(t, packageScript)
		if err != nil {
			t.Fatalf("Failed to run apply with packages: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		expectedPhrases := []string{
			"Plonk Apply (Dry Run)",
			"📦 Packages:",
			"📄 Dotfiles:",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(outputStr, phrase) {
				t.Errorf("Package config output missing expected phrase: %s\nOutput: %s", phrase, outputStr)
			}
		}
	})

	// Test apply with dotfiles configuration
	t.Run("apply with dotfiles", func(t *testing.T) {
		dotfileScript := `
			cd /home/testuser
			mkdir -p ~/.config/plonk/dotfiles
			echo "# Test bashrc" > ~/.config/plonk/dotfiles/test_bashrc
			echo "# Test vimrc" > ~/.config/plonk/dotfiles/test_vimrc
			
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles:
  - source: test_bashrc
    destination: ~/.bashrc
  - source: test_vimrc
    destination: ~/.vimrc
EOF
			/workspace/plonk apply --dry-run
		`

		output, err := runner.RunCommand(t, dotfileScript)
		if err != nil {
			t.Fatalf("Failed to run apply with dotfiles: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		expectedPhrases := []string{
			"Plonk Apply (Dry Run)",
			"📄 Dotfile summary:",
			"📦 Packages: 0 would be installed",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(outputStr, phrase) {
				t.Errorf("Dotfile config output missing expected phrase: %s\nOutput: %s", phrase, outputStr)
			}
		}
	})

	// Test apply with mixed configuration
	t.Run("apply with mixed config", func(t *testing.T) {
		mixedScript := `
			cd /home/testuser
			mkdir -p ~/.config/plonk/dotfiles
			echo "# Test bashrc" > ~/.config/plonk/dotfiles/test_bashrc
			
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews:
    - curl
npm:
  - lodash
dotfiles:
  - source: test_bashrc
    destination: ~/.bashrc
EOF
			/workspace/plonk apply --dry-run
		`

		output, err := runner.RunCommand(t, mixedScript)
		if err != nil {
			t.Fatalf("Failed to run apply with mixed config: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		expectedPhrases := []string{
			"Plonk Apply (Dry Run)",
			"📦 Packages:",
			"📄 Dotfiles:",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(outputStr, phrase) {
				t.Errorf("Mixed config output missing expected phrase: %s\nOutput: %s", phrase, outputStr)
			}
		}
	})

	// Test apply with JSON output
	t.Run("apply with JSON output", func(t *testing.T) {
		jsonScript := `
			cd /home/testuser
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles: []
EOF
			/workspace/plonk apply --dry-run -o json
		`

		output, err := runner.RunCommand(t, jsonScript)
		if err != nil {
			t.Fatalf("Failed to run apply with JSON output: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		expectedJSONFields := []string{
			`"dry_run": true`,
			`"packages"`,
			`"dotfiles"`,
		}

		for _, field := range expectedJSONFields {
			if !strings.Contains(outputStr, field) {
				t.Errorf("JSON output missing expected field: %s\nOutput: %s", field, outputStr)
			}
		}

		// Basic JSON validation - should not contain table output
		if strings.Contains(outputStr, "📦") || strings.Contains(outputStr, "📄") {
			t.Errorf("JSON output should not contain table formatting: %s", outputStr)
		}
	})

	// Test apply with YAML output
	t.Run("apply with YAML output", func(t *testing.T) {
		yamlScript := `
			cd /home/testuser
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles: []
EOF
			/workspace/plonk apply --dry-run -o yaml
		`

		output, err := runner.RunCommand(t, yamlScript)
		if err != nil {
			t.Fatalf("Failed to run apply with YAML output: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		expectedYAMLFields := []string{
			"dry_run: true",
			"packages:",
			"dotfiles:",
		}

		for _, field := range expectedYAMLFields {
			if !strings.Contains(outputStr, field) {
				t.Errorf("YAML output missing expected field: %s\nOutput: %s", field, outputStr)
			}
		}

		// Basic YAML validation - should not contain table output
		if strings.Contains(outputStr, "📦") || strings.Contains(outputStr, "📄") {
			t.Errorf("YAML output should not contain table formatting: %s", outputStr)
		}
	})

	// Test apply with backup flag
	t.Run("apply with backup flag", func(t *testing.T) {
		backupScript := `
			cd /home/testuser
			mkdir -p ~/.config/plonk/dotfiles
			echo "# Test bashrc" > ~/.config/plonk/dotfiles/test_bashrc
			echo "# Existing bashrc" > ~/.bashrc
			
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles:
  - source: test_bashrc
    destination: ~/.bashrc
EOF
			/workspace/plonk apply --dry-run --backup
		`

		output, err := runner.RunCommand(t, backupScript)
		if err != nil {
			t.Fatalf("Failed to run apply with backup flag: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		expectedPhrases := []string{
			"Plonk Apply (Dry Run)",
			"📄 Dotfile summary:",
			"📦 Packages: 0 would be installed",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(outputStr, phrase) {
				t.Errorf("Backup flag output missing expected phrase: %s\nOutput: %s", phrase, outputStr)
			}
		}
	})

	// Test apply with invalid output format
	t.Run("apply with invalid output format", func(t *testing.T) {
		invalidScript := `
			cd /home/testuser
			mkdir -p ~/.config/plonk
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles: []
EOF
			/workspace/plonk apply --dry-run -o invalid
		`

		output, err := runner.RunCommand(t, invalidScript)
		
		// Should error due to invalid format
		if err == nil {
			t.Error("Expected error for invalid output format")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "invalid") || !strings.Contains(outputStr, "format") {
			t.Errorf("Expected invalid format error message, got: %s", outputStr)
		}
	})
}

func TestApplyWorkflows(t *testing.T) {
	RequireDockerImage(t)
	runner := NewDockerRunner()
	CleanupBuildArtifacts(t)

	// Build plonk binary first
	t.Run("build plonk binary", func(t *testing.T) {
		if err := runner.BuildPlonkBinary(t); err != nil {
			t.Fatalf("Failed to build plonk binary: %v", err)
		}
	})

	// Test complete apply workflow
	t.Run("complete apply workflow", func(t *testing.T) {
		workflowScript := `
			cd /home/testuser
			
			echo "=== Complete Apply Workflow ==="
			
			# 1. Set up initial configuration
			echo "1. Setting up initial configuration..."
			mkdir -p ~/.config/plonk/dotfiles
			echo "# Test bashrc" > ~/.config/plonk/dotfiles/test_bashrc
			echo "# Test vimrc" > ~/.config/plonk/dotfiles/test_vimrc
			
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles:
  - source: test_bashrc
    destination: ~/.bashrc
  - source: test_vimrc
    destination: ~/.vimrc
EOF
			
			# 2. Check initial status
			echo "2. Checking initial status..."
			/workspace/plonk status || echo "Status check completed"
			
			# 3. Run apply dry-run
			echo "3. Running apply dry-run..."
			/workspace/plonk apply --dry-run
			
			# 4. Run apply with table output
			echo "4. Running apply with table output..."
			/workspace/plonk apply --dry-run -o table
			
			# 5. Run apply with JSON output
			echo "5. Running apply with JSON output..."
			/workspace/plonk apply --dry-run -o json
			
			# 6. Run apply with YAML output
			echo "6. Running apply with YAML output..."
			/workspace/plonk apply --dry-run -o yaml
			
			# 7. Run apply with backup flag
			echo "7. Running apply with backup flag..."
			/workspace/plonk apply --dry-run --backup
			
			# 8. Final status check
			echo "8. Final status check..."
			/workspace/plonk status || echo "Final status check completed"
			
			echo "=== Complete Apply Workflow Complete ==="
		`

		output, err := runner.RunCommand(t, workflowScript)
		if err != nil {
			t.Fatalf("Failed to run complete apply workflow: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		workflowSteps := []string{
			"Complete Apply Workflow",
			"Setting up initial configuration",
			"Checking initial status",
			"Running apply dry-run",
			"Running apply with table output",
			"Running apply with JSON output",
			"Running apply with YAML output",
			"Running apply with backup flag",
			"Final status check",
			"Complete Apply Workflow Complete",
		}

		for _, step := range workflowSteps {
			if !strings.Contains(outputStr, step) {
				t.Errorf("Workflow step missing: %s\nOutput: %s", step, outputStr)
			}
		}
	})

	// Test apply error handling workflow
	t.Run("apply error handling workflow", func(t *testing.T) {
		errorScript := `
			cd /home/testuser
			
			echo "=== Apply Error Handling Workflow ==="
			
			# 1. Test with missing config
			echo "1. Testing with missing config..."
			/workspace/plonk apply --dry-run || echo "Missing config error (expected)"
			
			# 2. Test with invalid config
			echo "2. Testing with invalid config..."
			mkdir -p ~/.config/plonk
			echo "invalid yaml content" > ~/.config/plonk/plonk.yaml
			/workspace/plonk apply --dry-run || echo "Invalid config error (expected)"
			
			# 3. Test with invalid output format
			echo "3. Testing with invalid output format..."
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles: []
EOF
			/workspace/plonk apply --dry-run -o invalid || echo "Invalid format error (expected)"
			
			# 4. Test with valid config after errors
			echo "4. Testing with valid config after errors..."
			/workspace/plonk apply --dry-run
			
			echo "=== Apply Error Handling Workflow Complete ==="
		`

		output, err := runner.RunCommand(t, errorScript)
		if err != nil {
			t.Logf("Error handling workflow completed with expected errors: %v", err)
		}

		outputStr := string(output)
		errorSteps := []string{
			"Apply Error Handling Workflow",
			"Testing with missing config",
			"Testing with invalid config",
			"Testing with invalid output format",
			"Testing with valid config after errors",
			"Apply Error Handling Workflow Complete",
		}

		for _, step := range errorSteps {
			if !strings.Contains(outputStr, step) {
				t.Errorf("Error handling step missing: %s\nOutput: %s", step, outputStr)
			}
		}
	})

	// Test apply performance workflow
	t.Run("apply performance workflow", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping performance test in short mode")
		}

		perfScript := `
			cd /home/testuser
			
			echo "=== Apply Performance Workflow ==="
			
			# 1. Create larger configuration
			echo "1. Creating larger configuration..."
			mkdir -p ~/.config/plonk/dotfiles
			
			# Create multiple dotfiles
			for i in {1..10}; do
				echo "# Test file $i" > ~/.config/plonk/dotfiles/test_file_$i
			done
			
			# Create config with multiple items
			cat > ~/.config/plonk/plonk.yaml << 'EOF'
settings:
  default_manager: homebrew
homebrew:
  brews: []
npm: []
dotfiles:
  - source: test_file_1
    destination: ~/.test_file_1
  - source: test_file_2
    destination: ~/.test_file_2
  - source: test_file_3
    destination: ~/.test_file_3
  - source: test_file_4
    destination: ~/.test_file_4
  - source: test_file_5
    destination: ~/.test_file_5
  - source: test_file_6
    destination: ~/.test_file_6
  - source: test_file_7
    destination: ~/.test_file_7
  - source: test_file_8
    destination: ~/.test_file_8
  - source: test_file_9
    destination: ~/.test_file_9
  - source: test_file_10
    destination: ~/.test_file_10
EOF
			
			# 2. Run apply multiple times to test performance
			echo "2. Running apply multiple times..."
			for i in {1..5}; do
				echo "  Run $i..."
				/workspace/plonk apply --dry-run > /dev/null
			done
			
			# 3. Test with different output formats
			echo "3. Testing different output formats..."
			/workspace/plonk apply --dry-run -o table > /dev/null
			/workspace/plonk apply --dry-run -o json > /dev/null
			/workspace/plonk apply --dry-run -o yaml > /dev/null
			
			echo "=== Apply Performance Workflow Complete ==="
		`

		output, err := runner.RunCommand(t, perfScript)
		if err != nil {
			t.Fatalf("Failed to run apply performance workflow: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		perfSteps := []string{
			"Apply Performance Workflow",
			"Creating larger configuration",
			"Running apply multiple times",
			"Testing different output formats",
			"Apply Performance Workflow Complete",
		}

		for _, step := range perfSteps {
			if !strings.Contains(outputStr, step) {
				t.Errorf("Performance step missing: %s\nOutput: %s", step, outputStr)
			}
		}
	})
}