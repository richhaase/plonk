package dotfiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestReconcile_RealFS(t *testing.T) {
	// Create temp directories
	configDir, err := os.MkdirTemp("", "plonk-config-*")
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	defer os.RemoveAll(configDir)

	homeDir, err := os.MkdirTemp("", "plonk-home-*")
	if err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	defer os.RemoveAll(homeDir)

	// Create some test files
	if err := os.WriteFile(filepath.Join(configDir, "zshrc"), []byte("# test"), 0644); err != nil {
		t.Fatalf("Failed to create zshrc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "vimrc"), []byte("# test"), 0644); err != nil {
		t.Fatalf("Failed to create vimrc: %v", err)
	}

	// Create manager with nil patterns
	dm := NewDotfileManager(configDir, homeDir, nil)
	statuses, err := dm.Reconcile()
	if err != nil {
		t.Fatalf("Reconcile() error: %v", err)
	}

	t.Logf("configDir: %s", configDir)
	t.Logf("homeDir: %s", homeDir)
	t.Logf("Statuses count: %d", len(statuses))

	missingCount := 0
	for _, s := range statuses {
		if s.State == SyncStateMissing {
			missingCount++
			t.Logf("  - Missing: %s (source: %s, target: %s)", s.Name, s.Source, s.Target)
		}
	}

	// Should find 2 missing files (not deployed yet)
	if missingCount != 2 {
		t.Errorf("Expected 2 missing files, got %d", missingCount)
	}
}

func TestReconcile_WithDefaultConfig(t *testing.T) {
	// Create temp directories
	configDir, err := os.MkdirTemp("", "plonk-config-*")
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	defer os.RemoveAll(configDir)

	homeDir, err := os.MkdirTemp("", "plonk-home-*")
	if err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}
	defer os.RemoveAll(homeDir)

	// Create some test files
	if err := os.WriteFile(filepath.Join(configDir, "zshrc"), []byte("# test"), 0644); err != nil {
		t.Fatalf("Failed to create zshrc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "vimrc"), []byte("# test"), 0644); err != nil {
		t.Fatalf("Failed to create vimrc: %v", err)
	}

	// Use LoadWithDefaults like the command does
	cfg := config.LoadWithDefaults(configDir)

	dm := NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)
	statuses, err := dm.Reconcile()
	if err != nil {
		t.Fatalf("Reconcile() error: %v", err)
	}

	t.Logf("configDir: %s", configDir)
	t.Logf("homeDir: %s", homeDir)
	t.Logf("IgnorePatterns: %v", cfg.IgnorePatterns)
	t.Logf("Statuses count: %d", len(statuses))

	missingCount := 0
	for _, s := range statuses {
		if s.State == SyncStateMissing {
			missingCount++
			t.Logf("  - Missing: %s (source: %s, target: %s)", s.Name, s.Source, s.Target)
		}
	}

	// Should find 2 missing files (not deployed yet)
	if missingCount != 2 {
		t.Errorf("Expected 2 missing files, got %d", missingCount)
	}
}

func TestReconcile_WithRealUserHome(t *testing.T) {
	// Create temp config directory
	configDir, err := os.MkdirTemp("", "plonk-config-*")
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	defer os.RemoveAll(configDir)

	// Use real user's home directory like the command does
	homeDir, err := config.GetHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Create some test files
	if err := os.WriteFile(filepath.Join(configDir, "test-plonk-file"), []byte("# test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Use LoadWithDefaults like the command does
	cfg := config.LoadWithDefaults(configDir)

	dm := NewDotfileManager(configDir, homeDir, cfg.IgnorePatterns)
	statuses, err := dm.Reconcile()
	if err != nil {
		t.Fatalf("Reconcile() error: %v", err)
	}

	t.Logf("configDir: %s", configDir)
	t.Logf("homeDir: %s", homeDir)
	t.Logf("Statuses count: %d", len(statuses))

	missingCount := 0
	for _, s := range statuses {
		if s.State == SyncStateMissing {
			missingCount++
			t.Logf("  - Missing: %s (source: %s, target: %s)", s.Name, s.Source, s.Target)
		}
	}

	// Should find 1 missing file
	if missingCount != 1 {
		t.Errorf("Expected 1 missing file, got %d", missingCount)
	}
}
