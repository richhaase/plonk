// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestNewPackageProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := NewMockPackageManager(ctrl)
	configLoader := NewMockPackageConfigLoader(ctrl)

	provider := NewPackageProvider("homebrew", manager, configLoader)

	if provider == nil {
		t.Fatal("NewPackageProvider() returned nil")
	}

	if provider.managerName != "homebrew" {
		t.Errorf("provider.managerName = %s, expected homebrew", provider.managerName)
	}

	// Note: Cannot directly compare interface types in tests
	// The manager is properly set through the constructor

	if provider.configLoader != configLoader {
		t.Error("provider.configLoader not set correctly")
	}
}

func TestPackageProvider_Domain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := NewMockPackageManager(ctrl)
	configLoader := NewMockPackageConfigLoader(ctrl)
	provider := NewPackageProvider("homebrew", manager, configLoader)

	domain := provider.Domain()
	if domain != "package" {
		t.Errorf("Domain() = %s, expected package", domain)
	}
}

func TestPackageProvider_GetConfiguredItems_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := NewMockPackageManager(ctrl)
	configLoader := NewMockPackageConfigLoader(ctrl)

	// Set up expectations
	configLoader.EXPECT().GetPackagesForManager("homebrew").Return([]PackageConfigItem{
		{Name: "git"},
		{Name: "curl"},
		{Name: "jq"},
	}, nil)

	provider := NewPackageProvider("homebrew", manager, configLoader)

	items, err := provider.GetConfiguredItems()
	if err != nil {
		t.Fatalf("GetConfiguredItems() failed: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("GetConfiguredItems() returned %d items, expected 3", len(items))
	}

	// Verify item structure
	expectedNames := map[string]bool{"git": true, "curl": true, "jq": true}
	for _, item := range items {
		if !expectedNames[item.Name] {
			t.Errorf("Unexpected item name: %s", item.Name)
		}

		if item.Metadata == nil {
			t.Errorf("Item %s has nil metadata", item.Name)
		} else if item.Metadata["manager"] != "homebrew" {
			t.Errorf("Item %s has manager %v, expected homebrew", item.Name, item.Metadata["manager"])
		}

		delete(expectedNames, item.Name)
	}

	if len(expectedNames) > 0 {
		t.Errorf("Missing expected items: %v", expectedNames)
	}
}

func TestPackageProvider_GetConfiguredItems_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := NewMockPackageManager(ctrl)
	configLoader := NewMockPackageConfigLoader(ctrl)

	configLoader.EXPECT().GetPackagesForManager("homebrew").Return(nil, errors.New("config load failed"))

	provider := NewPackageProvider("homebrew", manager, configLoader)

	_, err := provider.GetConfiguredItems()
	if err == nil {
		t.Error("GetConfiguredItems() should return error when config loading fails")
	}
}

func TestPackageProvider_GetActualItems_ManagerNotAvailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := NewMockPackageManager(ctrl)
	configLoader := NewMockPackageConfigLoader(ctrl)

	manager.EXPECT().IsAvailable(gomock.Any()).Return(false, nil)

	provider := NewPackageProvider("homebrew", manager, configLoader)

	items, err := provider.GetActualItems(context.Background())
	if err != nil {
		t.Fatalf("GetActualItems() failed: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("GetActualItems() with unavailable manager returned %d items, expected 0", len(items))
	}
}

func TestPackageProvider_GetActualItems_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := NewMockPackageManager(ctrl)
	configLoader := NewMockPackageConfigLoader(ctrl)

	manager.EXPECT().IsAvailable(gomock.Any()).Return(true, nil)
	manager.EXPECT().ListInstalled(gomock.Any()).Return([]string{"git", "curl", "wget"}, nil)

	provider := NewPackageProvider("homebrew", manager, configLoader)

	items, err := provider.GetActualItems(context.Background())
	if err != nil {
		t.Fatalf("GetActualItems() failed: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("GetActualItems() returned %d items, expected 3", len(items))
	}

	// Verify item structure
	expectedNames := map[string]bool{"git": true, "curl": true, "wget": true}
	for _, item := range items {
		if !expectedNames[item.Name] {
			t.Errorf("Unexpected item name: %s", item.Name)
		}

		if item.Metadata == nil {
			t.Errorf("Item %s has nil metadata", item.Name)
		} else if item.Metadata["manager"] != "homebrew" {
			t.Errorf("Item %s has manager %v, expected homebrew", item.Name, item.Metadata["manager"])
		}

		delete(expectedNames, item.Name)
	}

	if len(expectedNames) > 0 {
		t.Errorf("Missing expected items: %v", expectedNames)
	}
}

