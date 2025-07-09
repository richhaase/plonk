// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestApplyCommand_Creation(t *testing.T) {
	// Test that the apply command is created correctly
	if applyCmd == nil {
		t.Fatal("applyCmd is nil")
	}

	if applyCmd.Use != "apply" {
		t.Errorf("Expected Use to be 'apply', got '%s'", applyCmd.Use)
	}

	if applyCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if applyCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if applyCmd.RunE == nil {
		t.Error("RunE should not be nil")
	}
}

func TestApplyCommand_Flags(t *testing.T) {
	// Test that flags are set up correctly
	flag := applyCmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Error("dry-run flag not found")
	}

	flag = applyCmd.Flags().Lookup("backup")
	if flag == nil {
		t.Error("backup flag not found")
	}

	// Test flag defaults
	if applyDryRun != false {
		t.Error("applyDryRun should default to false")
	}

	if applyBackup != false {
		t.Error("applyBackup should default to false")
	}
}

func TestApplyCommand_HelpText(t *testing.T) {
	// Test that help text contains expected information
	longDesc := applyCmd.Long
	
	expectedPhrases := []string{
		"Apply the complete plonk configuration",
		"Install all missing packages",
		"Deploy all dotfiles",
		"single operation",
		"--dry-run",
		"--backup",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(longDesc, phrase) {
			t.Errorf("Long description missing expected phrase: %s", phrase)
		}
	}
}

