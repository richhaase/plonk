// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"errors"
	"testing"
)

// MockProvider implements Provider for testing
type MockProvider struct {
	domain         string
	configuredItems []ConfigItem
	actualItems    []ActualItem
	configError    error
	actualError    error
}

func NewMockProvider(domain string) *MockProvider {
	return &MockProvider{
		domain:         domain,
		configuredItems: []ConfigItem{},
		actualItems:    []ActualItem{},
	}
}

func (m *MockProvider) Domain() string {
	return m.domain
}

func (m *MockProvider) GetConfiguredItems() ([]ConfigItem, error) {
	if m.configError != nil {
		return nil, m.configError
	}
	return m.configuredItems, nil
}

func (m *MockProvider) GetActualItems(ctx context.Context) ([]ActualItem, error) {
	if m.actualError != nil {
		return nil, m.actualError
	}
	return m.actualItems, nil
}

func (m *MockProvider) CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item {
	item := Item{
		Name:   name,
		State:  state,
		Domain: m.domain,
	}
	
	if configured != nil {
		item.Metadata = configured.Metadata
	}
	if actual != nil {
		item.Path = actual.Path
		if item.Metadata == nil {
			item.Metadata = make(map[string]interface{})
		}
		for k, v := range actual.Metadata {
			item.Metadata[k] = v
		}
	}
	
	return item
}

func (m *MockProvider) SetConfiguredItems(items []ConfigItem) {
	m.configuredItems = items
}

func (m *MockProvider) SetActualItems(items []ActualItem) {
	m.actualItems = items
}

func (m *MockProvider) SetConfigError(err error) {
	m.configError = err
}

func (m *MockProvider) SetActualError(err error) {
	m.actualError = err
}

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
	reconciler := NewReconciler()
	provider := NewMockProvider("test")
	
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
	reconciler := NewReconciler()
	provider := NewMockProvider("test")
	provider.SetConfigError(errors.New("config load failed"))
	
	reconciler.RegisterProvider("test", provider)
	
	_, err := reconciler.ReconcileProvider(context.Background(), "test")
	if err == nil {
		t.Error("ReconcileProvider() should return error when config loading fails")
	}
	
	if !errors.Is(err, provider.configError) {
		t.Errorf("ReconcileProvider() should wrap config error")
	}
}

func TestReconciler_ReconcileProvider_ActualError(t *testing.T) {
	reconciler := NewReconciler()
	provider := NewMockProvider("test")
	provider.SetActualError(errors.New("actual items load failed"))
	
	reconciler.RegisterProvider("test", provider)
	
	_, err := reconciler.ReconcileProvider(context.Background(), "test")
	if err == nil {
		t.Error("ReconcileProvider() should return error when actual items loading fails")
	}
	
	if !errors.Is(err, provider.actualError) {
		t.Errorf("ReconcileProvider() should wrap actual error")
	}
}

