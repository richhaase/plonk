// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richhaase/plonk/internal/config"
)

func TestNewOrchestrator(t *testing.T) {
	cfg := &config.Config{}
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "plonk-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	orchestrator := NewOrchestrator(ctx, cfg, tmpDir, tmpDir)

	if orchestrator.ctx != ctx {
		t.Error("Expected context to be set")
	}
	if orchestrator.config != cfg {
		t.Error("Expected config to be set")
	}
	if orchestrator.configDir != tmpDir {
		t.Error("Expected configDir to be set")
	}
	if orchestrator.homeDir != tmpDir {
		t.Error("Expected homeDir to be set")
	}
	if orchestrator.hookRunner == nil {
		t.Error("Expected hookRunner to be set")
	}
	if orchestrator.lock == nil {
		t.Error("Expected lock to be set")
	}
}

func TestOrchestrator_HookConfiguration(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "plonk-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name   string
		config *config.Config
	}{
		{
			name:   "No hooks",
			config: &config.Config{},
		},
		{
			name: "Pre-sync hooks only",
			config: &config.Config{
				Hooks: config.Hooks{
					PreSync: []config.Hook{
						{Command: "echo 'pre-sync'"},
					},
				},
			},
		},
		{
			name: "Post-sync hooks only",
			config: &config.Config{
				Hooks: config.Hooks{
					PostSync: []config.Hook{
						{Command: "echo 'post-sync'"},
					},
				},
			},
		},
		{
			name: "Both pre and post hooks",
			config: &config.Config{
				Hooks: config.Hooks{
					PreSync: []config.Hook{
						{Command: "echo 'pre-sync'"},
					},
					PostSync: []config.Hook{
						{Command: "echo 'post-sync'"},
					},
				},
			},
		},
		{
			name: "Hooks with timeout and continue_on_error",
			config: &config.Config{
				Hooks: config.Hooks{
					PreSync: []config.Hook{
						{
							Command:         "echo 'test'",
							Timeout:         "30s",
							ContinueOnError: true,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orchestrator := NewOrchestrator(ctx, tc.config, tmpDir, tmpDir)

			if orchestrator.config != tc.config {
				t.Error("Expected config to be properly set")
			}

			// Verify hook configuration is accessible
			preSyncCount := len(tc.config.Hooks.PreSync)
			postSyncCount := len(tc.config.Hooks.PostSync)

			if len(orchestrator.config.Hooks.PreSync) != preSyncCount {
				t.Errorf("Expected %d pre-sync hooks, got %d", preSyncCount, len(orchestrator.config.Hooks.PreSync))
			}

			if len(orchestrator.config.Hooks.PostSync) != postSyncCount {
				t.Errorf("Expected %d post-sync hooks, got %d", postSyncCount, len(orchestrator.config.Hooks.PostSync))
			}
		})
	}
}

func TestOrchestrator_GetResources(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{}

	tmpDir, err := os.MkdirTemp("", "plonk-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, "config")
	homeDir := filepath.Join(tmpDir, "home")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	orchestrator := NewOrchestrator(ctx, cfg, configDir, homeDir)
	resources := orchestrator.GetResources()

	// Should have at least package and dotfile resources
	if len(resources) < 2 {
		t.Errorf("Expected at least 2 resources, got %d", len(resources))
	}

	// Verify resource types
	resourceTypes := make(map[string]bool)
	for _, resource := range resources {
		resourceTypes[resource.ID()] = true
	}

	expectedTypes := []string{"packages:all", "dotfiles"}
	for _, expectedType := range expectedTypes {
		if !resourceTypes[expectedType] {
			t.Errorf("Expected resource type '%s' to be present", expectedType)
		}
	}
}
