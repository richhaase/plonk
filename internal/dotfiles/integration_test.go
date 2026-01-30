package dotfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestReconcileWithConfig_RealFS(t *testing.T) {
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

	// Create config with nil patterns
	cfg := &config.Config{
		IgnorePatterns: nil,
	}

	ctx := context.Background()
	result, err := ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		t.Fatalf("ReconcileWithConfig() error: %v", err)
	}

	t.Logf("configDir: %s", configDir)
	t.Logf("homeDir: %s", homeDir)
	t.Logf("Domain: %s", result.Domain)
	t.Logf("Managed: %d", len(result.Managed))
	t.Logf("Missing: %d", len(result.Missing))
	t.Logf("Drifted: %d", len(result.Drifted))
	for _, item := range result.Missing {
		t.Logf("  - Missing: %s (source: %s, dest: %s)", item.Name, item.Source, item.Destination)
	}

	// Should find 2 missing files (not deployed yet)
	if len(result.Missing) != 2 {
		t.Errorf("Expected 2 missing files, got %d", len(result.Missing))
	}
}

func TestReconcileWithConfig_WithDefaultConfig(t *testing.T) {
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

	ctx := context.Background()
	result, err := ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		t.Fatalf("ReconcileWithConfig() error: %v", err)
	}

	t.Logf("configDir: %s", configDir)
	t.Logf("homeDir: %s", homeDir)
	t.Logf("IgnorePatterns: %v", cfg.IgnorePatterns)
	t.Logf("Domain: %s", result.Domain)
	t.Logf("Managed: %d", len(result.Managed))
	t.Logf("Missing: %d", len(result.Missing))
	t.Logf("Drifted: %d", len(result.Drifted))
	for _, item := range result.Missing {
		t.Logf("  - Missing: %s (source: %s, dest: %s)", item.Name, item.Source, item.Destination)
	}

	// Should find 2 missing files (not deployed yet)
	if len(result.Missing) != 2 {
		t.Errorf("Expected 2 missing files, got %d", len(result.Missing))
	}
}

func TestReconcileWithConfig_WithRealUserHome(t *testing.T) {
	// Create temp config directory
	configDir, err := os.MkdirTemp("", "plonk-config-*")
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	defer os.RemoveAll(configDir)

	// Use real user's home directory like the command does
	homeDir := config.GetHomeDir()

	// Create some test files
	if err := os.WriteFile(filepath.Join(configDir, "test-plonk-file"), []byte("# test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Use LoadWithDefaults like the command does
	cfg := config.LoadWithDefaults(configDir)

	ctx := context.Background()
	result, err := ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		t.Fatalf("ReconcileWithConfig() error: %v", err)
	}

	t.Logf("configDir: %s", configDir)
	t.Logf("homeDir: %s", homeDir)
	t.Logf("Domain: %s", result.Domain)
	t.Logf("Managed: %d", len(result.Managed))
	t.Logf("Missing: %d", len(result.Missing))
	t.Logf("Drifted: %d", len(result.Drifted))
	for _, item := range result.Missing {
		t.Logf("  - Missing: %s (source: %s, dest: %s)", item.Name, item.Source, item.Destination)
	}

	// Should find 1 missing file
	if len(result.Missing) != 1 {
		t.Errorf("Expected 1 missing file, got %d", len(result.Missing))
	}
}