func TestReconciler_ReconcileProvider_Success(t *testing.T) {
	reconciler := NewReconciler()
	provider := NewMockProvider("test")
	
	// Set up test data
	provider.SetConfiguredItems([]ConfigItem{
		{Name: "item1", Metadata: map[string]interface{}{"config": "data1"}},
		{Name: "item2", Metadata: map[string]interface{}{"config": "data2"}},
		{Name: "missing", Metadata: map[string]interface{}{"config": "missing"}},
	})
	
	provider.SetActualItems([]ActualItem{
		{Name: "item1", Path: "/path/item1", Metadata: map[string]interface{}{"actual": "data1"}},
		{Name: "item2", Path: "/path/item2", Metadata: map[string]interface{}{"actual": "data2"}},
		{Name: "untracked", Path: "/path/untracked", Metadata: map[string]interface{}{"actual": "untracked"}},
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
	reconciler := NewReconciler()
	
	// Set up first provider
	provider1 := NewMockProvider("package")
	provider1.SetConfiguredItems([]ConfigItem{
		{Name: "git", Metadata: map[string]interface{}{"manager": "homebrew"}},
		{Name: "curl", Metadata: map[string]interface{}{"manager": "homebrew"}},
	})
	provider1.SetActualItems([]ActualItem{
		{Name: "git", Path: "/usr/local/bin/git"},
		{Name: "wget", Path: "/usr/local/bin/wget"}, // untracked
	})
	
	// Set up second provider
	provider2 := NewMockProvider("dotfile")
	provider2.SetConfiguredItems([]ConfigItem{
		{Name: ".zshrc", Metadata: map[string]interface{}{"source": "zshrc"}},
	})
	provider2.SetActualItems([]ActualItem{
		{Name: ".zshrc", Path: "/home/user/.zshrc"},
		{Name: ".vimrc", Path: "/home/user/.vimrc"}, // untracked
	})
	
	reconciler.RegisterProvider("package", provider1)
	reconciler.RegisterProvider("dotfile", provider2)
	
	summary, err := reconciler.ReconcileAll(context.Background())
	if err != nil {
		t.Fatalf("ReconcileAll() failed: %v", err)
	}
	
	// Verify aggregated counts
	if summary.TotalManaged != 2 { // git + .zshrc
		t.Errorf("Summary.TotalManaged = %d, expected 2", summary.TotalManaged)
	}
	if summary.TotalMissing != 1 { // curl
		t.Errorf("Summary.TotalMissing = %d, expected 1", summary.TotalMissing)
	}
	if summary.TotalUntracked != 2 { // wget + .vimrc
		t.Errorf("Summary.TotalUntracked = %d, expected 2", summary.TotalUntracked)
	}
	
	// Verify results structure
	if len(summary.Results) != 2 {
		t.Errorf("len(Summary.Results) = %d, expected 2", len(summary.Results))
	}
	
	// Verify domain names are included
	domains := make(map[string]bool)
	for _, result := range summary.Results {
		domains[result.Domain] = true
	}
	
	if !domains["package"] || !domains["dotfile"] {
		t.Error("Expected both package and dotfile domains in results")
	}
}

func TestReconciler_ReconcileAll_ProviderError(t *testing.T) {
	reconciler := NewReconciler()
	
	provider1 := NewMockProvider("working")
	provider2 := NewMockProvider("failing")
	provider2.SetConfigError(errors.New("provider failed"))
	
	reconciler.RegisterProvider("working", provider1)
	reconciler.RegisterProvider("failing", provider2)
	
	_, err := reconciler.ReconcileAll(context.Background())
	if err == nil {
		t.Error("ReconcileAll() should fail when one provider fails")
	}
	
	expectedSubstring := "failed to reconcile failing"
	if !contains(err.Error(), expectedSubstring) {
		t.Errorf("ReconcileAll() error should contain %q, got: %s", expectedSubstring, err.Error())
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestReconciler_ContextCancellation(t *testing.T) {
	reconciler := NewReconciler()
	
	t.Run("ReconcileProvider_ContextCancellation", func(t *testing.T) {
		provider := NewMockProvider("test")
		provider.SetConfiguredItems([]ConfigItem{
			{Name: "item1", Metadata: map[string]interface{}{"config": "data1"}},
		})
		provider.SetActualItems([]ActualItem{
			{Name: "item1", Path: "/path/item1", Metadata: map[string]interface{}{"actual": "data1"}},
		})
		
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
	
	t.Run("ReconcileAll_ContextCancellation", func(t *testing.T) {
		provider := NewMockProvider("test")
		provider.SetConfiguredItems([]ConfigItem{
			{Name: "item1", Metadata: map[string]interface{}{"config": "data1"}},
		})
		provider.SetActualItems([]ActualItem{
			{Name: "item1", Path: "/path/item1", Metadata: map[string]interface{}{"actual": "data1"}},
		})
		
		reconciler.RegisterProvider("test", provider)
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := reconciler.ReconcileAll(ctx)
		if err == nil {
			t.Error("Expected error when context is cancelled")
		}
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
}