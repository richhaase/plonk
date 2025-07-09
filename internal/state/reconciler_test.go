// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestNewReconciler(t *testing.T) {
	reconciler := NewReconciler()
	
	if reconciler == nil {
		t.Fatal("NewReconciler() returned nil")
	}
	
	if reconciler.providers == nil {
		t.Fatal("NewReconciler() created reconciler with nil providers map")
	}
	
	if len(reconciler.providers) != 0 {
		t.Errorf("NewReconciler() created reconciler with %d providers, expected 0", len(reconciler.providers))
	}
}

func TestReconciler_RegisterProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reconciler := NewReconciler()
	provider := NewMockProvider(ctrl)
	
	reconciler.RegisterProvider("test", provider)
	
	domains := reconciler.GetDomains()
	if len(domains) != 1 {
		t.Errorf("GetDomains() returned %d domains, expected 1", len(domains))
	}
	if domains[0] != "test" {
		t.Errorf("GetDomains() returned %s, expected test", domains[0])
	}
	
	_, exists := reconciler.GetProvider("test")
	if !exists {
		t.Error("GetProvider() returned false for registered provider")
	}
	// Note: Cannot directly compare interface types in tests
	// The provider is properly registered through the RegisterProvider method
}

func TestReconciler_GetProvider_NotFound(t *testing.T) {
	reconciler := NewReconciler()
	
	_, exists := reconciler.GetProvider("nonexistent")
	if exists {
		t.Error("GetProvider() returned true for non-existent provider")
	}
}

func TestReconciler_ReconcileProvider_NotFound(t *testing.T) {
	reconciler := NewReconciler()
	
	_, err := reconciler.ReconcileProvider(context.Background(), "nonexistent")
	if err == nil {
		t.Error("ReconcileProvider() should return error for non-existent provider")
	}
	
	expectedError := "provider for domain nonexistent not found"
	if err.Error() != expectedError {
		t.Errorf("ReconcileProvider() error = %q, expected %q", err.Error(), expectedError)
	}
}

func TestReconciler_ReconcileProvider_ConfigError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reconciler := NewReconciler()
	provider := NewMockProvider(ctrl)
	configError := errors.New("config load failed")
	
	provider.EXPECT().GetConfiguredItems().Return(nil, configError)
	
	reconciler.RegisterProvider("test", provider)
	
	_, err := reconciler.ReconcileProvider(context.Background(), "test")
	if err == nil {
		t.Error("ReconcileProvider() should return error when config loading fails")
	}
	
	if !errors.Is(err, configError) {
		t.Errorf("ReconcileProvider() should wrap config error")
	}
}

func TestReconciler_ReconcileProvider_ActualError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reconciler := NewReconciler()
	provider := NewMockProvider(ctrl)
	actualError := errors.New("actual items load failed")
	
	provider.EXPECT().GetConfiguredItems().Return([]ConfigItem{}, nil)
	provider.EXPECT().GetActualItems(gomock.Any()).Return(nil, actualError)
	
	reconciler.RegisterProvider("test", provider)
	
	_, err := reconciler.ReconcileProvider(context.Background(), "test")
	if err == nil {
		t.Error("ReconcileProvider() should return error when actual items loading fails")
	}
	
	if !errors.Is(err, actualError) {
		t.Errorf("ReconcileProvider() should wrap actual error")
	}
}

