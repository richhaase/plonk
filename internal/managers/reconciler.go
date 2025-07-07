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
	Name            string
	Version         string
	State           PackageState
	ExpectedVersion string // For missing packages, what version was expected
	Manager         string // Which manager this package belongs to
}

// ConfigPackage represents a package defined in plonk.yaml
type ConfigPackage struct {
	Name    string
	Version string // Empty string means "any version"
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

// VersionChecker defines manager-specific version checking logic
type VersionChecker interface {
	CheckVersion(configPkg ConfigPackage, installedVersion string) bool
}

// StateReconciler handles comparing configuration vs installed packages
type StateReconciler struct {
	configLoader ConfigLoader
	managers     map[string]PackageManager
	checkers     map[string]VersionChecker
}

// NewStateReconciler creates a new state reconciler
func NewStateReconciler(configLoader ConfigLoader, managers map[string]PackageManager, checkers map[string]VersionChecker) *StateReconciler {
	return &StateReconciler{
		configLoader: configLoader,
		managers:     managers,
		checkers:     checkers,
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
	
	checker, hasChecker := s.checkers[managerName]
	if !hasChecker {
		return StateResult{}, fmt.Errorf("version checker for %s not found", managerName)
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
	return s.reconcilePackages(managerName, installed, configPackages, checker), nil
}

// reconcilePackages performs the actual reconciliation logic
func (s *StateReconciler) reconcilePackages(managerName string, installed []string, configPackages []ConfigPackage, checker VersionChecker) StateResult {
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
			// Package is installed, check version
			if checker.CheckVersion(configPkg, "") { // Pass empty version for now
				// Package is managed (in config AND correctly installed)
				result.Managed = append(result.Managed, Package{
					Name:    configPkg.Name,
					Version: "",
					State:   StateManaged,
					Manager: managerName,
				})
			} else {
				// Package is missing (in config BUT wrong version)
				result.Missing = append(result.Missing, Package{
					Name:            configPkg.Name,
					Version:         "",
					State:           StateMissing,
					ExpectedVersion: configPkg.Version,
					Manager:         managerName,
				})
			}
		} else {
			// Package is missing (in config BUT not installed)
			result.Missing = append(result.Missing, Package{
				Name:            configPkg.Name,
				Version:         "",
				State:           StateMissing,
				ExpectedVersion: configPkg.Version,
				Manager:         managerName,
			})
		}
	}
	
	// Check each installed package against config
	for _, pkg := range installed {
		if !configSet[pkg] {
			// Package is untracked (installed BUT not in config)
			result.Untracked = append(result.Untracked, Package{
				Name:    pkg,
				Version: "",
				State:   StateUntracked,
				Manager: managerName,
			})
		}
	}
	
	return result
}

// HomebrewVersionChecker implements version checking for Homebrew (version doesn't matter)
type HomebrewVersionChecker struct{}

func (h *HomebrewVersionChecker) CheckVersion(configPkg ConfigPackage, installedVersion string) bool {
	// For Homebrew, any installed version is acceptable
	return true
}

// AsdfVersionChecker implements version checking for ASDF (exact version required)
type AsdfVersionChecker struct{}

func (a *AsdfVersionChecker) CheckVersion(configPkg ConfigPackage, installedVersion string) bool {
	if configPkg.Version == "" {
		// Any version acceptable
		return true
	}
	// For ASDF, exact version match required
	return configPkg.Version == installedVersion
}

// NpmVersionChecker implements version checking for NPM (exact version required)
type NpmVersionChecker struct{}

func (n *NpmVersionChecker) CheckVersion(configPkg ConfigPackage, installedVersion string) bool {
	if configPkg.Version == "" {
		// Any version acceptable
		return true
	}
	// For NPM, exact version match required
	return configPkg.Version == installedVersion
}