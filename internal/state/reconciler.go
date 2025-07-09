// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package state

import (
	"context"
	"fmt"
)

// Provider defines the interface for any state provider (packages, dotfiles, etc.)
type Provider interface {
	// Domain returns the domain name (e.g., "package", "dotfile")
	Domain() string

	// GetConfiguredItems returns items defined in configuration
	GetConfiguredItems() ([]ConfigItem, error)

	// GetActualItems returns items currently present in the system
	GetActualItems(ctx context.Context) ([]ActualItem, error)

	// CreateItem creates an Item from configured and actual data
	CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item
}

// ConfigItem represents an item as defined in configuration
type ConfigItem struct {
	Name     string
	Metadata map[string]interface{}
}

// ActualItem represents an item as it exists in the system
type ActualItem struct {
	Name     string
	Path     string
	Metadata map[string]interface{}
}

// Reconciler performs state reconciliation for any provider
type Reconciler struct {
	providers map[string]Provider
}

// NewReconciler creates a new universal state reconciler
func NewReconciler() *Reconciler {
	return &Reconciler{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers a state provider for a specific domain
func (r *Reconciler) RegisterProvider(domain string, provider Provider) {
	r.providers[domain] = provider
}

// ReconcileAll reconciles state for all registered providers
func (r *Reconciler) ReconcileAll(ctx context.Context) (Summary, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return Summary{}, ctx.Err()
	default:
	}

	summary := Summary{
		Results: make([]Result, 0),
	}

	for domain := range r.providers {
		// Check context before each provider
		select {
		case <-ctx.Done():
			return Summary{}, ctx.Err()
		default:
		}

		result, err := r.ReconcileProvider(ctx, domain)
		if err != nil {
			return summary, fmt.Errorf("failed to reconcile %s: %w", domain, err)
		}

		result.AddToSummary(&summary)
	}

	return summary, nil
}

// ReconcileProvider reconciles state for a specific provider/domain
func (r *Reconciler) ReconcileProvider(ctx context.Context, domain string) (Result, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
	}

	provider, exists := r.providers[domain]
	if !exists {
		return Result{}, fmt.Errorf("provider for domain %s not found", domain)
	}

	// Get configured items
	configuredItems, err := provider.GetConfiguredItems()
	if err != nil {
		return Result{}, fmt.Errorf("failed to get configured items: %w", err)
	}

	// Check context before getting actual items
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
	}

	// Get actual items
	actualItems, err := provider.GetActualItems(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("failed to get actual items: %w", err)
	}

	// Perform reconciliation
	return r.reconcileItems(provider, configuredItems, actualItems), nil
}

// reconcileItems performs the core reconciliation logic
func (r *Reconciler) reconcileItems(provider Provider, configured []ConfigItem, actual []ActualItem) Result {
	// Build lookup sets
	actualSet := make(map[string]*ActualItem)
	for i := range actual {
		actualSet[actual[i].Name] = &actual[i]
	}

	configuredSet := make(map[string]*ConfigItem)
	for i := range configured {
		configuredSet[configured[i].Name] = &configured[i]
	}

	result := Result{
		Domain:    provider.Domain(),
		Managed:   make([]Item, 0),
		Missing:   make([]Item, 0),
		Untracked: make([]Item, 0),
	}

	// Check each configured item against actual
	for _, configItem := range configured {
		if actualItem, exists := actualSet[configItem.Name]; exists {
			// Item is managed (in config AND present)
			item := provider.CreateItem(configItem.Name, StateManaged, &configItem, actualItem)
			result.Managed = append(result.Managed, item)
		} else {
			// Item is missing (in config BUT not present)
			item := provider.CreateItem(configItem.Name, StateMissing, &configItem, nil)
			result.Missing = append(result.Missing, item)
		}
	}

	// Check each actual item against configured
	for _, actualItem := range actual {
		if _, exists := configuredSet[actualItem.Name]; !exists {
			// Item is untracked (present BUT not in config)
			item := provider.CreateItem(actualItem.Name, StateUntracked, nil, &actualItem)
			result.Untracked = append(result.Untracked, item)
		}
	}

	return result
}

// GetProvider returns the provider for a given domain
func (r *Reconciler) GetProvider(domain string) (Provider, bool) {
	provider, exists := r.providers[domain]
	return provider, exists
}

// GetDomains returns all registered domain names
func (r *Reconciler) GetDomains() []string {
	domains := make([]string, 0, len(r.providers))
	for domain := range r.providers {
		domains = append(domains, domain)
	}
	return domains
}