func TestApplyCommand_OutputFormat(t *testing.T) {
	// Test that invalid output formats are handled
	outputFormat = "invalid"
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "plonk-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal config file
	configDir := filepath.Join(tmpDir, ".config", "plonk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "plonk.yaml")
	configContent := `
settings:
  default_manager: homebrew
homebrew: []
npm: []
dotfiles: []
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Temporarily change home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test that invalid output format returns error
	err = runApply(applyCmd, []string{})
	if err == nil {
		t.Error("Expected error for invalid output format")
	}

	if !strings.Contains(err.Error(), "invalid output format") {
		t.Errorf("Expected 'invalid output format' error, got: %v", err)
	}

	// Reset output format
	outputFormat = "table"
}

func TestApplyCommand_NoConfig(t *testing.T) {
	// Test behavior when no config file exists
	tmpDir, err := os.MkdirTemp("", "plonk-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Temporarily change home directory to a location without config
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Reset flags
	applyDryRun = false
	applyBackup = false
	outputFormat = "table"

	// Test that missing config returns error
	err = runApply(applyCmd, []string{})
	if err == nil {
		t.Error("Expected error for missing config file")
	}

	// Should contain config-related error
	if !strings.Contains(err.Error(), "config") {
		t.Errorf("Expected config-related error, got: %v", err)
	}
}

func TestApplyCommand_EmptyConfig(t *testing.T) {
	// Test behavior with empty but valid config
	tmpDir, err := os.MkdirTemp("", "plonk-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create minimal config
	configDir := filepath.Join(tmpDir, ".config", "plonk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "plonk.yaml")
	configContent := `
settings:
  default_manager: homebrew
homebrew: []
npm: []
dotfiles: []
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Temporarily change home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Reset flags
	applyDryRun = true // Use dry-run to avoid actual operations
	applyBackup = false
	outputFormat = "table"

	// Test that empty config doesn't error (but does nothing)
	err = runApply(applyCmd, []string{})
	if err != nil {
		// Allow npm availability errors in test environment
		if !strings.Contains(err.Error(), "npm") {
			t.Errorf("Empty config should not error (except for npm), got: %v", err)
		}
	}
}

func TestApplyCommand_DryRunFlag(t *testing.T) {
	// Test that dry-run flag is respected
	tmpDir, err := os.MkdirTemp("", "plonk-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config with some packages
	configDir := filepath.Join(tmpDir, ".config", "plonk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "plonk.yaml")
	configContent := `
settings:
  default_manager: homebrew
homebrew:
  - nonexistent-test-package
npm: []
dotfiles: []
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Temporarily change home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test dry-run mode
	applyDryRun = true
	applyBackup = false
	outputFormat = "table"

	// This should not error even with non-existent packages in dry-run
	err = runApply(applyCmd, []string{})
	if err != nil {
		// Allow npm availability errors in test environment
		if !strings.Contains(err.Error(), "npm") {
			t.Errorf("Dry-run should not error for non-existent packages (except for npm), got: %v", err)
		}
	}
}

func TestCombinedApplyOutput_TableOutput(t *testing.T) {
	// Test table output formatting
	output := CombinedApplyOutput{
		DryRun: false,
		Packages: ApplyOutput{
			TotalInstalled: 2,
			TotalFailed:    1,
		},
		Dotfiles: DotfileApplyOutput{
			Deployed: 3,
			Skipped:  1,
		},
	}

	tableOutput := output.TableOutput()
	
	expectedElements := []string{
		"Plonk Apply",
		"📦 Packages: 2 installed, 1 failed",
		"📄 Dotfiles: 3 deployed, 1 skipped",
	}

	for _, element := range expectedElements {
		if !strings.Contains(tableOutput, element) {
			t.Errorf("Table output missing expected element: %s\nOutput: %s", element, tableOutput)
		}
	}
}

func TestCombinedApplyOutput_DryRunTableOutput(t *testing.T) {
	// Test dry-run table output formatting
	output := CombinedApplyOutput{
		DryRun: true,
		Packages: ApplyOutput{
			TotalWouldInstall: 2,
		},
		Dotfiles: DotfileApplyOutput{
			Deployed: 3,
			Skipped:  1,
		},
	}

	tableOutput := output.TableOutput()
	
	expectedElements := []string{
		"Plonk Apply (Dry Run)",
		"📦 Packages: 2 would be installed",
		"📄 Dotfiles: 3 would be deployed, 1 would be skipped",
	}

	for _, element := range expectedElements {
		if !strings.Contains(tableOutput, element) {
			t.Errorf("Dry-run table output missing expected element: %s\nOutput: %s", element, tableOutput)
		}
	}
}

func TestCombinedApplyOutput_StructuredData(t *testing.T) {
	// Test structured data output
	output := CombinedApplyOutput{
		DryRun: true,
		Packages: ApplyOutput{
			TotalWouldInstall: 2,
		},
		Dotfiles: DotfileApplyOutput{
			Deployed: 3,
			Skipped:  1,
		},
	}

	data := output.StructuredData()
	
	// Should return the same struct (test by type assertion)
	if _, ok := data.(CombinedApplyOutput); !ok {
		t.Error("StructuredData() should return CombinedApplyOutput type")
	}

	// Verify type assertion works
	if combinedOutput, ok := data.(CombinedApplyOutput); ok {
		if combinedOutput.DryRun != true {
			t.Error("StructuredData() should preserve DryRun field")
		}
	} else {
		t.Error("StructuredData() should return CombinedApplyOutput type")
	}
}

func TestApplyCommand_Integration(t *testing.T) {
	// Integration test that verifies the command works end-to-end
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "plonk-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a more complex config
	configDir := filepath.Join(tmpDir, ".config", "plonk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create dotfiles directory
	dotfilesDir := filepath.Join(configDir, "dotfiles")
	if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
		t.Fatalf("Failed to create dotfiles directory: %v", err)
	}

	// Create test dotfile
	testDotfile := filepath.Join(dotfilesDir, "test_bashrc")
	if err := os.WriteFile(testDotfile, []byte("# Test bashrc\necho 'test'\n"), 0644); err != nil {
		t.Fatalf("Failed to create test dotfile: %v", err)
	}

	configFile := filepath.Join(configDir, "plonk.yaml")
	configContent := `
settings:
  default_manager: homebrew
homebrew: []
npm: []
dotfiles:
  - source: test_bashrc
    destination: ~/.bashrc
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Temporarily change home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test with dry-run first
	applyDryRun = true
	applyBackup = false
	outputFormat = "table"

	err = runApply(applyCmd, []string{})
	if err != nil {
		// Allow npm availability errors in test environment
		if !strings.Contains(err.Error(), "npm") {
			t.Errorf("Apply with dry-run should not error (except for npm), got: %v", err)
		}
	}

	// Test with JSON output
	outputFormat = "json"
	err = runApply(applyCmd, []string{})
	if err != nil {
		// Allow npm availability errors in test environment
		if !strings.Contains(err.Error(), "npm") {
			t.Errorf("Apply with JSON output should not error (except for npm), got: %v", err)
		}
	}

	// Reset output format
	outputFormat = "table"
}

func TestApplyCommand_Timeout(t *testing.T) {
	// Test that the apply command doesn't hang indefinitely
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "plonk-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create minimal config
	configDir := filepath.Join(tmpDir, ".config", "plonk")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "plonk.yaml")
	configContent := `
settings:
  default_manager: homebrew
homebrew: []
npm: []
dotfiles: []
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Temporarily change home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Set up timeout test
	applyDryRun = true
	applyBackup = false
	outputFormat = "table"

	timeout := 30 * time.Second
	done := make(chan error, 1)

	go func() {
		done <- runApply(applyCmd, []string{})
	}()

	select {
	case err := <-done:
		if err != nil {
			// Allow npm availability errors in test environment
			if !strings.Contains(err.Error(), "npm") {
				t.Errorf("Apply command failed: %v", err)
			}
		}
	case <-time.After(timeout):
		t.Fatalf("Apply command timed out after %v", timeout)
	}
}

func TestApplyOutputStructures(t *testing.T) {
	// Test that all output structures implement the OutputData interface
	var _ OutputData = CombinedApplyOutput{}
	var _ OutputData = ApplyOutput{}
	var _ OutputData = DotfileApplyOutput{}

	// Test that methods exist and work
	combinedOutput := CombinedApplyOutput{
		DryRun: false,
		Packages: ApplyOutput{
			TotalInstalled: 1,
		},
		Dotfiles: DotfileApplyOutput{
			Deployed: 2,
		},
	}

	tableOutput := combinedOutput.TableOutput()
	if len(tableOutput) == 0 {
		t.Error("TableOutput() should not return empty string")
	}

	structuredData := combinedOutput.StructuredData()
	if structuredData == nil {
		t.Error("StructuredData() should not return nil")
	}
}