func TestReconciler_ReconcileProvider_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reconciler := NewReconciler()
	provider := NewMockProvider(ctrl)
	
	configItems := []ConfigItem{
		{Name: "item1", Metadata: map[string]interface{}{"config": "data1"}},
		{Name: "item2", Metadata: map[string]interface{}{"config": "data2"}},
		{Name: "missing", Metadata: map[string]interface{}{"config": "missing"}},
	}
	
	actualItems := []ActualItem{
		{Name: "item1", Path: "/path/item1", Metadata: map[string]interface{}{"actual": "data1"}},
		{Name: "item2", Path: "/path/item2", Metadata: map[string]interface{}{"actual": "data2"}},
		{Name: "untracked", Path: "/path/untracked", Metadata: map[string]interface{}{"actual": "untracked"}},
	}
	
	provider.EXPECT().GetConfiguredItems().Return(configItems, nil)
	provider.EXPECT().GetActualItems(gomock.Any()).Return(actualItems, nil)
	provider.EXPECT().Domain().Return("test")
	provider.EXPECT().CreateItem("item1", StateManaged, &configItems[0], &actualItems[0]).Return(Item{
		Name: "item1", State: StateManaged, Domain: "test",
	})
	provider.EXPECT().CreateItem("item2", StateManaged, &configItems[1], &actualItems[1]).Return(Item{
		Name: "item2", State: StateManaged, Domain: "test",
	})
	provider.EXPECT().CreateItem("missing", StateMissing, &configItems[2], nil).Return(Item{
		Name: "missing", State: StateMissing, Domain: "test",
	})
	provider.EXPECT().CreateItem("untracked", StateUntracked, nil, &actualItems[2]).Return(Item{
		Name: "untracked", State: StateUntracked, Domain: "test",
	})
	
	reconciler.RegisterProvider("test", provider)
	
	result, err := reconciler.ReconcileProvider(context.Background(), "test")
	if err != nil {
		t.Fatalf("ReconcileProvider() failed: %v", err)
	}
	
	// Verify result structure
	if result.Domain != "test" {
		t.Errorf("Result.Domain = %s, expected test", result.Domain)
	}
	
	// Verify managed items
	if len(result.Managed) != 2 {
		t.Errorf("len(result.Managed) = %d, expected 2", len(result.Managed))
	}
	
	managedNames := make(map[string]bool)
	for _, item := range result.Managed {
		managedNames[item.Name] = true
		if item.State != StateManaged {
			t.Errorf("Managed item %s has state %s, expected managed", item.Name, item.State)
		}
		if item.Domain != "test" {
			t.Errorf("Managed item %s has domain %s, expected test", item.Name, item.Domain)
		}
	}
	
	if !managedNames["item1"] || !managedNames["item2"] {
		t.Error("Expected item1 and item2 to be in managed items")
	}
	
	// Verify missing items
	if len(result.Missing) != 1 {
		t.Errorf("len(result.Missing) = %d, expected 1", len(result.Missing))
	}
	
	if result.Missing[0].Name != "missing" {
		t.Errorf("Missing item name = %s, expected missing", result.Missing[0].Name)
	}
	if result.Missing[0].State != StateMissing {
		t.Errorf("Missing item state = %s, expected missing", result.Missing[0].State)
	}
	
	// Verify untracked items
	if len(result.Untracked) != 1 {
		t.Errorf("len(result.Untracked) = %d, expected 1", len(result.Untracked))
	}
	
	if result.Untracked[0].Name != "untracked" {
		t.Errorf("Untracked item name = %s, expected untracked", result.Untracked[0].Name)
	}
	if result.Untracked[0].State != StateUntracked {
		t.Errorf("Untracked item state = %s, expected untracked", result.Untracked[0].State)
	}
}

func TestReconciler_ReconcileAll_EmptyProviders(t *testing.T) {
	reconciler := NewReconciler()
	
	summary, err := reconciler.ReconcileAll(context.Background())
	if err != nil {
		t.Fatalf("ReconcileAll() failed: %v", err)
	}
	
	if summary.TotalManaged != 0 || summary.TotalMissing != 0 || summary.TotalUntracked != 0 {
		t.Errorf("ReconcileAll() with empty providers should return zero counts")
	}
	
	if len(summary.Results) != 0 {
		t.Errorf("ReconcileAll() with empty providers should return empty results")
	}
}