func TestPackageProvider_GetActualItems_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := NewMockPackageManager(ctrl)
	configLoader := NewMockPackageConfigLoader(ctrl)

	manager.EXPECT().IsAvailable(gomock.Any()).Return(true, nil)
	manager.EXPECT().ListInstalled(gomock.Any()).Return(nil, errors.New("list failed"))

	provider := NewPackageProvider("homebrew", manager, configLoader)

	_, err := provider.GetActualItems(context.Background())
	if err == nil {
		t.Error("GetActualItems() should return error when listing fails")
	}
}

func TestPackageProvider_CreateItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := NewMockPackageManager(ctrl)
	configLoader := NewMockPackageConfigLoader(ctrl)
	provider := NewPackageProvider("npm", manager, configLoader)

	tests := []struct {
		name             string
		state            ItemState
		configured       *ConfigItem
		actual           *ActualItem
		expectedName     string
		expectedMetadata map[string]interface{}
	}{
		{
			name:         "managed item",
			state:        StateManaged,
			configured:   &ConfigItem{Name: "test", Metadata: map[string]interface{}{"config": "data"}},
			actual:       &ActualItem{Name: "test", Metadata: map[string]interface{}{"actual": "data"}},
			expectedName: "test",
			expectedMetadata: map[string]interface{}{
				"manager": "npm", // Always added by CreateItem
				"config":  "data",
				"actual":  "data",
			},
		},
		{
			name:         "missing item",
			state:        StateMissing,
			configured:   &ConfigItem{Name: "missing", Metadata: map[string]interface{}{"config": "data"}},
			actual:       nil,
			expectedName: "missing",
			expectedMetadata: map[string]interface{}{
				"manager": "npm", // Always added by CreateItem
				"config":  "data",
			},
		},
		{
			name:         "untracked item",
			state:        StateUntracked,
			configured:   nil,
			actual:       &ActualItem{Name: "untracked", Path: "/path", Metadata: map[string]interface{}{"actual": "data"}},
			expectedName: "untracked",
			expectedMetadata: map[string]interface{}{
				"manager": "npm", // Always added by CreateItem
				"actual":  "data",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := provider.CreateItem(test.expectedName, test.state, test.configured, test.actual)

			if item.Name != test.expectedName {
				t.Errorf("item.Name = %s, expected %s", item.Name, test.expectedName)
			}

			if item.State != test.state {
				t.Errorf("item.State = %s, expected %s", item.State, test.state)
			}

			if item.Domain != "package" {
				t.Errorf("item.Domain = %s, expected package", item.Domain)
			}

			if item.Manager != "npm" {
				t.Errorf("item.Manager = %s, expected npm", item.Manager)
			}

			// Package items don't typically set Path from ActualItem
			// The Path field is mainly used for dotfiles

			// Verify metadata
			if len(item.Metadata) != len(test.expectedMetadata) {
				t.Errorf("item.Metadata has %d keys, expected %d", len(item.Metadata), len(test.expectedMetadata))
			}

			for key, expectedValue := range test.expectedMetadata {
				if actualValue, exists := item.Metadata[key]; !exists {
					t.Errorf("item.Metadata missing key %s", key)
				} else if actualValue != expectedValue {
					t.Errorf("item.Metadata[%s] = %v, expected %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestPackageProvider_ContextCancellation(t *testing.T) {
	t.Run("GetActualItems_ContextCancellation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		manager := NewMockPackageManager(ctrl)
		configLoader := NewMockPackageConfigLoader(ctrl)
		provider := NewPackageProvider("homebrew", manager, configLoader)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Context is checked before calling manager methods, so they should not be called

		_, err := provider.GetActualItems(ctx)
		if err == nil {
			t.Error("Expected error when context is canceled")
		}
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
}
