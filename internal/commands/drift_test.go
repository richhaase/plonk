package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDriftDetection_NoConfig(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Test - should error when no config exists
	drift, err := detectConfigDrift()
	if err == nil {
		t.Error("Expected error when no config file exists")
	}
	if drift != nil {
		t.Error("Expected nil drift when no config exists")
	}
}

func TestDriftDetection_NoDrift(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create simple config (no packages)
	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create source file for dotfile
	err = os.WriteFile(filepath.Join(plonkDir, "zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// Create target file (already applied)
	err = os.WriteFile(filepath.Join(tempHome, ".zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}
	
	// Test drift detection
	drift, err := detectConfigDrift()
	if err != nil {
		t.Fatalf("Drift detection failed: %v", err)
	}
	
	if drift.HasDrift() {
		t.Error("Expected no drift when configs match")
	}
	
	if len(drift.MissingFiles) != 0 {
		t.Errorf("Expected no missing files, got %d", len(drift.MissingFiles))
	}
	
	if len(drift.ModifiedFiles) != 0 {
		t.Errorf("Expected no modified files, got %d", len(drift.ModifiedFiles))
	}
}

func TestDriftDetection_MissingFiles(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config with dotfiles
	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
  - gitconfig
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create source files
	err = os.WriteFile(filepath.Join(plonkDir, "zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create zshrc source: %v", err)
	}
	
	err = os.WriteFile(filepath.Join(plonkDir, "gitconfig"), []byte("# test gitconfig"), 0644)
	if err != nil {
		t.Fatalf("Failed to create gitconfig source: %v", err)
	}
	
	// Only create one target file (gitconfig missing)
	err = os.WriteFile(filepath.Join(tempHome, ".zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}
	
	// Test drift detection
	drift, err := detectConfigDrift()
	if err != nil {
		t.Fatalf("Drift detection failed: %v", err)
	}
	
	if !drift.HasDrift() {
		t.Error("Expected drift when files are missing")
	}
	
	if len(drift.MissingFiles) != 1 {
		t.Errorf("Expected 1 missing file, got %d", len(drift.MissingFiles))
	}
	
	if drift.MissingFiles[0] != "~/.gitconfig" {
		t.Errorf("Expected missing ~/.gitconfig, got %s", drift.MissingFiles[0])
	}
}

func TestDriftDetection_ModifiedFiles(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config
	configContent := `settings:
  default_manager: homebrew

dotfiles:
  - zshrc
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create source file
	err = os.WriteFile(filepath.Join(plonkDir, "zshrc"), []byte("# test zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	
	// Create different target file (modified)
	err = os.WriteFile(filepath.Join(tempHome, ".zshrc"), []byte("# modified zshrc"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}
	
	// Test drift detection
	drift, err := detectConfigDrift()
	if err != nil {
		t.Fatalf("Drift detection failed: %v", err)
	}
	
	if !drift.HasDrift() {
		t.Error("Expected drift when files are modified")
	}
	
	if len(drift.ModifiedFiles) != 1 {
		t.Errorf("Expected 1 modified file, got %d", len(drift.ModifiedFiles))
	}
	
	if drift.ModifiedFiles[0] != "~/.zshrc" {
		t.Errorf("Expected modified ~/.zshrc, got %s", drift.ModifiedFiles[0])
	}
}

func TestDriftDetection_MissingPackages(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config with packages (but package managers not available in test)
	configContent := `settings:
  default_manager: homebrew

homebrew:
  brews:
    - nonexistent-package
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test drift detection
	drift, err := detectConfigDrift()
	if err != nil {
		t.Fatalf("Drift detection failed: %v", err)
	}
	
	// Should detect missing packages (since package managers won't be available in test)
	if !drift.HasDrift() {
		t.Error("Expected drift when packages are missing")
	}
	
	if len(drift.MissingPackages) == 0 {
		t.Error("Expected missing packages to be detected")
	}
}