func TestReconciler_ReconcileAll_MultipleProviders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reconciler := NewReconciler()
	
	// Set up first provider
	provider1 := NewMockProvider(ctrl)
	provider1.EXPECT().GetConfiguredItems().Return([]ConfigItem{
		{Name: "git", Metadata: map[string]interface{}{"manager": "homebrew"}},
		{Name: "curl", Metadata: map[string]interface{}{"manager": "homebrew"}},
	}, nil)
	provider1.EXPECT().GetActualItems(gomock.Any()).Return([]ActualItem{
		{Name: "git", Path: "/usr/local/bin/git"},
		{Name: "wget", Path: "/usr/local/bin/wget"}, // untracked
	}, nil)
	provider1.EXPECT().Domain().Return("package")
	provider1.EXPECT().CreateItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(Item{
		Name: "git", State: StateManaged, Domain: "package",
	})
	provider1.EXPECT().CreateItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(Item{
		Name: "curl", State: StateMissing, Domain: "package",
	})
	provider1.EXPECT().CreateItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(Item{
		Name: "wget", State: StateUntracked, Domain: "package",
	})
	
	// Set up second provider
	provider2 := NewMockProvider(ctrl)
	provider2.EXPECT().GetConfiguredItems().Return([]ConfigItem{
		{Name: "vimrc", Metadata: map[string]interface{}{"destination": "~/.vimrc"}},
	}, nil)
	provider2.EXPECT().GetActualItems(gomock.Any()).Return([]ActualItem{
		{Name: "vimrc", Path: "/home/user/.vimrc"},
	}, nil)
	provider2.EXPECT().Domain().Return("dotfile")
	provider2.EXPECT().CreateItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(Item{
		Name: "vimrc", State: StateManaged, Domain: "dotfile",
	})
	
	reconciler.RegisterProvider("package", provider1)
	reconciler.RegisterProvider("dotfile", provider2)
	
	summary, err := reconciler.ReconcileAll(context.Background())
	if err != nil {
		t.Fatalf("ReconcileAll() failed: %v", err)
	}
	
	// Verify totals
	if summary.TotalManaged != 2 {
		t.Errorf("TotalManaged = %d, expected 2", summary.TotalManaged)
	}
	if summary.TotalMissing != 1 {
		t.Errorf("TotalMissing = %d, expected 1", summary.TotalMissing)
	}
	if summary.TotalUntracked != 1 {
		t.Errorf("TotalUntracked = %d, expected 1", summary.TotalUntracked)
	}
	
	// Verify results
	if len(summary.Results) != 2 {
		t.Errorf("len(Results) = %d, expected 2", len(summary.Results))
	}
	
	resultDomains := make(map[string]bool)
	for _, result := range summary.Results {
		resultDomains[result.Domain] = true
	}
	
	if !resultDomains["package"] || !resultDomains["dotfile"] {
		t.Error("Expected package and dotfile domains in results")
	}
}

func TestReconciler_ReconcileAll_ProviderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reconciler := NewReconciler()
	provider := NewMockProvider(ctrl)
	configError := errors.New("config load failed")
	
	provider.EXPECT().GetConfiguredItems().Return(nil, configError)
	
	reconciler.RegisterProvider("test", provider)
	
	_, err := reconciler.ReconcileAll(context.Background())
	if err == nil {
		t.Error("ReconcileAll() should return error when provider fails")
	}
	
	if !errors.Is(err, configError) {
		t.Errorf("ReconcileAll() should wrap provider error")
	}
}

func TestReconciler_ContextCancellation(t *testing.T) {
	t.Run("ReconcileProvider_ContextCancellation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		reconciler := NewReconciler()
		provider := NewMockProvider(ctrl)
		
		provider.EXPECT().GetConfiguredItems().Return([]ConfigItem{}, nil).MaxTimes(1)
		// Context check happens after GetConfiguredItems, so GetActualItems may not be called
		
		reconciler.RegisterProvider("test", provider)
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := reconciler.ReconcileProvider(ctx, "test")
		if err == nil {
			t.Error("Expected error when context is cancelled")
		}
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
}