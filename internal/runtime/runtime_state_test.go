// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestNewRuntimeState(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	rs := NewRuntimeState(tmpDir, homeDir)

	if rs == nil {
		t.Fatal("NewRuntimeState returned nil")
	}

	impl := rs.(*RuntimeStateImpl)
	if impl.configDir != tmpDir {
		t.Errorf("Expected configDir %s, got %s", tmpDir, impl.configDir)
	}
	if impl.homeDir != homeDir {
		t.Errorf("Expected homeDir %s, got %s", homeDir, impl.homeDir)
	}
}

func TestRuntimeState_LoadConfiguration(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	rs := NewRuntimeState(tmpDir, homeDir)

	// Should succeed even with no config file (creates default)
	err := rs.LoadConfiguration()
	if err != nil {
		t.Fatalf("LoadConfiguration failed: %v", err)
	}

	// Should be able to get config
	cfg := rs.GetConfig()
	if cfg == nil {
		t.Error("GetConfig returned nil after LoadConfiguration")
	}
}

func TestRuntimeState_SaveConfiguration(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	rs := NewRuntimeState(tmpDir, homeDir)

	// Should be able to save even without loading first
	err := rs.SaveConfiguration()
	if err != nil {
		t.Fatalf("SaveConfiguration failed: %v", err)
	}

	// Config file should exist
	configPath := filepath.Join(tmpDir, "plonk.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestRuntimeState_GetProviders(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	rs := NewRuntimeState(tmpDir, homeDir)

	// Load configuration first
	err := rs.LoadConfiguration()
	if err != nil {
		t.Fatalf("LoadConfiguration failed: %v", err)
	}

	// Should be able to get providers
	dotfileProvider := rs.GetDotfileProvider()
	if dotfileProvider == nil {
		t.Error("GetDotfileProvider returned nil")
	}

	packageProvider := rs.GetPackageProvider()
	if packageProvider == nil {
		t.Error("GetPackageProvider returned nil")
	}

	// Check domains
	if dotfileProvider.Domain() != "dotfile" {
		t.Errorf("Expected dotfile domain, got %s", dotfileProvider.Domain())
	}

	if packageProvider.Domain() != "package" {
		t.Errorf("Expected package domain, got %s", packageProvider.Domain())
	}
}

func TestRuntimeState_ReconcileDotfiles(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	rs := NewRuntimeState(tmpDir, homeDir)

	// Load configuration first
	err := rs.LoadConfiguration()
	if err != nil {
		t.Fatalf("LoadConfiguration failed: %v", err)
	}

	ctx := context.Background()
	result, err := rs.ReconcileDotfiles(ctx)
	if err != nil {
		t.Fatalf("ReconcileDotfiles failed: %v", err)
	}

	if result.Domain != "dotfile" {
		t.Errorf("Expected dotfile domain in result, got %s", result.Domain)
	}
}

func TestRuntimeState_ReconcilePackages(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	rs := NewRuntimeState(tmpDir, homeDir)

	// Load configuration first
	err := rs.LoadConfiguration()
	if err != nil {
		t.Fatalf("LoadConfiguration failed: %v", err)
	}

	ctx := context.Background()
	result, err := rs.ReconcilePackages(ctx)
	if err != nil {
		t.Fatalf("ReconcilePackages failed: %v", err)
	}

	if result.Domain != "package" {
		t.Errorf("Expected package domain in result, got %s", result.Domain)
	}
}

func TestRuntimeState_ReconcileAll(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	rs := NewRuntimeState(tmpDir, homeDir)

	// Load configuration first
	err := rs.LoadConfiguration()
	if err != nil {
		t.Fatalf("LoadConfiguration failed: %v", err)
	}

	ctx := context.Background()
	results, err := rs.ReconcileAll(ctx)
	if err != nil {
		t.Fatalf("ReconcileAll failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 domains in results, got %d", len(results))
	}

	if _, ok := results["dotfile"]; !ok {
		t.Error("Expected dotfile domain in results")
	}

	if _, ok := results["package"]; !ok {
		t.Error("Expected package domain in results")
	}
}

func TestRuntimeConfigAdapter(t *testing.T) {
	cfg := &config.Config{
		IgnorePatterns: []string{".git", "*.tmp"},
	}

	adapter := NewRuntimeConfigAdapter(cfg)

	// Test ignore patterns
	patterns := adapter.GetIgnorePatterns()
	if len(patterns) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(patterns))
	}

	// Note: GetDotfileTargets() auto-discovers files from the actual config directory,
	// so we skip testing it here as it would depend on the real filesystem state.

	// Test package methods (should return empty - packages in lock file)
	brews := adapter.GetHomebrewBrews()
	if len(brews) != 0 {
		t.Errorf("Expected empty homebrew brews, got %d", len(brews))
	}

	casks := adapter.GetHomebrewCasks()
	if len(casks) != 0 {
		t.Errorf("Expected empty homebrew casks, got %d", len(casks))
	}

	npm := adapter.GetNPMPackages()
	if len(npm) != 0 {
		t.Errorf("Expected empty npm packages, got %d", len(npm))
	}
}

func TestRuntimeConfigAdapter_WithDefaults(t *testing.T) {
	// Test with nil config fields to ensure defaults are used
	cfg := &config.Config{}

	adapter := NewRuntimeConfigAdapter(cfg)

	// Should get default ignore patterns
	patterns := adapter.GetIgnorePatterns()
	if len(patterns) == 0 {
		t.Error("Expected default ignore patterns when config is empty")
	}

	// Should get default expand directories
	dirs := adapter.GetExpandDirectories()
	if len(dirs) == 0 {
		t.Error("Expected default expand directories when config is empty")
	}
}
