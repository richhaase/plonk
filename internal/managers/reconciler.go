// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"fmt"
)

// PackageState represents the reconciliation state of a package
type PackageState int

const (
	StateManaged   PackageState = iota // In config AND correctly installed
	StateMissing                       // In config BUT not correctly installed
	StateUntracked                     // Installed BUT not in config
)

// Package represents a package with its current state
type Package struct {
	Name    string
	State   PackageState
	Manager string // Which manager this package belongs to
}

// ConfigPackage represents a package defined in plonk.yaml
type ConfigPackage struct {
	Name string
}

// StateResult contains the results of state reconciliation
type StateResult struct {
	Managed   []Package
	Missing   []Package
	Untracked []Package
}

// ConfigLoader defines how to load configuration
type ConfigLoader interface {
	GetPackagesForManager(managerName string) ([]ConfigPackage, error)
}

// StateReconciler handles comparing configuration vs installed packages
type StateReconciler struct {
	configLoader ConfigLoader
	managers     map[string]PackageManager
}

// NewStateReconciler creates a new state reconciler
func NewStateReconciler(configLoader ConfigLoader, managers map[string]PackageManager) *StateReconciler {
	return &StateReconciler{
		configLoader: configLoader,
		managers:     managers,
	}
}

// ReconcileAll reconciles state for all available managers
func (s *StateReconciler) ReconcileAll() (StateResult, error) {
	var allResult StateResult
	
	for managerName := range s.managers {
		result, err := s.ReconcileManager(managerName)
		if err != nil {
			return StateResult{}, fmt.Errorf("failed to reconcile %s: %w", managerName, err)
		}
		
		// Merge results
		allResult.Managed = append(allResult.Managed, result.Managed...)
		allResult.Missing = append(allResult.Missing, result.Missing...)
		allResult.Untracked = append(allResult.Untracked, result.Untracked...)
	}
	
	return allResult, nil
}

// ReconcileManager reconciles state for a specific manager
func (s *StateReconciler) ReconcileManager(managerName string) (StateResult, error) {
	manager, exists := s.managers[managerName]
	if !exists {
		return StateResult{}, fmt.Errorf("manager %s not found", managerName)
	}
	
	if !manager.IsAvailable() {
		// Manager not available, return empty result
		return StateResult{}, nil
	}
	
	// Get installed packages
	installed, err := manager.ListInstalled()
	if err != nil {
		return StateResult{}, fmt.Errorf("failed to list installed packages: %w", err)
	}
	
	// Get config packages
	configPackages, err := s.configLoader.GetPackagesForManager(managerName)
	if err != nil {
		return StateResult{}, fmt.Errorf("failed to load config packages: %w", err)
	}
	
	// Perform reconciliation
	return s.reconcilePackages(managerName, installed, configPackages), nil
}

// reconcilePackages performs the actual reconciliation logic
func (s *StateReconciler) reconcilePackages(managerName string, installed []string, configPackages []ConfigPackage) StateResult {
	// Build lookup set for installed packages
	installedSet := make(map[string]bool)
	for _, pkg := range installed {
		installedSet[pkg] = true
	}
	
	var result StateResult
	
	// Check each config package against installed
	configSet := make(map[string]bool)
	for _, configPkg := range configPackages {
		configSet[configPkg.Name] = true
		
		if installedSet[configPkg.Name] {
			// Package is managed (in config AND installed)
			result.Managed = append(result.Managed, Package{
				Name:    configPkg.Name,
				State:   StateManaged,
				Manager: managerName,
			})
		} else {
			// Package is missing (in config BUT not installed)
			result.Missing = append(result.Missing, Package{
				Name:    configPkg.Name,
				State:   StateMissing,
				Manager: managerName,
			})
		}
	}
	
	// Check each installed package against config
	for _, pkg := range installed {
		if !configSet[pkg] {
			// Package is untracked (installed BUT not in config)
			result.Untracked = append(result.Untracked, Package{
				Name:    pkg,
				State:   StateUntracked,
				Manager: managerName,
			})
		}
	}
	
	return result
}

